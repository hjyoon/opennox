package dialog

import (
	"go/build"
	"path/filepath"
	"testing"
)

// A dedicated server uses the disabled audio backend, not the client's
// OpenAL stream. Check both test contracts against the backend build tags.
func TestDialogAudioBuildTags(t *testing.T) {
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
		client    bool
	}{
		{".", "dialog_test.go", true},
		{".", "dialog_server_test.go", false},
		{filepath.Join("..", "client", "audio", "ail"), "audio_openal.go", true},
		{filepath.Join("..", "client", "audio", "ail"), "audio_null.go", false},
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
					want := file.client == mode.client
					if matches != want {
						t.Errorf("%s: included=%t, want %t with tags %v", filepath.Join(file.dir, file.name), matches, want, mode.tags)
					}
				}
			})
		}
	}
}
