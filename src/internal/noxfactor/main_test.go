package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoxFactor(t *testing.T) {
	dir := t.TempDir()
	goPath := filepath.Join(dir, "fixture.go")
	cPath := filepath.Join(dir, "fixture.c")
	require.NoError(t, os.WriteFile(goPath, []byte("package fixture\nfunc identity(x uint32_t) uint32_t { return x }\n"), 0644))
	require.NoError(t, os.WriteFile(cPath, []byte("extern int unused_value;\nint answer(void) { return 42; }\n"), 0644))
	require.NoError(t, run(dir))
	goData, err := os.ReadFile(goPath)
	require.NoError(t, err)
	require.Contains(t, string(goData), "uint32")
	require.NotContains(t, string(goData), "uint32_t")
	cData, err := os.ReadFile(cPath)
	require.NoError(t, err)
	require.NotContains(t, string(cData), "unused_value")
}

func TestFindExterns(t *testing.T) {
	got := findExterns([]byte(`
extern uint32_t some_var_123;
extern void *some_var_456 ;
extern int **arr1[5];
extern int **arr2 [1 + 5] ;
`))
	require.Equal(t, []SrcDecl{
		{Name: "some_var_123", Src: "extern uint32_t some_var_123;\n"},
		{Name: "some_var_456", Src: "extern void *some_var_456 ;\n"},
		{Name: "arr1", Src: "extern int **arr1[5];\n"},
		{Name: "arr2", Src: "extern int **arr2 [1 + 5] ;\n"},
	}, got)
}
