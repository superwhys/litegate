package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/litegate/config"
)

type debugRouter struct {
	configLoader config.ProxyConfigLoader
}

func DebugRouter(configLoader config.ProxyConfigLoader) *debugRouter {
	return &debugRouter{configLoader: configLoader}
}

func (r *debugRouter) Init(router gin.IRouter) {
	router.GET("/config", func(c *gin.Context) {
		routes, err := r.configLoader.GetAll()
		if err != nil {
			ginutils.ReturnError(c, http.StatusOK, err.Error())
			return
		}
		ginutils.ReturnSuccess(c, routes)
	})

	router.GET("/config/:serviceName", func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		route, err := r.configLoader.Get(serviceName)
		if err != nil {
			ginutils.ReturnError(c, http.StatusOK, err.Error())
			return
		}
		ginutils.ReturnSuccess(c, route)
	})
}
