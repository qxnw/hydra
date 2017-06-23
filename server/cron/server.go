package cron

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"fmt"

	"github.com/qxnw/hydra/server"
	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/utility"
)

type cronOption struct {
	ip           string
	registryRoot string
	registry     server.IServiceRegistry
	metric       *InfluxMetric
	startTime    time.Time
}

//CronOption 任务设置选项
type CronOption func(*cronOption)

//WithIP 设置为循环执行任务
func WithIP(ip string) CronOption {
	return func(o *cronOption) {
		o.ip = ip
	}
}

//WithStartTime 设置开始时间
func WithStartTime(tm time.Time) CronOption {
	return func(o *cronOption) {
		o.startTime = tm
	}
}

//WithRegistry 添加服务注册组件
func WithRegistry(r server.IServiceRegistry, root string) CronOption {
	return func(o *cronOption) {
		o.registry = r
		o.registryRoot = root
	}
}

type Handler interface {
	Handle(*Task)
}

type HandlerFunc func(ctx *Task)

func (h HandlerFunc) Handle(ctx *Task) {
	h(ctx)
}

//CronServer 基于HashedWheelTimer算法的定时任务
type CronServer struct {
	domain     string
	serverName string
	length     int
	index      int
	span       time.Duration
	done       bool
	*logger.Logger
	close       chan struct{}
	once        sync.Once
	slots       []cmap.ConcurrentMap
	clusterPath string
	handlers    []Handler
	mu          sync.Mutex
	taskCount   int32
	running     bool
	*cronOption
}

//NewCronServer 构建定时任务
func NewCronServer(domain string, name string, length int, span time.Duration, opts ...CronOption) (w *CronServer) {
	w = &CronServer{domain: domain, serverName: name, length: length, span: span, index: 0}
	w.Logger = logger.GetSession("hydra.cron", logger.CreateSession())
	w.cronOption = &cronOption{metric: NewInfluxMetric(), startTime: time.Now()}
	w.close = make(chan struct{})
	w.handlers = make([]Handler, 0, 3)
	w.slots = make([]cmap.ConcurrentMap, length, length)
	for _, opt := range opts {
		opt(w.cronOption)
	}
	for i := 0; i < length; i++ {
		w.slots[i] = cmap.New(2)
	}
	w.handlers = append(w.handlers, Recovery(), Logging(), w.metric)
	return w
}
func (w *CronServer) handle(task *cronTask) {
	task.task.Invoke()
	_, _, err := w.Add(task.task)
	if err != nil {
		fmt.Println("err:", err)
	}
}

//Start start cron server
func (w *CronServer) Start() error {
	w.Infof("start cron server(%s)[%d]", w.serverName, atomic.LoadInt32(&w.taskCount))
	err := w.registryServer()
	if err != nil {
		w.running = false
		return err
	}
	w.running = true
	go w.move()
	return nil
}
func (w *CronServer) getOffset(next time.Time) (pos int, circle int) {
	d := next.Sub(time.Now()) //剩余时间
	delaySeconds := int(d/1e9) + 1
	intervalSeconds := int(w.span.Seconds())
	circle = int(delaySeconds / intervalSeconds / w.length)
	pos = int(w.index+delaySeconds/intervalSeconds) % w.length
	return
}

/*
//GetOffset 获取当前任务的偏移量
func (w *CronServer) getOffset(next time.Time) (offset int, round int) {
	deadline := next.Sub(w.startTime) //剩余时间
	tick := int(deadline / w.span)    //总格数
	remain := w.length - w.index - 1
	offset = tick + w.index + 1 //相当于当前位置的偏移量
	round = 0
	if tick >= remain {
		round = (tick-remain)/w.length + 1
		offset = (tick - remain) % w.length
	} else {
		offset = offset % w.length
	}
	if offset < 0 {
		offset = 0
	}
	return
}
*/
//Add 添加任务
func (w *CronServer) Add(task ITask) (offset int, round int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	nextTime := task.NextTime()
	if nextTime.Sub(time.Now()) < 0 {
		return -1, -1, errors.New("next time less than now.1")
	}
	offset, round = w.getOffset(nextTime)
	if offset < 0 || round < 0 {
		return -1, -1, errors.New("next time less than now.2")
	}
	ctask := &cronTask{task: task, round: round}
	ctask.task.Reset(w, logger.GetSession(task.GetName(), logger.CreateSession()))
	if !w.done {
		w.slots[offset].Set(utility.GetGUID(), ctask)
		atomic.AddInt32(&w.taskCount, 1)
	}

	return
}

//Reset 重置
func (w *CronServer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	atomic.SwapInt32(&w.taskCount, 0)
	for i := 0; i < len(w.slots); i++ {
		w.slots[i].RemoveIterCb(func(k string, v interface{}) bool {
			return true
		})
	}
}

//Close 关闭
func (w *CronServer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running {
		return
	}
	w.running = false
	w.unregistryServer()
	w.done = true
	w.once.Do(func() {
		close(w.close)
	})
	atomic.SwapInt32(&w.taskCount, 0)
	for i := 0; i < len(w.slots); i++ {
		w.slots[i].RemoveIterCb(func(k string, v interface{}) bool {
			v = nil
			return true
		})
	}
	w.Infof("cron: Server closed(%s)", w.serverName)
}
func (w *CronServer) execute() {
	w.startTime = time.Now()
	w.mu.Lock()
	w.index = (w.index + 1) % w.length
	current := w.slots[w.index]
	current.RemoveIterCb(func(k string, value interface{}) bool {
		task := value.(*cronTask)
		if task.round == 0 {
			go w.handle(task)
			atomic.AddInt32(&w.taskCount, -1)
			return true
		}
		task.round--
		task.executed++
		return false
	})
	w.mu.Unlock()
}
func (w *CronServer) move() {
	time.Sleep(time.Second * 3) //延迟3秒执行，等待服务器就绪
START:
	for {
		select {
		case <-w.close:
			break START
		case <-time.After(w.span):
			w.execute()
		}
	}
}

//SetInfluxMetric 重置metric
func (w *CronServer) SetInfluxMetric(host string, dataBase string, userName string, password string, timeSpan time.Duration) {
	w.metric.RestartReport(host, dataBase, userName, password, timeSpan, w.Logger)
}

//StopInfluxMetric stop metric
func (w *CronServer) StopInfluxMetric() {
	w.metric.Stop()
}

//SetName 设置组件的server name
func (w *CronServer) SetName(name string) {
	w.serverName = name
}
