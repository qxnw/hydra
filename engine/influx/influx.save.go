package influx

import (
	"errors"
	"fmt"

	"github.com/qxnw/hydra/context"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
)

func (s *influxProxy) getSaveParams(ctx *context.Context) (measurement string, tags map[string]string, fields map[string]interface{}, err error) {
	if ctx.Input.Input == nil || ctx.Input.Args == nil || ctx.Input.Params == nil {
		err = fmt.Errorf("input,params,args不能为空:%v", ctx.Input)
		return
	}
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	input := ctx.Input.Input.(transform.ITransformGetter)
	measurement, err = input.Get("measurement")
	if ctx.Input.Body != nil && err != nil {
		inputMap := make(map[string]interface{})
		inputMap, err = jsons.Unmarshal([]byte(ctx.Input.Body.(string)))
		if err != nil {
			err = fmt.Errorf("engine:influx.输入的body不是有效的json数据，(err:%v)", err)
			return
		}
		msm, ok := inputMap["measurement"]
		if !ok {
			err = errors.New("engine:influx.body的内容中未包含measurement标签")
			return
		}

		if measurement, ok = msm.(string); !ok {
			err = fmt.Errorf("engine:influx.body的内容中measurement标签必须为字符串:(err:%v)", msm)
			return
		}
		tgs, ok := inputMap["tags"]
		if !ok {
			err = errors.New("engine:influx.body的内容中未包含tags标签")
			return
		}
		flds, ok := inputMap["fields"]
		if !ok {
			err = errors.New("engine:influx.body的内容中未包含fields标签")
			return
		}
		tgMap, ok := tgs.(map[string]interface{})
		if !ok {
			err = errors.New("engine:influx.body的内容中的tags标签必须为对象，并包含多个属性")
			return
		}
		fieldMap, ok := flds.(map[string]interface{})
		if !ok {
			err = errors.New("engine:influx.body的内容中的fields标签必须为对象，并包含多个属性")
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
		err = errors.New("engine:influx.form中未包含measurement标签")
		return
	}

	tagStr, err := input.Get("tags")
	if err != nil {
		err = errors.New("engine:influx.form中未包含tags标签")
		return
	}
	fieldStr, err := input.Get("fields")
	if err != nil {
		err = errors.New("engine:influx.form中未包含fields标签")
		return
	}
	tagMap, err := jsons.Unmarshal([]byte(tagStr))
	if err != nil {
		err = errors.New("engine:influx.form中的tags的值不是有效的json数据")
		return
	}

	fieldMap, err := jsons.Unmarshal([]byte(fieldStr))
	if err != nil {
		err = errors.New("engine:influx.form中的fields的值不是有效的json数据")
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

func (s *influxProxy) save(ctx *context.Context) (r string, err error) {
	measurement, t, f, err := s.getSaveParams(ctx)
	if err != nil {
		return "", err
	}
	client, err := s.getInfluxClient(ctx)
	if err != nil {
		return "", err
	}
	err = client.Send(measurement, t, f)
	if err != nil {
		err = fmt.Errorf("engine:influx.save(err:%v)", err)
	}
	r = "SUCCESS"
	return
}
