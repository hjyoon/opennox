//go:build darwin && arm64

package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type titleDynamicEffect415960 struct {
	Trigger    string `json:"trigger"`
	Occurrence int    `json:"occurrence"`
	Field      string `json:"field"`
	ReadField  string `json:"read_field"`
	Row        int    `json:"row"`
	Role       string `json:"role"`
	Value      uint32 `json:"value"`
}

type titleDynamicInput415960 struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Family    string `json:"family"`
	Kind      string `json:"kind"`
	Query     string `json:"query"`
	Table     uint32 `json:"table"`
	Flag      uint32 `json:"flag"`
	Mask      uint32 `json:"mask"`
	Ready     uint32 `json:"ready"`
	NilPolicy bool   `json:"native_nil_query_policy"`
	Rows      []struct {
		Title, Key string
		Mask       uint32
	} `json:"rows"`
	Effects []titleDynamicEffect415960 `json:"effects"`
}

type titleDynamicService415960 struct {
	Kind     string `json:"kind"`
	Ordinal  int    `json:"ordinal"`
	Query    string `json:"query"`
	Title    string `json:"title"`
	Key      string `json:"key"`
	Reply    string `json:"reply"`
	Result   uint32 `json:"result"`
	Returned bool   `json:"returned"`
}

type titleDynamicResult415960 struct {
	Case   int    `json:"case"`
	CW     uint32 `json:"cw"`
	Status string `json:"status"`
	Result uint32 `json:"result"`
	After  struct {
		Table string `json:"table_hex"`
		Ready uint32 `json:"ready"`
	} `json:"after"`
	Services      []titleDynamicService415960 `json:"services"`
	EffectsUsed   []int                       `json:"effects_used"`
	EffectsUnused []int                       `json:"effects_unused"`
}

