package main

import (
	"debug/buildinfo"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/opennox/opennox/v1/internal/vers/git"
)

const (
	BinOpenNox        = "opennox"
	BinOpenNoxHD      = "opennox-hd"
	BinOpenNoxDebug   = "opennox-debug"
	BinOpenNoxHDDebug = "opennox-hd-debug"
	BinServer         = "opennox-server"
)

const (
	versPackage       = "github.com/opennox/opennox/v1/internal/version"
	productBuildPath  = "github.com/opennox/opennox/v1/cmd/opennox"
	requiredGoVersion = "go1.26.5"
	cgoCFlagsAllow    = `(-fsigned-char)|(-fshort-wchar)|(-fno-strict-aliasing)|(-fno-strict-overflow)`
)

var (
	defTargets = []string{
		BinOpenNox, BinOpenNoxHD, BinServer,
	}
	fOut     = flag.String("o", "", "output directory")
	fSrc     = flag.String("s", "", "source directory")
	fOS      = flag.String("os", runtime.GOOS, "target OS to build for")
	fArch    = flag.String("arch", runtime.GOARCH, "target architecture to build for")
	fCC      = flag.String("cc", "", "C compiler for the target (required when no cross compiler default exists)")
	fCXX     = flag.String("cxx", "", "C++ compiler for the target")
	fGO386   = flag.String("go386", "sse2", "GO386 tuning value")
	fGOAMD64 = flag.String("goamd64", "v1", "GOAMD64 tuning value")
	fGOARM   = flag.String("goarm", "7,hardfloat", "GOARM tuning value")
	fGOARM64 = flag.String("goarm64", "v8.0", "GOARM64 tuning value")
	fSafe    = flag.Bool("safe", false, "build a safe version (will run significantly slower)")
	fDryRun  = flag.Bool("n", false, "print build commands without running them")
	fGo      = flag.String("go", "go", "go command to use")
	fVerify  = flag.Bool("verify", false, "verify product metadata against the clean current source revision instead of building")
	fVerbose = flag.Bool("v", false, "verbose mode")
)

