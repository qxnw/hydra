package registry

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConfig1(t *testing.T) {
	c := NewJSONConfWithHandle(getMap(), 100, nil)
	root := c.String("root")
	expect(t, len(c.data), 8+1)
	expect(t, root, "api/merchant.api")
}
func TestConfig2(t *testing.T) {
	c := NewJSONConfWithHandle(getMap(), 100, nil)
	qps, err := c.Int("QPS")
	if err != nil {
		t.Error(err)
	}
	expect(t, qps, 1000)
}
func TestConfig3(t *testing.T) {
	c := NewJSONConfWithHandle(getMap(), 100, nil)
	limiter, err := c.GetSections("limit")
	if err != nil {
		t.Error(err)
	}
	expect(t, len(limiter), 2)
}
func TestConfig4(t *testing.T) {
	c := NewJSONConfWithHandle(getMap(), 100, nil)
	routes, err := c.GetSections("routes")
	if err != nil {
		t.Error(err)
	}
	expect(t, len(routes), 1)
	expect(t, routes[0].String("to"), "../api/merchant.api/script/:request")

}
func TestConfig5(t *testing.T) {
	c := NewJSONConfWithHandle(getMap(), 100, nil)
	r := NewJSONConfWithHandle(getMap2(), 100, nil)
	expect(t, c.Len(), r.Len())
	router1, err := c.GetSections("routes")
	if err != nil {
		t.Error(err)
	}
	router2, err := r.GetSections("routes")

	if err != nil {
		t.Error(err)
	}
	expect(t, len(router1), len(router2))
	v, err := r.Int("QPS")
	expect(t, err, nil)
	expect(t, v, 1000)
	expect(t, router1[0].String("to"), router2[0].String("to"))

}
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func getMap2() map[string]interface{} {
	str := `{
    "type": "api",
    "name": "merchant.api",
    "status": "starting",
    "package": "1.0.0.1",
	"root":    "@type/@name",
    "QPS": 1000,
    "limit": [
        {
            "local": "@client like 192.168*",
            "Operation": "deny",
            "to": "@service like 192.168*"
        },
        {
            "local": "@client like 192.168* && @service == /order/request",
            "Operation": "assign",
            "to": "@ip like 192.168.1*"
        }
    ],
    "routes": [
        {
            "from": "/:module/:action/:id",
            "method": "request",
            "to": "../@type/@name/script/@module_@action:@method",
            "params": "db=@var_weixin"
        }
    ]
}`
	c := make(map[string]interface{})
	json.Unmarshal([]byte(str), &c)
	return c
}
func getMap() map[string]interface{} {
	data := map[string]interface{}{
		"type":    "api",
		"name":    "merchant.api",
		"status":  "starting",
		"package": "1.0.0.1",
		"QPS":     1000,
		"root":    "@type/@name",
		"limit": []interface{}{
			map[string]interface{}{
				"local":     "@client like 192.168*",
				"Operation": "deny",
				"to":        "@service like 192.168*",
			},
			map[string]interface{}{
				"local":     "@client like 192.168* && @service == /order/request",
				"Operation": "assign",
				"to":        "@ip like 192.168.1*",
			},
		},
		"routes": []interface{}{
			map[string]interface{}{
				"from":   "/:module/:action/:id",
				"method": "request",
				"to":     "../@type/@name/script/@module_@action:@method",
				"params": "db=@var_weixin",
			},
		},
	}

	return data
}
