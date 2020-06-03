.PHONY: all
all: bin/imagefunnel

dist:
	mkdir -p bin

bin:
	mkdir -p bin

bin/imagefunnel: $(shell find . -name '*.go') go.mod bin
	cd cmd/imagefunnel && go build -o ../../$@

.PHONY: linux
linux: dist
	docker run --rm -i -t -v $(PWD):/src -w /src/cmd/imagefunnel golang:1.14 go build -o ../../dist/imagefunnel

.PHONY: test
test:
	go test ./... -v

.PHONY: clean
clean:
	rm -rf bin dist
