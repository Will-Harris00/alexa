#!/bin/sh
JSON="{\"speech\":\"`base64 -i question.wav`\"}"
JSON2=`curl -s -X POST -d "$JSON" localhost:3000/alexa`
echo $JSON2 | cut -d '"' -f4 | base64 -d > answer.wav

