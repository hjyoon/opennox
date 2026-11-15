package legacy

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/intrusive_list_original_425760_425960_v2.json.gz
var intrusiveListOriginal []byte

type intrusiveListOriginalValue struct {
	Pointer string `json:"pointer"`
	Value   uint32 `json:"value"`
}

type intrusiveListOriginalWrite struct {
	Owner  string                     `json:"owner"`
	Slot   int                        `json:"slot"`
	Stored intrusiveListOriginalValue `json:"stored"`
}

type intrusiveListOriginalCase struct {
	Kind        string                                  `json:"kind"`
	First       *string                                 `json:"first"`
	Second      *string                                 `json:"second"`
	Index       int32                                   `json:"index"`
	Before      map[string][]intrusiveListOriginalValue `json:"graph_before"`
	After       map[string][]intrusiveListOriginalValue `json:"graph_after"`
	ResultType  string                                  `json:"result_type"`
	Result      json.RawMessage                         `json:"result"`
	OwnerBefore string                                  `json:"owner_before_hex"`
	OwnerAfter  string                                  `json:"owner_after_hex"`
	Writes      []intrusiveListOriginalWrite            `json:"writes"`
}

var intrusiveListLabels = []string{"head", "a", "b", "c", "d", "e", "f", "new", "other", "x", "y", "z"}
var intrusiveListModes = map[string]int{
	"clear": 0, "init": 1, "insert_ascending": 2, "insert_descending": 3,
	"first_safe": 4, "next_safe": 5, "at_index": 6, "insert_before": 7,
	"insert_after": 8, "remove": 9, "next": 10, "previous": 11,
}

func intrusiveListCases(t *testing.T) []intrusiveListOriginalCase {
	t.Helper()
	var out []intrusiveListOriginalCase
	npcVoiceOriginalDecode(t, intrusiveListOriginal, "3067c68649caf7ce365d8f25d35ec93311b58d814520585befd59b5fc054735d", "2f15b036f9b37d63180a4132570c1954df6df8f7bf45d0f11996aa5344586792", 723, &out)
	return out
}

type intrusiveListOwners struct {
	raw   []byte
	start int
	ptrs  map[string]unsafe.Pointer
	index map[string]int
}

func intrusiveListOwnersNew(t *testing.T, placement int) *intrusiveListOwners {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 12*128+64)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("actual C list arena must be foreign storage above4GiB")
	}
	v := &intrusiveListOwners{raw: raw, start: 32 + placement, ptrs: make(map[string]unsafe.Pointer), index: make(map[string]int)}
	for i, name := range intrusiveListLabels {
		v.index[name] = i
		v.ptrs[name] = unsafe.Pointer(&raw[v.start+128*i])
	}
	return v
}

func (v *intrusiveListOwners) pointer(name *string) unsafe.Pointer {
	if name == nil {
		return nil
	}
	return v.ptrs[*name]
}

func intrusiveListRawValue(t *testing.T, val intrusiveListOriginalValue, pointers map[string]uint32) uint32 {
	t.Helper()
	if val.Pointer != "" {
		p, ok := pointers[val.Pointer]
		if !ok {
			t.Fatal("unknown original pointer label", val.Pointer)
		}
		return p
	}
	return val.Value
}

