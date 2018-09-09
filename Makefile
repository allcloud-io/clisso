GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=clisso
VERSION=`git describe --tags --always`

.PHONY: build
build:
	$(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) -v

.PHONY: test
test:
	$(GOCMD) test -v ./...

.PHONY: darwin-386
darwin-386:
	GOOS=darwin GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-darwin-386 -v

.PHONY: darwin-amd64
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-darwin-amd64 -v

.PHONY: linux-386
linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-linux-386 -v

.PHONY: linux-amd64
linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-linux-amd64 -v

.PHONY: windows-386
windows-386:
	GOOS=windows GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-windows-386.exe -v

.PHONY: windows-amd64
windows-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-windows-amd64.exe -v

.PHONY: all
all: darwin-386 darwin-amd64 linux-386 linux-amd64 windows-386 windows-amd64

.PHONY: zip
zip:
	for i in `ls -1 $(BINARY_NAME)* | grep -v '.zip'`; do zip $$i.zip $$i; done

# See https://gist.github.com/maelvalais/068af21911c7debc4655cdaa41bbf092 for a rough guide on CI/CD for Brew.
.PHONY: brew
brew:
	# There is a dependency on the version in the formula. Follow the following steps.
	# 1. Update the formula to the new source tar.gz.
	# 2. Change the sha256 to match the tar.gz.
	# 3. Remove all sha256 lines in the bottle section.
	# 4. Commit and push.
	# 5. Add the outputed bottles below.
	brew tap allcloud-io/tools
	# If the next command fails, the formula isn't matching the version we are building.
	grep "archive/$(VERSION).tar.gz" `brew --repo allcloud-io/tools`/$(BINARY_NAME).rb >/dev/null
	brew test-bot allcloud-io/tools/$(BINARY_NAME)
	for json in `ls -1 *bottle.json`; do MAC_VERSION=$$(echo $$json | cut -d. -f4);SHA=$$(cat $$json | jq ".\"allcloud-io/tools/$(BINARY_NAME)\".bottle.tags.$$MAC_VERSION.sha256");LOCAL=$$(cat $$json | jq -r ".\"allcloud-io/tools/$(BINARY_NAME)\".bottle.tags.$$MAC_VERSION.local_filename");REMOTE=$$(cat $$json | jq -r ".\"allcloud-io/tools/$(BINARY_NAME)\".bottle.tags.$$MAC_VERSION.filename"); mv $$LOCAL $$REMOTE; echo "sha256 $$SHA => :$$MAC_VERSION"; rm $$json; done

.PHONY: release
release: clean all zip brew

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)"

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)*
