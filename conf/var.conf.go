package conf

type DBConf struct {
	Provider   string `json:"provider"`
	ConnString string `json:"connString"`
	Max        int    `json:"max" valid:"range(1|1000)"`
}
