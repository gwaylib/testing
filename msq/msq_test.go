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
	end := make(chan bool, eventSize)

	// 消费者
	c := NewConsumer(addr, tube)
	handle := func(ctx context.Context, job *beans.Job, tried int) bool {
		end <- true
		fmt.Printf("receive job, tried:%d, job:%+v\n", tried, string(job.Body))
		return true
	}
	for i := 10; i > 0; i-- {
		go c.Reserve(10*time.Minute, handle)
	}

	// 生产者
	p := NewProducer(100, addr, tube)
	for i := eventSize; i > 0; i-- {
		go func(in int) {
			if err := p.Put([]byte(fmt.Sprintf("%d", in))); err != nil {
				t.Fatal(err)
			}
		}(i)
	}

	// 等待结束事件
	for i := eventSize; i > 0; i-- {
		<-end
	}

	fmt.Println("end 1")
	c.Close()
	fmt.Println("end 2")
	p.Close()
	fmt.Println("end 3")
}
