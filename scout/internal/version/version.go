// Package version 由 -ldflags 在 build 時注入版本資訊。
package version

var (
	Version = "0.1.0-dev"
	Commit  = "none"
)
