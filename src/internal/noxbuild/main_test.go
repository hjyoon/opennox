package main

import (
	"os"
	"runtime/debug"
	"strings"
	"testing"
)

func TestARMCharSignednessFlagIsAllowed(t *testing.T) {
	if !strings.Contains(cgoCFlagsAllow, "-fsigned-char") {
		t.Fatal("ARM portability requires -fsigned-char in CGO_CFLAGS_ALLOW")
	}
}

func TestCheckGoVersion(t *testing.T) {
	if err := checkGoVersion("test", requiredGoVersion); err != nil {
		t.Fatalf("exact version rejected: %v", err)
	}

	err := checkGoVersion("test", "go1.26.6")
	if err == nil {
		t.Fatal("mismatched patch version accepted")
	}
	if !strings.Contains(err.Error(), requiredGoVersion) {
		t.Fatalf("error does not name required version: %v", err)
	}
}

func TestRequiredGoVersionFile(t *testing.T) {
	b, err := os.ReadFile("../../../toolchain/go-version.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != requiredGoVersion {
		t.Fatalf("toolchain/go-version.txt is %q, noxbuild requires %q", got, requiredGoVersion)
	}
}

func TestBuildPlatformValidate(t *testing.T) {
	valid := []buildPlatform{
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
	for _, target := range valid {
		if err := target.validate(); err != nil {
			t.Errorf("valid target %s rejected: %v", target, err)
		}
	}

	invalid := []buildPlatform{
		{GOOS: "windows", GOARCH: "arm"},
		{GOOS: "darwin", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "mips"},
		{GOOS: "plan9", GOARCH: "amd64"},
	}
	for _, target := range invalid {
		if err := target.validate(); err == nil {
			t.Errorf("invalid target %s accepted", target)
		}
	}
}

func TestBuildPlatformTuningEnv(t *testing.T) {
	tests := []struct {
		target buildPlatform
		want   string
	}{
		{buildPlatform{GOOS: "linux", GOARCH: "386"}, "GO386=sse2"},
		{buildPlatform{GOOS: "linux", GOARCH: "amd64"}, "GOAMD64=v1"},
		{buildPlatform{GOOS: "linux", GOARCH: "arm"}, "GOARM=7,hardfloat"},
		{buildPlatform{GOOS: "linux", GOARCH: "arm64"}, "GOARM64=v8.0"},
	}
	for _, tc := range tests {
		if got := tc.target.tuningEnv("sse2", "v1", "7,hardfloat", "v8.0"); got != tc.want {
			t.Errorf("%s tuning = %q, want %q", tc.target, got, tc.want)
		}
	}
}

func TestDefaultCrossCompilers(t *testing.T) {
	host := buildPlatform{GOOS: "linux", GOARCH: "amd64"}
	tests := []struct {
		target buildPlatform
		wantCC string
	}{
		{buildPlatform{GOOS: "linux", GOARCH: "386"}, "gcc -m32"},
		{buildPlatform{GOOS: "linux", GOARCH: "arm"}, "arm-linux-gnueabihf-gcc"},
		{buildPlatform{GOOS: "linux", GOARCH: "arm64"}, "aarch64-linux-gnu-gcc"},
		{buildPlatform{GOOS: "windows", GOARCH: "386"}, "i686-w64-mingw32-gcc"},
		{buildPlatform{GOOS: "windows", GOARCH: "amd64"}, "x86_64-w64-mingw32-gcc"},
	}
	for _, tc := range tests {
		got, _ := defaultCrossCompilers(host, tc.target)
		if got != tc.wantCC {
			t.Errorf("compiler for %s = %q, want %q", tc.target, got, tc.wantCC)
		}
	}

	if cc, _ := defaultCrossCompilers(host, buildPlatform{GOOS: "windows", GOARCH: "arm64"}); cc != "" {
		t.Fatalf("Windows/ARM64 must require an explicit compiler, got %q", cc)
	}
}

func TestCGOCompilerEnvRequiresUnknownCrossCompiler(t *testing.T) {
	host := buildPlatform{GOOS: "linux", GOARCH: "amd64"}
	target := buildPlatform{GOOS: "windows", GOARCH: "arm64"}
	if _, err := cgoCompilerEnv(host, target, "", ""); err == nil {
		t.Fatal("unknown cross compiler accepted")
	}

	envs, err := cgoCompilerEnv(host, target, "clang --target=aarch64-w64-windows-gnu", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(envs, "\n"); !strings.Contains(got, "CC=clang --target=aarch64-w64-windows-gnu") {
		t.Fatalf("explicit compiler missing from environment: %q", got)
	}
}

func TestParseAndValidateProductMetadata(t *testing.T) {
	const revision = "dbec4fe67f2a8263091f2f101a1487094cb1a1c3"
	info := &debug.BuildInfo{
		GoVersion: requiredGoVersion,
		Path:      productBuildPath,
		Settings: []debug.BuildSetting{
			{Key: "GOOS", Value: "linux"},
			{Key: "GOARCH", Value: "amd64"},
			{Key: "vcs.revision", Value: revision},
			{Key: "vcs.modified", Value: "false"},
		},
	}
	meta, err := parseProductMetadata(info)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProductMetadata(meta, revision); err != nil {
		t.Fatal(err)
	}
	if meta.GOOS != "linux" || meta.GOARCH != "amd64" || meta.Modified {
		t.Fatalf("unexpected product metadata: %+v", meta)
	}
}

func TestValidateProductMetadataRejectsUnsafeProducts(t *testing.T) {
	const revision = "dbec4fe67f2a8263091f2f101a1487094cb1a1c3"
	base := productMetadata{
		GoVersion: requiredGoVersion,
		GOOS:      "linux",
		GOARCH:    "amd64",
		Revision:  revision,
	}
	tests := []struct {
		name string
		edit func(*productMetadata)
		want string
	}{
		{"stale revision", func(meta *productMetadata) { meta.Revision = "24d2db56cdd6c0e8097eda61b46e475eca3d006a" }, "stale product revision"},
		{"dirty product", func(meta *productMetadata) { meta.Modified = true }, "modified source tree"},
		{"wrong Go", func(meta *productMetadata) { meta.GoVersion = "go1.26.4" }, requiredGoVersion},
		{"unsupported target", func(meta *productMetadata) { meta.GOARCH = "mips" }, "unsupported target platform"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta := base
			tc.edit(&meta)
			err := validateProductMetadata(meta, revision)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestParseProductMetadataRejectsIncompleteBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{"nil", nil, "missing Go build information"},
		{"wrong path", &debug.BuildInfo{Path: "example.com/not-opennox"}, "build path"},
		{"missing setting", &debug.BuildInfo{Path: productBuildPath}, "missing GOOS"},
		{"invalid dirty bit", &debug.BuildInfo{
			Path: productBuildPath,
			Settings: []debug.BuildSetting{
				{Key: "GOOS", Value: "linux"},
				{Key: "GOARCH", Value: "amd64"},
				{Key: "vcs.revision", Value: "revision"},
				{Key: "vcs.modified", Value: "maybe"},
			},
		}, "invalid vcs.modified"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseProductMetadata(tc.info)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}
