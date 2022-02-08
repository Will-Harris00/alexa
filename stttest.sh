#!/bin/sh
JSON="{\"speech\":\"`base64 -i speech.wav`\"}"
JSON2=`curl -s -X POST -d "$JSON" localhost:3002/stt`
echo $JSON2

