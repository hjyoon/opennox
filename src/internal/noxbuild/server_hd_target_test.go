package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

func TestDedicatedHighResServerBuildTargets(t *testing.T) {
	for _, target := range []string{"server-hd", "server-highres", "opennox-server-hd"} {
		for _, platform := range []string{"darwin", "windows"} {
			t.Run(target+"/"+platform, func(t *testing.T) {
				testDedicatedHighResServerCommand(t, target, platform, false)
			})
		}
	}
	t.Run("safe", func(t *testing.T) {
		testDedicatedHighResServerCommand(t, "server-hd", "darwin", true)
	})
}

func testDedicatedHighResServerCommand(t *testing.T, target, goos string, safe bool) {
	t.Helper()
	oldOS, oldArch, oldOut, oldSrc, oldGo := *fOS, *fArch, *fOut, *fSrc, *fGo
	oldCC, oldCXX, oldDry, oldSafe, oldVerbose := *fCC, *fCXX, *fDryRun, *fSafe, *fVerbose
	defer func() {
		*fOS, *fArch, *fOut, *fSrc, *fGo = oldOS, oldArch, oldOut, oldSrc, oldGo
		*fCC, *fCXX, *fDryRun, *fSafe, *fVerbose = oldCC, oldCXX, oldDry, oldSafe, oldVerbose
	}()
	out := t.TempDir()
	*fOS, *fArch, *fOut, *fSrc, *fGo = goos, "arm64", out, "", "fixture-go"
	*fCC, *fCXX, *fDryRun, *fSafe, *fVerbose = "fixture-cc", "fixture-cxx", true, safe, false
	before := append([]string(nil), defTargets...)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	oldStdout := os.Stdout
	os.Stdout = w
	buildErr := buildTarget(target)
	os.Stdout = oldStdout
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if buildErr != nil {
		t.Fatalf("HD dedicated target rejected: %v", buildErr)
	}
	tags := "server,highres"
	if safe {
		tags += ",safe"
	}
	name := "opennox-server-hd"
	if goos == "windows" {
		name += ".exe"
	}
	command := string(data)
	for _, part := range []string{"GOOS=" + goos, "GOARCH=arm64", "CGO_ENABLED=1", "fixture-go build", "-tags " + tags, "-o " + filepath.Join(out, name), "./cmd/opennox-server"} {
		if !strings.Contains(command, part) {
			t.Errorf("dedicated HD command missing %q: %s", part, command)
		}
	}
	if strings.Contains(command, "windowsgui") || strings.Contains(command, "guiapp") || strings.Count(command, "fixture-go build") != 1 {
		t.Fatalf("HD dedicated server must remain one console/server build: %s", command)
	}
	if !reflect.DeepEqual(defTargets, before) || !reflect.DeepEqual(defTargets, []string{BinOpenNox, BinOpenNoxHD, BinServer}) {
		t.Fatalf("opt-in target changed the existing default product set: %v", defTargets)
	}
}

func TestDedicatedHighResServerMetadataEntrypoint(t *testing.T) {
	for _, file := range []string{"build/opennox-server-hd", "build/opennox-server-hd.exe", "build with spaces/opennox-server-hd"} {
		t.Run(file, func(t *testing.T) {
			if got := productBuildPathForFile(file); got != serverBuildPath {
				t.Fatalf("HD dedicated build path=%q, want %q", got, serverBuildPath)
			}
			info := &debug.BuildInfo{
				GoVersion: requiredGoVersion,
				Path:      serverBuildPath,
				Settings: []debug.BuildSetting{
					{Key: "GOOS", Value: "darwin"},
					{Key: "GOARCH", Value: "arm64"},
					{Key: "vcs.revision", Value: "fixture-revision"},
					{Key: "vcs.modified", Value: "false"},
				},
			}
			meta, err := parseProductMetadata(info, productBuildPathForFile(file))
			if err != nil {
				t.Fatal(err)
			}
			if err := validateProductMetadata(meta, "fixture-revision"); err != nil {
				t.Fatal(err)
			}
			info.Path = productBuildPath
			if _, err := parseProductMetadata(info, productBuildPathForFile(file)); err == nil {
				t.Fatal("client executable accepted as an HD dedicated server")
			}
		})
	}
}
