package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

type AlexaRequest struct {
	Version string `json:"version"`
	Session struct {
		Attributes map[string]interface{} `json:"attributes"`
	} `json:"session"`
	Request struct {
		Type   string `json:"type"`
		Time   string `json:"timestamp"`
		Intent struct {
			Name               string `json:"name"`
			ConfirmationStatus string `json:"confirmationstatus"`
		} `json:"intent"`
	} `json:"request"`
}

type AlexaResponse struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes"`
	Response          struct {
		ShouldEndSession bool `json:"shouldEndSession"`
		OutputSpeech     struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"outputSpeech"`
	} `json:"response"`
}

func CreateResponse() *AlexaResponse {
	var resp AlexaResponse
	resp.Version = "1.0"
	resp.Response.OutputSpeech.Type = "PlainText"
	resp.Response.OutputSpeech.Text = "Hello.  Please override this default output."
	return &resp
}

func (resp *AlexaResponse) Say(text string, endSession bool) {
	resp.Response.OutputSpeech.Text = text
	resp.Response.ShouldEndSession = endSession
}

func HandleRequest(ctx context.Context, r AlexaRequest) (AlexaResponse, error) {
	log.Println("Request:")
	log.Println(r)
	log.Printf("Request type is: ", r.Request.Intent.Name)

	resp := CreateResponse()

	if r.Request.Type == "LaunchRequest" {
		resp.Say("Welcome to containers world. Say 'Deploy a new Swarm cluster' to begin", false)
	}

	switch r.Request.Intent.Name {
	case "Deploy":
		resp.SessionAttributes["clusterSize"] = 4
		resp.Say("How many nodes do you want?", false)
	case "ClusterSize":
		resp.Say(fmt.Sprintf("Thanks for selecting the size %d", r.Session.Attributes["clusterSize"]), true)
	case "AMAZON.HelpIntent":
		resp.Say("Welcome to containers world. Say 'Deploy a new Swarm cluster' to begin", false)
	default:
		resp.Say("Welcome to containers world. Say 'Deploy a new Swarm cluster' to begin", false)
	}
	return *resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}
