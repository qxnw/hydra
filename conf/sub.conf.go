package conf

type HttpServerConf struct {
	Address           string `json:"address" valid:"dialstring"`
	Status            string `json:"status" valid:"in(start|stop)"`
	Engines           string `json:"engines"`
	ReadTimeout       int    `json:"readTimeout"`
	WriteTimeout      int    `json:"writeTimeout"`
	ReadHeaderTimeout int    `json:"readHeaderTimeout"`
	Name              string
	metadata          metadata
}

func (s *HttpServerConf) GetMetadata(key string) interface{} {
	return s.metadata.Get(key)
}
func (s *HttpServerConf) SetMetadata(key string, v interface{}) {
	s.metadata.Set(key, v)
}

type Metric struct {
	Host     string `json:"host" valid:"requrl,required"`
	DataBase string `json:"dataBase" valid:"ascii,required"`
	Cron     string `json:"cron" valid:"ascii,required"`
	UserName string `json:"userName" valid:"ascii"`
	Password string `json:"password" valid:"ascii"`
	Disable  bool   `json:"disable"`
}
type Authes map[string]*Auth

type Auth struct {
	Name     string   `json:"name" valid:"ascii,required"`
	ExpireAt int64    `json:"expireAt" valid:"int,required"`
	Mode     string   `json:"mode" valid:"in(HS256|HS384|HS512|RS256|ES256|ES384|ES512|RS384|RS512|PS256|PS384|PS512),required"`
	Secret   string   `json:"secret" valid:"ascii,required"`
	Exclude  []string `json:"exclude"`
	Disable  bool     `json:"disable"`
}
type Routers struct {
	Setting map[string]string `json:"args"`
	Routers []*Router         `json:"routers"`
}
type Router struct {
	Name    string            `json:"name" valid:"ascii,required"`
	Action  []string          `json:"action" valid:"uppercase,in(GET|POST|PUT|DELETE|HEAD|TRACE|OPTIONS)"`
	Engine  string            `json:"engine" valid:"ascii"`
	Service string            `json:"service" valid:"ascii,required"`
	Setting map[string]string `json:"args"`
	Disable bool              `json:"disable"`
	Handler interface{}
}

type View struct {
	Path  string `json:"path" valid:"ascii,required"`
	Left  string `json:"left" valid:"ascii"`
	Right string `json:"right" valid:"ascii"`
	Files []string
}

type CircuitBreaker struct {
	ForceBreak      bool                `json:"force-break"`
	Disable         bool                `json:"disable"`
	SwitchWindow    int                 `json:"swith-window" valid:"int"`
	CircuitBreakers map[string]*Breaker `json:"circuit-breakers"`
}
type Breaker struct {
	URL              string `json:"url" valid:"ascii,required"`
	RequestPerSecond int    `json:"request-per-second"`
	FailedPercent    int    `json:"failed-request"`
	RejectPerSecond  int    `json:"reject-per-second"`
	Disable          bool   `json:"disable"`
}
type Static struct {
	Path    string `json:"path" valid:"ascii,required"`
	Left    string `json:"left" valid:"ascii"`
	Right   string `json:"right" valid:"ascii"`
	Files   []string
	Disable bool
}
type Tasks struct {
	Setting map[string]string `json:"args"`
	Tasks   []*Task           `json:"tasks"`
}
type Task struct {
	Name    string            `json:"name" valid:"ascii,required"`
	Cron    string            `json:"cron" valid:"ascii,required"`
	Input   string            `json:"input,omitempty" valid:"ascii"`
	Body    string            `json:"body,omitempty" valid:"ascii"`
	Engine  string            `json:"engine,omitempty"  valid:"ascii"`
	Service string            `json:"service"  valid:"ascii"`
	Setting map[string]string `json:"args"`
	Next    string            `json:"next"`
	Last    string            `json:"last"`
	Handler interface{}       `json:"handler,omitempty"`
}
type Headers map[string]string
type Hosts []string

type Queues struct {
	Setting map[string]string `json:"args"`
	Queues  []*Queue          `json:"queue"`
}

type Queue struct {
	Name        string            `json:"name" valid:"ascii,required"`
	Queue       string            `json:"queueue" valid:"ascii,required"`
	Engine      string            `json:"engine,omitempty"  valid:"ascii"`
	Service     string            `json:"service" valid:"ascii,required"`
	Setting     map[string]string `json:"args"`
	Concurrency int               `json:"concurrency"`
	Handler     interface{}
}
