package job

import (
	"reflect"
	"testing"
	"time"
)

func TestGetOffset1(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	tsk := NewTask(time.Second*8, func() {})
	offset, round := timer.getOffset(tsk)
	expect(t, offset, 7)
	expect(t, round, 0)

}
func TestGetOffset2(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	tsk := NewTask(time.Second*91, func() {})
	offset, round := timer.getOffset(tsk)
	expect(t, offset, 1)
	expect(t, round, 9)
}
func TestGetOffset3(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	timer.index = 4
	tsk := NewTask(time.Second*10, func() {})
	offset, round := timer.getOffset(tsk)
	expect(t, offset, 5)
	expect(t, round, 1)
}
func TestGetOffset4(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	timer.index = 4
	task := NewTask(time.Second*10, func() {})
	offset, round := timer.Add(task)
	expect(t, offset, 5)
	expect(t, round, 1)
	expect(t, len(timer.slots[offset]), 1)
}
func TestGetOffset5(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	timer.index = 4
	task := NewTask(time.Second*2, func() {})
	offset, round := timer.Add(task)
	expect(t, offset, 6)
	expect(t, round, 0)
	expect(t, len(timer.slots[offset]), 1)
	timer.execute()
	expect(t, len(timer.slots[offset]), 1)
	timer.execute()
	expect(t, len(timer.slots[offset]), 0)
}
func TestGetOffset6(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	timer.index = 4
	value := 0
	task := NewTask(time.Second*2, func() { value++ }, WithCycle())
	offset, round := timer.Add(task)
	expect(t, offset, 6)
	expect(t, round, 0)
	timer.execute()
	timer.execute()
	expect(t, len(timer.slots[offset]), 0)
	time.Sleep(time.Millisecond * 10)
	expect(t, offset+2, 8)
	expect(t, value, 1)
}
func TestGetOffset7(t *testing.T) {
	timer := NewJobServer(10, time.Second)
	timer.index = 4
	task := NewTask(time.Second*2, func() {})
	offset, round := timer.Add(task)
	expect(t, offset, 6)
	expect(t, round, 0)
	expect(t, len(timer.slots[offset]), 1)
	timer.Reset()
	expect(t, len(timer.slots[offset]), 0)
}
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
