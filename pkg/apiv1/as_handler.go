package apiv1

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionutils"
)

func GetTopologyHandler(eng *gin.RouterGroup, configDir string) {

	eng.GET("as/topology", func(c *gin.Context) {

		// Load the topology file
		topology, err := scionutils.LoadSCIONTopology(filepath.Join(configDir, "/topology.json"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the topology file"})
			return
		}

		c.JSON(http.StatusOK, topology)
	})
}

func GetSCIONLinksHandler(eng *gin.RouterGroup, configDir string) {

	eng.GET("as/links", func(c *gin.Context) {

		// Load the topology file
		topology, err := scionutils.LoadSCIONTopology(filepath.Join(configDir, "/topology.json"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the topology file"})
			return
		}

		links, err := scionutils.GetSCIONLinksWithStatus("http://localhost:30401/metrics", topology)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get links status"})
			return
		}

		c.JSON(http.StatusOK, links)
	})
}

type SCIONLink struct {
	BorderRouter string `json:"border_router"`
	Neighbor     string `json:"neighbor"`
	LinkType     string `json:"link_type"`
	MTU          int    `json:"mtu"`
	Local        string `json:"local"`
	Remote       string `json:"remote"`
}

func AddSCIONLinksHandler(eng *gin.RouterGroup, configDir string) {

	eng.POST("as/links", func(c *gin.Context) {

		link := &SCIONLink{}

		err := c.BindJSON(link)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Load the topology file
		topology, err := scionutils.LoadSCIONTopology(filepath.Join(configDir, "topology.json"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the topology file"})
			return
		}

		nextInterface := topology.NextInterfaceID()

		// Add the link to the topology
		topology.BorderRouters[link.BorderRouter].Interfaces[fmt.Sprintf("%d", nextInterface)] = scionutils.RouterInterface{
			ISD_AS:   link.Neighbor,
			LinkTo:   link.LinkType,
			MTU:      link.MTU,
			Underlay: scionutils.Underlay{Public: link.Local, Remote: link.Remote},
		}

		err = scionutils.SaveSCIONTopology(filepath.Join(configDir, "topology.json"), topology)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save the topology file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Links added successfully"})
	})
}
