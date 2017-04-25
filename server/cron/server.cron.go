package cron

type cronTask struct {
	task     ITask
	round    int
	executed int
	idx      int
}
