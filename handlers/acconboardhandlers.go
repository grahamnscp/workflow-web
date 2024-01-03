package handlers

import (
  "context"
  "fmt"
  "log"
  "math/rand"
  "net/http"
  "strconv"
  "strings"

  "go.temporal.io/sdk/client"

  ao "webapp/accountonboarding"
  u "webapp/utils"
)

/* Flat local struct to pass account application form to handlers */
type AccFormData struct {
  FirstName string
  Surname   string
  Email     string
}

/* Flat local struct to pass standing order data to template */
type AccAppFraudReviewData struct {
  Token      string
  WorkflowID string
  FirstName  string
  Surname    string
  Email      string
  Fraud      string
  Risk       string
  Credit     string
  Reviewer   string
  Approval   string
  Comment    string
  Success    bool
}

/* NewAccApp */
func NewAccApp(w http.ResponseWriter, r *http.Request) {

  log.Println("NewAccApp: called")

  log.Println("NewAccApp: method:", r.Method) //get request method
  if r.Method == "GET" {
    u.Render(w, "templates/NewAccApp.html", nil)
    return
  }

  r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
  log.Println("NewAccApp: Received form values:", r.FormValue("firstname"), r.FormValue("surname"), r.FormValue("email"))

  appl := ao.ApplicationForm{
    FirstName: r.FormValue("firstname"),
    Surname:   r.FormValue("surname"),
    Email:     r.FormValue("email"),
  }
  log.Println("NewAccApp: Starting new Account Application Workflow for:", appl.Email)

  // have data from form, create new Temporal workflow

  // Temporal Client Start Workflow Options
  wkflowid := fmt.Sprintf("go-accapp-wkfl-%d", rand.Intn(99999))

  // Load the Temporal Cloud from env
  clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
  if err != nil {
    log.Fatalf("NewAccApp: Failed to load Temporal Cloud environment: %v", err)
  }
  // Connect to Temporal Cloud
  log.Println("NewAccApp: connecting to temporal server..")
  c, err := client.Dial(clientOptions)
  if err != nil {
    log.Fatalln("NewAccApp: Unable to create Temporal client.", err)
  }
  defer c.Close()

  // Start Workflow
  log.Println("NewAppApp: Starting AccountApplicationWorkflow..")

  workflowOptions := client.StartWorkflowOptions{
    ID:        wkflowid,
    TaskQueue: ao.AccOnboardingTaskQueueName,
  }
  we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, ao.AccountApplicationWorkflow, appl)

  if err != nil {
    log.Fatalln("NewAccApp: Unable to execute workflow", err)
  }
  log.Printf("NewAccApp: %sWorkflow started:%s (WorkflowID: %s, RunID: %s)", u.ColorYellow, u.ColorReset, we.GetID(), we.GetRunID())

  // Render acknowledgement page
  u.Render(w, "templates/NewAccApp.html", struct{ Success bool }{true})
}

