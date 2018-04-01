package msq

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMsq(t *testing.T) {
	addr := "127.0.0.1:11300"
	tube := "testing_tube"
	end := make(chan bool, 1)
	dealing := make(chan bool, 1)

	// 消费者
	c := NewConsumer(addr, tube)
	result := sync.Map{}
	_ = result
	handle := func(ctx context.Context, job *Job, tried int) bool {
		// 检查并发消息的安全性, 若出现重复，说明读取是不安全的
		oldID, ok := result.LoadOrStore(string(job.Body), job.ID)
		if ok {
			t.Fatal(fmt.Sprintf("repeated:%d,%d,%s", oldID, job.ID, string(job.Body)))
		}
		fmt.Printf("receive job, tried:%d, job:%+v\n", tried, string(job.Body))
		dealing <- true
		return true
	}
	for i := 1; i > 0; i-- {
		go c.Reserve(10*time.Minute, handle)
	}
	// 等待消费者就绪
	time.Sleep(1e9)

	// 生产者
	p := NewProducer(1000, addr, tube)
	eventSize := 1 // 50000
	seed := time.Now().UnixNano()
	for i := eventSize; i > 0; i-- {
		in := i
		// go func(in int) {
		if err := p.Put([]byte(fmt.Sprintf("%d", seed+int64(in)))); err != nil {
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
