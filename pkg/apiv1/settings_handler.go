package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
)

func AddSettingsHandler(eng *gin.RouterGroup, config *conf.Config) {
	eng.GET("settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, config)
	})
}
