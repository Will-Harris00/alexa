#!/bin/sh
JSON="{\"text\":\"What is the melting point of silver?\"}"
JSON2=`curl -s -X POST -d "$JSON" localhost:3003/tts`
echo $JSON2

