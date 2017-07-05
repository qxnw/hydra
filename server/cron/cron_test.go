package cron

import (
	"testing"

	"time"

	"github.com/qxnw/lib4go/ut"
	"github.com/zkfy/cron"
)

func TestCron1(t *testing.T) {
	s, err := cron.ParseStandard("@every 1h")
	ut.Expect(t, err, nil)
	now := time.Now()
	next := s.Next(now)
	ut.Expect(t, int(next.Sub(now)/1e9)+1, 3600)
}
func TestCron2(t *testing.T) {

	server := NewCronServer("-", "-", 60, time.Second)

	s, err := cron.ParseStandard("@every 1h")
	ut.Expect(t, err, nil)
	now := time.Now()
	next := s.Next(now)
	p, c := server.getOffset(next)

	ut.Expect(t, p, 0)
	ut.Expect(t, c, 60)

}
