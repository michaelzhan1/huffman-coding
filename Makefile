GO := go

SRC := $(shell find . -name '*.go')
COMP_BIN := hcomp
DECOMP_BIN := hdecomp

.PHONY: all
all: $(COMP_BIN) $(DECOMP_BIN)

$(COMP_BIN): $(SRC) go.mod
	$(GO) build -o $@ cmd/compress/main.go

$(DECOMP_BIN): $(SRC) go.mod
	$(GO) build -o $@ cmd/decompress/main.go

.PHONY: clean
clean:
	rm hcomp