#!/bin/bash -eux

ROOTURL=localhost:8080

echo "ReadAll"

curl -X GET "$ROOTURL/jobs" | jq .

