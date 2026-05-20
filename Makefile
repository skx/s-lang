BINARY := s-lang

GOFILES := $(shell find . -type f -name '*.go')


$(BINARY): $(GOFILES)
	go build -o $(BINARY) .


.PHONY: clean
clean:
	rm -f $(BINARY)


.PHONY: test
test:
	go test ./...
	cd test && make
