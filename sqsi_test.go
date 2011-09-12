package sqs

import (
	"fmt"
	"launchpad.net/goamz/aws"
	. "launchpad.net/gocheck"
)

var _ = Suite(&SI{})

type SI struct {
	SuiteI
	sqs *SQS
}

func (s *SI) SetUpSuite(c *C) {
	s.SuiteI.SetUpSuite(c)
	s.sqs = New(s.auth, aws.USEast)
}

func (s *SI) Queue(name string) string {
	return name + "-" + s.sqs.Auth.AccessKey
}

const testQueue = "goamz-test-queue"

func (s *SI) TestBasicFunctionality(c *C) {
	q, err := s.sqs.CreateQueue(s.Queue(testQueue), nil)
	c.Assert(err, IsNil)

	queues, err := s.sqs.ListQueues("")
	c.Assert(err, IsNil)
	fmt.Println(queues)

	_, err = q.SendMessage("hi")
	c.Assert(err, IsNil)

	msg, err := q.ReceiveMessage()
	c.Assert(err, IsNil)
	c.Assert(msg.Body, Equals, "hi")

	err = q.DeleteQueue()
	c.Assert(err, IsNil)
}
