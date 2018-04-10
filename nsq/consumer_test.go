package nsq

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	nsq "github.com/nsqio/go-nsq"
)

type MyTestHandler struct {
	t                *testing.T
	q                *nsq.Consumer
	messagesSent     int
	messagesReceived int
	messagesFailed   int
}

var nullLogger = log.New(ioutil.Discard, "", log.LstdFlags)

func (h *MyTestHandler) LogFailedMessage(message *nsq.Message) {
	fmt.Printf("handle error msg:%s\n", string(message.Body))
	h.messagesFailed++
	h.q.Stop()
}

func (h *MyTestHandler) HandleMessage(message *nsq.Message) error {
	fmt.Printf("handle suc msg:%s\n", string(message.Body))
	return nil
}

var (
	addr      = "127.0.0.1:4150"
	topicName = "rdr_test"
	testP     = NewProducer(1, addr, topicName)
)

func SendMessage(t *testing.T, port int, topic string, method string, body []byte) {
	if err := testP.Put(body); err != nil {
		t.Fatal(err)
	}
}

func TestConsumer(t *testing.T) {
	consumerTest(t, nil)
}

func consumerTest(t *testing.T, cb func(c *nsq.Config)) {

	c := NewConsumer(addr, topicName)
	defer c.Close()
	go c.Reserve(10*time.Minute, func(ctx context.Context, job *Job, tried int) bool {
		fmt.Println(string(job.Body))
		return true
	})
	time.Sleep(1e9)
	SendMessage(t, 4151, topicName, "pub", []byte("testing2"))

	time.Sleep(1e9)
}
