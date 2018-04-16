package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
)

type Request struct {
	Records []struct {
		SNS struct {
			Type       string `json:"Type"`
			Timestamp  string `json:"Timestamp"`
			SNSMessage string `json:"Message"`
		} `json:"Sns"`
	} `json:"Records"`
}

func HandleRequest(ctx context.Context, r Request) error {
	log.Println("SNS received")

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}

	receiptHandler, cluster, err := getSQSMessage(cfg)
	if err != nil {
		return err
	}

	master := cluster.Instances[0]
	workers := cluster.Instances[1:]

	err = swarmInit(master.ID)
	if err != nil {
		return err
	}

	token, err := swarmWorkerToken(master.ID)
	if err != nil {
		return err
	}

	log.Println(token)

	for _, worker := range workers {
		err = swarmJoinWorker([]string{worker.ID}, master.IP, token)
		if err != nil {
			return err
		}
	}

	err = updateDynamoDB(cfg, cluster.ID)
	if err != nil {
		return err
	}

	err = deleteSQSMessage(cfg, receiptHandler)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
