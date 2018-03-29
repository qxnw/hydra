package log

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/logger"
)

//RemoteAppender 文件输出器
type RemoteAppender struct {
	name      string
	buffer    *bytes.Buffer
	lastWrite time.Time
	layout    *logger.Appender
	ticker    *time.Ticker
	locker    sync.Mutex
	writer    io.WriteCloser
	Level     int
}

//NewRemoteAppender 构建writer日志输出对象
func NewRemoteAppender(writer io.WriteCloser, layout *logger.Appender) (fa *RemoteAppender, err error) {
	fa = &RemoteAppender{layout: layout, writer: writer}
	fa.Level = logger.GetLevel(layout.Level)
	intervalStr := layout.Interval
	fa.buffer = bytes.NewBufferString("")
	interval, _ := time.ParseDuration(intervalStr)
	fa.ticker = time.NewTicker(interval)
	go fa.writeTo()
	return
}

//Write 写入日志
func (f *RemoteAppender) Write(event *logger.LogEvent) {
	current := logger.GetLevel(event.Level)
	if current < f.Level {
		return
	}
	f.locker.Lock()
	f.buffer.WriteString(",")
	f.buffer.WriteString(jsons.Escape(event.Output))
	f.locker.Unlock()
	f.lastWrite = time.Now()
}

//Close 关闭当前appender
func (f *RemoteAppender) Close() {
	f.Level = logger.ILevel_OFF
	f.ticker.Stop()
	f.locker.Lock()
	f.buffer.WriteTo(f.writer)
	f.writer.Close()
	f.locker.Unlock()
}

//writeTo 定时写入
func (f *RemoteAppender) writeTo() {
START:
	for {
		select {
		case _, ok := <-f.ticker.C:
			if ok {
				f.locker.Lock()
				f.buffer.WriteTo(f.writer)
				f.buffer.Reset()
				f.locker.Unlock()
			} else {
				break START
			}
		}
	}
}
