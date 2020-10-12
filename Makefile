BUILD_PATH	?= .
EXAMPLE_SUBDIR	= example
EXAMPLE_BIN	= example
V3_SUBDIR	= v3

.PHONY: all vet build test clean

all: vet test build

vet:
	cd $(BUILD_PATH) && go vet -all -vettool=$(shell which shadow)

build:
	cd $(BUILD_PATH) && go build .
	cd $(BUILD_PATH)/$(V3_SUBDIR) && go build .

test:
	cd $(BUILD_PATH) && go test -race .
	cd $(BUILD_PATH)/$(V3_SUBDIR) && go test -race .

clean:
	rm -f $(EXAMPLE_SUBDIR)/$(EXAMPLE_BIN)
	rm -f $(V3_SUBDIR)/$(EXAMPLE_SUBDIR)/$(EXAMPLE_BIN)
