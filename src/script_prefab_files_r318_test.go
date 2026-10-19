package opennox

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/strman"
	"github.com/opennox/noxscript/ns/asm"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

// Authored SCRIPT03 records exercise the actual native merger and live VM.
// The expected bytes follow the file grammar, independently of the C writer.
func prefabScriptStreamR318(names, strings []string) []byte {
	var out bytes.Buffer
	word := func(v uint32) { _ = binary.Write(&out, binary.LittleEndian, v) }
	text := func(v string) { word(uint32(len(v))); out.WriteString(v) }
	out.WriteString("SCRIPT03STRG")
	word(uint32(len(strings)))
	for _, s := range strings {
		text(s)
	}
	out.WriteString("CODE")
	word(uint32(2 + len(names)))
	for index, name := range append([]string{"GLOBAL", "GLOBAL"}, names...) {
		out.WriteString("FUNC")
		text(name)
		word(0)
		word(0)
		out.WriteString("SYMB")
		count := uint32(0)
		if index == 1 {
			count = 4
		}
		word(count)
		word(0)
		for i := uint32(0); i < count; i++ {
			word(1)
		}
		out.WriteString("DATA")
		word(4)
		word(uint32(asm.OpReturn))
	}
	out.WriteString("DONE")
	return out.Bytes()
}

func TestScriptPrefabFilesR318(t *testing.T) {
	// Native merger scratch and game globals belong to a separate test process.
	const marker = "NOX_R318_PREFAB_FILES_CHILD"
	if os.Getenv(marker) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestScriptPrefabFilesR318$", "-test.v")
		cmd.Env = append(os.Environ(), marker+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("actual ScriptObject/native merge subprocess: %v\n%s", err, out)
		}
		t.Logf("actual ScriptObject/native merge subprocess:\n%s", out)
		return
	}

	root := t.TempDir()
	handles.Init()
	defer handles.Release()
	datapath.SetData(root)
	noxflags.ResetGame()
	s := NewServer(nil, nil, strman.New())
	noxServer = s
	defer s.Close()
	nativePath := (*byte)(memmap.PtrOff(0x973F18, 30760))
	payload := prefabScriptStreamR318([]string{"Probe%7%13%-29"}, nil)
	for _, tc := range []struct {
		name string
		load bool
		data []byte
	}{
		{"save absent script", false, nil},
		{"save nonempty script", false, payload},
		{"load empty script", true, nil},
		{"load nonempty script", true, payload},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, asm.NCobjName)
			_ = os.Remove(path)
			alloc.StrCopy(unsafe.Slice(nativePath, 1024), "unrelated-native-path")
			section := filepath.Join(t.TempDir(), "ScriptObject.section")
			var want bytes.Buffer
			_ = binary.Write(&want, binary.LittleEndian, uint16(1))
			_ = binary.Write(&want, binary.LittleEndian, uint32(len(tc.data)))
			want.Write(tc.data)
			mode := cryptfile.WriteOnly
			if tc.load {
				mode = cryptfile.ReadOnly
				if err := os.WriteFile(section, want.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
			} else if tc.data != nil {
				if err := os.WriteFile(path, tc.data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			cf, err := cryptfile.OpenFile(section, mode, -1)
			if err != nil {
				t.Fatal(err)
			}
			err = nox_server_mapRWScriptObject_505A40(cf, nil)
			closeErr := cf.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("section operation/close: %v / %v", err, closeErr)
			}
			if got := alloc.GoString(nativePath); got != path {
				t.Fatalf("native current path = %q, want actual section file %q", got, path)
			}
			got, err := os.ReadFile(section)
			if err != nil || !bytes.Equal(got, want.Bytes()) {
				t.Fatalf("ScriptObject bytes changed: %x / %v", got, err)
			}
			if tc.load {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, tc.data) {
					t.Fatalf("loaded compiled file: %x / %v", got, err)
				}
				if len(tc.data) != 0 {
					fn := s.noxScript.Funcs()[2]
					if fn.Name() != "Probe%7%13%-29" || fn.NamePref != "%7" || fn.PosOff != image.Pt(13, -29) {
						t.Fatalf("actual section-loaded function = %q/%q/%v", fn.Name(), fn.NamePref, fn.PosOff)
					}
				}
			}
		})
	}

	first := prefabScriptStreamR318([]string{"First%0%13%-29"}, []string{"first string"})
	second := prefabScriptStreamR318([]string{"Second%1%-91%37"}, []string{"second string"})
	want := prefabScriptStreamR318([]string{"First%0%13%-29", "Second%1%-91%37"}, []string{"first string", "second string"})
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "merge creates destination", true: "merge truncates previous destination"}[existing], func(t *testing.T) {
			dir := t.TempDir()
			a, b, dst := filepath.Join(dir, "a.obj"), filepath.Join(dir, "b.obj"), filepath.Join(dir, "blend.obj")
			for path, data := range map[string][]byte{a: first, b: second} {
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if existing {
				if err := os.WriteFile(dst, bytes.Repeat([]byte{0xcc}, 4096), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ca, fa := alloc.CString(a)
			defer fa()
			cb, fb := alloc.CString(b)
			defer fb()
			cdst, fdst := alloc.CString(dst)
			defer fdst()
			if result := nox_script_readWriteZzz_541670(ca, cb, cdst); result != 1 {
				t.Fatalf("actual merge returned %d", result)
			}
			got, err := os.ReadFile(dst)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("actual merged bytes (%d) differ from grammar (%d): %v", len(got), len(want), err)
			}
			var vm server.NoxScriptVM
			if err := vm.ReadScript(bytes.NewReader(got)); err != nil {
				t.Fatal(err)
			}
			funcs := vm.Funcs()
			if len(funcs) != 4 || funcs[2].PosOff != image.Pt(13, -29) || funcs[3].PosOff != image.Pt(-91, 37) {
				t.Fatal("merged live VM did not retain both independent coordinate pairs")
			}
			for path, original := range map[string][]byte{a: first, b: second} {
				data, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(data, original) {
					t.Fatalf("merge changed input %q: %v", path, err)
				}
			}
		})
	}
}
