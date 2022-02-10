#!/bin/sh
# echo "{\"text\":\"What is the melting point of silver?\"}" > input
echo "{\"text\":\"How far is Los Angeles from New York?\"}" > input
JSON2=`curl -s -X POST -d @input localhost:3001/alpha`
echo $JSON2

# curl -v -s -X POST -d '{"text":"How far is Los Angeles from New York?"}' localhost:3001/alpha
