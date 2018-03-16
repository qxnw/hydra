package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/qxnw/hydra/conf"
)

func JwtWriter(cnf *conf.MetadataConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if response := getResponse(ctx); response != nil {
			setJwtResponse(ctx, cnf, response.GetParams()["__jwt_"])
		}
	}
}
