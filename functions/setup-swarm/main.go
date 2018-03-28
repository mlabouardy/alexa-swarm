package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type CommandOutput struct {
	Status ssm.CommandInvocationStatus
	Output string
}

type Cluster struct {
	ID        string
	Name      string
	Instances []Instance
}

type Instance struct {
	Name string
	IP   string
	ID   string
}

func executeCommand(instanceIds []string, command string) (string, error) {
	document := "AWS-RunShellScript"
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}
	svc := ssm.New(cfg)
	req := svc.SendCommandRequest(&ssm.SendCommandInput{
		InstanceIds:  instanceIds,
		DocumentName: &document,
		Parameters: map[string][]string{
			"commands": []string{command},
		},
	})
	result, err := req.Send()
	if err != nil {
		return "", err
	}
	return *result.Command.CommandId, nil
}

func getCommandOutput(instanceId string, commandId string) (CommandOutput, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return CommandOutput{}, err
	}
	svc := ssm.New(cfg)
	req := svc.GetCommandInvocationRequest(&ssm.GetCommandInvocationInput{
		CommandId:  &commandId,
		InstanceId: &instanceId,
	})
	result, err := req.Send()
	if err != nil {
		return CommandOutput{}, err
	}
	return CommandOutput{
		Status: result.Status,
		Output: *result.StandardOutputContent,
	}, nil
}

func swarmInit(instanceId string) error {
	commandId, err := executeCommand([]string{instanceId}, "docker swarm init")
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 5)
	output, err := getCommandOutput(instanceId, commandId)
	if err != nil {
		return err
	}
	if output.Status != ssm.CommandInvocationStatusSuccess {
		return errors.New("Cannot setup a swarm cluster")
	}
	return nil
}

func swarmWorkerToken(instanceId string) (string, error) {
	commandId, err := executeCommand([]string{instanceId}, "docker swarm join-token worker -q")
	if err != nil {
		return "", err
	}
	time.Sleep(time.Second * 5)
	output, err := getCommandOutput(instanceId, commandId)
	if err != nil {
		return "", err
	}
	if output.Status != ssm.CommandInvocationStatusSuccess {
		return "", errors.New("Cannot get a swarm worker token")
	}
	return strings.TrimSpace(output.Output), nil
}

func swarmJoinWorker(instanceIds []string, masterIp string, token string) error {
	command := fmt.Sprintf("docker swarm join --token %s %s:2377", token, masterIp)
	commandId, err := executeCommand(instanceIds, command)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 5)
	for _, id := range instanceIds {
		output, err := getCommandOutput(id, commandId)
		if err != nil {
			return err
		}
		if output.Status != ssm.CommandInvocationStatusSuccess {
			return errors.New("Cannot join swarm cluster")
		}
	}
	return nil
}

func getSQSMessage(cfg aws.Config) (string, Cluster, error) {
	cluster := Cluster{}
	queueURL := os.Getenv("SQS_URL")
	svc := sqs.New(cfg)
	req := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
		QueueUrl: &queueURL,
	})
	result, err := req.Send()
	if err != nil {
		return "", cluster, err
	}
	json.Unmarshal([]byte(*result.Messages[0].Body), &cluster)
	return *result.Messages[0].ReceiptHandle, cluster, nil
}

func deleteSQSMessage(cfg aws.Config, receiptHandler string) error {
	queueURL := os.Getenv("SQS_URL")
	svc := sqs.New(cfg)
	req := svc.DeleteMessageRequest(&sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &receiptHandler,
	})
	_, err := req.Send()
	if err != nil {
		return err
	}
	return nil
}

func updateDynamoDB(cfg aws.Config, id string) error {
	tableName := os.Getenv("TABLE_NAME")

	updateExpression := "set ClusterStatus = :s"
	clusterStatus := "Done"
	svc := dynamodb.New(cfg)
	req := svc.UpdateItemRequest(&dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]dynamodb.AttributeValue{
			"ID": dynamodb.AttributeValue{
				S: &id,
			},
		},
		UpdateExpression: &updateExpression,
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":s": dynamodb.AttributeValue{
				S: &clusterStatus,
			},
		},
	})
	_, err := req.Send()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}

	receiptHandler, cluster, err := getSQSMessage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	master := cluster.Instances[0]
	workers := cluster.Instances[1:]

	err = swarmInit(master.ID)
	if err != nil {
		log.Fatal(err)
	}

	token, err := swarmWorkerToken(master.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)

	for _, worker := range workers {
		err = swarmJoinWorker([]string{worker.ID}, master.IP, token)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = updateDynamoDB(cfg, cluster.ID)
	if err != nil {
		log.Fatal(err)
	}

	err = deleteSQSMessage(cfg, receiptHandler)
	if err != nil {
		log.Fatal(err)
	}
}
