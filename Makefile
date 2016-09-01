BUILD_PATH ?= .

.PHONY: all restore rebase resave build example test clean

all: test build example

restore:
	cd $(BUILD_PATH) && godep restore

rebase:
	godep update `cat Godeps/Godeps.json | jq -r .Deps[].ImportPath`

resave:
	godep save ./...

build:
	cd $(BUILD_PATH) && go tool vet -all -composites=false -shadow=true . && go build -a *.go

example:
	cd $(BUILD_PATH)/example && go build -a -o example *.go

test:
	cd $(BUILD_PATH) && go test -race *.go

clean:
	rm -f $(EXAMPLE_BIN)
