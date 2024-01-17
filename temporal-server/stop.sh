#!/bin/bash

RUNNING=`ps -ef | grep temporal | grep -v grep | wc -l | sed 's/^[ \t]*//;s/[ \t]*$//'`

if [ $RUNNING -eq "1" ] 
then
  TSPID=`ps -ef | grep temporal | grep server | awk '{print $2}'`
  echo "Stopping Temporal server (PID:$TSPID).."
  kill -TERM $TSPID
  RETVAL=$?
  exit $RETVAL
else
  echo "Temoral Server is not running"
  exit 0
fi

echo done.
exit $RETVAL
