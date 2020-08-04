GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

all: test

test:
	$(GOTEST) -v ./...

run: 
	$(GORUN) cmd/gopad/main.go

tsc: $(shell find src -name "*.ts")
	npm run compile

clean:
	rm -rf static
