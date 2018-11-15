#!/usr/bin/env bash

echo "Building the binary"
GOOS=linux GOARCH=amd64 go build -o main main.go

echo "Creating a zip file"
zip deployment.zip main

echo "Updating lambda function code"
read -p "Input lambda function name: " function_name
aws lambda update-function-code --function-name ${function_name}  --zip-file fileb://./deployment.zip

echo "Cleaning up"
rm main
rm deployment.zip