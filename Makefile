.PHONY: all binary binary-in-docker build tag release

all: binary-in-docker build release

binary-in-docker:
	gcloud docker -- pull us.gcr.io/sharpspring-us/golang:build
	docker run \
		--rm \
		--volume $(shell pwd):/go/src/github-listener \
		--workdir /go/src/github-listener \
		us.gcr.io/sharpspring-us/golang:build sh -c " \
			go get \
				&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags -w"

build:
	docker build -t us.gcr.io/sharpspring-us/github-listener:`git rev-parse --short HEAD` .

release:
	gcloud docker -- push us.gcr.io/sharpspring-us/github-listener:`git rev-parse --short HEAD`

deploy:
	kubectl --context=global --namespace=jenkins patch deployment github-listener -p '{"spec":{"template":{"spec":{"containers":[{"name":"github-listener", "image": "us.gcr.io/sharpspring-us/github-listener:$(shell git rev-parse --short HEAD)"}]}}}}'
