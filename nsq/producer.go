package nsq

import (
	"io"
	"strings"
	"sync"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
	nsq "github.com/nsqio/go-nsq"
)

// Producer is the interface that wrap the Put method
type Producer interface {
	io.Closer
	Put(data []byte) error
}

type producer struct {
	addr        string
	tube        string
	borrowEvent chan bool
	poolSync    sync.Mutex
	curPoolSize int
	maxPoolSize int
	isClosed    bool
	pool        sync.Pool
}

// 通过一个连接池发送数据给beanstalkd，若需要顺序发送，请将池设定为1
func (p *producer) Put(data []byte) error {
	p.poolSync.Lock()
	if p.isClosed {
		return errors.New("producer has closed")
	}
	p.poolSync.Unlock()

	// 借调事件, 若超过池的大小，需要等待池的归还后才能继续
	p.borrowEvent <- true
	defer func() {
		<-p.borrowEvent
	}()

	conn := p.pool.Get().(*conn)
	defer p.pool.Put(conn)

	if err := conn.put(data); err != nil {
		return errors.As(err)
	}
	return nil
}

func (p *producer) Close() error {
	p.poolSync.Lock()
	p.isClosed = true
	p.poolSync.Unlock()

	for i := p.maxPoolSize; i > 0; i-- {
		p.borrowEvent <- true
	}
	// 等待所有输出都完成后关闭各个连接
	for i := p.maxPoolSize; i > 0; i-- {
		<-p.borrowEvent
	}

	for i := p.curPoolSize; i > 0; i-- {
		conn := p.pool.Get().(*conn)
		conn.disconn()
	}
	return nil
}

// NewProducer create Producer object.
func NewProducer(size int, addr, tube string) Producer {
	if size < 1 {
		panic("need size > 0")
	}
	p := &producer{
		addr:        addr,
		tube:        tube,
		borrowEvent: make(chan bool, size),
	}
	p.pool.New = func() interface{} {
		p.poolSync.Lock()
		defer p.poolSync.Unlock()
		p.curPoolSize += 1
		return newConn(addr, tube)
	}
	return p
}

// ErrClosed closed by Close
var ErrClosed = errors.New("msq: closed")

func isBrokenPipeErr(err error) bool {
	if err != nil && strings.Contains(err.Error(), "broken pipe") {
		return true
	}
	return false
}

// conn implements Producer and Workder and io.Closer
type conn struct {
	addr, tube string
	mu         sync.Mutex
	conn       *nsq.Conn
	closed     bool
}

func newConn(addr, tube string) *conn {
	return &conn{
		addr: addr,
		tube: tube,
	}
}

func (p *conn) connect() error {
	if p.conn != nil {
		return nil
	}

	c := nsq.NewConn(p.addr, nsq.NewConfig(), NewDelegate("producer"))
	_, err := c.Connect()
	if err != nil {
		return err
	}
	p.conn = c
	return nil
}

func (p *conn) disconn() error {
	if p.conn != nil {
		if err := p.conn.Flush(); err != nil {
			log.Warn(errors.As(err))
		}
		if err := p.conn.Close(); err != nil {
			log.Warn(errors.As(err))
		}
		p.conn = nil
	}
	p.closed = true
	return nil
}

func (p *conn) isClosed() bool {
	return p.closed
}

func (p *conn) put(data []byte) error {
	if p.isClosed() {
		return ErrClosed
	}

	if err := p.connect(); err != nil {
		p.disconn()
		return errors.As(err)
	}

	if err := p.conn.WriteCommand(nsq.Publish(p.tube, data)); err != nil {
		p.disconn()
		return errors.As(err)
	}
	return nil
}
