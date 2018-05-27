GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=clisso

.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

.PHONY: test
test:
	$(GOCMD) test -v ./...

.PHONY: darwin-386
darwin-386:
	GOOS=darwin GOARCH=386 $(GOBUILD) -o $(BINARY_NAME)-darwin-386 -v

.PHONY: darwin-amd64
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64 -v

.PHONY: linux-386
linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINARY_NAME)-linux-386 -v

.PHONY: linux-amd64
linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 -v

.PHONY: all
all: darwin-386 darwin-amd64 linux-386 linux-amd64

.PHONY: zip
zip:
	for i in `ls -1 $(BINARY_NAME)* | grep -v '.zip'`; do zip $$i.zip $$i; done

.PHONY: release
release: clean all zip

.PHONY: install
install:
	go install

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)*