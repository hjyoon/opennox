package legacy

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
)

func TestScriptMergeGlobals541E40Native(t *testing.T) {
	fixture, err := os.ReadFile("testdata/script_merge_globals_541e40_test.c")
	if err != nil {
		t.Fatal(err)
	}
	body := teamReportBody210(t, "server__script__file.c", "int nox_script_readWriteVvv_541E40(")
	program := strings.Replace(string(fixture), "// PRODUCTION_GLOBAL_MERGE_541E40", body, 1)
	var input bytes.Buffer
	word := func(dst *bytes.Buffer, value uint32) { _ = binary.Write(dst, binary.LittleEndian, value) }
	global := func(name string, symbols uint32) []byte {
		var out bytes.Buffer
		out.WriteString("FUNC")
		word(&out, uint32(len(name)))
		out.WriteString(name)
		word(&out, 0) // Parameters.
		word(&out, 0) // Return values.
		out.WriteString("SYMB")
		word(&out, symbols)
		word(&out, 0)
		for range symbols {
			word(&out, 1)
		}
		out.WriteString("DATA")
		word(&out, 4)
		word(&out, uint32(asm.OpReturn))
		return out.Bytes()
	}
	word(&input, uint32(asm.OpReturn))
	word(&input, 3)
	for _, names := range [][2]string{{"GLOBAL", "GLOBAL"}, {"GLOBAL", strings.Repeat("G", 64)}, {strings.Repeat("A", 4096), strings.Repeat("B", 4096)}} {
		for _, name := range []string{names[0], names[1], names[1]} {
			data := append(global(name, 0), global(name, 4)...)
			word(&input, uint32(len(data)))
			input.Write(data)
		}
	}
	modes := []struct {
		name  string
		flags []string
	}{{"O0", []string{"-O0"}}, {"O2", []string{"-O2"}}}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		modes = append(modes, struct {
			name  string
			flags []string
		}{"sanitized", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}})
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			binary := filepath.Join(t.TempDir(), "script-merge-globals")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-Wall", "-Wextra", "-Werror", "-Wno-sign-compare")
			args = append(args, mode.flags...)
			args = append(args, "-x", "c", "-", "-o", binary)
			cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
			cmd.Stdin = strings.NewReader(program)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("compile complete production GLOBAL merger: %v\n%s", err, out)
			}
			cmd = exec.Command(binary)
			cmd.Stdin = bytes.NewReader(input.Bytes())
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("native GLOBAL name/metadata merge: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
