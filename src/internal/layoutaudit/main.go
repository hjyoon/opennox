// Command layoutaudit reports target-specific Go type sizes and field offsets.
// It is a diagnostic tool for separating the original 32-bit ABI from native
// runtime layouts; it does not certify C/Go ABI compatibility by itself.
package main

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const requiredGoVersion = "go1.26.5"

var (
	targetOS   = flag.String("os", runtime.GOOS, "target GOOS")
	targetArch = flag.String("arch", runtime.GOARCH, "target GOARCH")
	pkgPattern = flag.String("package", "./server", "package pattern to inspect")
	typeNames  = flag.String("types", "ObjectIndex,Object,PlayerJournal,MinimapItem,EquipmentData,Player,NPC", "comma-separated named types")
	fieldTypes = flag.String("fields", "", "comma-separated types whose field offsets should be printed")
	showErrors = flag.Bool("show-errors", false, "print package loading/type errors")
)

func main() {
	flag.Parse()
	if runtime.Version() != requiredGoVersion {
		fmt.Fprintf(os.Stderr, "layoutaudit must be built with exactly %s (got %q)\n", requiredGoVersion, runtime.Version())
		os.Exit(2)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesSizes,
		Env: append(os.Environ(),
			"GOOS="+*targetOS,
			"GOARCH="+*targetArch,
		),
	}
	pkgs, err := packages.Load(cfg, *pkgPattern)
	if err != nil {
		return fmt.Errorf("load %s for %s/%s: %w", *pkgPattern, *targetOS, *targetArch, err)
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("package pattern %q resolved to %d packages, want 1", *pkgPattern, len(pkgs))
	}
	pkg := pkgs[0]
	if pkg.Types == nil || pkg.TypesSizes == nil {
		return fmt.Errorf("type information unavailable for %s", *pkgPattern)
	}
	if len(pkg.Errors) != 0 {
		fmt.Fprintf(os.Stderr, "layoutaudit: %d package errors retained for diagnostic layout extraction\n", len(pkg.Errors))
		if *showErrors {
			for _, err := range pkg.Errors {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	fields := nameSet(*fieldTypes)
	fmt.Printf("target\t%s/%s\n", *targetOS, *targetArch)
	fmt.Printf("package\t%s\n", pkg.PkgPath)
	fmt.Printf("pointer_size\t%d\n", pkg.TypesSizes.Sizeof(types.Typ[types.UnsafePointer]))
	fmt.Printf("package_errors\t%d\n", len(pkg.Errors))
	for _, name := range names(*typeNames) {
		obj := pkg.Types.Scope().Lookup(name)
		if obj == nil {
			return fmt.Errorf("type %q not found in %s", name, pkg.PkgPath)
		}
		fmt.Printf("type\t%s\t%d\n", name, pkg.TypesSizes.Sizeof(obj.Type()))
		if !fields[name] {
			continue
		}
		st, err := structType(obj.Type())
		if err != nil {
			return fmt.Errorf("type %s: %w", name, err)
		}
		vars := make([]*types.Var, st.NumFields())
		for i := range vars {
			vars[i] = st.Field(i)
		}
		offsets := pkg.TypesSizes.Offsetsof(vars)
		for i, field := range vars {
			fmt.Printf("field\t%s.%s\t%d\n", name, field.Name(), offsets[i])
		}
	}
	return nil
}

func structType(t types.Type) (*types.Struct, error) {
	if named, ok := t.(*types.Named); ok {
		t = named.Underlying()
	}
	st, ok := t.(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("underlying type is %T, not struct", t)
	}
	return st, nil
}

func names(s string) []string {
	set := nameSet(s)
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func nameSet(s string) map[string]bool {
	out := make(map[string]bool)
	for _, name := range strings.Split(s, ",") {
		if name = strings.TrimSpace(name); name != "" {
			out[name] = true
		}
	}
	return out
}
