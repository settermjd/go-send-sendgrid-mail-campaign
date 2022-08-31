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
	"github.com/sendgrid/sendgrid-go"
)

type SendGridMessage struct {
	Name        string      `json:"name"`
	Categories  []string    `json:"categories,omitempty"`
	SendAt      string      `json:"send_at,omitempty"`
	SendTo      Recipient   `json:"send_to,omitempty"`
	EmailConfig EmailConfig `json:"email_config,omitempty"`
}

type Recipient struct {
	ListIds    []string `json:"list_ids,omitempty"`
	SegmentIds []string `json:"segment_ids,omitempty"`
	All        bool     `json:"all,omitempty"`
}

type EmailConfig struct {
	Subject              string
	HtmlContent          string `json:"html_content"`
	PlainContent         string `json:"plain_content"`
	GeneratePlainContent bool   `json:"generate_plain_content,omitempty"`
	DesignId             string `json:"design_id,omitempty"`
	Editor               string `json:"omitempty"`
	SuppressionGroupId   int    `json:"suppression_group_id,omitempty"`
	CustomUnsubscribeUrl string `json:"custom_unsubscribe_url,omitempty"`
	SenderId             int    `json:"sender_id"`
	IpPool               string `json:"ip_pool,omitempty"`
}

type CreateMessageResponse struct {
	ID string `json:"id"`
}

type SendMessageBody struct {
	SendAt string `json:"send_at"`
}

func main() {
	// Load environment variables from .env
	godotenv.Load()

	// Get the content to send
	htmlBodyTemplate, err := template.ParseFiles("./templates/campaign/html-body.html")
	if err != nil {
		log.Fatalf("Unable to parse HTML template: %s", "templates/campaign/html-body.html")
	}
	var tpl bytes.Buffer
	err = htmlBodyTemplate.Execute(&tpl, nil)
	if err != nil {
		log.Fatalf("Could not retrieve template data. Reason: %s", err)
	}

	sendAt := time.Now().Format(time.RFC3339)
	//sendAt := "now"

	// Send 20 minutes from now
	message := &SendGridMessage{
		Name:   "Summer Products Campaign",
		SendAt: sendAt,
		SendTo: Recipient{
			ListIds: []string{"459f7e42-8c0d-4720-9b0f-7a45d8ba2a19"},
		},
		EmailConfig: EmailConfig{
			Subject:              "New Products for Summer!",
			HtmlContent:          tpl.String(),
			CustomUnsubscribeUrl: "https://matthewsetter.com/unsubscribe",
			SenderId:             1409812,
		},
	}

	apiKey := os.Getenv("SENDGRID_API_KEY")
	request := sendgrid.GetRequest(apiKey, "/v3/marketing/singlesends", "https://api.sendgrid.com")
	request.Method = "POST"

	body, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Unable to create email message body in JSON format.")
	}

	request.Body = []byte(body)
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		var responseBody CreateMessageResponse
		json.Unmarshal([]byte(response.Body), &responseBody)
		fmt.Printf("Response : %+v", responseBody.ID)

		request := sendgrid.GetRequest(
			apiKey,
			fmt.Sprintf("/v3/marketing/singlesends/%s/schedule", responseBody.ID),
			"https://api.sendgrid.com",
		)
		request.Method = "PUT"
		sendCampaignBody := &SendMessageBody{SendAt: time.Now().Add(time.Minute * 3).Format(time.RFC3339)}
		sendBody, err := json.Marshal(sendCampaignBody)
		if err != nil {
			log.Fatal("Could not create send campaign body")
		}
		request.Body = []byte(sendBody)
		response, err := sendgrid.API(request)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Body)
			fmt.Println(response.Headers)
		}
	}
}
