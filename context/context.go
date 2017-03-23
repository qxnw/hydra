package context

type IContext interface {
	Params() *Params
	Service() string
	Method() string
	IP() string
}
