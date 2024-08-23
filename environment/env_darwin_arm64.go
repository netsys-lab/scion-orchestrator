package environment

import "embed"

func init() {
	EndhostEnv = &EndhostEnvironment{
		BasePath:         "/Applications/scion/",
		ConfigPath:       "/Applications/scion/",
		DaemonConfigPath: "/Applications/scion/",
	}
}
