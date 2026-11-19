GO := go

SRC := $(shell find . -name '*.go')
BIN := hcomp

.PHONY: all
all: $(BIN)

$(BIN): $(SRC) go.mod
	$(GO) build -o $@ cmd/main/main.go

.PHONY: clean
clean:
	rm hcomp