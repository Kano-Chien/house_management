#!/bin/bash
TOKEN="4hW2ocz2jjziSoVrJAXbyKs5Hw3ZzMd302oWITyErQT06R4Ja8NsBmW2PItsCt6hBNAP8gji6pZh2wO5jWkVUQjUwhSp4bRInC34hU1nT7qWELxncsjF6+dTIC+kJbyemmuOphdU+F5BwkWlJ1QkWQdB04t89/1O/w1cDnyilFU="
USER_ID="Uccf36e7addbe098015379bfddea266b2"
MESSAGE=$1

curl -X POST https://api.line.me/v2/bot/message/multicast \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d "{
    \"to\": [\"$USER_ID\"],
    \"messages\": [{\"type\":\"text\",\"text\":\"$MESSAGE\"}]
}"
