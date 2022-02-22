
proj := $(shell basename $(shell pwd))

default: runtest

all: $(proj)

# do something or other when testing it
runtest: $(proj)
	time ./$(proj) -s -0 / | xargs -r0 ls -lt

test:
	go test

format:
	gofmt -s=true -w *.go

clean:
	rm -f *~ */*~ .*~ $(proj)
	if [ -f go.mod ] ; then go mod tidy ; fi
	go clean

$(proj): Makefile *.go
	if [ -f go.mod ] ; then go mod tidy ; fi
	go build -ldflags="-w -s"
	go vet

.PHONY: all runtest test format clean
