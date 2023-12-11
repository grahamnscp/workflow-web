package standingorder

import (
	"os"
	"time"

  mt "webapp/moneytransfer"
)

var StandingOrdersTaskQueueName = os.Getenv("STANDING_ORDERS_TASK_QUEUE")

type PaymentSchedule struct {
	PeriodDuration time.Duration // seconds
	Active         bool
}

type StandingOrder struct {
	Schedule PaymentSchedule
	Details  mt.PaymentDetails
}

var maxSOrdersPerWF int = 12
