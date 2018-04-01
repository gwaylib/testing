package msq

import nsq "github.com/nsqio/go-nsq"

type Delegate struct {
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

func NewDelegate() *Delegate {
	return &Delegate{
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
}

// OnError is called when the connection
// receives a FrameTypeError from nsqd
func (d *Delegate) OnError(conn *nsq.Conn, data []byte) {
}

// OnMessage is called when the connection
// receives a FrameTypeMessage from nsqd
func (d *Delegate) OnMessage(conn *nsq.Conn, msg *nsq.Message) {
	d.msg <- msg
}

// OnMessageFinished is called when the connection
// handles a FIN command from a message handler
func (d *Delegate) OnMessageFinished(*nsq.Conn, *nsq.Message) {
}

// OnMessageRequeued is called when the connection
// handles a REQ command from a message handler
func (d *Delegate) OnMessageRequeued(*nsq.Conn, *nsq.Message) {
}

// OnBackoff is called when the connection triggers a backoff state
func (d *Delegate) OnBackoff(*nsq.Conn) {
}

// OnContinue is called when the connection finishes a message without adjusting backoff state
func (d *Delegate) OnContinue(*nsq.Conn) {
}

// OnResume is called when the connection triggers a resume state
func (d *Delegate) OnResume(*nsq.Conn) {
}

// OnIOError is called when the connection experiences
// a low-level TCP transport error
func (d *Delegate) OnIOError(*nsq.Conn, error) {
}

// OnHeartbeat is called when the connection
// receives a heartbeat from nsqd
func (d *Delegate) OnHeartbeat(*nsq.Conn) {
}

// OnClose is called when the connection
// closes, after all cleanup
func (d *Delegate) OnClose(*nsq.Conn) {
}
