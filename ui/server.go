package ui

import (
	"encoding/json"
	"log"
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

	/*
			const paths = [
		            "/",
		            "/cryptography",
		            "/modules",
		            "/troubleshooting",
		            "/bootstrapping",
		            "/certificate-authority",
		            "/settings"
		        ];
	*/

	authorized.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})

	authorized.GET("/cryptography", func(c *gin.Context) {
		csr, err := json.MarshalIndent(gin.H{"subject": gin.H{"common_name": "1-150 AS Certificate", "isd_as": "1-150"}}, "", "  ")
		if err != nil {
			log.Println(err)
		}
		c.HTML(http.StatusOK, "cryptography.html", gin.H{
			"title": "SCION Orchestrator",
			"csr":   string(csr),
			// "Content": "cryptograhpy",
		})
	})

	authorized.GET("/modules", func(c *gin.Context) {
		c.HTML(http.StatusOK, "modules.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})

	authorized.GET("/troubleshooting", func(c *gin.Context) {
		c.HTML(http.StatusOK, "troubleshooting.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})

	authorized.GET("/connectivity", func(c *gin.Context) {
		c.HTML(http.StatusOK, "connectivity.html", gin.H{
			"title": "SCION Orchestrator",
		})
	})

	authorized.GET("/certificate-authority", func(c *gin.Context) {
		c.HTML(http.StatusOK, "certificate-authority.html", gin.H{
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
