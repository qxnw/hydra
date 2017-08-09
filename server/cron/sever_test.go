package cron

import (
	"testing"
	"time"

	"github.com/qxnw/lib4go/ut"
	"github.com/zkfy/cron"
)

//time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
func TestGetOffset1(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 10, time.Second, WithStartTime(start))
	cronStr := "@every 8s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)
	next := s.Next(start)
	offset, round := timer.getOffset(start, next)
	ut.Expect(t, offset, 9)
	ut.Expect(t, round, 0)

}
func TestGetOffset2(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 10, time.Second, WithStartTime(start))
	cronStr := "@every 11s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)
	next := s.Next(start)
	offset, round := timer.getOffset(start, next)
	ut.Expect(t, offset, 2)
	ut.Expect(t, round, 1)

}
func TestGetOffset3(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 10, time.Second, WithStartTime(start))
	timer.index = 4
	cronStr := "@every 10s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)
	next := s.Next(start)
	offset, round := timer.getOffset(start, next)
	ut.Expect(t, offset, 5)
	ut.Expect(t, round, 1)
}

type offsetTask struct {
	*Task
}

func (ctx *offsetTask) NextTime(start time.Time) time.Time {
	return ctx.schedule.Next(start)
}
func TestGetOffset5(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 10, time.Second, WithStartTime(start))
	timer.index = 4

	cronStr := "@every 2s"
	s, err := cron.ParseStandard(cronStr)
	ut.Expect(t, err, nil)

	offTask := &offsetTask{}
	offTask.Task = NewTask("cron", s, func(t *Task) error { return nil }, "order.report")
	offset, round, err := timer.Add(offTask)
	ut.ExpectSkip(t, err, nil)
	ut.ExpectSkip(t, offset, 6)
	ut.ExpectSkip(t, round, 0)
	ut.Expect(t, timer.slots[offset].Count(), 1)
	//timer.execute()
	//ut.Expect(t, timer.slots[offset].Count(), 1)
	timer.execute()
	ut.Expect(t, timer.slots[offset].Count(), 1)
	timer.execute()
	ut.Expect(t, timer.slots[offset].Count(), 0)

}
func TestGetOffset6(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 10, time.Second, WithStartTime(start))
	timer.index = 4
	value := 0

	cronStr := "@every 2s"
	s, err := cron.ParseStandard(cronStr)
	ut.Expect(t, err, nil)

	task := NewTask("order.report", s, func(t *Task) error {
		value++
		t.Result = struct {
			id   int
			name string
		}{
			id:   1,
			name: "colin"}
		return nil
	}, "order.report")
	offTask := &offsetTask{Task: task}
	offset, round, err := timer.Add(offTask)
	ut.ExpectSkip(t, err, nil)
	ut.ExpectSkip(t, offset, 6)
	ut.ExpectSkip(t, round, 0)
	timer.execute()
	timer.execute()
	timer.execute()
	ut.ExpectSkip(t, timer.slots[offset].Count(), 0)
	time.Sleep(time.Millisecond * 10)
	ut.ExpectSkip(t, offset+3, 9)
	ut.ExpectSkip(t, value, 1)
}
func TestGetOffset7(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron.server", 10, time.Second, WithStartTime(start))
	timer.index = 4

	cronStr := "@every 2s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)

	task := NewTask("-", s, func(t *Task) error { return nil }, "order.report")
	offTask := &offsetTask{Task: task}
	offset, round, err := timer.Add(offTask)
	ut.ExpectSkip(t, err, nil)
	ut.ExpectSkip(t, offset, 6)
	ut.ExpectSkip(t, round, 0)
	ut.ExpectSkip(t, timer.slots[offset].Count(), 1)
	timer.Reset()
	ut.ExpectSkip(t, timer.slots[offset].Count(), 0)
}
func TestGetOffset8(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 2, time.Second, WithStartTime(start))
	cronStr := "@every 2s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)
	next := s.Next(start)
	offset, round := timer.getOffset(start, next)
	ut.Expect(t, offset, 1)
	ut.Expect(t, round, 1)
}
func TestGetOffset9(t *testing.T) {
	start, _ := time.Parse("2006/01/02 15:04:05", "2099/10/10 10:11:00")
	timer := NewCronServer("hydra", "cron", 2, time.Second, WithStartTime(start))
	cronStr := "@every 4s"
	s, err := cron.ParseStandard(cronStr)
	ut.ExpectSkip(t, err, nil)
	next := s.Next(start)
	offset, round := timer.getOffset(start, next)
	ut.Expect(t, offset, 1)
	ut.Expect(t, round, 2)
}
