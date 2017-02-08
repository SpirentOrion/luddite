BUILD_PATH ?= .

.PHONY: all build example test clean

all: test build example

build:
	cd $(BUILD_PATH) && go tool vet -all -composites=false -shadow=true *.go && go build -a *.go

example:
	cd $(BUILD_PATH)/example && go build -a -o example *.go

test:
	cd $(BUILD_PATH) && go test -race *.go

clean:
	rm -f $(EXAMPLE_BIN)
