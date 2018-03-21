// example
// worker := NewWorker("localhost:11130", "test",func(job *beans.Job)bool{
// 	// deal job
//	// return a signal of anything has deal and can to delete job.
//	return true
// })
//
// // start a gorutine to work
// if err := worker.Work(); err != nil{
//		log.Println(err)
//		return
// }
//
// // worker.Stop() // Stop func to stop working
//
package msq

import (
	"context"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log/logger"
	"github.com/gwaylib/log/logger/adapter/stdio"
	"github.com/gwaylib/log/logger/proto"
	beans "github.com/iwanbk/gobeanstalk"
)

// 已不建议使用此回调，应使用HandleContext以便可以解决推送方法超时的问题
type HandleMsg func(job *beans.Job) bool

// HandleMsg
// function callback when reader readed a job.
//
// Spec if function return false, worker will keep job in beanstalk.
//
// 若发送不成功
// 分别间隔以1次3秒钟、30次每分钟、48次每小时再次尝试发送, 若48小时后未能发送成功，数据将被删除
type HandleContext func(ctx context.Context, job *beans.Job) bool

type Worker interface {
	// 开始工作，若正在工作中，返回正在工作中的错误
	// 如果没有工作，打开一条gorutine开始读取队列数据进行推送
	Work() error
	// 通知正在工作中的协程停止工作
	Stop()
	// 运行状态
	Status() int
	// 检查接口
	// 输入参数与预期输出进行校验
	Restart() error
}

type worker struct {
	// 日志器
	log proto.Logger
	// lock for Reader operator.
	mutex sync.Mutex

	// server addr
	addr string

	// channle name
	tubename string

	// handle which is pushed
	handle HandleContext

	// server connection
	conn *beans.Conn

	// worker status.
	deamoning bool
	// doing status
	doing bool

	// work timeout for dealock
	workout time.Duration

	// log error times
	connErrTimes int

	// log id which delete job fail.
	rmErrHistor map[uint64]bool

	relHistor map[uint64]int

	// 运行状态
	status int

	// 运行结果
	job_queue chan *beans.Job
	job_done  chan bool

	// signal command.
	sig_restart      chan bool
	sig_exit         chan bool
	sig_exit_reserve chan bool
	sig_exit_deliver chan bool
	sig_end          chan bool
}

func New(addr, tubename string, handle HandleMsg) Worker {
	return NewWorker(addr, tubename, handle)
}

// 兼容原设计而保留
func NewWorker(addr, tubename string, handle HandleMsg) Worker {
	fn := func(ctx context.Context, job *beans.Job) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return handle(job)
		}
	}
	return NewTimeOutWorker(addr, tubename, fn, 10*time.Minute)
}

func NewContextWorker(addr, tubename string, handle HandleContext) Worker {
	return NewTimeOutWorker(addr, tubename, handle, 10*time.Minute)
}

func NewTimeOutWorker(addr, tubename string, handle HandleContext, timeout time.Duration) Worker {
	return &worker{
		log:              logger.New(tubename, stdio.New(os.Stderr)),
		mutex:            sync.Mutex{},
		addr:             addr,
		tubename:         tubename,
		handle:           handle,
		workout:          timeout,
		rmErrHistor:      make(map[uint64]bool),
		relHistor:        make(map[uint64]int),
		job_queue:        make(chan *beans.Job, 1),
		job_done:         make(chan bool, 1),
		sig_exit:         make(chan bool, 1),
		sig_exit_reserve: make(chan bool, 1),
		sig_exit_deliver: make(chan bool, 1),
		sig_restart:      make(chan bool, 1),
		sig_end:          make(chan bool, 2),
	}

}

func (c *worker) Status() int {
	return c.status
}

func (c *worker) Restart() error {
	c.restart()
	return nil
}

func (c *worker) restart() {
	c.sig_exit_reserve <- true
	c.sig_exit_deliver <- true
	if !c.isDoing() {
		c.close()
	}
	<-c.sig_end // for deliver
	<-c.sig_end // for reserve

	// try again close for reserve run
	c.close()

	go c.reserve()
	go c.deliver()
}

func (c *worker) deamon() {
	for {
		// 读取数据
		select {
		case <-c.sig_restart:
			c.restart()
		case <-c.sig_exit:
			c.sig_exit_deliver <- true
			c.sig_exit_reserve <- true
			if !c.isDoing() {
				c.close()
			}
			return
		}
	}
}

func (c *worker) deliver() {
	for {
		select {
		case <-time.After(c.workout * 3):
			// reserve timeout
			c.log.Info(errors.New("reserve timeout"))
			c.sig_restart <- true
		case <-c.sig_exit_deliver:
			c.sig_end <- true
			return
		case job := <-c.job_queue:
			ctx, _ := context.WithTimeout(context.Background(), c.workout)
			if err := c.do(ctx, job); err != nil {
				c.log.Warn(err.Error())
				time.Sleep(10e9)
			}
			c.job_done <- true
		}
	}
}

func (c *worker) reserve() {
	for {
		select {
		case <-c.sig_exit_reserve:
			c.sig_end <- true
			return
		default:
			c.mutex.Lock()
			// 检查连接
			if c.conn == nil {
				if err := c.connect(); err != nil {
					c.mutex.Unlock()
					c.log.Warn(errors.As(err))
					continue
				}
			}
			c.mutex.Unlock()
			job, err := c.conn.Reserve()
			if err != nil {
				if strings.Index(err.Error(), "use of closed network connection") < 0 {
					c.log.Warn(errors.As(err))
					// 重连
					c.close()
				}
				continue
			}
			c.job_queue <- job
			<-c.job_done
		}
	}
}

