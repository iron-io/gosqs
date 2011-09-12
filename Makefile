include $(GOROOT)/src/Make.inc

TARG=github.com/iron-io/gosqs

GOFILES=\
	sqs.go\
	sign.go\

include $(GOROOT)/src/Make.pkg

GOFMT=gofmt
BADFMT=$(shell $(GOFMT) -l $(GOFILES) $(wildcard *_test.go) 2> /dev/null)

gofmt: $(BADFMT)
	@for F in $(BADFMT); do $(GOFMT) -w $$F && echo $$F; done
