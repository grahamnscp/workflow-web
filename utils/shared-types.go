package utils

import (
	"fmt"
	"os"
	"strings"
)

var log_level = strings.ToLower(os.Getenv("LOG_LEVEL"))
var NoSDKMetrics bool = false
var SDKMetrics bool = true

// AccountApplication uses mongodb
var MongoDBName = "bank"
var MongoURI = fmt.Sprintf("mongodb://bankuser:bankuserpwd@localhost:27017/%s", MongoDBName)
