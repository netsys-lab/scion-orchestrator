package scionutils

import "strings"

func GetISDFromISDAS(isdAS string) string {
	parts := strings.Split(isdAS, "-")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
