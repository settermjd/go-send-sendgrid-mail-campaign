package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
)

const SendGridBaseUrl string = "https://api.sendgrid.com"
const SendGridBasePath string = "/v3/marketing/"

type SendGridMessage struct {
	Categories  []string    `json:"categories,omitempty"`
	EmailConfig EmailConfig `json:"email_config,omitempty"`
	Name        string      `json:"name"`
	SendAt      string      `json:"send_at,omitempty"`
	SendTo      Recipient   `json:"send_to,omitempty"`
}

type Recipient struct {
	All        bool     `json:"all,omitempty"`
	ListIds    []string `json:"list_ids,omitempty"`
	SegmentIds []string `json:"segment_ids,omitempty"`
}

type EmailConfig struct {
	CustomUnsubscribeUrl string `json:"custom_unsubscribe_url,omitempty"`
	DesignId             string `json:"design_id,omitempty"`
	Editor               string `json:"omitempty"`
	GeneratePlainContent bool   `json:"generate_plain_content,omitempty"`
	HtmlContent          string `json:"html_content"`
	IpPool               string `json:"ip_pool,omitempty"`
	PlainContent         string `json:"plain_content"`
	SenderId             int    `json:"sender_id"`
	Subject              string
	SuppressionGroupId   int `json:"suppression_group_id,omitempty"`
}

type CreateMessageResponse struct {
	ID string `json:"id"`
}

type SendMessageBody struct {
	SendAt string `json:"send_at"`
}

// CreateCampaign creates a campaign to be sent at a later time
func CreateCampaign(apiKey string, listId string) (*rest.Response, error) {
	// Get the content to send
	templateFile := "./templates/campaign/html-body.html"
	htmlBodyTemplate, err := template.ParseFiles(templateFile)
	if err != nil {
		log.Fatalf("Unable to parse HTML template: %s", templateFile)
	}
	var tpl bytes.Buffer
	err = htmlBodyTemplate.Execute(&tpl, nil)
	if err != nil {
		log.Fatalf("Could not retrieve template data. Reason: %s", err)
	}

	message := &SendGridMessage{
		Name: "Summer Products Campaign",
		SendTo: Recipient{
			ListIds: []string{listId},
		},
		EmailConfig: EmailConfig{
			Subject:              "New Products for Summer!",
			HtmlContent:          tpl.String(),
			CustomUnsubscribeUrl: "https://matthewsetter.com/unsubscribe",
			SenderId:             1409812,
		},
	}

	request := sendgrid.GetRequest(apiKey, SendGridBasePath+"singlesends", SendGridBaseUrl)
	request.Method = "POST"

	body, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Unable to create email message body in JSON format.")
	}

	request.Body = []byte(body)
	response, err := sendgrid.API(request)

	return response, err
}

// ScheduleCampaign schedules a created campaign to be sent
func ScheduleCampaign(apiKey string, response *rest.Response) (*rest.Response, error) {
	var responseBody CreateMessageResponse
	json.Unmarshal([]byte(response.Body), &responseBody)
	fmt.Printf("Response : %+v", responseBody.ID)

	sendAt := time.Now().Add(time.Minute * 3).Format(time.RFC3339)
	sendCampaignBody := &SendMessageBody{SendAt: sendAt}
	sendBody, err := json.Marshal(sendCampaignBody)
	if err != nil {
		log.Fatal("Could not create send campaign body")
	}

	request := sendgrid.GetRequest(
		apiKey,
		fmt.Sprintf(SendGridBasePath+"singlesends/%s/schedule", responseBody.ID),
		SendGridBaseUrl,
	)
	request.Method = "PUT"
	request.Body = []byte(sendBody)
	response, err = sendgrid.API(request)

	return response, err
}

func main() {
	godotenv.Load()

	apiKey := os.Getenv("SENDGRID_API_KEY")

	response, err := CreateCampaign(apiKey, os.Getenv("SENDGRID_LIST_ID"))
	if err != nil {
		fmt.Printf("Could not schedule campaign. Status: %i. Reason: %s", response.StatusCode, err)
	}

	response, err = ScheduleCampaign(apiKey, response)
	if err != nil {
		fmt.Printf("Could not send campaign. Status: %i. Reason: %s", response.StatusCode, err)
	}
}
