//
// goamz - Go packages to interact with the Amazon Web Services.
//
// https://wiki.ubuntu.com/goamz
//
// Copyright (c) 2011 Iron.io
//
// Written by Evan Shaw <evan@iron.io>
//
package sqs

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// The SQS type encapsulates operations with a specific SQS region.
type SQS struct {
	aws.Auth
	aws.Region
	private byte // Reserve the right of using private data.
}

// The Queue type encapsulates operations with an SQS queue.
type Queue struct {
	*SQS
	path string
}

// An Attribute specifies which attribute of a message to set or receive.
type Attribute string

const (
	All                                   Attribute = "All"
	ApproximateNumberOfMessages           Attribute = "ApproximateNumberOfMessages"
	ApproximateNumberOfMessagesNotVisible Attribute = "ApproximateNumberOfMessagesNotVisible"
	VisibilityTimeout                     Attribute = "VisibilityTimeout"
	CreatedTimestamp                      Attribute = "CreatedTimestamp"
	LastModifiedTimestamp                 Attribute = "LastModifiedTimestamp"
	Policy                                Attribute = "Policy"
	MaximumMessageSize                    Attribute = "MaximumMessageSize"
	MessageRetentionPeriod                Attribute = "MessageRetentionPeriod"
	QueueArn                              Attribute = "QueueArn"
)

// New creates a new SQS.
func New(auth aws.Auth, region aws.Region) *SQS {
	return &SQS{auth, region, 0}
}

type ResponseMetadata struct {
	RequestId string
}

func (sqs *SQS) Queue(name string) (*Queue, error) {
	qs, err := sqs.ListQueues(name)
	if err != nil {
		return nil, err
	}
	for _, q := range qs {
		if q.Name() == name {
			return q, nil
		}
	}
	// TODO: return error
	return nil, nil
}

type listQueuesResponse struct {
	Queues []string `xml:"ListQueuesResult>QueueUrl"`
	ResponseMetadata
}

// ListQueues returns a list of your queues.
//
// See http://goo.gl/q1ue9 for more details.
func (sqs *SQS) ListQueues(namePrefix string) ([]*Queue, error) {
	params := url.Values{}
	if namePrefix != "" {
		params.Set("QueueNamePrefix", namePrefix)
	}
	var resp listQueuesResponse
	if err := sqs.get("ListQueues", "/", params, &resp); err != nil {
		return nil, err
	}
	queues := make([]*Queue, len(resp.Queues))
	for i, queue := range resp.Queues {
		u, err := url.Parse(queue)
		if err != nil {
			return nil, err
		}
		queues[i] = &Queue{sqs, u.Path}
	}
	return queues, nil
}

func (sqs *SQS) newRequest(method, action, url_ string, params url.Values) (*http.Request, error) {
	req, err := http.NewRequest("GET", url_, nil)
	if err != nil {
		return nil, err
	}

	params["Action"] = []string{action}
	params["Timestamp"] = []string{time.Now().UTC().Format(time.RFC3339)}
	params["Version"] = []string{"2009-02-01"}

	req.Header.Set("Host", req.Host)

	sign(sqs.Auth, method, req.URL.Path, params, req.Header)
	return req, nil
}

// Error encapsulates an error returned by SDB.
type Error struct {
	StatusCode int    // HTTP status code (200, 403, ...)
	StatusMsg  string // HTTP status message ("Service Unavailable", "Bad Request", ...)
	Type       string // Whether the error was a receiver or sender error
	Code       string // SQS error code ("InvalidParameterValue", ...)
	Message    string // The human-oriented error message
	RequestId  string // A unique ID for this request
}

func (err *Error) Error() string {
	return err.Message
}

func buildError(r *http.Response) error {
	err := Error{}
	err.StatusCode = r.StatusCode
	err.StatusMsg = r.Status
  body, _ := ioutil.ReadAll(r.Body)
	xml.Unmarshal(body, &err)
	return &err
}

func (sqs *SQS) doRequest(req *http.Request, resp interface{}) error {
	/*dump, _ := http.DumpRequest(req, true)
	println("req DUMP:\n", string(dump))*/

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	/*str, _ := http.DumpResponse(r, true)
	fmt.Printf("response text: %s\n", str)
	fmt.Printf("response struct: %+v\n", resp)*/
	if r.StatusCode != 200 {
		return buildError(r)
	}
  body, _ := ioutil.ReadAll(r.Body)
	return xml.Unmarshal(body, resp)
}

