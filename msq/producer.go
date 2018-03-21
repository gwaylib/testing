package msq

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	beans "github.com/iwanbk/gobeanstalk"
)

// Producer is the interface that wrap the Put method
type Producer interface {
	io.Closer
	Put(data []byte) error
}

// ErrClosed closed by Close
var ErrClosed = errors.New("msq: closed")

func isBrokenPipeErr(err error) bool {
	if err != nil && strings.Contains(err.Error(), "broken pipe") {
		return true
	}
	return false
}

// NewProducer create Producer object.
func NewProducer(uri, tube string) Producer {
	return newConn(uri, tube)
}

var pSync = sync.Mutex{}
var pConns = map[string]Producer{}

var conns = &sync.Pool{
	New: func() interface{} {
		return nil
	},
}

func GetProducer(uri, tube string) Producer {
	if len(uri) == 0 {
		panic("need uri")
	}
	if len(tube) == 0 {
		panic("need tube name")
	}

	key := uri + "/" + tube

	p := NewProducer(uri, tube)
	pConns[key] = p.(*conn)
	return p
}

// conn implements Producer and Workder and io.Closer
type conn struct {
	uri, tube string
	conn      *beans.Conn
	mu        sync.Mutex
	closed    bool
}

func newConn(uri, tube string) *conn {
	return &conn{
		uri:  uri,
		tube: tube,
	}
}

func (p *conn) dial() error {
	if p.conn != nil {
		return nil
	}

	// conn
	kon, err := net.DialTimeout("tcp", p.uri, 20*1e9)
	if err != nil {
		return err
	}
	p.conn, _ = beans.NewConn(kon, p.uri)
	// tube
	if err := p.conn.Use(p.tube); err != nil {
		p.disconn()
		return err
	}
	return nil
}

func (p *conn) disconn() error {
	if p.conn != nil {
		p.conn.Quit()
		p.conn = nil
	}
	return nil
}

func (p *conn) Put(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed() {
		return ErrClosed
	}

	if err := p.dial(); err != nil {
		p.disconn()
		return err
	}

	_, err := p.conn.Put(data, 0, 0, 30)
	if err != nil {
		p.disconn()
		return err
	}
	return nil
}

func (p *conn) isClosed() bool {
	return p.closed
}

func (p *conn) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.disconn()
}