func TestTitleTablesOriginalDynamic415960(t *testing.T) {
	// Expected values come only from the independently read actual PE recording.
	// No original producer or native algorithm is imported to construct them.
	inputBytes, err := os.ReadFile("testdata/title_original_dynamic_v1/inputs-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var in struct {
		Roles map[string]struct {
			Address uint32 `json:"address"`
			Kind    string `json:"kind"`
			Text    string `json:"text"`
		} `json:"roles"`
		Cases []titleDynamicInput415960 `json:"cases"`
	}
	if err := json.Unmarshal(inputBytes, &in); err != nil {
		t.Fatal(err)
	}
	resultBytes, err := os.ReadFile("testdata/title_original_dynamic_v1/results-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var ref struct {
		Qualified          bool                       `json:"qualified"`
		InputSHA           string                     `json:"inputs_SHA"`
		OriginalExecutions int                        `json:"original_executions"`
		RootBytes          int                        `json:"root_code_bytes"`
		HelperBytes        int                        `json:"compare_helper_bytes"`
		Rows               []titleDynamicResult415960 `json:"rows"`
	}
	if err := json.Unmarshal(resultBytes, &ref); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(inputBytes))
	if !ref.Qualified || digest != "871a0d1171455847f352405c10fb0277f640ff681eb25adb2ad323e7a15c64b3" || ref.InputSHA != digest || len(in.Cases) != 32 || len(ref.Rows) != 64 || ref.OriginalExecutions != 64 || ref.RootBytes != 406 || ref.HelperBytes != 223 {
		t.Fatal("source-bound independently qualified actual original64 scope")
	}
	for index, expected := range ref.Rows {
		if expected.Case != index/2 || expected.CW != []uint32{0x027f, 0x0e7f}[index%2] {
			t.Fatal("original identity order")
		}
		item := in.Cases[expected.Case]
		for _, backend := range []string{"Go", "C"} {
			t.Run(fmt.Sprintf("%s/CW%04x/%s", item.Name, expected.CW, backend), func(t *testing.T) {
				armor := item.Family == "armor"
				table := titleTestTable415960(armor)
				if item.Table != uint32(0x587000+table.Base) || item.Flag != uint32(0x5D4594+table.Ready) {
					t.Fatal("original/native owner correspondence")
				}
				owners := map[string]unsafe.Pointer{"nil": nil}
				originalRole := map[uint32]string{0: "nil"}
				for name, r := range in.Roles {
					if r.Kind == "wide" {
						owners[name] = titleTestWide415960(t, r.Text)
					} else {
						owners[name] = titleTestKey415960(t, r.Text)
					}
					if uintptr(owners[name]) <= 0xffffffff {
						t.Fatal("actual >4GiB role owner required")
					}
					originalRole[r.Address] = name
				}
				rows := make([]titleTestRow415960, len(item.Rows))
				for i, r := range item.Rows {
					rows[i] = titleTestRow415960{owners[r.Title], owners[r.Key], r.Mask}
				}
				packed := titleTestInstall415960(t, table, rows)
				for i := 0; i < 4; i++ {
					v := uint32(0)
					if i < len(item.Rows) {
						v = item.Rows[i].Mask
					}
					*memmap.PtrUint32(0x587000, table.Base+uintptr(12*i+8)) = v
				}
				*memmap.PtrUint32(0x5D4594, table.Ready) = item.Ready
				// Expected physical canaries are input-only; never clone actual BEFORE.
				physicalWant := make([]byte, len(packed))
				for i := range physicalWant {
					physicalWant[i] = byte(0xa1 + i%19)
				}
				for i := 0; i < 4; i++ {
					v := uint32(0)
					if i < len(item.Rows) {
						v = item.Rows[i].Mask
					}
					binary.LittleEndian.PutUint32(physicalWant[12*i+8:12*i+12], v)
				}
				if !bytes.Equal(packed, physicalWant) {
					t.Fatal("independent physical BEFORE canaries")
				}
				used := make([]int, 0)
				service := 0
				reads := map[uintptr]int{}
				fieldOff := func(field string, row int) uintptr {
					if field == "flag" {
						return table.Ready
					}
					return table.Base + uintptr(12*row) + map[string]uintptr{"title": 0, "key": 4, "mask": 8}[field]
				}
				apply := func(trigger string, occurrence int, readOff uintptr) {
					for i, effect := range item.Effects {
						already := false
						for _, v := range used {
							if v == i {
								already = true
							}
						}
						if already || effect.Trigger != trigger || effect.Occurrence != occurrence {
							continue
						}
						if trigger == "after_root_read" && fieldOff(effect.ReadField, effect.Row) != readOff {
							continue
						}
						off := fieldOff(effect.Field, effect.Row)
						switch effect.Field {
						case "title", "key":
							*memmap.PtrPtr(0x587000, off) = owners[effect.Role]
						case "mask":
							*memmap.PtrUint32(0x587000, off) = effect.Value
						case "flag":
							*memmap.PtrUint32(0x5D4594, off) = effect.Value
						default:
							t.Fatal("unknown input effect")
						}
						used = append(used, i)
					}
				}
				accesses := 0
				hook := &titleObserver415960{Access: func(e titleAccess415960) {
					accesses++
					if !e.Write && e.Complete {
						reads[e.Offset]++
						apply("after_root_read", reads[e.Offset], e.Offset)
					}
				}}
				hook.Compare = func(query, title unsafe.Pointer) (int32, bool) {
					if service >= len(expected.Services) {
						t.Fatal("extra native compare")
					}
					s := expected.Services[service]
					service++
					if s.Kind != "compare" || !s.Returned || query != owners[s.Query] || title != owners[s.Title] {
						t.Fatal("actual original comparison argument identities")
					}
					value := titleDefaultCompare415960(query, title)
					if uint32(value) != s.Result {
						t.Fatal("actual native comparison differs from actual original C-locale helper")
					}
					apply("after_compare", s.Ordinal, 0)
					return value, true
				}
				hook.Load = func(key unsafe.Pointer, source string, line int32) (unsafe.Pointer, bool) {
					if service >= len(expected.Services) {
						t.Fatal("extra native strman call")
					}
					s := expected.Services[service]
					service++
					if s.Kind != "load" || !s.Returned || key != owners[s.Key] || source != table.Source || line != table.Line {
						t.Fatal("actual original strman boundary")
					}
					apply("after_load", s.Ordinal, 0)
					return owners[s.Reply], true
				}
				pop := titlePushObserver415960(hook)
				defer pop()
				var word uint32
				var pointer unsafe.Pointer
				var opaque uintptr
				switch item.Kind {
				case "to":
					word = titleTestTo415960(backend, armor, owners[item.Query])
				case "from":
					pointer = titleTestFrom415960(backend, armor, item.Mask)
				case "load":
					opaque = titleTestLoad415960(backend, armor)
				default:
					t.Fatal("unknown entry kind")
				}
				if item.NilPolicy {
					if item.Kind != "to" || item.Query != "nil" || word != 0 || accesses != 0 || service != 0 {
						t.Fatal("explicit native nil-query compatibility policy")
					}
					// Original nonempty-table nil-read fault is recorded and retained,
					// but intentionally has no native fault/equivalence credit.
				} else {
					if expected.Status != "returned" || service != len(expected.Services) {
						t.Fatal("incomplete original/native service closure")
					}
					switch item.Kind {
					case "to":
						if word != expected.Result {
							t.Fatal("actual original scalar return")
						}
					case "from":
						name, ok := originalRole[expected.Result]
						if !ok || pointer != owners[name] {
							t.Fatal("actual original full native pointer return")
						}
					case "load":
						if expected.Result >= item.Table+4 && expected.Result < item.Table+48 && (expected.Result-item.Table-4)%12 == 0 {
							off := table.Base + uintptr(expected.Result-item.Table)
							if opaque != uintptr(unsafe.Pointer(memmap.PtrPtr(0x587000, off))) {
								t.Fatal("actual original terminal cell mapped to genuine full native side-slot")
							}
						} else if opaque != uintptr(expected.Result) {
							t.Fatal("actual original mixed scalar loader return")
						}
					}
				}
				sort.Ints(used)
				unused := make([]int, 0)
				for i := range item.Effects {
					found := false
					for _, v := range used {
						if v == i {
							found = true
						}
					}
					if !found {
						unused = append(unused, i)
					}
				}
				if !reflect.DeepEqual(used, expected.EffectsUsed) || !reflect.DeepEqual(unused, expected.EffectsUnused) {
					t.Fatal("actual original effect usage accounting")
				}
				after, err := hex.DecodeString(expected.After.Table)
				if err != nil || len(after) != 48 {
					t.Fatal("original whole field projection")
				}
				for i := 0; i < 4; i++ {
					for _, f := range []uintptr{0, 4} {
						v := binary.LittleEndian.Uint32(after[12*i+int(f) : 12*i+int(f)+4])
						name, ok := originalRole[v]
						if !ok || *memmap.PtrPtr(0x587000, table.Base+uintptr(12*i)+f) != owners[name] {
							t.Fatal("actual original field -> full native owner projection")
						}
					}
					v := binary.LittleEndian.Uint32(after[12*i+8 : 12*i+12])
					if *memmap.PtrUint32(0x587000, table.Base+uintptr(12*i+8)) != v {
						t.Fatal("actual original DWORD field projection")
					}
					binary.LittleEndian.PutUint32(physicalWant[12*i+8:12*i+12], v)
				}
				if !bytes.Equal(packed, physicalWant) || *memmap.PtrUint32(0x5D4594, table.Ready) != expected.After.Ready {
					t.Fatal("whole physical packed canary/DWORD projection")
				}
				guard := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, table.Ready-4)), 12)
				guardWant := make([]byte, 12)
				for i := range guardWant {
					guardWant[i] = byte(0x61 + i)
				}
				binary.LittleEndian.PutUint32(guardWant[4:8], expected.After.Ready)
				if !bytes.Equal(guard, guardWant) {
					t.Fatal("actual ready DWORD neighbor canaries")
				}
			})
		}
	}
}
