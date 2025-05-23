package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
)

// RegisterRoutes sets up the routes using Gin
func RegisterRoutes(env *environment.HostEnvironment, config *conf.Config, r *gin.Engine, installWizard bool, scionConfig *conf.SCIONConfig) error {
	// Load HTML templates from the "ui/templates" directory

	uiGlob := "ui/templates"
	if _, err := os.Stat(uiGlob); err != nil && os.IsNotExist(err) {
		uiGlob = filepath.Join(env.ConfigPath, "ui/templates")
	}

	r.LoadHTMLGlob(fmt.Sprintf("%s/**/*", uiGlob))

	uiStatic := "ui/static"
	if _, err := os.Stat(uiStatic); err != nil && os.IsNotExist(err) {
		uiStatic = filepath.Join(env.ConfigPath, "ui/static")
	}

	r.Static("/static", uiStatic)

	accs := make(gin.Accounts)

	for _, user := range config.Api.Users {
		parts := strings.Split(user, ":")
		accs[parts[0]] = parts[1]
	}

	var routerGroup *gin.RouterGroup

	if installWizard {
		routerGroup = r.Group("/")
		routerGroup.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "install.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})
	} else {
		// Apply the BasicAuth middleware to a specific route group
		routerGroup = r.Group("/", gin.BasicAuth(accs))
		routerGroup.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

		routerGroup.GET("/cryptography", func(c *gin.Context) {
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

		routerGroup.GET("/modules", func(c *gin.Context) {
			c.HTML(http.StatusOK, "modules.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

		routerGroup.GET("/troubleshooting", func(c *gin.Context) {
			c.HTML(http.StatusOK, "troubleshooting.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

		routerGroup.GET("/connectivity", func(c *gin.Context) {
			c.HTML(http.StatusOK, "connectivity.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

		routerGroup.GET("/certificate-authority", func(c *gin.Context) {
			c.HTML(http.StatusOK, "certificate-authority.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

		routerGroup.GET("/settings", func(c *gin.Context) {
			c.HTML(http.StatusOK, "settings.html", gin.H{
				"title": "SCION Orchestrator",
			})
		})

	}

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
