package cron

import (
	"sync"
	"time"

	"os"

	"github.com/qxnw/hydra/context"
)

const (
	//Once 单次执行任务
	Once = iota
	//Cycle 循环执行任务
	Cycle
)

type taskOption struct {
	ip       string
	tp       int
	max      int
	logger   context.Logger
	registry context.IServiceRegistry
	metric   *InfluxMetric
}

//TaskOption 任务设置选项
type TaskOption func(*taskOption)

//WithOnce 设置为单次执行任务
func WithOnce() TaskOption {
	return func(o *taskOption) {
		o.tp = Once
	}
}

//WithCycle 设置为循环执行任务
func WithCycle() TaskOption {
	return func(o *taskOption) {
		o.tp = Cycle
	}
}

//WithIP 设置为循环执行任务
func WithIP(ip string) TaskOption {
	return func(o *taskOption) {
		o.ip = ip
	}
}

//WithRegister 设置服务注册组件
func WithRegister(i context.IServiceRegistry) TaskOption {
	return func(o *taskOption) {
		o.registry = i
	}
}

//WithMax 设置最大执行次数
func WithMax(max int) TaskOption {
	return func(o *taskOption) {
		o.max = max
	}
}

//WithLogger 设置日志记录组件
func WithLogger(logger context.Logger) TaskOption {
	return func(o *taskOption) {
		o.logger = logger
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
	serverName string
	length     int
	index      int

	span      time.Duration
	done      bool
	close     chan struct{}
	slots     [][]*Task
	startTime time.Time

	handlers []Handler
	mu       sync.Mutex
	*taskOption
}

//NewCronServer 构建定时任务
func NewCronServer(name string, length int, span time.Duration, opts ...TaskOption) (w *CronServer) {
	w = &CronServer{serverName: name, length: length, span: span, index: -1, startTime: time.Now()}
	w.taskOption = &taskOption{metric: NewInfluxMetric(), logger: NewLogger(name, os.Stdout)}
	w.close = make(chan struct{}, 1)
	w.handlers = make([]Handler, 0, 3)
	w.slots = make([][]*Task, length, length)
	for _, opt := range opts {
		opt(w.taskOption)
	}
	w.handlers = append(w.handlers, Logging(), Recovery(), w.metric)
	return w
}
func (w *CronServer) handle(task *Task) {
	task.invoke()
	if task.tp == Cycle && (task.max == 0 || task.executed < task.max) {
		task.next = task.span
		w.Add(task)
	}
}
func (w *CronServer) Start() {
	go w.move()
}

//GetOffset 获取当前任务的偏移量
func (w *CronServer) getOffset(span time.Duration) (offset int, round int) {
	deadline := time.Now().Add(span).Sub(w.startTime) //剩余时间
	tick := int(deadline / w.span)                    //总格数
	remain := w.length - w.index - 1
	offset = tick + w.index //相当于当前位置的偏移量
	round = 0
	if tick > remain {
		round = (tick-remain)/w.length + 1
		offset = (tick - remain) % w.length
	}
	if offset < 0 {
		offset = 0
	}
	return
}

//Add 添加任务
func (w *CronServer) Add(task *Task) (offset int, round int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	task.server = w
	offset, round = w.getOffset(task.next)
	task.round = round
	w.slots[offset] = append(w.slots[offset], task)
	return
}

//Reset 重置
func (w *CronServer) Reset() {
	w.mu.Lock()
	w.slots = make([][]*Task, w.length, w.length)
	w.mu.Unlock()
}

//Close 关闭
func (w *CronServer) Close() {
	w.done = true
	w.close <- struct{}{}
	w.slots = make([][]*Task, 0, 0)
}
func (w *CronServer) execute() {
	w.startTime = time.Now()
	w.mu.Lock()
	w.index = (w.index + 1) % w.length
	for i, task := range w.slots[w.index] {
		task.round--
		task.executed++
		if task.round <= 0 {
			go w.handle(task)
			copy(w.slots[w.index][i:], w.slots[w.index][i+1:])
			w.slots[w.index] = w.slots[w.index][:len(w.slots[w.index])-1]
		}
	}
	w.mu.Unlock()
}
func (w *CronServer) move() {
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
	w.metric.RestartReport(host, dataBase, userName, password, timeSpan)
}

//StopInfluxMetric stop metric
func (w *CronServer) StopInfluxMetric() {
	w.metric.Stop()
}

//SetName 设置组件的server name
func (w *CronServer) SetName(name string) {
	w.serverName = name
}
