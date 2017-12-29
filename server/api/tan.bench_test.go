package api

/*
type benchtHandler struct {
	version int32
}

func (h benchtHandler) Handle(name string, method []string, s string, p string, c context.Context) (r context.Response, err error) {
	return &context.Response{Content: "success"}, nil
}
func (h benchtHandler) GetPath(p string) (conf.Conf, error) {
	h.version++
	if strings.HasSuffix(p, "influxdb1") {
		return conf.NewJSONConfWithJson(metricStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router1") {
		return conf.NewJSONConfWithJson(routerStr1, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "influxdb2") {
		return conf.NewJSONConfWithJson(metricStr2, h.version, h.GetPath)
	} else if strings.HasSuffix(p, "router2") {
		return conf.NewJSONConfWithJson(routerStr2, h.version, h.GetPath)
	}
	return nil, errors.New("not find")
}


func TestNotify(t *testing.T) {
	handler := &benchtHandler{version: 101}
	conf, err := registry.NewJSONConfWithJson(confstr_b1, 100, handler.GetPath)
	ut.Expect(t, err, nil)
	server, err := newHydraWebServer(handler, conf)
	ut.ExpectSkip(t, err, nil)
	err = server.Start()
	ut.ExpectSkip(t, err, nil)
	for i := 0; i < 300; i++ {
		conf, err := registry.NewJSONConfWithJson(confstr_b1, int32(2000+i), handler.GetPath)
		ut.ExpectSkip(t, err, nil)
		err = server.Notify(conf)
		ut.Expect(t, err, nil)
		time.Sleep(time.Second)
	}
	fmt.Println("----------------------------")
	time.Sleep(time.Hour)

}
*/
var confstr_b1 = `{
    "type": "api.server01",
    "name": "merchant.api",
	"address":":1091",
    "status": "starting",
    "package": "1.0.0.1",
    "QPS": 1000,
    "metric": "#/@domain/var/db/influxdb",
    "limiter": "#@path/limiter1",
    "router": "#@path/router1"
}`
