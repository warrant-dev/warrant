NAME    = warrant
BUILD_PATH = bin/$(NAME)
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
