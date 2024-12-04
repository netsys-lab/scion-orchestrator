//go:build !windows

package main

func init() {
	RunFunc = runUnix
}

func runUnix() error {
	return nil
}
