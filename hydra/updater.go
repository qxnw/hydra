package hydra

/*
func update(domain string, address string, log *logger.Logger) (err error) {
	rgst, err := registry.NewRegistryWithAddress(address, log)
	if err != nil {
		err = fmt.Errorf("初始化注册中心失败：%s:%v", address, err)
		return err
	}
	path := fmt.Sprintf("%s/var/global/logger", domain)
	buff, err := getConfig(rgst, path)
	if err != nil {
		return err
	}
	loggerConf, err := conf.NewJSONConfWithJson(string(buff), 0, nil)
	if err != nil {
		err = fmt.Errorf("rpc日志配置错误:%s,%v", string(buff), err)
		return
	}
	return
}

func getConfig(rgst registry.Registry, path string) ([]byte, error) {
LOOP:
	for {
		select {
		case <-r.closeChan:
			break LOOP
		case <-time.After(time.Second):
			if b, err := rgst.Exists(path); err == nil && b {
				buff, _, err := rgst.GetValue(path)
				if err != nil {
					err = fmt.Errorf("无法获取RPC日志配置:%v", err)
					return nil, err
				}
				return buff, nil
			}
		}
	}
	return nil, fmt.Errorf("关闭监听:%s", path)

}
*/
