package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/xid"
)

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

type DBItem struct {
	ID            string
	Name          string
	Size          int
	ClusterStatus string
}

func setupInfrastructure(cfg aws.Config, name string, count int64) (Cluster, error) {
	imageID := os.Getenv("AMI")
	keyName := os.Getenv("KEYPAIR")
	ssmRoleName := os.Getenv("SSM_ROLE_NAME")
	securityGroupID := os.Getenv("SECURITY_GROUP_ID")

	svc := ec2.New(cfg)
	req := svc.RunInstancesRequest(&ec2.RunInstancesInput{
		ImageId:      &imageID,
		InstanceType: ec2.InstanceTypeT1Micro,
		KeyName:      &keyName,
		MaxCount:     &count,
		MinCount:     &count,
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: &ssmRoleName,
		},
		SecurityGroupIds: []string{securityGroupID},
	})
	result, err := req.Send()
	if err != nil {
		return Cluster{}, err
	}

	instances := make([]Instance, 0, len(result.Instances))
	for index, instance := range result.Instances {
		instances = append(instances, Instance{
			Name: fmt.Sprintf("node-%d", (index + 1)),
			IP:   *instance.PrivateIpAddress,
			ID:   *instance.InstanceId,
		})
	}
	return Cluster{
		ID:        xid.New().String(),
		Name:      name,
		Instances: instances,
	}, nil
}

func pushToSQS(cfg aws.Config, cluster Cluster) error {
	queueURL := os.Getenv("SQS_URL")
	data, _ := json.Marshal(cluster)
	raw := string(data)
	svc := sqs.New(cfg)
	req := svc.SendMessageRequest(&sqs.SendMessageInput{
		MessageBody: &raw,
		QueueUrl:    &queueURL,
	})
	_, err := req.Send()
	if err != nil {
		return err
	}
	return nil
}

func insertToDynamoDB(cfg aws.Config, cluster Cluster) error {
	tableName := os.Getenv("TABLE_NAME")

	item := DBItem{
		ID:            cluster.ID,
		Name:          cluster.Name,
		Size:          len(cluster.Instances),
		ClusterStatus: "Pending",
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		fmt.Println(err)
		return err
	}

	svc := dynamodb.New(cfg)
	req := svc.PutItemRequest(&dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	_, err = req.Send()
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

	cluster, err := setupInfrastructure(cfg, "staging", 2)
	if err != nil {
		log.Fatal(err)
	}

	err = pushToSQS(cfg, cluster)
	if err != nil {
		log.Fatal(err)
	}

	err = insertToDynamoDB(cfg, cluster)
	if err != nil {
		log.Fatal(err)
	}
}
