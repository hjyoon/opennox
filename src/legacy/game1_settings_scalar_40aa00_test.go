package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type game1SettingsCase40AA00 struct {
	Name                                string
	Entry, Argument, FPS, Frame, Global uint32
	Window                              [80]byte
	IndependentBlob                     [20]byte
}

func game1SettingsInputs40AA00() ([]game1SettingsCase40AA00, []byte) {
	var rows []game1SettingsCase40AA00
	add := func(name string, entry, argument, fps, frame, start, cell uint32) {
		i := len(rows)
		r := game1SettingsCase40AA00{Name: name, Entry: entry, Argument: argument, FPS: fps, Frame: frame, Global: cell}
		for j := range r.Window {
			r.Window[j] = byte(17*i + 29*j + 0x53)
		}
		for j := range r.IndependentBlob {
			r.IndependentBlob[j] = byte(31*i + 11*j + 0xa7)
		}
		binary.LittleEndian.PutUint32(r.Window[8:], cell)             // DWORD at 3476
		binary.LittleEndian.PutUint32(r.Window[40:], cell^0xa5a55a5a) // DWORD at 3508
		binary.LittleEndian.PutUint32(r.Window[52:], start)           // DWORD at 3520
		binary.LittleEndian.PutUint32(r.Window[68:], cell)            // DWORD at 3536
		binary.LittleEndian.PutUint32(r.IndependentBlob[8:], cell^0x69c3965a)
		rows = append(rows, r)
	}
	fpsWords := []uint32{0, 1, 2, 30, 60, 0x0ccccccc, 0x0ccccccd, 0x40000000, 0x80000000, 0xffffffff}
	pairs := [][2]uint32{{0, 0}, {1, 0}, {0, 1}, {0xffffffff, 0}, {0, 0xffffffff}, {0x80000000, 1}}
	for i, fps := range fpsWords {
		for j, pair := range pairs {
			add(fmt.Sprintf("AA00/independent-%02d-%02d", i, j), 0x40aa00, 0, fps, pair[0], pair[1], uint32(0x12000000+i*16+j))
		}
	}
	// Literal boundary inputs only; no return value is calculated here.
	for i, v := range [][3]uint32{
		{10, 199, 0}, {10, 200, 0}, {10, 201, 0}, {30, 599, 0}, {30, 600, 0}, {30, 601, 0},
		{1, 9, 0xfffffff5}, {1, 9, 0xfffffff4}, {1, 9, 0xfffffff3},
		{0x0ccccccd, 3, 0}, {0x0ccccccd, 4, 0}, {0x0ccccccd, 5, 0},
	} {
		add(fmt.Sprintf("AA00/literal-boundary-%02d", i), 0x40aa00, 0, v[0], v[1], v[2], 0xfedc0000+uint32(i))
	}
	words := []uint32{0, 1, 2, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff, 0x01020304}
	for _, entry := range []uint32{0x40aa20, 0x40aa30, 0x40aa40, 0x40aa60, 0x40e090} {
		for i, word := range words {
			add(fmt.Sprintf("%06X/word-%08x", entry, word), entry, word, 30, uint32(i), 0, word)
		}
	}
	var input bytes.Buffer
	put := func(v uint32) { var word [4]byte; binary.LittleEndian.PutUint32(word[:], v); input.Write(word[:]) }
	put(0x36414153)
	put(uint32(len(rows)))
	for _, r := range rows {
		for _, v := range []uint32{r.Entry, r.Argument, r.FPS, r.Frame, r.Global} {
			put(v)
		}
		input.Write(r.Window[:])
		input.Write(r.IndependentBlob[:])
	}
	return rows, input.Bytes()
}

