package activity

//IActivity 活动状态
type IActivity interface {
	CreateActivity(title string, start string, end string, tag string, url string, amount int, desc string) (id string, err error)
}
