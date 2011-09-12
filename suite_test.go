package sqs

import (
	"flag"
	. "launchpad.net/gocheck"
	"launchpad.net/goamz/aws"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

var integration = flag.Bool("i", false, "Enable integration tests")

type SuiteI struct {
	auth aws.Auth
}

func (s *SuiteI) SetUpSuite(c *C) {
	if !*integration {
		c.Skip("Integration tests not enabled (-int flag)")
	}
	auth, err := aws.EnvAuth()
	if err != nil {
		c.Fatal(err.String())
	}
	s.auth = auth
}
