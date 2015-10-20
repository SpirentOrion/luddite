BUILD_PATH	?= .
EXAMPLE_DIR	= ./example
EXAMPLE_BIN	= $(EXAMPLE_DIR)/example

.PHONY: all restore build rebase clean

all: build

restore:
	cd $(BUILD_PATH) && godep restore

build:
	cd $(BUILD_PATH) && \
	go build -a -v ./... && \
	go build -a -v -o $(EXAMPLE_BIN) ./example && \
	go test ./...

rebase:
	godep update `cat Godeps/Godeps.json | jq -r .Deps[].ImportPath`

clean:
	rm -f $(EXAMPLE_BIN)
