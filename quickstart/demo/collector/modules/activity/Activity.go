package activity

import (
	"energy/coupon-services/conf/sql/activity"
	"fmt"

	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/logger"
)

//Activity 活动模块
type Activity struct {
	db     *db.DB
	logger *logger.Logger
}

func (n *Activity) CreateActivity(title string, start string, end string, tag string, url string, amount int, desc string) (id string, err error) {
	activityId, _, _, err := n.db.Scalar(activity.SqlGetActivityId, nil)
	if err != nil {
		err = fmt.Errorf("创建活动时.获取新的活动编号失败：%+v", err)
		return
	}
	params := map[string]interface{}{
		"activity_id":  activityId,
		"title":        title,
		"begin_time":   start,
		"end_time":     end,
		"activity_url": url,
		"activity_tag": tag,
		"description":  desc,
		"total_amount": amount,
	}

	_, _, _, err = n.db.Execute(sqlCreate, params)
	if err != nil {
		err = fmt.Errorf("创建活动时.保存活动数据到数据库失败：%+v", err)
		return
	}
	id = fmt.Sprintf("%v", activityId)
	return
}
