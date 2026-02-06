.PHONY: build
build:
	mkdir -p build/
	CGO_ENABLED=0 go build -o build/hades
	cp build/hades ~/bin/hades

.PHONY: test
test:
	go test -v ./...

.PHONY: init
init:
	@CGO_ENABLED=0 go install github.com/fatih/gomodifytags@latest
