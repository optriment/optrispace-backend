#!/bin/bash -eux

ROOTURL=localhost:8080

echo "ReadAll"

curl -vv -X GET "$ROOTURL/jobs" | jq .

