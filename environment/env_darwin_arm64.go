package environment

func init() {
	HostEnv = &HostEnvironment{
		BasePath:          "/Applications/scion/",
		ConfigPath:        "/Applications/scion/",
		DaemonConfigPath:  "/Applications/scion/",
		ControlConfigPath: "/Applications/scion/",
		RouterConfigPath:  "/Applications/scion/",
		DatabasePath:      "/Applications/scion/database/",
		LogPath:           "/Applications/scion/logs/",
	}
}
