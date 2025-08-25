package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/litegate/agent"
	"github.com/superwhys/litegate/api/middleware"
	"github.com/superwhys/litegate/config"
)

func ProxyRouter(gatewayConf *config.GatewayConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxyConfig := middleware.GetProxyConfig(c)
		if proxyConfig == nil {
			ginutils.ReturnError(c, http.StatusOK, "proxy config not found")
			return
		}

		// 1. parse route config
		upstreamConf := proxyConfig.MatchRequest(c, c.Request)
		if upstreamConf == nil {
			ginutils.ReturnError(c, http.StatusOK, "route config not found")
			return
		}

		// 2. create agent
		proxyAgent, err := agent.NewAgent(upstreamConf, gatewayConf)
		if err != nil {
			ginutils.ReturnError(c, http.StatusOK, err.Error())
			return
		}

		// 3. auth route
		proxyAgent.Auth(c.Writer, c.Request)

		// 4. proxy request
		proxyAgent.ServeHTTP(c.Writer, c.Request)
	}
}
