package cron

import (
	"time"

	"github.com/qxnw/lib4go/logger"
)

//Task 任务
type Schedule interface {
	Next(time.Time) time.Time
}
type ITask interface {
	logger.ILogger
	Reset(s *CronServer, l *logger.Logger)
	GetName() string
	DoNext()
	NextTime() time.Time
	Invoke()
}

type Task struct {
	taskName string
	*logger.Logger
	params     interface{}
	server     *CronServer
	schedule   Schedule
	idx        int
	handle     func(*Task) error
	err        error
	Result     interface{}
	statusCode int
}

//NewTask 构建执行任务
func NewTask(taskName string, s Schedule, handle func(*Task) error, params interface{}) *Task {
	t := &Task{taskName: taskName, schedule: s, params: params, handle: handle}
	return t
}
func (ctx *Task) GetName() string {
	return ctx.taskName
}
func (ctx *Task) Reset(s *CronServer, l *logger.Logger) {
	ctx.idx = 0
	ctx.server = s
	ctx.Logger = l
}
func (ctx *Task) DoNext() {
	ctx.idx += 1
	ctx.Invoke()
}
func (ctx *Task) NextTime() time.Time {
	return ctx.schedule.Next(time.Now())
}

func (ctx *Task) Invoke() {
	if ctx.idx < len(ctx.server.handlers) {
		ctx.server.handlers[ctx.idx].Handle(ctx)
	} else {
		ctx.err = ctx.handle(ctx)
	}
}