func game1SettingsOldSource40AA00(t *testing.T) []byte {
	t.Helper()
	program, err := os.ReadFile("testdata/game1_settings_scalar_40aa00_old_c.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME1.c")
	if err != nil {
		t.Fatal(err)
	}
	entries := []struct{ name, signature, digest string }{
		// Digests are the exact braces-inclusive current C definitions, not a model.
		{"sub_40AA00", "int sub_40AA00() {", "8a720b501cf4f1748660ea731d2ee951f06b24ee0562443cabd8eed72d7161ee"},
		{"sub_40AA20", "int sub_40AA20() {", "fc9759581d0f36b8a06eebba8600ea64f56d4c663a81c851fe17ae2bd3f39fbb"},
		{"sub_40AA30", "int sub_40AA30(int a1) {", "99dc2c61ac3b9eb8ee12588aab2442e06a069f2ab9dfa3b88a6f3ab3baa5c56d"},
		{"sub_40AA40", "int sub_40AA40() {", "9a94c6d92a166a99bb711ab1e929618988f046eed6a83cafa803dcfd9da9b44e"},
		{"sub_40AA60", "int sub_40AA60(int a1) {", "d7ea1ce4b9b12d0ab12ac66b1573cebf5be20241a9f10c4b2db0c29bcd16172f"},
		{"sub_40E090", "void sub_40E090() {", "e65f447cab5961c7d6e79ab4ab181487312e96d7c26d90f748a9793af3b1dc62"},
	}
	for _, e := range entries {
		if bytes.Count(source, []byte(e.signature)) != 1 {
			t.Fatal("unique preserved C definition", e.name)
		}
		start := bytes.Index(source, []byte(e.signature))
		opening := start + strings.Index(e.signature, "{")
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
		if end < 0 || game1LeafHash40A180(source[start:end]) != e.digest {
			t.Fatal("exact preserved old C SHA", e.name)
		}
		marker := []byte("// ORIGINAL_" + e.name)
		if bytes.Count(program, marker) != 1 {
			t.Fatal("unique old C template marker", e.name)
		}
		program = bytes.Replace(program, marker, source[start:end], 1)
	}
	return program
}

func game1SettingsCallGo40AA00(entry, argument uint32) uint32 {
	switch entry {
	case 0x40aa00:
		return uint32(sub40AA00Native())
	case 0x40aa20:
		return uint32(sub40AA20Native())
	case 0x40aa30:
		return uint32(sub40AA30Native(int32(argument)))
	case 0x40aa40:
		return uint32(sub40AA40Native())
	case 0x40aa60:
		return uint32(sub40AA60Native(int32(argument)))
	case 0x40e090:
		sub40E090Native()
		return 0 // transport placeholder only
	default:
		panic("selected settings scalar entry")
	}
}

func game1SettingsGlobalWindow40AA00(t *testing.T) (unsafe.Pointer, int, int) {
	t.Helper()
	cell := game1SettingsGlobal40E090()
	base, length, center, err := settingFlagsFixtureWindow409E40(cell)
	if err != nil || length < 4 || length > 36 || center < 0 || center > length-4 || unsafe.Add(base, center) != cell {
		t.Fatal("actual C global VM-bounded neighbour window", err)
	}
	return base, length, center
}

func game1SettingsSave40AA00(t *testing.T) func() {
	t.Helper()
	window := game1SettingsBlobC40AA00(3468, 80)
	other := game1SettingsBlobC40AA00(10976, 20)
	savedWindow, savedOther := bytes.Clone(window), bytes.Clone(other)
	globalBase, globalLength, _ := game1SettingsGlobalWindow40AA00(t)
	savedGlobal := settingFlagsFixtureSnapshot409E40(globalBase, globalLength)
	oldFrame, oldFPS := gameFrameHook, gameFPSHook
	return func() {
		copy(window, savedWindow)
		copy(other, savedOther)
		settingFlagsFixtureCopy409E40(globalBase, savedGlobal)
		gameFrameHook, gameFPSHook = oldFrame, oldFPS
	}
}

func TestGame1SettingsScalarABI40AA00(t *testing.T) {
	for _, offset := range []uint32{3476, 3508, 3520, 3536} {
		p := memmap.PtrOff(0x5d4594, uintptr(offset))
		if (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff) || p != game1SettingsCellC40AA00(offset) {
			t.Fatalf("actual full native live cell alias offset=%d", offset)
		}
	}
	p, width, align := game1SettingsGlobalC40AA00()
	rawBlob := game1SettingsBlobC40AA00(10976, 20)
	if p != game1SettingsGlobal40E090() || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= 0xffffffff) || width != 4 || align != 4 || uintptr(p)%align != 0 || p == unsafe.Pointer(&rawBlob[8]) {
		t.Fatal("actual C global DWORD width/alignment/address and independent packed blob")
	}
	t.Logf("GAME1_SETTINGS_SCALAR_ABI_OK public-C=6 live-aliases=4 separate-C-global=1 native-pointer-width=%d high-address-qualified=%t original-pointer-truncation-claim=false", unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(uintptr(0)) == 8)
}

