GO := go

SRC := $(shell find . -name '*.go')
COMP_BIN := hcomp
DECOMP_BIN := hdecomp

.PHONY: all
all: $(COMP_BIN) $(DECOMP_BIN)

$(COMP_BIN): $(SRC) go.mod
	$(GO) build -o $@ ./cmd/compress

$(DECOMP_BIN): $(SRC) go.mod
	$(GO) build -o $@ ./cmd/decompress

.PHONY: clean
clean:
	rm -f $(COMP_BIN) $(DECOMP_BIN)