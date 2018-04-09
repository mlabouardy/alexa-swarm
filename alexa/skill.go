package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
)

const welcomeMessage = "Welcome to containers world. Say 'Deploy a new Swarm cluster' to begin"
const exitMessage = "See you later till then may the containers be with you"

func HandleRequest(ctx context.Context, r AlexaRequest) (AlexaResponse, error) {
	resp := CreateResponse()

	switch r.Request.Intent.Name {
	case "Deploy":
		resp.Say("How many nodes do you want?", false)
	case "ClusterSize":
		resp.SessionAttributes = make(map[string]interface{}, 1)
		resp.SessionAttributes["clusterSize"] = r.Request.Intent.Slots["size"].Value
		resp.Say("What do you want to name your cluster?", false)
	case "ClusterName":
		clusterSize := r.Session.Attributes["clusterSize"]
		clusterName := r.Request.Intent.Slots["cluster"].Value
		count, _ := strconv.ParseInt(clusterSize.(string), 10, 64)
		err := deployInfrastructure(count, clusterName)
		if err != nil {
			resp.Say(err.Error(), true)
		} else {
			resp.Say(fmt.Sprintf("Sure, your %v cluster with %v nodes is being created", clusterName, clusterSize), true)
		}
	case "AMAZON.HelpIntent":
		resp.Say(welcomeMessage, false)
	case "AMAZON.StopIntent":
		resp.Say(exitMessage, true)
	case "AMAZON.CancelIntent":
		resp.Say(exitMessage, true)
	default:
		resp.Say(welcomeMessage, false)
	}
	return *resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}
