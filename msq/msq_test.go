package msq

import (
	"context"
	"fmt"
	"testing"
	"time"

	beans "github.com/iwanbk/gobeanstalk"
)

func TestMsq(t *testing.T) {
	addr := "127.0.0.1:11300"
	tube := "testing_tube"
	eventSize := 50000
	end := make(chan bool, 1)
	dealing := make(chan bool, 1)

	// 消费者
	c := NewConsumer(addr, tube)
	handle := func(ctx context.Context, job *beans.Job, tried int) bool {
		fmt.Printf("receive job, tried:%d, job:%+v\n", tried, string(job.Body))
		dealing <- true
		return true
	}
	for i := 100; i > 0; i-- {
		go c.Reserve(10*time.Minute, handle)
	}

	// 生产者
	p := NewProducer(1000, addr, tube)
	for i := eventSize; i > 0; i-- {
		in := i
		// go func(in int) {
		if err := p.Put([]byte(fmt.Sprintf("%d", in))); err != nil {
			t.Fatal(err)
		}
		// }(i)
	}

	// 等待结束事件
	// 若1秒钟读不到数据，认为已经没有数据可读
	go func() {
		for {
			select {
			case <-time.After(1e9):
				end <- true
				return
			case <-dealing:
				// ignore
			}
		}
	}()
	<-end

	c.Close()
	p.Close()
}
