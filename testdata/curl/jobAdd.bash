#!/bin/bash -eux

ROOTURL=localhost:8080

echo "Creating"

curl -vv -X POST "$ROOTURL/jobs" \
    -H 'Content-Type: application/json' \
    --data @- \
    > .stdout \
    2> .stderr <<< '{    "title": "Заголовок задачи","description": "Описание задачи, возможно, \nв несколько строк","customer": {"id": "ih7vRdVVUpZiVe4zn3pZNA"}}'

cat .stdout

NEWLOC=$(awk '/Location:/ {print $3}' .stderr | tr -d '\r')

echo "$ROOTURL$NEWLOC"

curl -vv "$ROOTURL$NEWLOC"

# echo "Read all"
# curl -X GET "$ROOTURL/jobs" | jq .

# echo $entity | jq .

# echo "Viewing $entity"
