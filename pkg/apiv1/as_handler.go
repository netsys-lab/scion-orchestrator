package apiv1

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/netutils"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON, please check your request"})
			return
		}

		// Validate the link
		if link.MTU <= 1000 || link.MTU > 8952 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link MTU, please set it between 1000 and 8952"})
			return
		}

		// Validate the link
		if link.BorderRouter == "" || link.Neighbor == "" || link.LinkType == "" || link.Remote == "" || link.Local == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link parameters, please fill all fields"})
			return
		}

		// TODO: Ensure remote Addr is a valid UDP addr

		remoteUDPAddr, err := net.ResolveUDPAddr("udp", link.Remote)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid remote IP %s", link.Local)})
			return
		}

		if remoteUDPAddr.Port == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid remote port %s", link.Local)})
			return
		}

		udpAddr, err := net.ResolveUDPAddr("udp", link.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid local IP %s", link.Local)})
			return
		}

		if udpAddr.Port == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid local port %s", link.Local)})
			return
		}

		isValidLink, err := netutils.IsLocalIPWithMTU(udpAddr.IP.String(), link.MTU)
		if !isValidLink && err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Local IP %s not found on this host or MTU exceeds interface MTU", link.Local)})
			return
		}
		if err != nil {
			fmt.Println("IP %s is invalid on this host", link.Local, "error: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not detect if IP %s is valid on this host", link.Local)})
			return
		}

		if !scionutils.IsValidISDAS(link.Neighbor) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid remote ISD-AS %s", link.Remote)})
			return
		}

		if !netutils.IsUDPPortFree(link.Local) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("UDP port %s is already in use", link.Local)})
			return
		}

		// Load the topology file
		topology, err := scionutils.LoadSCIONTopology(filepath.Join(configDir, "topology.json"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not load the topology file"})
			return
		}

		if !slices.Contains(topology.Attributes, "core") && strings.Contains(link.LinkType, "CORE") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add core link to non-core AS"})
			return
		}

		if slices.Contains(topology.Attributes, "core") && strings.Contains(link.LinkType, "PARENT") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add parent link to core AS"})
			return
		}

		if slices.Contains(topology.Attributes, "core") && strings.Contains(link.LinkType, "PEER") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add peering link to core AS"})
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

		// Standalone mode
		if len(environment.StandaloneServices) > 0 {
			routers := environment.GetStandaloneBorderRouters()
			for _, router := range routers {
				err := router.Restart()
				if err != nil {
					fmt.Println("Could not restart border router %s", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not restart border router %s", router)})
					return
				}
			}

			controls := environment.GetStandaloneControlServices()
			for _, control := range controls {
				err := control.Restart()
				if err != nil {
					fmt.Println("Could not restart control service %s", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not restart control service %s", control)})
					return
				}
			}
		} else {
			routers := environment.GetBorderRouters()
			for _, router := range routers {
				err := router.Restart()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not restart border router %s", router)})
					return
				}
			}

			controls := environment.GetControlServices()
			for _, control := range controls {
				err := control.Restart()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not restart control service %s", control)})
					return
				}
			}
		}

		c.JSON(http.StatusOK, link)
	})
}
