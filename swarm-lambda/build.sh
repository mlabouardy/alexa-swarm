#/bin/sh

echo "Build binary"
GOOS=linux GOARCH=amd64 go build -o main *.go

echo "Create deployment package"
zip deployment.zip main

echo "Cleanup"
rm main