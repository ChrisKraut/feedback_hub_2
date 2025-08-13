SWAG_BIN := $(shell go env GOPATH)/bin/swag

.PHONY: swag-install swag-init

swag-install:
	go install github.com/swaggo/swag/cmd/swag@v1.16.3

swag-init: swag-install
	$(SWAG_BIN) init -g cmd/api/main.go -o docs