func TestGame1SettingsScalarOldC40AA00(t *testing.T) {
	rows, input := game1SettingsInputs40AA00()
	if len(rows) != 112 || len(input) != 8+112*120 {
		t.Fatal("independent input protocol")
	}
	reference, directory := game1LeafReference40A180(t, game1SettingsOldSource40AA00(t), input)
	if len(reference) != len(rows)*108 {
		t.Fatal("exact old C output protocol")
	}
	restore := game1SettingsSave40AA00(t)
	defer restore()
	window := game1SettingsBlobC40AA00(3468, 80)
	other := game1SettingsBlobC40AA00(10976, 20)
	global := (*uint32)(game1SettingsGlobal40E090())
	globalBase, globalLength, globalCenter := game1SettingsGlobalWindow40AA00(t)
	trace, err := os.OpenFile(filepath.Join(directory, "actual-settings-scalar-results.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer trace.Close()
	encoder := json.NewEncoder(trace)
	calls := 0
	for index, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			old := reference[index*108 : (index+1)*108]
			wanted := binary.LittleEndian.Uint32(old)
			for _, backend := range []string{"Go", "C"} {
				copy(window, row.Window[:])
				copy(other, row.IndependentBlob[:])
				*global = row.Global
				globalBefore := settingFlagsFixtureSnapshot409E40(globalBase, globalLength)
				gameFPSHook = func() uint32 { return row.FPS }
				gameFrameHook = func() uint32 { return row.Frame }
				var got uint32
				if backend == "Go" {
					got = game1SettingsCallGo40AA00(row.Entry, row.Argument)
				} else {
					got = game1SettingsCallC40AA00(row.Entry, row.Argument)
				}
				calls++
				actualWindow, actualOther := bytes.Clone(window), bytes.Clone(other)
				actualGlobal := *global
				globalAfter := settingFlagsFixtureSnapshot409E40(globalBase, globalLength)
				globalExpected := bytes.Clone(globalBefore)
				copy(globalExpected[globalCenter:globalCenter+4], old[84:88])
				returnDefined := row.Entry != 0x40e090
				if err := encoder.Encode(map[string]any{"name": row.Name, "backend": backend, "entry": row.Entry, "argument_u32": row.Argument, "FPS": row.FPS, "frame": row.Frame, "return_defined": returnDefined, "actual_return_transport_u32": got, "old_C_return_transport_u32": wanted, "BEFORE_window": row.Window[:], "AFTER_window": actualWindow, "old_C_AFTER_window": old[4:84], "BEFORE_actual_global_u32": row.Global, "AFTER_actual_global_u32": actualGlobal, "old_C_AFTER_actual_global_u32": binary.LittleEndian.Uint32(old[84:88]), "actual_C_global_window_center": globalCenter, "BEFORE_actual_C_global_window": globalBefore, "AFTER_actual_C_global_window": globalAfter, "expected_actual_C_global_window": globalExpected, "BEFORE_independent_blob": row.IndependentBlob[:], "AFTER_independent_blob": actualOther, "old_C_AFTER_independent_blob": old[88:108]}); err != nil {
					t.Fatal(err)
				}
				if (returnDefined && got != wanted) || !bytes.Equal(actualWindow, old[4:84]) || actualGlobal != binary.LittleEndian.Uint32(old[84:88]) || !bytes.Equal(globalAfter, globalExpected) || !bytes.Equal(actualOther, old[88:108]) {
					t.Fatalf("exact old C full return/live100bytes/global4 backend=%s got=%08x old=%08x", backend, got, wanted)
				}
				t.Logf("GAME1_SETTINGS_SCALAR_OLD_C_OK case=%s backend=%s return-defined=%t raw-return=%08x", row.Name, backend, returnDefined, got)
			}
		})
	}
	if err := trace.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	game1LeafPut40A180(t, filepath.Join(directory, "actual-settings-scalar-summary.json"), map[string]any{"cases": len(rows), "native_calls": calls, "old_C_stdout_sha256": game1LeafHash40A180(reference), "whole_state_bytes_each": 104, "actual_separate_C_global": true, "E090_void_return_compared": false, "C_executable_preserved_before_first_EXEC": true, "new_original_CPU_execution": 0, "scope": "75B static PE and compiled exact preserved C6; portable C stable clock data only; PE-order controls reported separately; no original stack/FPU or incoming caller-body credit"})
	if calls != 224 {
		t.Fatal("exact actual native calls")
	}
}

