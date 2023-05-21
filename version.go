package main

import (
	"runtime/debug"
	"time"
)

var (
	vcsCommit   string
	vcsTime     time.Time
	vcsModified bool
	compiler    string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			vcsCommit = setting.Value
			continue
		}

		if setting.Key == "vcs.time" {
			vcsTime, _ = time.Parse(time.RFC3339, setting.Value)
			vcsTime = vcsTime.Local()
			continue
		}

		if setting.Key == "vcs.modified" {
			if setting.Value == "true" {
				vcsModified = true
			} else {
				vcsModified = false
			}
			continue
		}

		if setting.Key == "-compiler" {
			compiler = setting.Value
			continue
		}
	}
}
