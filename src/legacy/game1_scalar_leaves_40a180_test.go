package legacy

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/common/memmap"
)

type game1LeafConfig40A180 struct {
	Name     string
	Settings []uint32
	Limits   [6]byte
}

type game1LeafCase40A180 struct {
	Name         string
	Kind         uint32
	Entry        uint32
	Argument     uint32
	WindowA      [76]byte
	WindowB      [92]byte
	OriginalRate *uint32
}

func game1LeafHash40A180(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func game1LeafPut40A180(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}

func game1LeafArchive40A180(t *testing.T, directory string, files []string) {
	t.Helper()
	manifest := make(map[string]map[string]any)
	archive, err := os.OpenFile(filepath.Join(directory, "old-C-before-first-EXEC.zip"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(archive)
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		manifest[name] = map[string]any{"bytes": len(data), "sha256": game1LeafHash40A180(data)}
		member, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = member.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	member, err := z.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = member.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if err = archive.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(filepath.Join(directory, "old-C-before-first-EXEC.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if len(reader.File) != len(files)+1 {
		t.Fatal("exact old-C PREEXEC member count")
	}
	for _, member := range reader.File {
		stream, err := member.Open()
		if err != nil {
			t.Fatal(err)
		}
		actual, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		if member.Name == "manifest.json" {
			if !bytes.Equal(actual, data) {
				t.Fatal("PREEXEC manifest")
			}
			continue
		}
		source, err := os.ReadFile(filepath.Join(directory, member.Name))
		if err != nil || !bytes.Equal(actual, source) {
			t.Fatal("old-C PREEXEC whole member bytes", member.Name, err)
		}
	}
	game1LeafPut40A180(t, filepath.Join(directory, "old-C-PREEXEC-byte-verified.json"), map[string]any{"members": manifest, "before_first_execution": true, "actual_old_C_started": false})
}

func game1LeafOldSource40A180(t *testing.T) []byte {
	t.Helper()
	program, err := os.ReadFile("testdata/game1_scalar_leaves_40a180_old_c.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1.c")
	if err != nil {
		t.Fatal(err)
	}
	entries := []struct{ name, signature, digest string }{
		{"sub_409A70", "int sub_409A70(short a1) {", "1b9cffc7b8e4c2e8b016246fdfc4b73e4595d479788cfd4b98c7caf1ff99c6d5"},
		{"sub_40A180", "unsigned char sub_40A180(short a1) {", "a859fcb8b5c18e1c966c0e6f2df7c774a6bc5a5cd8c2e7406a69f8050d752b1b"},
		{"sub_40A220", "int sub_40A220() {", "765319633bab542a6de8b7b73bdeafc06878565565edafe4467333ae3445022d"},
		{"sub_40A3C0", "unsigned int sub_40A3C0(unsigned int a1) {", "4fa88c8031449a9e3b0ead4033faed7b49e05bee267c93184050a61f989d8c85"},
		{"sub_40A6A0", "int sub_40A6A0(int a1) {", "208ed2f2b05926e2574186ce953e30c3e4da40db16b0a9882db67d23b9443cdf"},
		{"sub_40A6B0", "int sub_40A6B0() {", "a1da7528517bc3de361e1f203701a4435afb9d22b6663ffb563637435bfbf95a"},
		{"sub_40A710", "int sub_40A710(int a1) {", "33adc12701852ecc9e4c1c7d5934433fd131b443b21d5d4692e0889ae659fc42"},
	}
	for _, entry := range entries {
		if bytes.Count(source, []byte(entry.signature)) != 1 {
			t.Fatal("unique preserved C definition", entry.name)
		}
		start := bytes.Index(source, []byte(entry.signature))
		opening := start + strings.Index(entry.signature, "{")
		depth, end := 0, -1
		for i := opening; i < len(source); i++ {
			if source[i] == '{' {
				depth++
			}
			if source[i] == '}' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}
		if end < 0 || game1LeafHash40A180(source[start:end]) != entry.digest {
			t.Fatal("exact preserved old C braces-inclusive SHA", entry.name)
		}
		marker := []byte("// ORIGINAL_" + entry.name)
		if bytes.Count(program, marker) != 1 {
			t.Fatal("exact standalone C marker", entry.name)
		}
		program = bytes.Replace(program, marker, source[start:end], 1)
	}
	return program
}

func game1LeafInputs40A180(t *testing.T) ([]game1LeafConfig40A180, []game1LeafCase40A180, []byte) {
	t.Helper()
	raw, err := os.ReadFile("testdata/network_tables_r256_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if game1LeafHash40A180(raw) != "95430f633e7c21378aa2e0abb08fe941b7aeb9e85be0d4d477be5b0fa33c4532" {
		t.Fatal("unchanged original network table fixture")
	}
	original := networkTableOriginal256(t)
	limits := [][6]byte{{0, 0, 0, 0, 0, 0}, {0, 1, 127, 128, 254, 255}, {255, 0, 128, 1, 127, 254}}
	var configs []game1LeafConfig40A180
	for _, row := range original.Rows {
		if row.Name != "settings" {
			continue
		}
		for i, values := range limits {
			configs = append(configs, game1LeafConfig40A180{Name: fmt.Sprintf("%s/limits-%d", row.Variant, i), Settings: append([]uint32(nil), row.Table...), Limits: values})
		}
	}
	words := []uint32{0, 1, 0x7ff, 0x800, 0x801, 0x7fff, 0x8000, 0x8001, 0xffff, 0x10000, 0x7fffffff, 0x80000000, 0x80000001, 0xfffffffe, 0xffffffff}
	var cases []game1LeafCase40A180
	add := func(kind, entry, arg, state uint32, name string, table []uint32, recorded *uint32) {
		c := game1LeafCase40A180{Name: name, Kind: kind, Entry: entry, Argument: arg, OriginalRate: recorded}
		for i := range c.WindowA {
			c.WindowA[i] = byte(i*29 + 17)
		}
		for i := range c.WindowB {
			c.WindowB[i] = byte(i*31 + 23)
		}
		binary.LittleEndian.PutUint32(c.WindowA[0:], state)
		binary.LittleEndian.PutUint32(c.WindowA[8:], state)
		binary.LittleEndian.PutUint32(c.WindowB[88:], state)
		if table != nil {
			if len(table) != 10 {
				t.Fatal("five rate pairs")
			}
			for i, v := range table {
				binary.LittleEndian.PutUint32(c.WindowA[12+4*i:], v)
			}
		}
		cases = append(cases, c)
	}
	for _, state := range words {
		add(0, 0x40a220, 0, state, fmt.Sprintf("A220/state-%08x", state), nil, nil)
		add(3, 0x40a6b0, 0, state, fmt.Sprintf("A6B0/state-%08x", state), nil, nil)
		for _, arg := range words {
			add(1, 0x40a3c0, arg, state, fmt.Sprintf("A3C0/arg-%08x-state-%08x", arg, state), nil, nil)
			add(2, 0x40a6a0, arg, state, fmt.Sprintf("A6A0/arg-%08x-state-%08x", arg, state), nil, nil)
		}
	}
	for _, row := range original.Rows {
		if row.Name != "rate" {
			continue
		}
		for _, pair := range row.Cases {
			value := pair[1]
			add(4, 0x40a710, pair[0], 0xa55aa55a, fmt.Sprintf("A710/%s/input-%08x", row.Variant, pair[0]), row.Table, &value)
		}
	}
	if len(configs) != 9 || len(cases) != 516 {
		t.Fatalf("input-only configurations=%d scalar-cases=%d", len(configs), len(cases))
	}
	var input bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&input, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	word(0x36313041)
	word(uint32(len(configs)))
	for _, c := range configs {
		if len(c.Settings) != 6 {
			t.Fatal("six live setting keys")
		}
		for _, v := range c.Settings {
			word(v)
		}
		input.Write(c.Limits[:])
	}
	word(uint32(len(cases)))
	for _, c := range cases {
		word(c.Kind)
		word(c.Argument)
		input.Write(c.WindowA[:])
		input.Write(c.WindowB[:])
	}
	return configs, cases, input.Bytes()
}

func game1LeafReference40A180(t *testing.T, source, input []byte) ([]byte, string) {
	t.Helper()
	root := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS")
	if root == "" {
		root = os.Getenv("OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS")
	}
	if root == "" || !filepath.IsAbs(root) {
		t.Fatal("absolute retained prepared-test artifact directory required")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(root, "game1-six-leaves-old-C-")
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(directory, "old-C-six-leaves")
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	args := append(compiler[1:], "-std=c11", "-O2", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Wno-sign-compare", "-x", "c", "-", "-o", binaryPath)
	args = nativeFixtureCompileArgs(t, args)
	for name, data := range map[string][]byte{"source.c": source, "input.bin": input} {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	compile := exec.CommandContext(ctx, compiler[0], args...)
	compile.Stdin = bytes.NewReader(source)
	compile.WaitDelay = 5 * time.Second
	output, compileErr := compile.CombinedOutput()
	if err := os.WriteFile(filepath.Join(directory, "compile.combined.log"), output, 0600); err != nil {
		t.Fatal(err)
	}
	game1LeafPut40A180(t, filepath.Join(directory, "compile.Wait.json"), map[string]any{"argv": compile.Args, "Wait_terminal": true, "error": fmt.Sprint(compileErr), "context_error": fmt.Sprint(ctx.Err()), "source_sha256": game1LeafHash40A180(source), "input_sha256": game1LeafHash40A180(input)})
	if compileErr != nil || bytes.Contains(bytes.ToLower(output), []byte("warning:")) {
		t.Fatalf("compile exact old C: %v\n%s", compileErr, output)
	}
	exe, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && (len(exe) < 8 || !bytes.Equal(exe[:8], []byte{0xcf, 0xfa, 0xed, 0xfe, 0x0c, 0, 0, 1})) {
		t.Fatal("actual standalone Mach-O ARM64 executable")
	}
	mainPath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	mainBytes, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	game1LeafPut40A180(t, filepath.Join(directory, "actual-C-before-first-EXEC.json"), map[string]any{"executable": binaryPath, "bytes": len(exe), "sha256": game1LeafHash40A180(exe), "source_sha256": game1LeafHash40A180(source), "input_sha256": game1LeafHash40A180(input), "actual_old_C_started": false, "Go_main": mainPath, "Go_main_sha256": game1LeafHash40A180(mainBytes), "Go_PID": os.Getpid(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "new_original_CPU_execution": 0})
	game1LeafArchive40A180(t, directory, []string{"old-C-six-leaves", "source.c", "input.bin", "compile.combined.log", "compile.Wait.json", "actual-C-before-first-EXEC.json"})
	runCtx, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer runCancel()
	cmd := exec.CommandContext(runCtx, binaryPath)
	cmd.WaitDelay = 5 * time.Second
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	startErr := cmd.Start()
	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	runErr := startErr
	if startErr == nil {
		runErr = cmd.Wait()
	}
	for name, data := range map[string][]byte{"actual.stdout.bin": stdout.Bytes(), "actual.stderr.log": stderr.Bytes()} {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	exit := -999
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	game1LeafPut40A180(t, filepath.Join(directory, "actual.Wait.json"), map[string]any{"argv": cmd.Args, "PID": pid, "Wait_terminal": true, "exit_code": exit, "error": fmt.Sprint(runErr), "context_error": fmt.Sprint(runCtx.Err()), "forced": runCtx.Err() != nil, "stdout_bytes": stdout.Len(), "stdout_sha256": game1LeafHash40A180(stdout.Bytes()), "stderr_bytes": stderr.Len(), "executable_sha256": game1LeafHash40A180(exe)})
	if runErr != nil || exit != 0 || pid <= 0 || runCtx.Err() != nil || stderr.Len() != 0 {
		t.Fatalf("actual retained old C execution %v exit=%d pid=%d stderr=%s", runErr, exit, pid, stderr.Bytes())
	}
	return stdout.Bytes(), directory
}

func game1LeafCallGo40A180(entry, argument uint32) uint32 {
	switch entry {
	case 0x40a220:
		return uint32(sub40A220Native())
	case 0x40a3c0:
		return sub40A3C0Native(argument)
	case 0x40a6a0:
		return uint32(sub40A6A0Native(int32(argument)))
	case 0x40a6b0:
		return uint32(sub40A6B0Native())
	case 0x40a710:
		return uint32(sub40A710Native(int32(argument)))
	}
	panic("selected scalar native leaf")
}

func TestGame1ScalarLeavesABI40A180(t *testing.T) {
	for _, cell := range []struct{ base, offset uint32 }{{0x587000, 4652}, {0x587000, 4660}, {0x587000, 4664}, {0x587000, 4704}, {0x5d4594, 3500}, {0x5d4594, 3588}} {
		p := memmap.PtrOff(uintptr(cell.base), uintptr(cell.offset))
		if uintptr(p) <= 0xffffffff || p != game1LeafCellC40A180(cell.base, cell.offset) {
			t.Fatalf("actual native C/Go memmap alias base=%x offset=%d", cell.base, cell.offset)
		}
	}
	t.Log("GAME1_SIX_LEAVES_ABI_OK public-typed-C-entries=6 actual-live-memmap-aliases=6 native>4GiB pointer-truncation-claim=false")
}

func TestGame1ScalarLeavesOldC40A180(t *testing.T) {
	configs, cases, input := game1LeafInputs40A180(t)
	reference, directory := game1LeafReference40A180(t, game1LeafOldSource40A180(t), input)
	if len(reference) != 9*65536+len(cases)*(4+76+92) {
		t.Fatal("exact old C output protocol byte count")
	}
	a := memmap.Slice(0x587000, 4652)[:76]
	b := memmap.Slice(0x5d4594, 3500)[:92]
	savedA, savedB := bytes.Clone(a), bytes.Clone(b)
	defer copy(a, savedA)
	defer copy(b, savedB)
	tracePath := filepath.Join(directory, "actual-native-scalar-results.jsonl")
	trace, err := os.OpenFile(tracePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer trace.Close()
	encoder := json.NewEncoder(trace)
	nativeCalls := 0
	for index, config := range configs {
		t.Run("time-limits/"+config.Name, func(t *testing.T) {
			for i, v := range config.Settings {
				binary.LittleEndian.PutUint32(a[52+4*i:], v)
			}
			copy(b[:6], config.Limits[:])
			beforeA, beforeB := bytes.Clone(a), bytes.Clone(b)
			for value := 0; value <= 65535; value++ {
				wanted := reference[index*65536+value]
				goValue := sub40A180Native(int16(uint16(value)))
				cValue := game1LeafCallC40A180(0x40a180, uint32(value))
				nativeCalls += 2
				if goValue != wanted || cValue != uint32(wanted) {
					t.Fatalf("old C A180 input=%04x Go=%02x C=%08x old=%02x", value, goValue, cValue, wanted)
				}
			}
			if !bytes.Equal(a, beforeA) || !bytes.Equal(b, beforeB) {
				t.Fatal("read-only lookup changed live table/cell bytes")
			}
			t.Logf("GAME1_A180_ALL16_OK configuration=%s inputs=65536 Go-and-actual-C=131072 old-C-output-sha256=%s", config.Name, game1LeafHash40A180(reference[index*65536:(index+1)*65536]))
		})
	}
	for index, row := range cases {
		t.Run("scalar/"+row.Name, func(t *testing.T) {
			position := 9*65536 + index*172
			wanted := binary.LittleEndian.Uint32(reference[position:])
			afterA := reference[position+4 : position+80]
			afterB := reference[position+80 : position+172]
			if row.OriginalRate != nil && wanted != *row.OriginalRate {
				t.Fatal("old C changed frozen actual GAME.EXE rate-table result")
			}
			for _, backend := range []string{"Go", "C"} {
				copy(a, row.WindowA[:])
				copy(b, row.WindowB[:])
				var got uint32
				if backend == "Go" {
					got = game1LeafCallGo40A180(row.Entry, row.Argument)
				} else {
					got = game1LeafCallC40A180(row.Entry, row.Argument)
				}
				nativeCalls++
				actualA, actualB := bytes.Clone(a), bytes.Clone(b)
				e := encoder.Encode(map[string]any{"name": row.Name, "entry": row.Entry, "argument_u32": row.Argument, "backend": backend, "actual_return_u32": got, "old_C_actual_return_u32": wanted, "BEFORE_A": row.WindowA[:], "BEFORE_B": row.WindowB[:], "AFTER_A": actualA, "AFTER_B": actualB, "old_C_AFTER_A": afterA, "old_C_AFTER_B": afterB})
				if e != nil {
					t.Fatal(e)
				}
				if got != wanted || !bytes.Equal(actualA, afterA) || !bytes.Equal(actualB, afterB) {
					t.Fatalf("actual old C full return/168byte live state/canaries backend=%s got=%08x old=%08x", backend, got, wanted)
				}
			}
		})
	}
	game1LeafPut40A180(t, filepath.Join(directory, "actual-six-leaves-summary.json"), map[string]any{"A180_configurations": 9, "A180_inputs_each": 65536, "A180_Go_and_C_native_calls": 9 * 65536 * 2, "scalar_cases": len(cases), "scalar_Go_and_C_native_calls": len(cases) * 2, "actual_native_calls": nativeCalls, "old_C_reference_stdout_sha256": game1LeafHash40A180(reference), "C_executable_preserved_before_first_EXEC": true, "new_GAME_EXE_CPU_execution": 0, "scope": "immutable106B PE static facts and compiled exact preserved old C7 reference; no new original opcode execution or physical original stack/FPU claim"})
	if nativeCalls != 9*65536*2+len(cases)*2 {
		t.Fatal("exact selected actual native call count")
	}
	t.Logf("GAME1_SIX_LEAVES_ALL_OK actual-native-calls=%d scalar-cases=%d old-C-artifacts=%s new-original-execution=0", nativeCalls, len(cases), directory)
}
