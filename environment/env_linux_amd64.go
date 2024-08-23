package environment

func init() {
	EndhostEnv = &EndhostEnvironment{
		BasePath:          "/etc/scion/",
		ConfigPath:        "/etc/scion/",
		DaemonConfigPath:  "/etc/scion/",
		ControlConfigPath: "/etc/scion/",
		RouterConfigPath:  "/etc/scion/",
	}
}
