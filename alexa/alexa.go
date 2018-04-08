package main

type AlexaRequest struct {
	Version string `json:"version"`
	Request struct {
		Type   string `json:"type"`
		Time   string `json:"timestamp"`
		Intent struct {
			Name  string `json:"name"`
			Slots map[string]struct {
				Value string `json:"value"`
			} `json:"slots"`
		} `json:"intent"`
	} `json:"request"`
	Session struct {
		Attributes map[string]interface{} `json:"attributes"`
	}
}

type AlexaResponse struct {
	Version  string `json:"version"`
	Response struct {
		ShouldEndSession bool `json:"shouldEndSession"`
		OutputSpeech     struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"outputSpeech"`
	} `json:"response"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes"`
}

func CreateResponse() *AlexaResponse {
	var resp AlexaResponse
	resp.Version = "1.0"
	resp.Response.OutputSpeech.Type = "PlainText"
	return &resp
}

func (resp *AlexaResponse) Say(text string, shouldEndSession bool) {
	resp.Response.OutputSpeech.Text = text
	resp.Response.ShouldEndSession = shouldEndSession
}
