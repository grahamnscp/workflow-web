#!/bin/bash

function create-namespace {
  NAMESPACE=$1

  NAMESP=`temporal operator namespace list | grep $NAMESPACE| wc -l | sed 's/^[ \t]*//'`
  while [[ "$NAMESP" != "1" ]]
  do
    echo "Creating namespace $NAMESPACE.."
    temporal operator namespace create $NAMESPACE
    sleep 10
    NAMESP=`temporal operator namespace list | grep $NAMESPACE| wc -l | sed 's/^[ \t]*//'`
    echo "Namespace $NAMESPACE query present: $NAMESP"
  done
}

function create-search-attribute {
  NAMESPACE=$1
  ATTRIBUTE_NAME=$2
  ATTRIBUTE_TYPE=$3

  ATTR=`temporal operator search-attribute list -n $NAMESPACE | grep $ATTRIBUTE_NAME | wc -l | sed 's/^[ \t]*//'`
  echo "Attribute $ATTRIBUTE_NAME query present: $ATTR"
  while [[ "$ATTR" != "1" ]]
  do
    echo "Creating search attribute $ATTRIBUTE_NAME on namespace $NAMESPACE.."
    temporal operator search-attribute create --namespace $NAMESPACE --name $ATTRIBUTE_NAME --type $ATTRIBUTE_TYPE
    sleep 5
    ATTR=`temporal operator search-attribute list -n $NAMESPACE | grep $ATTRIBUTE_NAME | wc -l | sed 's/^[ \t]*//'`
    echo "Attribute $ATTRIBUTE_NAME query present: $ATTR"
  done
}


## Main
export TEMPORAL_SERVER_LOG=$(pwd)/temporal-server.log
export TEMPORAL_SERVER_DB=$(pwd)/temporal-server.db

RUNNING=`ps -ef | grep temporal | grep -v grep | wc -l | sed 's/^[ \t]*//;s/[ \t]*$//'`
echo RUNNING=$RUNNING

if [ $RUNNING -eq "1" ] 
then
  echo "Temporal server is already running, no action taken.."
  ps -ef | grep temporal | grep server
else
  echo "Starting Temporal Server using CLI start-dev mode.."
  nohup temporal server start-dev --ui-port 8080 --ip 0.0.0.0 --db-filename $TEMPORAL_SERVER_DB --dynamic-config-value frontend.enableUpdateWorkflowExecution=true >> $TEMPORAL_SERVER_LOG 2>&1 &
  RETVAL=$?
  echo sleeping for server to initialise..
  sleep 10

  # create namespaces
  echo 
  echo "Creating namespaces.."
  create-namespace hello-ns
  #create-namespace background-check
  create-namespace workflow-web

  # create custom search attributes
  # Type = [Text Keyword Int Double Bool Datetime KeywordList]
  echo 

  echo "Creating search fields for default.."
  create-search-attribute default CustomDatetimeField Datetime
  create-search-attribute default CustomIntField Int
  create-search-attribute default CustomBoolField Bool
  create-search-attribute default CustomDoubleField Double
  create-search-attribute default CustomStringField Text
  create-search-attribute default CustomKeywordField Keyword

  #echo "Creating search fields for background-check.."
  #create-search-attribute background-check CandidateEmail Keyword
  #create-search-attribute background-check BackgroundCheckStatus Keyword

  echo "Creating search fields for workflow-web.."
  create-search-attribute workflow-web CustomStringField Text
fi

# List search attributes
echo
echo listing Custom search attributes..
echo   default:
temporal operator search-attribute list -n default | egrep 'Custom'
temporal operator search-attribute list -n workflow-web | egrep 'Custom'

echo done.
exit $RETVAL