func intrusiveListVerifyOriginalOwners(t *testing.T, c intrusiveListOriginalCase) {
	t.Helper()
	before, after := npcVoiceUnhex(t, c.OwnerBefore), npcVoiceUnhex(t, c.OwnerAfter)
	if len(before) != 384 || len(after) != 384 || len(c.Before) != 12 || len(c.After) != 12 {
		t.Fatal("sealed original graph/owner extent")
	}
	addresses := make(map[string]uint32)
	indices := make(map[string]int)
	for i, label := range intrusiveListLabels {
		addresses[label] = 0x21001000 + uint32(32*i)
		indices[label] = i
	}
	for i, label := range intrusiveListLabels {
		for slot := 0; slot < 3; slot++ {
			if binary.LittleEndian.Uint32(before[32*i+4*slot:]) != intrusiveListRawValue(t, c.Before[label][slot], addresses) || binary.LittleEndian.Uint32(after[32*i+4*slot:]) != intrusiveListRawValue(t, c.After[label][slot], addresses) {
				t.Fatal("sealed original logical graph differs from raw owner", label, slot)
			}
		}
	}
	expected := bytes.Clone(before)
	for _, w := range c.Writes {
		if w.Slot < 0 || w.Slot >= 3 {
			t.Fatal("write escaped original list header", w)
		}
		binary.LittleEndian.PutUint32(expected[32*indices[w.Owner]+4*w.Slot:], intrusiveListRawValue(t, w.Stored, addresses))
	}
	if !bytes.Equal(expected, after) {
		t.Fatal("sealed original writes differ from whole original owner")
	}
}

func (v *intrusiveListOwners) writeNative(t *testing.T, data []byte, owner string, slot int, value intrusiveListOriginalValue) {
	t.Helper()
	pos := v.start + 128*v.index[owner] + slot*int(unsafe.Sizeof(uintptr(0)))
	var bits uintptr
	if value.Pointer != "" {
		bits = uintptr(v.ptrs[value.Pointer])
	} else if value.Value != 0 {
		t.Fatal("original pointer write had a scalar payload", owner, slot, value)
	}
	*(*uintptr)(unsafe.Pointer(&data[pos])) = bits
}

func (v *intrusiveListOwners) reset(t *testing.T, c intrusiveListOriginalCase) []byte {
	t.Helper()
	for i := range v.raw {
		v.raw[i] = 0xa5
	}
	width := int(unsafe.Sizeof(uintptr(0)))
	for _, label := range intrusiveListLabels {
		for slot, value := range c.Before[label] {
			if slot == 2 && value.Pointer == "" {
				// Keep a contradictory upper half on64bit. Scalar comparisons
				// read only the original signed32bit key; most operations retain
				// every other byte in this native-width tag slot.
				*(*uint32)(unsafe.Add(v.ptrs[label], 2*width)) = value.Value
			} else {
				v.writeNative(t, v.raw, label, slot, value)
			}
		}
	}
	want := bytes.Clone(v.raw)
	for _, w := range c.Writes {
		v.writeNative(t, want, w.Owner, w.Slot, w.Stored)
	}
	return want
}

func TestIntrusiveListNativeLayoutCABI(t *testing.T) {
	width := unsafe.Sizeof(uintptr(0))
	want := [5]uintptr{3 * width, 0, width, 2 * width, width}
	if got := intrusiveListLayoutC(); got != want {
		t.Fatal("independent public C list layout", got, want)
	}
	if unsafe.Sizeof(intrusiveListNode425760{}) != want[0] || unsafe.Offsetof(intrusiveListNode425760{}.next) != want[1] || unsafe.Offsetof(intrusiveListNode425760{}.prev) != want[2] || unsafe.Offsetof(intrusiveListNode425760{}.tag) != want[3] {
		t.Fatal("Go native pointer links/scalar tag differ from C layout")
	}
}

func TestIntrusiveListNativeOriginalCABI(t *testing.T) {
	cases := intrusiveListCases(t)
	for _, c := range cases {
		intrusiveListVerifyOriginalOwners(t, c)
	}
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := intrusiveListOwnersNew(t, placement)
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					want := v.reset(t, c)
					mode, ok := intrusiveListModes[c.Kind]
					if !ok {
						t.Fatal("unknown original entry", c.Kind)
					}
					got := intrusiveListCallC(mode, v.pointer(c.First), v.pointer(c.Second), c.Index)
					switch c.ResultType {
					case "pointer":
						var result intrusiveListOriginalValue
						if err := json.Unmarshal(c.Result, &result); err != nil {
							t.Fatal(err)
						}
						if result.Value != 0 || got != uintptr(v.ptrs[result.Pointer]) {
							t.Fatal("actual C pointer result", got, result)
						}
					case "int32":
						var result int32
						if err := json.Unmarshal(c.Result, &result); err != nil {
							t.Fatal(err)
						}
						if int32(got) != result {
							t.Fatal("original signed32 skipped count", int32(got), result)
						}
					case "void":
						if string(c.Result) != "null" {
							t.Fatal("void entry must not infer original EAX")
						}
					default:
						t.Fatal("unknown original return ABI", c.ResultType)
					}
					if !bytes.Equal(v.raw, want) {
						t.Fatal("native links, scalar tag or whole-owner canaries differ from original graph")
					}
				})
			}
			runtime.GC()
		})
	}
	t.Log("unchanged original intrusive-list cases=723, instruction bytes=361, declared services=0; whole original PE image immutable; actual C calls/native links/scalar sort keys/whole graph canaries")
}

