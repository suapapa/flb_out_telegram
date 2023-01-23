build:
	GOOS=linux GOARCH=amd64 \
	go build -buildmode=c-shared -o out_telegram.so *.go

all: build