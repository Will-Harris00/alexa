#!/bin/sh
# JSON="{\"text\":\"What is the melting point of silver?\"}"
JSON="{\"text\":\"How far is Los Angeles from New York?\"}"
JSON2=`curl -s -X POST -d "$JSON" localhost:3001/alpha`
echo $JSON2