func (c *worker) Work() error {
	if c.isRunning() {
		return errors.New("already in working")
	}
	c.setRunning(true)

	// 推送器守护线程
	go c.deliver()
	go c.reserve()
	go c.deamon()
	return nil
}

func (c *worker) Stop() {
	if c.isRunning() {
		c.setRunning(false)
		// insert exit signal.
		c.sig_exit <- true
		<-c.sig_end // for deliver
		<-c.sig_end // for reserve
	}
	c.close()
}

var ErrNotFound = errors.New("Not Found")

//
// if connection is exist, do close at first.
//
// 1, make a connection to server:
// if it is fail, make a warn log and do sleep(error times affect the sleepping time), then reconnect again;
// if it is successful, do a checking for the tube.
//
// 2, prepare tube.
// if watch tube fail, make a warn log and do sleep(error times affect the sleepping time),
// then reconnect again;
// if it is successful, do a checking for history of delete fail, and deal with the history.
//
// 3, deal the fail history of delete
// if deal the history fail, make a warn log and do sleepping(error times affect the sleeping time),
// then reconnect again;
//
// if all above is ok, make a read signal go to begin read job
func (c *worker) connect() error {
	// do close at first
	if c.conn != nil {
		c.close()
	}

	// connect
	c.log.Info("bean connect:" + c.tubename)
	if conn, err := beans.Dial(c.addr); err != nil {
		c.connErrTimes++
		c.dealConnErrTimes(c.connErrTimes, errors.As(err))
		return errors.As(err)
	} else {
		c.connErrTimes = 0
		c.conn = conn
	}

	// prepare
	if _, err := c.conn.Watch(c.tubename); err != nil {
		err := errors.As(err, c.tubename)
		c.log.Warn(err)
		return err
	} else {
		c.connErrTimes = 0
	}

	//deal the fail history of delete
	for id, _ := range c.rmErrHistor {
		// if delete history id is found, do delete again.
		if err := c.conn.Delete(id); err != nil {
			err := errors.As(err, id)
			c.log.Warn(err)
			if err.Equal(ErrNotFound) {
				return nil
			}
			return err
		} else {
			delete(c.rmErrHistor, id)
			delete(c.relHistor, id)
		}
	}

	c.connErrTimes = 0
	return nil
}

// deal connction error
func (c *worker) dealConnErrTimes(times int, err error) {
	if err == nil {
		err = errors.New("unknow error")
	}
	switch times {
	case 10:
		// if error times equal 10, make a error log.
		// TODO:
		// send an sms or a mail
		c.log.Error(errors.As(err))
	case 3:
		// if error times equal 3, make warnning log.
		c.log.Warn(errors.As(err))
	}

	// if error times more than 10 times, do a sleep 30 sec.
	if times > 10 {
		time.Sleep(10 * 1e9)
	} else {
		time.Sleep(1 * 1e9)
	}
}

// do job
func (c *worker) do(ctx context.Context, job *beans.Job) error {
	result := make(chan error, 1)

	// push to handle.
	// if user not confirm delete the job, not delete the job.
	go func(ctx context.Context) {
		c.setDoing(true)
		defer c.setDoing(false)

		deal := false
		defer func() {
			defer close(result)
			// recover for handle
			if r := recover(); r != nil {
				c.log.Warn(errors.New("handle panic").As(r))
				debug.PrintStack()
				time.Sleep(10e9)
				deal = false
			}
			if deal {
				// delete job
				result <- c.delJob(job.ID)
				return
			}

			// fail deal
			times, ok := c.relHistor[job.ID]
			if !ok {
				times = 0
			}
			times = times + 1
			c.relHistor[job.ID] = times
			// 若发送不成功
			// 分别间隔以1次3秒钟、30次每分钟、48次每小时再次尝试发送, 若48小时后未能发送成功，数据将被删除
			sleep := 10
			if times < 2 {
				sleep = 3 // 3 sec.
			} else if times < 30 {
				sleep = 60 // 1 minute
			} else if times < 78 {
				sleep = 60 * 60 // 1 hour
			} else {
				c.log.Warn(errors.New("delete data").As(string(job.Body)))
				// 48+6次后删除数据
				// delete job
				result <- c.delJob(job.ID)
				return
			}
			if err := c.conn.Release(job.ID, 0, time.Duration(sleep*1e9)); err != nil {
				if ErrNotFound.Equal(err) {
					result <- nil
					return
				}
				c.log.Error(errors.As(err, job))
			}
			result <- nil
			return
		}()
		deal = c.handle(ctx, job)
	}(ctx)
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return errors.New("handle time out").As(job)
	}
}

func (c *worker) delJob(id uint64) error {
	if err := c.conn.Delete(id); err != nil {
		if ErrNotFound.Equal(err) {
			delete(c.rmErrHistor, id)
			return nil
		}

		err := errors.As(err, id)
		c.log.Warn(err)
		c.rmErrHistor[id] = true
		return err
	}

	delete(c.rmErrHistor, id)
	return nil
}

func (c *worker) close() {
	if c.conn != nil {
		c.conn.Quit()
		c.conn = nil
		c.log.Info("bean closed:" + c.tubename)
	}
}

func (c *worker) isRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.deamoning
}

func (c *worker) setRunning(run bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.deamoning = run
}

func (c *worker) isDoing() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.doing
}

func (c *worker) setDoing(do bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.doing = do
}
