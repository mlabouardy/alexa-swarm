package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}
	svc := ssm.New(cfg)
	req := svc.SendCommandRequest(&ssm.SendCommandInput{
		InstanceIds:  instanceIds,
		DocumentName: aws.String("AWS-RunShellScript"),
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
	svc := sqs.New(cfg)
	req := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
		QueueUrl: aws.String(os.Getenv("SQS_URL")),
	})
	result, err := req.Send()
	if err != nil {
		return "", cluster, err
	}
	json.Unmarshal([]byte(*result.Messages[0].Body), &cluster)
	return *result.Messages[0].ReceiptHandle, cluster, nil
}

func deleteSQSMessage(cfg aws.Config, receiptHandler string) error {
	svc := sqs.New(cfg)
	req := svc.DeleteMessageRequest(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(os.Getenv("SQS_URL")),
		ReceiptHandle: &receiptHandler,
	})
	_, err := req.Send()
	if err != nil {
		return err
	}
	return nil
}

func updateDynamoDB(cfg aws.Config, id string) error {
	svc := dynamodb.New(cfg)
	req := svc.UpdateItemRequest(&dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Key: map[string]dynamodb.AttributeValue{
			"ID": dynamodb.AttributeValue{
				S: &id,
			},
		},
		UpdateExpression: aws.String("set ClusterStatus = :s"),
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":s": dynamodb.AttributeValue{
				S: aws.String("Done"),
			},
		},
	})
	_, err := req.Send()
	if err != nil {
		return err
	}
	return nil
}
