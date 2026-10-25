IMAGE ?= quay.io/oshoval/groutdra:dev
CONTAINER_ENGINE ?= docker

.PHONY: build image test
build:
	go build -o bin/groutdra ./cmd/groutdra

image:
	$(CONTAINER_ENGINE) build -t $(IMAGE) .

test:
	go test ./...
