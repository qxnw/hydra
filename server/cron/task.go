package cron

import (
	"time"

	"github.com/qxnw/lib4go/logger"
)

//Task 任务
type Task struct {
	taskName string
	*logger.Logger
	params     interface{}
	server     *CronServer
	next       time.Duration
	span       time.Duration
	round      int
	executed   int
	idx        int
	handle     func(*Task) error
	err        error
	Result     interface{}
	statusCode int
}

//NewTask 构建执行任务
func NewTask(taskName string, next time.Duration, span time.Duration, handle func(*Task) error, params interface{}) *Task {
	t := &Task{taskName: taskName, span: span, next: next, params: params, handle: handle}

	return t
}
func (ctx *Task) Next() {
	ctx.idx += 1
	ctx.invoke()
}
func (ctx *Task) invoke() {
	if ctx.idx < len(ctx.server.handlers) {
		ctx.server.handlers[ctx.idx].Handle(ctx)
	} else {
		ctx.err = ctx.handle(ctx)
	}
}
