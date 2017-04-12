package cron

import (
	"testing"
	"time"

	"github.com/qxnw/lib4go/ut"
)

func TestGetOffset1(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	tsk := NewTask("-", time.Second*8, time.Second*8, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.getOffset(tsk.next)
	ut.Expect(t, offset, 7)
	ut.Expect(t, round, 0)

}
func TestGetOffset2(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	tsk := NewTask("-", time.Second*91, time.Second*91, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.getOffset(tsk.next)
	ut.Expect(t, offset, 1)
	ut.Expect(t, round, 9)
}
func TestGetOffset3(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	timer.index = 4
	tsk := NewTask("-", time.Second*10, time.Second*10, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.getOffset(tsk.next)
	ut.Expect(t, offset, 5)
	ut.Expect(t, round, 1)
}
func TestGetOffset4(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	timer.index = 4
	task := NewTask("-", time.Second*10, time.Second*10, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.Add(task)
	ut.Expect(t, offset, 5)
	ut.Expect(t, round, 1)
	ut.Expect(t, len(timer.slots[offset]), 1)
}
func TestGetOffset5(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	timer.index = 4
	task := NewTask("cron.server", time.Second*2, time.Second*2, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.Add(task)
	ut.Expect(t, offset, 6)
	ut.Expect(t, round, 0)
	ut.Expect(t, len(timer.slots[offset]), 1)
	timer.execute()
	ut.Expect(t, len(timer.slots[offset]), 1)
	timer.execute()
	ut.Expect(t, len(timer.slots[offset]), 0)
}
func TestGetOffset6(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	timer.index = 4
	value := 0
	task := NewTask("order.report", time.Second*2, time.Second*2, func(t *Task) error {
		value++
		t.Result = struct {
			id   int
			name string
		}{
			id:   1,
			name: "colin"}
		return nil
	}, "order.report")
	offset, round := timer.Add(task)
	ut.Expect(t, offset, 6)
	ut.Expect(t, round, 0)
	timer.execute()
	timer.execute()
	ut.Expect(t, len(timer.slots[offset]), 0)
	time.Sleep(time.Millisecond * 10)
	ut.Expect(t, offset+2, 8)
	ut.Expect(t, value, 1)
}
func TestGetOffset7(t *testing.T) {
	timer := NewCronServer("cron.server", 10, time.Second)
	timer.index = 4
	task := NewTask("-", time.Second*2, time.Second*2, func(t *Task) error { return nil }, "order.report")
	offset, round := timer.Add(task)
	ut.Expect(t, offset, 6)
	ut.Expect(t, round, 0)
	ut.Expect(t, len(timer.slots[offset]), 1)
	timer.Reset()
	ut.Expect(t, len(timer.slots[offset]), 0)
}
