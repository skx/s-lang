BINARY := s-lang

GOFILES := $(shell find . -type f -name '*.go')


$(BINARY): $(GOFILES)
	go build -o $(BINARY) .


.PHONY: clean
clean:
	rm -f $(BINARY)
	cd test && make clean
	cd examples && make clean

.PHONY: fuzz
fuzz:
	nice -n 19 go test -parallel=1 -fuzz=FuzzProject -v

.PHONY: test
test:
	go test ./...
	cd test/ && make test
	cd examples/ && make test
