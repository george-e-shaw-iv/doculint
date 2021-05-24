.PHONY: build
build: bin
	go build -o bin ./...

.PHONY: bin
bin:
	mkdir -p bin
