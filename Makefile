GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BUILDPATH=build
ASSETPATH=assets
BINARY_NAME=clisso
VERSION=`git describe --tags --always`

.PHONY: build
build:
	$(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME) -v

.PHONY: test
test:
	$(GOCMD) test -v ./...

.PHONY: darwin-amd64
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-darwin-amd64 -v

.PHONY: linux-386
linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-linux-386 -v

.PHONY: linux-amd64
linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-linux-amd64 -v

.PHONY: windows-386
windows-386:
	GOOS=windows GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-windows-386.exe -v

.PHONY: windows-amd64
windows-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-windows-amd64.exe -v

.PHONY: all
all: darwin-amd64 linux-386 linux-amd64 windows-386 windows-amd64

.PHONY: sign
sign: darwin-amd64
	# sign
	gon -log-level=info ./gon.json

.PHONY: zip
zip: all sign
	mkdir -p $(ASSETPATH)
	cd $(BUILDPATH) && \
	for i in `ls -1 $(BINARY_NAME)* | grep -v '.zip' | grep -v darwin`; do zip ../$(ASSETPATH)/$$i.zip $$i; done
	cd $(ASSETPATH) && \
	sha256sum *zip > SHASUMS256.txt

.PHONY: unsigned-darwin-amd64-zip
unsigned-darwin-amd64-zip: darwin-amd64
	# used by brew if signing isn't setup
	mkdir -p $(ASSETPATH)
	cd $(BUILDPATH) && \
	zip ../$(ASSETPATH)/clisso-darwin-amd64.zip clisso-darwin-amd64

.PHONY: brew
brew:
	bash make_brew_release.sh $(BINARY_NAME) $(VERSION)

.PHONY: release
release: clean all zip brew

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)"

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BUILDPATH)/$(BINARY_NAME)* $(ASSETPATH)/*