func main() {
	flag.Parse()
	if err := checkToolchains(*fGo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *fVerify {
		if err := verifyProducts(flag.Args()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := build(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkToolchains(goCommand string) error {
	if err := checkGoVersion("noxbuild compiler", runtime.Version()); err != nil {
		return err
	}
	out, err := exec.Command(goCommand, "env", "GOVERSION").CombinedOutput()
	if err != nil {
		return fmt.Errorf("check child Go toolchain %q: %w: %s", goCommand, err, strings.TrimSpace(string(out)))
	}
	return checkGoVersion("child Go toolchain", strings.TrimSpace(string(out)))
}

func checkGoVersion(name, got string) error {
	if got != requiredGoVersion {
		return fmt.Errorf("%s must be exactly %s (got %q); use the repository Go wrapper", name, requiredGoVersion, got)
	}
	return nil
}

type productMetadata struct {
	GoVersion string
	GOOS      string
	GOARCH    string
	Revision  string
	Modified  bool
}

func parseProductMetadata(info *buildinfo.BuildInfo) (productMetadata, error) {
	if info == nil {
		return productMetadata{}, fmt.Errorf("missing Go build information")
	}
	if info.Path != productBuildPath {
		return productMetadata{}, fmt.Errorf("build path is %q, want %q", info.Path, productBuildPath)
	}
	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	required := []string{"GOOS", "GOARCH", "vcs.revision", "vcs.modified"}
	for _, key := range required {
		if _, ok := settings[key]; !ok {
			return productMetadata{}, fmt.Errorf("missing %s build setting", key)
		}
	}
	var modified bool
	switch settings["vcs.modified"] {
	case "false":
	case "true":
		modified = true
	default:
		return productMetadata{}, fmt.Errorf("invalid vcs.modified setting %q", settings["vcs.modified"])
	}
	return productMetadata{
		GoVersion: info.GoVersion,
		GOOS:      settings["GOOS"],
		GOARCH:    settings["GOARCH"],
		Revision:  settings["vcs.revision"],
		Modified:  modified,
	}, nil
}

func validateProductMetadata(meta productMetadata, revision string) error {
	if meta.GoVersion != requiredGoVersion {
		return fmt.Errorf("Go version is %q, want %q", meta.GoVersion, requiredGoVersion)
	}
	platform := buildPlatform{GOOS: meta.GOOS, GOARCH: meta.GOARCH}
	if err := platform.validate(); err != nil {
		return err
	}
	if meta.Revision != revision {
		return fmt.Errorf("stale product revision %q, current source revision is %q; stop old OpenNox processes and rebuild", meta.Revision, revision)
	}
	if meta.Modified {
		return fmt.Errorf("product was built from a modified source tree; commit the source and rebuild")
	}
	return nil
}

func cleanSourceRevision() (string, error) {
	revision := git.SHA()
	if revision == "" {
		return "", fmt.Errorf("cannot determine current source revision")
	}
	out, err := exec.Command("git", "status", "--porcelain=v1", "--untracked-files=normal").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("check current source status: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if status := strings.TrimSpace(string(out)); status != "" {
		return "", fmt.Errorf("current source tree is not clean; commit or stash these changes before verifying products:\n%s", status)
	}
	return revision, nil
}

func verifyProducts(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("-verify requires at least one product path")
	}
	revision, err := cleanSourceRevision()
	if err != nil {
		return err
	}
	for _, path := range paths {
		info, err := buildinfo.ReadFile(path)
		if err != nil {
			return err
		}
		meta, err := parseProductMetadata(info)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := validateProductMetadata(meta, revision); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fmt.Printf("verified %s: %s %s/%s revision=%s clean\n", path, meta.GoVersion, meta.GOOS, meta.GOARCH, meta.Revision)
	}
	return nil
}

func build(args []string) error {
	if *fOS == "windows" {
		defTargets = append(defTargets, BinOpenNoxDebug, BinOpenNoxHDDebug)
	}
	targets := args
	if len(targets) == 0 {
		targets = defTargets
	}
	start := time.Now()
	// TODO: run in parallel
	for _, target := range targets {
		if err := buildTarget(target); err != nil {
			return err
		}
	}
	fmt.Printf("=== DONE in %v ===\n", time.Since(start))
	vers, sha := git.Version()
	fmt.Printf("version: %s (%s)\n", vers, sha)
	return nil
}

func buildTarget(target string) error {
	switch target {
	default:
		return fmt.Errorf("unsupported target: %q", target)
	case BinServer, "server":
		return goBuild("opennox", BinServer, &buildOpts{
			CGO: true, Tags: []string{"server"},
		})
	case BinOpenNox, "client":
		return goBuild("opennox", BinOpenNox, &buildOpts{
			CGO: true, GUI: true,
		})
	case BinOpenNoxHD, "client-hd", "hd":
		return goBuild("opennox", BinOpenNoxHD, &buildOpts{
			CGO: true, GUI: true, Tags: []string{"highres"},
		})
	case BinOpenNoxDebug, "client-debug":
		return goBuild("opennox", BinOpenNoxDebug, &buildOpts{
			CGO: true,
		})
	case BinOpenNoxHDDebug, "client-hd-debug":
		return goBuild("opennox", BinOpenNoxHDDebug, &buildOpts{
			CGO: true, Tags: []string{"highres"},
		})
	}
}

type buildOpts struct {
	CGO  bool
	GUI  bool
	Tags []string
}

type buildPlatform struct {
	GOOS   string
	GOARCH string
}

func (p buildPlatform) String() string {
	return p.GOOS + "/" + p.GOARCH
}

func (p buildPlatform) validate() error {
	valid := false
	switch p.GOOS {
	case "linux":
		valid = p.GOARCH == "386" || p.GOARCH == "amd64" || p.GOARCH == "arm" || p.GOARCH == "arm64"
	case "windows":
		valid = p.GOARCH == "386" || p.GOARCH == "amd64" || p.GOARCH == "arm64"
	case "darwin":
		valid = p.GOARCH == "amd64" || p.GOARCH == "arm64"
	}
	if !valid {
		return fmt.Errorf("unsupported target platform %s", p)
	}
	return nil
}

func (p buildPlatform) tuningEnv(go386, goamd64, goarm, goarm64 string) string {
	switch p.GOARCH {
	case "386":
		return "GO386=" + go386
	case "amd64":
		return "GOAMD64=" + goamd64
	case "arm":
		return "GOARM=" + goarm
	case "arm64":
		return "GOARM64=" + goarm64
	default:
		panic("validated platform has no architecture tuning")
	}
}

func defaultCrossCompilers(host, target buildPlatform) (cc, cxx string) {
	switch target.GOOS {
	case "windows":
		switch target.GOARCH {
		case "386":
			return "i686-w64-mingw32-gcc", "i686-w64-mingw32-g++"
		case "amd64":
			return "x86_64-w64-mingw32-gcc", "x86_64-w64-mingw32-g++"
		}
	case "linux":
		switch target.GOARCH {
		case "386":
			if host.GOOS == "linux" && host.GOARCH == "amd64" {
				return "gcc -m32", "g++ -m32"
			}
			return "i686-linux-gnu-gcc", "i686-linux-gnu-g++"
		case "amd64":
			if host.GOOS == "linux" && host.GOARCH == "386" {
				return "gcc -m64", "g++ -m64"
			}
			return "x86_64-linux-gnu-gcc", "x86_64-linux-gnu-g++"
		case "arm":
			return "arm-linux-gnueabihf-gcc", "arm-linux-gnueabihf-g++"
		case "arm64":
			return "aarch64-linux-gnu-gcc", "aarch64-linux-gnu-g++"
		}
	case "darwin":
		if host.GOOS == "darwin" {
			switch target.GOARCH {
			case "amd64":
				return "clang -arch x86_64", "clang++ -arch x86_64"
			case "arm64":
				return "clang -arch arm64", "clang++ -arch arm64"
			}
		}
	}
	return "", ""
}

func cgoCompilerEnv(host, target buildPlatform, cc, cxx string) ([]string, error) {
	isCross := host != target
	if isCross {
		defCC, defCXX := defaultCrossCompilers(host, target)
		if cc == "" {
			cc = defCC
		}
		if cxx == "" {
			cxx = defCXX
		}
		if cc == "" {
			return nil, fmt.Errorf("cross CGo target %s from %s requires -cc (and usually -cxx)", target, host)
		}
	}
	var envs []string
	if cc != "" {
		envs = append(envs, "CC="+cc, "CC_FOR_TARGET="+cc)
	}
	if cxx != "" {
		envs = append(envs, "CXX="+cxx, "CXX_FOR_TARGET="+cxx)
	}
	return envs, nil
}

func goBuild(cmd string, bin string, opts *buildOpts) error {
	target := buildPlatform{GOOS: *fOS, GOARCH: *fArch}
	if err := target.validate(); err != nil {
		return err
	}
	host := buildPlatform{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	goos := target.GOOS
	switch goos {
	case "windows":
		if filepath.Ext(bin) == "" {
			bin += ".exe"
		}
	}
	bin = filepath.Join(*fOut, bin)
	if opts == nil {
		opts = &buildOpts{}
	}
	if *fSafe {
		opts.Tags = append(opts.Tags, "safe")
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	vers, sha := git.Version()
	var (
		LDFLAGS = []string{
			"-X '" + versPackage + ".commit=" + sha + "'",
			"-X '" + versPackage + ".version=" + vers + "'",
		}
		GCFLAGS  = "-trimpath=" + wd
		ASMFLAGS = "-trimpath=" + wd
	)
	if opts.GUI {
		switch goos {
		case "windows":
			opts.Tags = append(opts.Tags, "guiapp")
			LDFLAGS = append(LDFLAGS, "-H windowsgui")
		}
	}
	var (
		envs []string
		args = []string{
			*fGo, "build", "-v",
			`-ldflags=` + strings.Join(LDFLAGS, " "),
			"-gcflags=" + GCFLAGS,
			"-asmflags=" + ASMFLAGS,
		}
	)
	if *fVerbose {
		args = append(args, "-work", "-x")
		work := filepath.Join(os.TempDir(), "go-opennox-build")
		if err := os.RemoveAll(work); err != nil {
			return err
		}
		if err := os.MkdirAll(work, 0755); err != nil {
			return err
		}
		envs = append(envs,
			"GOTMPDIR="+work,
		)
	}
	envs = append(envs,
		"GOOS="+goos,
		"GOARCH="+target.GOARCH,
		target.tuningEnv(*fGO386, *fGOAMD64, *fGOARM, *fGOARM64),
	)
	if opts.CGO {
		envs = append(envs,
			"CGO_ENABLED=1",
			"CGO_CFLAGS_ALLOW="+cgoCFlagsAllow,
		)
		compilerEnvs, err := cgoCompilerEnv(host, target, *fCC, *fCXX)
		if err != nil {
			return err
		}
		envs = append(envs, compilerEnvs...)
	} else {
		envs = append(envs, "CGO_ENABLED=0")
	}
	if len(opts.Tags) != 0 {
		args = append(args, "-tags", strings.Join(opts.Tags, ","))
	}
	src := filepath.Join(*fSrc, "cmd", cmd)
	if !filepath.IsAbs(src) {
		src = "." + string(filepath.Separator) + src
	}
	args = append(args,
		"-o", bin,
		src,
	)
	return doEnvs(wd, envs, args...)
}

func do(cmd ...string) error {
	fmt.Println("+ " + strings.Join(cmd, " "))
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	return c.Run()
}

func doEnvs(wd string, envs []string, cmd ...string) error {
	logArgs := make([]string, 0, len(envs)+len(cmd))
	logArgs = append(logArgs, envs...)
	logArgs = append(logArgs, cmd...)
	fmt.Println("+ " + strings.Join(logArgs, " "))
	if *fDryRun {
		return nil
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Dir = wd
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), envs...)
	return c.Run()
}
