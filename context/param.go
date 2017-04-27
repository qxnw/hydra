package context

import (
	"errors"
	"html/template"
	"strconv"
)

type (
	Param struct {
		Name  string
		Value string
	}
	Params []Param
)

var Set = &Params{}

func (p *Params) Each(f func(string, string)) {
	for _, v := range *p {
		f(v.Name, v.Value)
	}
}
func (p *Params) Get(key string) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for _, v := range *p {
		if v.Name == key {
			return v.Value, nil
		}
	}
	return "", nil
}

func (p *Params) String(key string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("not exist")
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for _, v := range *p {
		if v.Name == key {
			return v.Value, nil
		}
	}
	return "", errors.New("not exist")
}

func (p *Params) Strings(key string) ([]string, error) {
	if len(key) == 0 {
		return nil, errors.New("not exist")
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	var s = make([]string, 0)
	for _, v := range *p {
		if v.Name == key {
			s = append(s, v.Value)
		}
	}
	if len(s) > 0 {
		return s, nil
	}
	return nil, errors.New("not exist")
}

func (p *Params) Escape(key string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("not exist")
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for _, v := range *p {
		if v.Name == key {
			return template.HTMLEscapeString(v.Value), nil
		}
	}
	return "", errors.New("not exist")
}

func (p *Params) Int(key string) (int, error) {
	v, _ := p.Get(key)
	return strconv.Atoi(v)
}

func (p *Params) Int32(key string) (int32, error) {
	value, _ := p.Get(key)
	v, err := strconv.ParseInt(value, 10, 32)
	return int32(v), err
}

func (p *Params) Int64(key string) (int64, error) {
	value, _ := p.Get(key)
	return strconv.ParseInt(value, 10, 64)
}

func (p *Params) Uint(key string) (uint, error) {
	value, _ := p.Get(key)
	v, err := strconv.ParseUint(value, 10, 64)
	return uint(v), err
}

func (p *Params) Uint32(key string) (uint32, error) {
	value, _ := p.Get(key)
	v, err := strconv.ParseUint(value, 10, 32)
	return uint32(v), err
}

func (p *Params) Uint64(key string) (uint64, error) {
	value, _ := p.Get(key)
	return strconv.ParseUint(value, 10, 64)
}

func (p *Params) Bool(key string) (bool, error) {
	value, _ := p.Get(key)
	return strconv.ParseBool(value)
}

func (p *Params) Float32(key string) (float32, error) {
	value, _ := p.Get(key)
	v, err := strconv.ParseFloat(value, 32)
	return float32(v), err
}

func (p *Params) Float64(key string) (float64, error) {
	value, _ := p.Get(key)
	return strconv.ParseFloat(value, 64)
}

func (p *Params) MustString(key string, defaults ...string) string {
	if len(key) == 0 {
		return ""
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for _, v := range *p {
		if v.Name == key {
			return v.Value
		}
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

func (p *Params) MustStrings(key string, defaults ...[]string) []string {
	if len(key) == 0 {
		return []string{}
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	var s = make([]string, 0)
	for _, v := range *p {
		if v.Name == key {
			s = append(s, v.Value)
		}
	}
	if len(s) > 0 {
		return s
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return []string{}
}

func (p *Params) MustEscape(key string, defaults ...string) string {
	if len(key) == 0 {
		return ""
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for _, v := range *p {
		if v.Name == key {
			return template.HTMLEscapeString(v.Value)
		}
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

func (p *Params) MustInt(key string, defaults ...int) int {
	value, _ := p.Get(key)
	v, err := strconv.Atoi(value)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return v
}

func (p *Params) MustInt32(key string, defaults ...int32) int32 {
	value, _ := p.Get(key)
	r, err := strconv.ParseInt(value, 10, 32)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}

	return int32(r)
}

func (p *Params) MustInt64(key string, defaults ...int64) int64 {
	value, _ := p.Get(key)
	r, err := strconv.ParseInt(value, 10, 64)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return r
}

func (p *Params) MustUint(key string, defaults ...uint) uint {
	value, _ := p.Get(key)
	v, err := strconv.ParseUint(value, 10, 64)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return uint(v)
}

func (p *Params) MustUint32(key string, defaults ...uint32) uint32 {
	value, _ := p.Get(key)
	r, err := strconv.ParseUint(value, 10, 32)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}

	return uint32(r)
}

func (p *Params) MustUint64(key string, defaults ...uint64) uint64 {
	value, _ := p.Get(key)
	r, err := strconv.ParseUint(value, 10, 64)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return r
}

func (p *Params) MustFloat32(key string, defaults ...float32) float32 {
	value, _ := p.Get(key)
	r, err := strconv.ParseFloat(value, 32)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return float32(r)
}

func (p *Params) MustFloat64(key string, defaults ...float64) float64 {
	value, _ := p.Get(key)
	r, err := strconv.ParseFloat(value, 64)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return r
}

func (p *Params) MustBool(key string, defaults ...bool) bool {
	value, _ := p.Get(key)
	r, err := strconv.ParseBool(value)
	if len(defaults) > 0 && err != nil {
		return defaults[0]
	}
	return r
}

func (p *Params) Set(key, value string) {
	if len(key) == 0 {
		return
	}
	if key[0] != ':' && key[0] != '*' {
		key = ":" + key
	}

	for i, v := range *p {
		if v.Name == key {
			(*p)[i].Value = value
			return
		}
	}

	*p = append(*p, Param{key, value})
}

func (p *Params) SetParams(params []Param) {
	*p = params
}
