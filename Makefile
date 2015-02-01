GOPATH := $(shell pwd)

all: bootsmann


bootsmann:
	GOPATH=$(GOPATH) go get $@
	GOPATH=$(GOPATH) go build $@

clean:
	GOPATH=$(GOPATH) go clean
	${RM} -r pkg/

example:
	rm -rf example
	cp -a example-pre example
	./bootsmann -config example-pre/bootsmann.conf
	find example -type f | xargs cat

.PHONY: bootsmann
