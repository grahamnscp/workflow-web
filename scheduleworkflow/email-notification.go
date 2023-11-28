package scheduleworkflow

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	u "webapp/utils"

	mail "github.com/xhit/go-simple-mail/v2"
	"go.temporal.io/sdk/client"
)

func SendEmailNotification(ctx context.Context, processStage int, sd ScheduleDetails) error {

	log.Printf("%sSendEmailNotification:%s ScheduleID: %s, email: %s\n", u.ColorGreen, u.ColorReset, sd.Id, sd.Email)

	var emailTemplate, emailSubject string

	switch processStage {
	case EmailNotificationStageStarted:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageStartedTemplate)
		emailSubject = EmailNotificationStageStartedSubject

	case EmailNotificationStageRunning:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageRunningTemplate)
		emailSubject = EmailNotificationStageRunningSubject

	case EmailNotificationStageComplete:
		emailTemplate = fmt.Sprintf("templates/%s", EmailNotificationStageCompleteTemplate)
		emailSubject = EmailNotificationStageCompleteSubject

	}
	log.Printf("%sSendEmailNotification:%s email: %s, subject: %s", u.ColorGreen, u.ColorReset, sd.Email, emailSubject)

	// Schedule Details may have changed since passed to workflow at creation
	clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
	if err != nil {
		log.Fatalf("%sSendEmailNotification:%s Failed to load Temporal Cloud environment:%s %v\n", u.ColorGreen, u.ColorRed, u.ColorReset, err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("%sSendEmailNotification:%s Unable to create client,%s %v\n", u.ColorGreen, u.ColorRed, u.ColorReset, err)
	}
	defer c.Close()

	log.Printf("%sSendEmailNotification:%s Checking scheduleID: %s values -\n", u.ColorGreen, u.ColorReset, sd.Id)
	//ctx := context.Background()
	scheduleHandle := c.ScheduleClient().GetHandle(ctx, sd.Id)
	description, err := scheduleHandle.Describe(ctx)
	if err != nil {
		log.Printf("%sSendEmailNotification:%s Failed to get scheduleHandle.Describe, %s %v\n", u.ColorGreen, u.ColorRed, u.ColorReset, err)
	}
	sd.Description = description.Schedule.Spec.Calendars[0].Comment
	sd.Minutes = description.Schedule.Spec.Calendars[0].Minute[0].Start

	// Generate the content
	//var htmlContentTemplate = template.Must(template.New(emailTemplate).Parse(emailTemplate))
	htmlContentTemplate, err := template.ParseFiles(emailTemplate)
	if err != nil {
		log.Printf("%sSendEmailNotification:%s Failed to Parse template file,%s %v", u.ColorGreen, u.ColorRed, u.ColorReset, err)
		return err
	}

	// local stream variable for template content
	var htmlContent bytes.Buffer

	err = htmlContentTemplate.Execute(&htmlContent, sd)
	if err != nil {
		log.Printf("%sSendEmailNotification:%s Failed to Execute template,%s %v", u.ColorGreen, u.ColorRed, u.ColorReset, err)
		return err
	}

	// Create the email
	email := mail.NewMSG()
	email.SetFrom(emailFromAddress).
		AddTo(sd.Email).
		SetSubject(emailSubject).
		SetBody(mail.TextHTML, htmlContent.String())

	if email.Error != nil {
		return email.Error
	}

	// email server
	server := mail.NewSMTPClient()
	server.Host = SMTPHost
	server.Port = SMTPPort
	server.ConnectTimeout = time.Second
	server.SendTimeout = time.Second

	client, err := server.Connect()
	if err != nil {
		return err
	}

	// send and return result
	return email.Send(client)
}
