GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
TSC=./node_modules/typescript/bin/tsc

all: test
test:
	$(GOTEST) -v ./...
run: 
	$(GORUN) cmd/gopad/main.go
ts: $(shell find . -name "*.ts")
	$(TSC)
clean:
	rm static/*.js