// These are explicit pre-existing C compatibility guards. Their successful
// behavior is separate from the unchanged original executable's valid graphs.
func TestIntrusiveListNativeCompatibilityCABI(t *testing.T) {
	c := intrusiveListCases(t)[0]
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := intrusiveListOwnersNew(t, placement)
			for _, mode := range []string{"nil-owner", "nil-next", "nil-previous", "nil-links", "before-without-previous"} {
				t.Run(mode, func(t *testing.T) {
					v.reset(t, c)
					head, node := v.ptrs["head"], v.ptrs["new"]
					switch mode {
					case "nil-next":
						*(*unsafe.Pointer)(head) = nil
					case "nil-previous", "before-without-previous":
						*(*unsafe.Pointer)(unsafe.Add(head, unsafe.Sizeof(uintptr(0)))) = nil
					case "nil-links":
						*(*unsafe.Pointer)(head) = nil
						*(*unsafe.Pointer)(unsafe.Add(head, unsafe.Sizeof(uintptr(0)))) = nil
					}
					want := bytes.Clone(v.raw)
					if mode == "before-without-previous" {
						v.writeNative(t, want, "new", 0, intrusiveListOriginalValue{Pointer: "head"})
						v.writeNative(t, want, "new", 1, intrusiveListOriginalValue{})
						v.writeNative(t, want, "head", 1, intrusiveListOriginalValue{Pointer: "new"})
						intrusiveListCallC(7, head, node, 0)
					} else {
						if mode == "nil-owner" {
							head = nil
						}
						intrusiveListCallC(9, head, nil, 0)
					}
					if !bytes.Equal(v.raw, want) {
						t.Fatal("current explicit compatibility guard changed native owner/canaries")
					}
				})
			}
		})
	}
}

func TestIntrusiveListNativeMapListCABI(t *testing.T) {
	width := int(unsafe.Sizeof(uintptr(0)))
	head := Get_nox_common_maplist()
	headBytes := unsafe.Slice((*byte)(head), 3*width)
	oldHead := bytes.Clone(headBytes)
	t.Cleanup(func() { copy(headBytes, oldHead) })
	names := []string{"war03b", "con02a", "012345678901"}
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := intrusiveListOwnersNew(t, placement)
			intrusiveListCallC(0, head, nil, 0)
			if intrusiveListCallC(4, head, nil, 0) != 0 {
				t.Fatal("public C clear did not initialize the actual native map-list owner")
			}
			var payloads [][]byte
			for i, name := range names {
				p := v.ptrs[intrusiveListLabels[i+1]]
				mp := (*Nox_map_list_item)(p)
				*mp = Nox_map_list_item{}
				copy(mp.Name[:], name)
				mp.Field_6, mp.Field_7, mp.Field_8_0, mp.Field_8_1, mp.Field_8_2 = -10-int32(i), 0xdeadface+uint32(i), 0xe1, 0xf2, 0xa55a
				Sub_425770(mp)
				goInitialized := bytes.Clone(unsafe.Slice((*byte)(p), int(unsafe.Sizeof(*mp))))
				if intrusiveListCallC(1, p, nil, 0) != uintptr(p) || !bytes.Equal(goInitialized, unsafe.Slice((*byte)(p), len(goInitialized))) {
					t.Fatal("existing Go wrapper and actual public C initializer disagree")
				}
				payloads = append(payloads, bytes.Clone(goInitialized[3*width:]))
				intrusiveListCallC(7, head, p, 0)
			}
			if got := ConsoleMapNames4D09B0(); !reflect.DeepEqual(got, names) {
				t.Fatal("actual C list operations and native Go map consumer", got, names)
			}
			if intrusiveListCallC(6, head, nil, 1) != uintptr(v.ptrs["b"]) || intrusiveListCallC(11, head, nil, 0) != uintptr(v.ptrs["c"]) {
				t.Fatal("actual map-list index/reverse pointers narrowed")
			}
			intrusiveListCallC(9, v.ptrs["b"], nil, 0)
			if got := ConsoleMapNames4D09B0(); !reflect.DeepEqual(got, []string{names[0], names[2]}) {
				t.Fatal("actual C removal and native map names", got)
			}
			for _, label := range []string{"a", "c"} {
				intrusiveListCallC(9, v.ptrs[label], nil, 0)
			}
			if intrusiveListCallC(4, head, nil, 0) != 0 || len(ConsoleMapNames4D09B0()) != 0 {
				t.Fatal("actual native map list did not empty")
			}
			for i, label := range []string{"a", "b", "c"} {
				p := v.ptrs[label]
				if !bytes.Equal(payloads[i], unsafe.Slice((*byte)(unsafe.Add(p, 3*width)), len(payloads[i]))) {
					t.Fatal("list operations changed bounded map name/data", label)
				}
			}
		})
	}
}

