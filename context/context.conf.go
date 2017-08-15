package context

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ContextConf struct {
	data map[string]interface{}
	json string
}

func NewConf(js string) (conf *ContextConf, err error) {
	if js == "" {
		return nil, fmt.Errorf("输入json串为空")
	}
	conf = &ContextConf{}
	conf.json = js
	conf.data = make(map[string]interface{})
	err = json.Unmarshal([]byte(js), &conf.data)
	if err != nil {
		return
	}
	return
}

func (c *ContextConf) Has(keys ...string) error {
	for _, key := range keys {
		if _, ok := c.data[key]; !ok {
			return fmt.Errorf("不包含:%s  %s", key, c.json)
		}
	}
	return nil
}

func (c *ContextConf) Get(key string) interface{} {
	return c.data[key]
}

func (c *ContextConf) GetString(key string) (string, error) {
	if v, ok := c.data[key]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
		return "", fmt.Errorf("%s的值不是字符串:(%s)", key, c.json)
	}
	return "", fmt.Errorf("不包含key:%s (%s)", key, c.json)
}
func (c *ContextConf) GetInt(key string) (int, error) {
	if v, ok := c.data[key]; ok {
		ix, err := strconv.Atoi(fmt.Sprint(v))
		if err != nil {
			return 0, fmt.Errorf("%s的值不是有效的数字:(%s)", key, c.json)
		}
		return ix, err
	}
	return 0, fmt.Errorf("不包含%s (%s)", key, c.json)
}
