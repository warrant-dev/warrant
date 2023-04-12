NAME       = warrant
BUILD_PATH = bin/$(NAME)
GOENV      = GOARCH=amd64 GOOS=linux CGO_ENABLED=0
GOCMD      = go
GOBUILD    = $(GOCMD) build -v -o $(BUILD_PATH)

.PHONY: clean
clean:
	rm -f $(BUILD_PATH)

.PHONY: dev
dev: clean
	$(GOCMD) get
	$(GOBUILD) -tags="sqlite" main.go

.PHONY: build
build: clean
	$(GOCMD) get
	$(GOENV) $(GOBUILD) -ldflags="-s -w" main.go