func TestIntrusiveListNativeAbortCABI(t *testing.T) {
	const env = "OPENNOX_INTRUSIVE_LIST_ABORT_CHILD"
	if mode := os.Getenv(env); mode != "" {
		v := intrusiveListOwnersNew(t, 0)
		v.reset(t, intrusiveListCases(t)[0])
		a, b := v.ptrs["head"], v.ptrs["new"]
		op := 7
		switch mode {
		case "direct-CRT-control":
			op = 12 // The fixture's default branch directly calls the same CRT abort.
		case "nil-head":
			a = nil
		case "nil-item":
			b = nil
		case "both-nil":
			a, b = nil, nil
		default:
			t.Fatal("unknown abort child", mode)
		}
		fmt.Fprintln(os.Stderr, "INTRUSIVE LIST BEFORE ABORT:", mode)
		intrusiveListCallC(op, a, b, 0)
		t.Fatal("INTRUSIVE LIST RETURNED FROM ABORT")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS")
	if root == "" {
		root = os.TempDir()
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(root, "intrusive-list-native-abort-")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("retained actual C abort logs:", dir)
	run := func(mode string) ([]byte, *exec.ExitError) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "-test.run=^TestIntrusiveListNativeAbortCABI$", "-test.v")
		command.Env = append(os.Environ(), env+"="+mode)
		out, err := command.CombinedOutput()
		if writeErr := os.WriteFile(filepath.Join(dir, mode+".log"), out, 0600); writeErr != nil {
			t.Fatal(writeErr)
		}
		exit, ok := err.(*exec.ExitError)
		if ctx.Err() != nil || !ok || !bytes.Contains(out, []byte("INTRUSIVE LIST BEFORE ABORT: "+mode)) || bytes.Contains(out, []byte("INTRUSIVE LIST RETURNED FROM ABORT")) {
			t.Fatalf("actual C abort child %s: %v\n%s", mode, err, out)
		}
		return out, exit
	}
	control, controlExit := run("direct-CRT-control")
	for _, mode := range []string{"nil-head", "nil-item", "both-nil"} {
		t.Run(mode, func(t *testing.T) {
			out, exit := run(mode)
			if exit.ExitCode() != controlExit.ExitCode() || (bytes.Contains(control, []byte("SIGABRT")) && !bytes.Contains(out, []byte("SIGABRT"))) {
				t.Fatalf("public C guard differs from direct same-CRT abort: control %v, actual %v\n%s", controlExit, exit, out)
			}
		})
	}
}