func TestGame1SettingsScalarPEOrder40AA00(t *testing.T) {
	restore := game1SettingsSave40AA00(t)
	defer restore()
	start := memmap.PtrUint32(0x5d4594, 3520)
	for _, backend := range []string{"Go", "C"} {
		for _, control := range []string{"frame-before-start", "start-before-FPS"} {
			t.Run(backend+"/"+control, func(t *testing.T) {
				window := game1SettingsBlobC40AA00(3468, 80)
				before := bytes.Clone(window)
				var order []string
				*start = 0
				if control == "frame-before-start" {
					*start = 0xffffffff
				}
				gameFrameHook = func() uint32 {
					order = append(order, "frame")
					if control == "frame-before-start" {
						*start = 0
					}
					return 200
				}
				gameFPSHook = func() uint32 {
					order = append(order, "FPS")
					if control == "start-before-FPS" {
						*start = 0xffffffff
					}
					return 10
				}
				var got uint32
				if backend == "Go" {
					got = uint32(sub40AA00Native())
				} else {
					got = game1SettingsCallC40AA00(0x40aa00, 0)
				}
				wantedStart := uint32(0)
				if control == "start-before-FPS" {
					wantedStart = 0xffffffff
				}
				if got != 0 || strings.Join(order, ",") != "frame,FPS" || *start != wantedStart {
					t.Fatalf("static PE read order backend=%s control=%s actual=%08x order=%v start=%08x", backend, control, got, order, *start)
				}
				copy(window[52:56], before[52:56])
				if !bytes.Equal(window, before) {
					t.Fatal("PE-order control changed bytes outside input start DWORD")
				}
				t.Logf("GAME1_SETTINGS_SCALAR_PE_ORDER_OK backend=%s control=%s static-PE-only=true portable-C-order-claim=false", backend, control)
			})
		}
	}
}

func TestGame1SettingsScalarLiveAliases40AA00(t *testing.T) {
	restore := game1SettingsSave40AA00(t)
	defer restore()
	for _, v := range []uint32{0x80000001, 0xffffffff} {
		if uint32(sub40AA30Native(int32(v))) != v || game1SettingsCallC40AA00(0x40aa20, 0) != v {
			t.Fatal("Go setter immediately visible to actual C getter")
		}
		v ^= 0x01020304
		if game1SettingsCallC40AA00(0x40aa30, v) != v || uint32(sub40AA20Native()) != v {
			t.Fatal("actual C setter immediately visible to Go getter")
		}
	}
	for _, backend := range []string{"Go", "C"} {
		const v = uint32(0x8badf00d)
		if backend == "Go" {
			if uint32(sub40AA60Native(int32(-1951535091))) != v {
				t.Fatal("raw Go setter return")
			}
		} else if game1SettingsCallC40AA00(0x40aa60, v) != v {
			t.Fatal("raw actual C setter return")
		}
		if *memmap.PtrUint32(0x5d4594, 3508) != v || *(*uint32)(game1SettingsCellC40AA00(3508)) != v {
			t.Fatal("live cell setter visibility")
		}
	}
	t.Log("GAME1_SETTINGS_SCALAR_LIVE_ALIAS_OK setter-getter-crossings=4 setter-cell-checks=2 raw-signed-DWORD=true")
}
