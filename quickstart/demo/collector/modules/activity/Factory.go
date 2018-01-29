package activity

import (
	"sync"

	"github.com/qxnw/hydra/component"
)

var standActivity *Activity
var locker sync.Mutex

//GetActivity 获取GetActivity
func GetActivity(c component.IContainer) (ac IActivity, err error) {
	if standActivity != nil {
		return standActivity, nil
	}
	locker.Lock()
	defer locker.Unlock()
	if standActivity != nil {
		return standActivity, nil
	}
	customer := &Activity{}
	customer.db, err = c.GetDefaultDB()
	if err != nil {
		return
	}
	standActivity = customer
	return standActivity, nil
}
