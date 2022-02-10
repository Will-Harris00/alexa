#!/bin/sh
echo "{\"speech\":\"`base64 -i speech.wav`\"}" > input
JSON2=`curl -s -X POST -d @input localhost:3002/stt`
echo $JSON2
