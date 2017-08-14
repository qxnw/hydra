package context

import (
	"fmt"

	"github.com/qxnw/lib4go/utility"
)

type tLogger struct {
	session string
}

func (t *tLogger) SetTag(name string, value string) {

}
func (t *tLogger) Printf(format string, content ...interface{}) {
	fmt.Printf(format, content...)
}
func (t *tLogger) Print(content ...interface{}) {
	fmt.Print(content...)

}
func (t *tLogger) Println(args ...interface{}) {
	fmt.Println(args...)
}

func (t *tLogger) Infof(format string, content ...interface{}) {
	fmt.Printf(format, content)
}
func (t *tLogger) Info(content ...interface{}) {
	fmt.Print(content...)
}

func (t *tLogger) Errorf(format string, content ...interface{}) {
	fmt.Printf(format, content...)
}
func (t *tLogger) Error(content ...interface{}) {
	fmt.Println(content...)
}

func (t *tLogger) Debugf(format string, content ...interface{}) {
	fmt.Printf(format, content...)
}
func (t *tLogger) Debug(content ...interface{}) {
	fmt.Println(content...)
}

func (t *tLogger) Fatalf(format string, content ...interface{}) {
	fmt.Printf(format, content...)
}
func (t *tLogger) Fatal(content ...interface{}) {
	fmt.Println(content...)
}
func (t *tLogger) Fatalln(args ...interface{}) {
	fmt.Println(args...)
}

func (t *tLogger) Warnf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
func (t *tLogger) Warn(v ...interface{}) {
	fmt.Println(v...)
}
func (t *tLogger) GetSessionID() string {
	if t.session == "" {
		t.session = utility.GetGUID()[0:8]
	}
	return t.session
}
