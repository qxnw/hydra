package influx

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/types"
)

func (s *influxProxy) getSaveParams(ctx *context.Context) (measurement string, tags map[string]string, fields map[string]interface{}, err error) {

	tags = make(map[string]string)
	fields = make(map[string]interface{})
	measurement, err = ctx.GetInput().Get("measurement")
	if err != nil && !types.IsEmpty(ctx.GetBody()) {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(ctx.GetBody()))
		if err != nil {
			err = fmt.Errorf("输入的body不是有效的json数据，(err:%v)", err)
			return
		}
		msm, ok := inputMap["measurement"]
		if !ok {
			err = errors.New("body的内容中未包含measurement标签")
			return
		}

		if measurement, ok = msm.(string); !ok {
			err = fmt.Errorf("body的内容中measurement标签必须为字符串:(err:%v)", msm)
			return
		}
		tgs, ok := inputMap["tags"]
		if !ok {
			err = errors.New("body的内容中未包含tags标签")
			return
		}
		flds, ok := inputMap["fields"]
		if !ok {
			err = errors.New("body的内容中未包含fields标签")
			return
		}
		tgMap, ok := tgs.(map[string]interface{})
		if !ok {
			err = errors.New("body的内容中的tags标签必须为对象，并包含多个属性")
			return
		}
		fieldMap, ok := flds.(map[string]interface{})
		if !ok {
			err = errors.New("body的内容中的fields标签必须为对象，并包含多个属性")
			return
		}

		for k, v := range tgMap {
			tags[k] = fmt.Sprintf("%v", v)
		}
		for k, v := range fieldMap {
			fields[k] = fmt.Sprintf("%v", v)
		}
		return
	}
	if err != nil {
		err = errors.New("form中未包含measurement标签")
		return
	}

	tagStr, err := ctx.GetInput().Get("tags")
	if err != nil {
		err = errors.New("form中未包含tags标签")
		return
	}
	fieldStr, err := ctx.GetInput().Get("fields")
	if err != nil {
		err = errors.New("form中未包含fields标签")
		return
	}
	tagMap, err := jsons.Unmarshal([]byte(tagStr))
	if err != nil {
		err = errors.New("form中的tags的值不是有效的json数据")
		return
	}

	fieldMap, err := jsons.Unmarshal([]byte(fieldStr))
	if err != nil {
		err = errors.New("form中的fields的值不是有效的json数据")
		return
	}
	for k, v := range tagMap {
		tags[k] = fmt.Sprintf("%v", v)
	}
	for k, v := range fieldMap {
		fields[k] = fmt.Sprintf("%v", v)
	}
	return
}

func (s *influxProxy) save(ctx *context.Context) (r string, st int, err error) {
	measurement, t, f, err := s.getSaveParams(ctx)
	if err != nil {
		return
	}
	client, err := s.getInfluxClient(ctx)
	if err != nil {
		return
	}
	err = client.Send(measurement, t, f)
	if err != nil {
		err = fmt.Errorf("save(err:%v)", err)
	}
	r = "SUCCESS"
	return
}
