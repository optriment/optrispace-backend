#!/bin/bash -eux

ROOTURL=localhost:8080

echo "Creating"

curl -vv -X POST "$ROOTURL/persons" \
    -H 'Content-Type: application/json' \
    --data @- \
    > .stdout \
    2> .stderr <<< '{"address":"adfasdf12231231212"}'

cat .stdout

NEWLOC=$(awk '/Location:/ {print $3}' .stderr | tr -d '\r')

echo "$ROOTURL$NEWLOC"

curl -vv "$ROOTURL$NEWLOC"

