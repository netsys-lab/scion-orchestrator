package environment

import "embed"

func init() {
	EndhostEnv = &EndhostEnvironment{
		BasePath:          "/Applications/scion/",
		ConfigPath:        "/Applications/scion/",
		DaemonConfigPath:  "/Applications/scion/",
		ControlConfigPath: "/Applications/scion/",
		RouterConfigPath:  "/Applications/scion/",
		DatabasePath:      "/Applications/scion/database/",
		LogPath:           "/Applications/scion/logs/",
	}
}
