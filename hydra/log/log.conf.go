package log

import "encoding/json"

type LogConf struct {
	Level         string      `json:"level" valid:"required"`
	Layout        interface{} `json:"layout" valid:"required"`
	WriteInterval string      `json:"interval" valid:"required"`
	Service       string      `json:"service" valid:"required"`
}

func (l LogConf) GetLayout() (string, error) {
	buff, err := json.Marshal(l.Layout)
	if err != nil {
		return "", nil
	}
	return string(buff), nil
}
