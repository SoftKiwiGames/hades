.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	mkdir -p build/
	CGO_ENABLED=0 go build -o build/hades

.PHONY: init
init:
	@CGO_ENABLED=0 go install github.com/fatih/gomodifytags@latest
