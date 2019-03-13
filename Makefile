.PHONY: build clean deploy

build:
	godep save -v ./...
	env GOOS=linux go build -ldflags="-s -w" -o bin/trainchecker trainchecker/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
