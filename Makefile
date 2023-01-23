build:
	go build -buildmode=c-shared -o out_telegram.so *.go

all: build