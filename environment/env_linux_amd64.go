package environment

func init() {
	HostEnv = &HostEnvironment{
		BasePath:          "/etc/scion/",
		ConfigPath:        "/etc/scion/",
		DaemonConfigPath:  "/etc/scion/",
		ControlConfigPath: "/etc/scion/",
		RouterConfigPath:  "/etc/scion/",
		DatabasePath:      "/var/lib/scion/",
		LogPath:           "/var/log/scion/",
	}
}