func (sqs *SQS) post(action, path string, params url.Values, body []byte, resp interface{}) error {
	endpoint := strings.Replace(sqs.Region.EC2Endpoint, "ec2", "sqs", 1) + path
	req, err := sqs.newRequest("POST", action, endpoint, params)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "x-www-form-urlencoded")

	encodedParams := params.Encode()
	req.Body = ioutil.NopCloser(strings.NewReader(encodedParams))
	req.ContentLength = int64(len(encodedParams))

	return sqs.doRequest(req, resp)
}

func (sqs *SQS) get(action, path string, params url.Values, resp interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	endpoint := strings.Replace(sqs.Region.EC2Endpoint, "ec2", "sqs", 1) + path
	req, err := sqs.newRequest("GET", action, endpoint, params)
	if err != nil {
		return err
	}

	if len(params) > 0 {
		req.URL.RawQuery = params.Encode()
	}

	return sqs.doRequest(req, resp)
}

func (q *Queue) Name() string {
	return path.Base(q.path)
}

// AddPermission adds a permission to a queue for a specific principal.
//
// See http://goo.gl/vG4CP for more details.
func (q *Queue) AddPermission() error {
	return nil
}

// ChangeMessageVisibility changes the visibility timeout of a specified message
// in a queue to a new value.
//
// See http://goo.gl/tORrh for more details.
func (q *Queue) ChangeMessageVisibility() error {
	return nil
}

type CreateQueueOpt struct {
	DefaultVisibilityTimeout int
}

type createQueuesResponse struct {
	QueueUrl string `xml:"CreateQueueResult>QueueUrl"`
	ResponseMetadata
}

// CreateQueue creates a new queue.
//
// See http://goo.gl/EwNUK for more details.
func (sqs *SQS) CreateQueue(name string, opt *CreateQueueOpt) (*Queue, error) {
	params := url.Values{
		"QueueName": []string{name},
	}
	if opt != nil {
		dvt := strconv.Itoa(opt.DefaultVisibilityTimeout)
		params["DefaultVisibilityTimeout"] = []string{dvt}
	}
	var resp createQueuesResponse
	if err := sqs.get("CreateQueue", "/", params, &resp); err != nil {
		return nil, err
	}
	u, err := url.Parse(resp.QueueUrl)
	if err != nil {
		return nil, err
	}
	return &Queue{sqs, u.Path}, nil
}

// DeleteQueue deletes a queue.
//
// See http://goo.gl/zc45Q for more details.
func (q *Queue) DeleteQueue() error {
	params := url.Values{}
	var resp ResponseMetadata
	if err := q.SQS.get("DeleteQueue", q.path, params, &resp); err != nil {
		return err
	}
	return nil
}

// DeleteMessage deletes a message from the queue.
//
// See http://goo.gl/t8jnk for more details.
func (q *Queue) DeleteMessage() error {
	return nil
}

type QueueAttributes struct {
	Attributes []struct {
		Name  string
		Value string
	}
	ResponseMetadata
}

// GetQueueAttributes returns one or all attributes of a queue.
//
// See http://goo.gl/X01zD for more details.
func (q *Queue) GetQueueAttributes(attrs ...Attribute) (*QueueAttributes, error) {
	params := url.Values{}
	for i, attr := range attrs {
		key := fmt.Sprintf("Attribute.%d", i)
		params[key] = []string{string(attr)}
	}
	var resp QueueAttributes
	if err := q.get("GetQueueAttributes", q.path, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type Message struct {
	Id   string `xml:"ReceiveMessageResult>Message>MessageId"`
	Body string `xml:"ReceiveMessageResult>Message>Body"`
}

// ReceiveMessage retrieves one or more messages from the queue.
//
// See http://goo.gl/8RLI4 for more details.
func (q *Queue) ReceiveMessage() (*Message, error) {
	var resp Message
	if err := q.get("ReceiveMessage", q.path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemovePermission removes a permission from a queue for a specific principal.
//
// See http://goo.gl/5QB9W for more details.
func (q *Queue) RemovePermission() error {
	return nil
}

type sendMessageResponse struct {
	Id string `xml:"SendMessageResult>MessageId"`
	ResponseMetadata
}

// SendMessage delivers a message to the specified queue.
// It returns the sent message's ID.
//
// See http://goo.gl/ThjJG for more details.
func (q *Queue) SendMessage(body string) (string, error) {
	params := url.Values{
		"MessageBody": []string{body},
	}
	var resp sendMessageResponse
	if err := q.get("SendMessage", q.path, params, &resp); err != nil {
		return "", err
	}
	return resp.Id, nil
}

// SetQueueAttributes sets one attribute of a queue.
//
// See http://goo.gl/YtIjs for more details.
func (q *Queue) SetQueueAttributes() error {
	return nil
}
