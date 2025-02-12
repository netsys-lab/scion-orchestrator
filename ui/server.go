package ui

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
)

// RegisterRoutes sets up the routes using Gin
func RegisterRoutes(env *environment.HostEnvironment, config *conf.Config, r *gin.Engine) error {
	// Load HTML templates from the "ui/templates" directory
	r.LoadHTMLGlob("ui/templates/**/*")

	accs := make(gin.Accounts)

	for _, user := range config.Api.Users {
		parts := strings.Split(user, ":")
		accs[parts[0]] = parts[1]
	}

	// Apply the BasicAuth middleware to a specific route group
	authorized := r.Group("/", gin.BasicAuth(accs))

	authorized.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})
	authorized.GET("/settings", func(c *gin.Context) {
		c.HTML(http.StatusOK, "settings.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})

	// log.Println("[UI] Server running")
	return nil
}

// Handler for the index page
func renderIndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Index Page",
	})
}

// Handler for the settings page
func renderSettingsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title": "Settings Page",
	})
}
