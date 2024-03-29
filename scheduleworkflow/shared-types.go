package scheduleworkflow

import "os"

var ScheduleWFTaskQueueName = os.Getenv("SCHEDULE_WF_TASK_QUEUE")

// Schedule type
type ScheduleDetails struct {
	Id          string
	Description string
	Email       string
	Minutes     int
}

// Pseudo process state email notification stages
var EmailNotificationStageStarted int = 1
var EmailNotificationStageRunning int = 2
var EmailNotificationStageComplete int = 3

var EmailNotificationStageStartedTemplate string = "ScheduleStarted.html"
var EmailNotificationStageRunningTemplate string = "ScheduleRunning.html"
var EmailNotificationStageCompleteTemplate string = "ScheduleComplete.html"

var EmailNotificationStageStartedSubject string = "Schedule Workflow Started"
var EmailNotificationStageRunningSubject string = "Schedule Workflow Running"
var EmailNotificationStageCompleteSubject string = "Schedule Workflow Completed"

var emailFromAddress string = "noreply@webapp.domain"

// localhost mailserver (mailhog alias to localhost)
// var SMTPHost string = "mailhog"
var SMTPHost string = "localhost"
var SMTPPort int = 1025

