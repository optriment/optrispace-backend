#!/bin/bash -eux

ROOTURL=localhost:8080

echo "Creating"

curl -vv -X POST "$ROOTURL/jobs" \
    -H 'Content-Type: application/json' \
    --data @job.json \
    > .stdout \
    2> .stderr \

cat .stdout

NEWLOC=$(awk '/Location:/ {print $3}' .stderr | tr -d '\r')

echo "$ROOTURL$NEWLOC"

curl -vv "$ROOTURL$NEWLOC"

# echo "Read all"
# curl -X GET "$ROOTURL/jobs" | jq .

# echo $entity | jq .

# echo "Viewing $entity"
