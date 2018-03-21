package msq

import (
	"testing"
	"time"

	"github.com/gwaylib/log"
	beans "github.com/iwanbk/gobeanstalk"
)

func TestNewWorker(t *testing.T) {
	handle := func(job *beans.Job) bool {
		println("reviece a job:" + string(job.Body))
		return true
	}
	log.Debug("test")
	worker := NewWorker("127.0.0.1:11300", "test_tube", handle)
	go worker.Work()
	time.Sleep(10 * 1e9)
	worker.Stop()
	time.Sleep(3 * 1e9)
}
