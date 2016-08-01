BUILD_PATH	?= .
EXAMPLE_DIR	= ./example
EXAMPLE_BIN	= example

.PHONY: all restore rebase resave build example test clean

all: test build example

restore:
	cd $(BUILD_PATH) && godep restore

rebase:
	godep update `cat Godeps/Godeps.json | jq -r .Deps[].ImportPath`

resave:
	godep save ./...

build:
	cd $(BUILD_PATH) && go build -a *.go

example:
	cd $(BUILD_PATH)/example && go build -a -o $(EXAMPLE_BIN) *.go

test:
	cd $(BUILD_PATH) && go test -race *.go

clean:
	rm -f $(EXAMPLE_BIN)
