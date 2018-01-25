package alarm

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/qxnw/hydra/component"

	"github.com/qxnw/lib4go/influxdb"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/types"
	"github.com/qxnw/lib4go/utility"
)

func checkAndSave(c component.IContainer, tf *transform.Transform, t int, tp string) (status int, err error) {
	status = 204
	db, err := c.GetInflux("alarm")
	if err != nil {
		return
	}
	value, err := db.QueryMaps(tf.Translate(queryMap[tp]))
	if err != nil {
		return
	}
	if t == 0 {
		//上次无消息，则不上报
		if len(value) == 0 || len(value[0]) == 0 {
			return
		}
		//上次消息是成功不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "0" {
			return
		}
		//其它情况，上次消息是失败则上报
	} else {
		//上次消息是失败，但记录时间小于5分钟，则不上报
		if len(value) > 0 && len(value[0]) > 0 && types.GetString(value[0][0]["value"]) == "1" {
			lastTime, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", fmt.Sprintf("%v", value[0][0]["time"]))
			if err != nil {
				return 204, err
			}
			if time.Now().Sub(lastTime).Minutes() < 5 {
				return 204, nil
			}
		}
	}

	return save2Influxdb(reportMap[tp], tf, db)
}
func save2Influxdb(sql string, tf *transform.Transform, db *influxdb.InfluxClient) (int, error) {
	sqls := strings.Split(sql, " ")
	measurement := sqls[0]
	tagsMap, err := utility.GetMapWithQuery(strings.Replace(sqls[1], ",", "&", -1))
	if err != nil {
		return 500, err
	}
	filedsMap, err := utility.GetMapWithQuery(strings.Replace(sqls[2], ",", "&", -1))
	if err != nil {
		return 500, err
	}
	tags := make(map[string]string)
	for k, v := range tagsMap {
		tags[k] = tf.TranslateAll(v, true)
	}
	fileds := make(map[string]interface{})
	for k, v := range filedsMap {
		fileds[k] = tf.TranslateAll(v, true)
	}
	if len(tags) == 0 || len(fileds) == 0 {
		err = fmt.Errorf("tags 或 fileds的个数不能为0")
		return 500, err
	}
	err = db.Send(measurement, tags, fileds)
	if err != nil {
		return 500, err
	}
	return 200, nil
}

func query(c component.IContainer, sql string, tf *transform.Transform) (domain []string, count []int, err error) {
	db, err := c.GetInflux("metricdb")
	if err != nil {
		return
	}

	data, err := db.QueryResponse(sql)
	if err != nil {
		return
	}
	if err = data.Error(); err != nil {
		return
	}
	domain = make([]string, 0, 2)
	count = make([]int, 0, 2)
	for _, row := range data.Results {
		for _, ser := range row.Series {
			if len(ser.Tags) > 1 {
				err = fmt.Errorf("返回的数据集包含我个tag:%v", ser.Tags)
				return nil, nil, err
			}
			for _, v := range ser.Tags {
				domain = append(domain, v)
			}
			value, err := strconv.ParseFloat(types.GetString(ser.Values[0][1]), 64)
			if err != nil {
				err = fmt.Errorf("查询返回的数据不是数字:%v", data)
				return nil, nil, err
			}
			count = append(count, int(math.Floor(value)))
		}
	}
	return
}
