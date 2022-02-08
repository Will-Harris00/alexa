#!/bin/sh
JSON="{\"text\":\"What is the melting point of silver?\"}"
JSON2=`curl -s -X POST -d "$JSON" localhost:3001/alpha`
echo $JSON2
