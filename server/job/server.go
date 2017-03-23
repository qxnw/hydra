package job

import (
	"sync"
	"time"
)

const (
	//Once 单次执行任务
	Once = iota
	//Cycle 循环执行任务
	Cycle
)

//Task 任务
type Task struct {
	span     time.Duration
	round    int
	executed int
	handle   func()
	*taskOption
}
type taskOption struct {
	tp  int
	max int
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

//WithMax 设置最大执行次数
func WithMax(max int) TaskOption {
	return func(o *taskOption) {
		o.max = max
	}
}

//NewTask 构建执行任务
func NewTask(span time.Duration, handle func(), opts ...TaskOption) *Task {
	t := &Task{span: span, handle: handle, taskOption: &taskOption{}}
	for _, opt := range opts {
		opt(t.taskOption)
	}
	return t
}

//WheelTimer 基于HashedWheelTimer算法的定时任务
type WheelTimer struct {
	length    int
	index     int
	span      time.Duration
	done      bool
	close     chan struct{}
	slots     [][]*Task
	startTime time.Time
	mu        sync.Mutex
}

//NewWheelTimer 构建定时任务
func NewWheelTimer(length int, span time.Duration) (w *WheelTimer) {
	w = &WheelTimer{length: length, span: span, index: -1, startTime: time.Now()}
	w.close = make(chan struct{}, 1)
	w.slots = make([][]*Task, length, length)
	go w.move()
	return w
}
func (w *WheelTimer) handle(task *Task) {
	task.handle()
	if task.tp == Cycle && (task.max == 0 || task.executed < task.max) {
		w.Add(task)
	}
}

//GetOffset 获取当前任务的偏移量
func (w *WheelTimer) getOffset(task *Task) (offset int, round int) {
	deadline := time.Now().Add(task.span).Sub(w.startTime) //剩余时间
	tick := int(deadline / w.span)                         //总格数
	remain := w.length - w.index - 1
	offset = tick + w.index //相当于当前位置的偏移量
	round = 0
	if tick > remain {
		round = (tick-remain)/w.length + 1
		offset = (tick - remain) % w.length
	}
	return
}

//Add 添加任务
func (w *WheelTimer) Add(task *Task) (offset int, round int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	offset, round = w.getOffset(task)
	task.round = round
	w.slots[offset] = append(w.slots[offset], task)
	return
}

//Reset 重置
func (w *WheelTimer) Reset() {
	w.mu.Lock()
	w.slots = make([][]*Task, w.length, w.length)
	w.mu.Unlock()
}

//Close 关闭
func (w *WheelTimer) Close() {
	w.done = true
	w.close <- struct{}{}
}
func (w *WheelTimer) execute() {
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
func (w *WheelTimer) move() {
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
