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

const sendGridBaseURL string = "https://api.sendgrid.com"
const sendGridBasePath string = "/v3/marketing/"

type sendGridCampaignMessage struct {
	categories  []string    `json:"categories,omitempty"`
	emailConfig emailConfig `json:"email_config,omitempty"`
	name        string      `json:"name"`
	sendAt      string      `json:"send_at,omitempty"`
	sendTo      recipient   `json:"send_to,omitempty"`
}

type recipient struct {
	all        bool     `json:"all,omitempty"`
	listIDs    []string `json:"list_ids,omitempty"`
	segmentIDs []string `json:"segment_ids,omitempty"`
}

type emailConfig struct {
	customUnsubscribeURL string `json:"custom_unsubscribe_url,omitempty"`
	designId             string `json:"design_id,omitempty"`
	editor               string `json:"editor,omitempty"`
	generatePlainContent bool   `json:"generate_plain_content,omitempty"`
	htmlContent          string `json:"html_content"`
	ipPool               string `json:"ip_pool,omitempty"`
	plainContent         string `json:"plain_content"`
	senderID             int    `json:"sender_id"`
	subject              string `json:"subject"`
	suppressionGroupId   int    `json:"suppression_group_id,omitempty"`
}

type createMessageResponse struct {
	ID string `json:"id"`
}

type sendMessageBody struct {
	sendAt string `json:"send_at"`
}

type CampaignManager struct {
	apiKey string
}

// CreateCampaign creates a SendGrid email campaign to be sent at a later time
func (c *CampaignManager) CreateCampaign(listID string) (*rest.Response, error) {
	// Get the content to send
	templateFile := "./templates/campaign/html-body.html"
	htmlBodyTemplate, err := template.ParseFiles(templateFile)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	err = htmlBodyTemplate.Execute(&tpl, nil)
	if err != nil {
		return nil, err
	}

	message := &sendGridCampaignMessage{
		name: "Summer Products Campaign",
		sendTo: recipient{
			listIDs: []string{listID},
		},
		emailConfig: emailConfig{
			subject:              "New Products for Summer!",
			htmlContent:          tpl.String(),
			customUnsubscribeURL: "https://matthewsetter.com/unsubscribe",
			senderID:             1409812,
		},
	}

	request := sendgrid.GetRequest(c.apiKey, sendGridBasePath+"singlesends", sendGridBaseURL)
	request.Method = "POST"

	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	request.Body = []byte(body)

	return sendgrid.API(request)
}

// ScheduleCampaign schedules a created SendGrid email campaign to be sent
func (c *CampaignManager) ScheduleCampaign(response *rest.Response) (*rest.Response, error) {
	var responseBody createMessageResponse
	err := json.Unmarshal([]byte(response.Body), &responseBody)
	if err != nil {
		return nil, err
	}

	sendAt := time.Now().Add(time.Minute * 3).Format(time.RFC3339)
	sendCampaignBody := &sendMessageBody{sendAt: sendAt}
	sendBody, err := json.Marshal(sendCampaignBody)
	if err != nil {
		return nil, err
	}

	request := sendgrid.GetRequest(
		c.apiKey,
		fmt.Sprintf(sendGridBasePath+"singlesends/%s/schedule", responseBody.ID),
		sendGridBaseURL,
	)
	request.Method = "PUT"
	request.Body = []byte(sendBody)

	return sendgrid.API(request)
}

func (c *CampaignManager) SendCampaign() error {
	response, err := c.CreateCampaign(os.Getenv("SENDGRID_LIST_ID"))
	if err != nil {
		return fmt.Errorf("Could not schedule campaign. Reason: %w", err)
	}
	log.Println("Campaign scheduled")

	response, err = c.ScheduleCampaign(response)
	if err != nil {
		return fmt.Errorf("Could not send campaign. Reason: %w", err)
	}
	log.Println("Campaign sent")

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cm := CampaignManager{apiKey: os.Getenv("SENDGRID_API_KEY")}
	cm.SendCampaign()
}
