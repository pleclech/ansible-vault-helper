.phony: linux-avh windows-avh

all: linux-avh windows-avh

linux-avh:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/releases/linux-amd64/avh main.go

windows-avh:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/releases/windows-amd64/avh.exe main.go
