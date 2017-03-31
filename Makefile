BUILD_PATH ?= .

ifdef COMSPEC
	EXAMPLE := example.exe
else
  EXAMPLE := example
endif

.PHONY: all build example test clean

all: test build example

build:
	cd $(BUILD_PATH) && go tool vet -all -composites=false -shadow=true *.go && go build -a

example:
	cd $(BUILD_PATH)/example && go build -a -o $(EXAMPLE) *.go

test:
	cd $(BUILD_PATH) && go test -race 

clean:
	rm -f $(EXAMPLE_BIN)
