NAME    = warrant
BUILD_PATH = bin/$(NAME)
GOENV   = GOARCH=amd64 GOOS=linux CGO_ENABLED=0
GOCMD   = go
GOBUILD = $(GOCMD) build -o
VERSION = $(shell cat VERSION)

.PHONY: clean
clean:
	rm -f $(BUILD_PATH)

.PHONY: dev
dev: clean
	$(GOCMD) get
	$(GOBUILD) $(BUILD_PATH) main.go

.PHONY: build
build: clean
	$(GOCMD) get
	$(GOENV) $(GOBUILD) $(BUILD_PATH) -ldflags="-s -w" main.go
