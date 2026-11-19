GO := go

.PHONY: build
build:
	$(GO) build -o hcomp cmd/main/main.go