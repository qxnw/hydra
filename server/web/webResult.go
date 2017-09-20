package web

import "github.com/qxnw/hydra/context"

type webResult struct {
	Response context.Response
	Error    error
}
