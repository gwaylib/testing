package nsq

import (
	"fmt"

	nsq "github.com/nsqio/go-nsq"
)

type Delegate struct {
	name     string
	resp     chan []byte
	err      chan []byte
	msg      chan *nsq.Message
	finished chan *nsq.Message
	requeue  chan *nsq.Message
	backoff  chan bool
	goOn     chan bool
	resume   chan bool
	ioErr    chan error
	hearbeat chan bool
	close    chan bool
}

func NewDelegate(name string) *Delegate {
	return &Delegate{
		name:     name,
		resp:     make(chan []byte, 1),
		err:      make(chan []byte, 1),
		msg:      make(chan *nsq.Message, 1),
		finished: make(chan *nsq.Message, 1),
		requeue:  make(chan *nsq.Message, 1),
		backoff:  make(chan bool, 1),
		goOn:     make(chan bool, 1),
		resume:   make(chan bool, 1),
		ioErr:    make(chan error, 1),
		hearbeat: make(chan bool, 1),
		close:    make(chan bool, 1),
	}
}

// OnResponse is called when the connection
// receives a FrameTypeResponse from nsqd
func (d *Delegate) OnResponse(conn *nsq.Conn, data []byte) {
	// fmt.Println(d.name + " on response:" + string(data))
}

// OnError is called when the connection
// receives a FrameTypeError from nsqd
func (d *Delegate) OnError(conn *nsq.Conn, data []byte) {
	fmt.Println(d.name + " on error:" + string(data))
}

// OnMessage is called when the connection
// receives a FrameTypeMessage from nsqd
func (d *Delegate) OnMessage(conn *nsq.Conn, msg *nsq.Message) {
	d.msg <- msg
}

// OnMessageFinished is called when the connection
// handles a FIN command from a message handler
func (d *Delegate) OnMessageFinished(conn *nsq.Conn, msg *nsq.Message) {
	fmt.Printf("%s on msg finished:%+v\n", d.name, *msg)
}

// OnMessageRequeued is called when the connection
// handles a REQ command from a message handler
func (d *Delegate) OnMessageRequeued(conn *nsq.Conn, msg *nsq.Message) {
	fmt.Printf("%s on msg requeue:%+v\n", d.name, *msg)
}

// OnBackoff is called when the connection triggers a backoff state
func (d *Delegate) OnBackoff(*nsq.Conn) {
	fmt.Println(d.name + " on backoff")
}

// OnContinue is called when the connection finishes a message without adjusting backoff state
func (d *Delegate) OnContinue(*nsq.Conn) {
	fmt.Println(d.name + " on continue")
}

// OnResume is called when the connection triggers a resume state
func (d *Delegate) OnResume(*nsq.Conn) {
	fmt.Println("on resume")
}

// OnIOError is called when the connection experiences
// a low-level TCP transport error
func (d *Delegate) OnIOError(conn *nsq.Conn, err error) {
	fmt.Println(d.name + " OnIOError:" + err.Error())
}

// OnHeartbeat is called when the connection
// receives a heartbeat from nsqd
func (d *Delegate) OnHeartbeat(*nsq.Conn) {
	fmt.Println(d.name + " on heart beat ")
}

// OnClose is called when the connection
// closes, after all cleanup
func (d *Delegate) OnClose(*nsq.Conn) {
	fmt.Println(d.name + " on close")
}
