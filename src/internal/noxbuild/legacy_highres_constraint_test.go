package main

import (
	"go/build"
	"testing"
)

func TestLegacyHighResBuildConstraintsMatchGoVariant(t *testing.T) {
	platforms := []buildPlatform{
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "386"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
	}
	for _, platform := range platforms {
		for _, hd := range []bool{false, true} {
			name := platform.String() + "/normal"
			if hd {
				name = platform.String() + "/highres"
			}
			t.Run(name, func(t *testing.T) {
				ctx := build.Default
				ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled = platform.GOOS, platform.GOARCH, true
				ctx.BuildTags = nil
				if hd {
					ctx.BuildTags = []string{"highres"}
				}
				for _, file := range []struct {
					dir  string
					name string
					want bool
				}{
					{"../../legacy", "video_highres.go", hd},
					{"../..", "video_highres.go", hd},
					{"../..", "video_vanilla.go", !hd},
				} {
					matches, err := ctx.MatchFile(file.dir, file.name)
					if err != nil {
						t.Fatal(err)
					}
					if matches != file.want {
						t.Errorf("%s/%s selected=%v, want %v: legacy C protocol/limits must match the Go build variant", file.dir, file.name, matches, file.want)
					}
				}
			})
		}
	}
}