/* AccAppFraudReview */
func AccAppFraudReview(w http.ResponseWriter, r *http.Request) {

  log.Println("AccAppFraudReview: called")

  // URL Parameters
  var token string
  params := r.URL.Query()
  for k, v := range params {
    if k == "ref" {
      token = strings.Join(v, "")
    }
    log.Println("AccAppFraudReview: url params:", k, " => ", v)
  }

  log.Println("AccAppFraudReview: method:", r.Method) //get request method

  // Obtain workflow details from token
  wkflId, runId, err := ao.TokenDecodeAPI(token)
  if err != nil {
    log.Println("AccAppFraudReview: Failed to decode token, err:", err)
  }
  log.Println("AccAppFraudReview: Found WorkflowID:", wkflId, ", RunID:", runId)

  // Load the Temporal Cloud from env
  clientOptions, err := u.LoadClientOptions(u.NoSDKMetrics, "")
  if err != nil {
    log.Fatalf("AccAppFraudReview: Failed to load Temporal Cloud environment: %v", err)
  }
  log.Println("AccAppFraudReview: connecting to temporal server..")
  c, err := client.Dial(clientOptions)
  if err != nil {
    log.Fatalln("AccAppFraudReview: Unable to create Temporal client.", err)
  }
  defer c.Close()

  if r.Method == "GET" {

    // Query the workflow for application information
    log.Println("AccAppFraudReview: Query Fraud Application Form Details..")
    resp, err := c.QueryWorkflow(context.Background(), wkflId, "", "fraud.applicationdetails")
    if err != nil {
      log.Fatalln("Unable to query workflow", err)
    }
    var appdetails ao.ApplicationForm
    if err := resp.Get(&appdetails); err != nil {
      log.Fatalln("Unable to decode appdetails query result", err)
    }

    log.Println("AccAppFraudReview: Query Fraud Approval Status..")
    resp, err = c.QueryWorkflow(context.Background(), wkflId, "", "fraud.approvalstatus")
    if err != nil {
      log.Fatalln("Unable to query workflow", err)
    }
    var approvalstatus ao.ApprovalStatus
    if err := resp.Get(&approvalstatus); err != nil {
      log.Fatalln("Unable to decode approvalstatus query result", err)
    }
    log.Println("AccAppFraudReview: Fraud Approval Status Details:", approvalstatus)

    if approvalstatus.Approved {
      // If workflow has already been approved, send to POST screen again 
      // - TODO: maybe change this to info screen
      u.Render(w, "templates/AccAppFraudReview.html", struct{ Success bool }{true})
    } else {

      log.Println("AccAppFraudReview: Query Fraud Parent Workflow ID..")
      resp, err = c.QueryWorkflow(context.Background(), wkflId, "", "fraud.parentworkflowid")
      if err != nil {
        log.Fatalln("Unable to query workflow", err)
      }
      var parentwkflid string
      if err := resp.Get(&parentwkflid); err != nil {
        log.Fatalln("Unable to decode parentwkflid query result", err)
      }

      // Using parent workflow ID get Credit Status Check Details
      log.Println("AccAppFraudReview: Query Parent Workflow Application Status..")
      resp, err = c.QueryWorkflow(context.Background(), parentwkflid, "", "application.status")
      if err != nil {
        log.Fatalln("Unable to query workflow", err)
      }
      var appstatus ao.ApplicationStatus
      if err := resp.Get(&appstatus); err != nil {
        log.Fatalln("Unable to decode appstatus query result", err)
      }

      // Render Approval Screen
      fraudreview := AccAppFraudReviewData{}

      fraudreview.Token = token
      fraudreview.WorkflowID = wkflId

      fraudreview.FirstName = appdetails.FirstName
      fraudreview.Surname = appdetails.Surname
      fraudreview.Email = appdetails.Email
      fraudreview.Fraud = "L2 Approval Required"
      fraudreview.Risk = strconv.Itoa(approvalstatus.FraudRisk)
      fraudreview.Credit = appstatus.Credit
      fraudreview.Reviewer = approvalstatus.Approver
      fraudreview.Approval = "Approved or Denied"
      fraudreview.Comment = ""
      fraudreview.Success = false

      log.Println("AccAppFraudReview: Rendering Approval Screen", fraudreview)

      u.Render(w, "templates/AccAppFraudReview.html", fraudreview)
    }
  } else if r.Method == "POST" {

    // Process Approval screen response
    r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
    fraudresult := AccAppFraudReviewData{}
    fraudresult.Token = strings.TrimSpace(r.FormValue("token"))
    fraudresult.WorkflowID = strings.TrimSpace(r.FormValue("wkflid"))
    fraudresult.Email = strings.TrimSpace(r.FormValue("email"))
    fraudresult.Approval = strings.TrimSpace(r.FormValue("approval"))
    fraudresult.Comment = strings.TrimSpace(r.FormValue("comment"))

    log.Println("AccAppFraudReview: Processing Approval Result", fraudresult)

    // Parse result and signal fraud workflow

    // Approved or Denied signal
    approvalresult := fraudresult.Comment

    if fraudresult.Approval == "Denied" {
      // Only Approved is approval else response is the reason comment
      if approvalresult == "" {
        approvalresult = "Denied"
      }
    } else if fraudresult.Approval == "Approved" {
      approvalresult = "Approved"
    }

    err = c.SignalWorkflow(context.Background(), wkflId, "", "fraudapproverresult", approvalresult)
    if err != nil {
      log.Println("AccAppFraudReview: Unable to signal fraudapproverresult to workflow", err)
    }
    u.Render(w, "templates/AccAppFraudReview.html", struct{ Success bool }{true})
  }
}
