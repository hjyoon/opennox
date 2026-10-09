//go:build !server

package main

import (
	"go/build"
	"path/filepath"
	"testing"
)

// The standalone movie player uses the same client-only implementation as
// OpenNox. Keep it available to both clients and out of dedicated-server builds.
func TestMovieCommandClientBuildTags(t *testing.T) {
	platforms := []struct {
		os, arch string
	}{
		{"darwin", "arm64"},
		{"darwin", "amd64"},
		{"linux", "amd64"},
		{"linux", "386"},
		{"linux", "arm"},
		{"linux", "arm64"},
		{"windows", "386"},
		{"windows", "amd64"},
		{"windows", "arm64"},
	}
	modes := []struct {
		name   string
		tags   []string
		client bool
	}{
		{"client", nil, true},
		{"client-hd", []string{"highres"}, true},
		{"server", []string{"server"}, false},
		{"server-highres", []string{"server", "highres"}, false},
	}
	files := []struct {
		dir, name string
	}{
		{".", "main.go"},
		{".", "main_build_test.go"},
		{filepath.Join("..", "..", "client", "noxmovie"), "noxmovie.go"},
	}
	for _, platform := range platforms {
		for _, mode := range modes {
			t.Run(platform.os+"-"+platform.arch+"/"+mode.name, func(t *testing.T) {
				ctx := build.Default
				ctx.GOOS, ctx.GOARCH = platform.os, platform.arch
				ctx.CgoEnabled = true
				ctx.BuildTags = mode.tags
				for _, file := range files {
					matches, err := ctx.MatchFile(file.dir, file.name)
					if err != nil {
						t.Fatal(err)
					}
					if matches != mode.client {
						t.Errorf("%s: included=%t, want %t with tags %v", filepath.Join(file.dir, file.name), matches, mode.client, mode.tags)
					}
				}
			})
		}
	}
}
