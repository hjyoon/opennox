package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

//go:embed testdata/soundset_parser_original_4240f0_424170_v2.json.gz
var soundSetParserOriginal []byte

type soundSetParserOriginalNode struct {
	NameHex  string   `json:"name_hex"`
	Sounds   []uint32 `json:"sound_words"`
	Reserved uint32   `json:"reserved"`
}

type soundSetParserOriginalCase struct {
	Kind     string                       `json:"kind"`
	Table    [][]json.RawMessage          `json:"table"`
	QueryHex string                       `json:"query_hex"`
	Value    uint32                       `json:"value"`
	Result   int32                        `json:"result"`
	Before   string                       `json:"before"`
	After    string                       `json:"after"`
	Tokens   []string                     `json:"tokens_hex"`
	Open     bool                         `json:"open"`
	Key      int32                        `json:"key_result"`
	Sounds   map[string]uint32            `json:"sounds"`
	Nodes    []soundSetParserOriginalNode `json:"nodes"`
	Trace    []map[string]any             `json:"trace"`
	Consumed int                          `json:"tokens_consumed"`
}

func soundSetParserCases(t *testing.T) []soundSetParserOriginalCase {
	t.Helper()
	var out []soundSetParserOriginalCase
	npcVoiceOriginalDecode(t, soundSetParserOriginal, "ab319c82ad0a2f4d176605df105eb5f06a697d73a2638bac9d7a298b197c682b", "b860a1a8bc97158a96e05ab068afcedb6db7e3b3ec713e90689ca192aa91b558", 1998, &out)
	return out
}

type soundSetParserOwners struct {
	names, query, data    []byte
	start                 int
	oldPacked             []byte
	oldPointers           [19]unsafe.Pointer
	oldCHead, oldBlobHead uint32
}

func soundSetParserOwnersNew(t *testing.T, placement int) *soundSetParserOwners {
	t.Helper()
	v := &soundSetParserOwners{start: 32 + placement}
	for _, out := range []*[]byte{&v.names, &v.query, &v.data} {
		b, free := alloc.MakeHigh([]byte{}, 8192)
		t.Cleanup(free)
		*out = b
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
			t.Fatal("actual public C input must be foreign memory above4GiB")
		}
	}
	npcVoicePreserveRegistry(t)
	v.oldPacked = append([]byte(nil), memmap.Slice(0x587000, soundSetFieldTable-16)[:19*8+32]...)
	for i := range v.oldPointers {
		v.oldPointers[i] = *memmap.PtrPtr(0x587000, soundSetFieldTable+uintptr(i*8))
	}
	v.oldCHead = *soundSetOldHeadC()
	v.oldBlobHead = *memmap.PtrUint32(0x5d4594, 588120)
	t.Cleanup(func() {
		copy(memmap.Slice(0x587000, soundSetFieldTable-16), v.oldPacked)
		for i, p := range v.oldPointers {
			*memmap.PtrPtr(0x587000, soundSetFieldTable+uintptr(i*8)) = p
		}
		*soundSetOldHeadC() = v.oldCHead
		*memmap.PtrUint32(0x5d4594, 588120) = v.oldBlobHead
	})
	return v
}

func (v *soundSetParserOwners) reset(t *testing.T, c soundSetParserOriginalCase) {
	t.Helper()
	for _, b := range [][]byte{v.names, v.query, v.data} {
		for i := range b {
			b[i] = 0xa5
		}
	}
	for i := 0; i < 19; i++ {
		off := soundSetFieldTable + uintptr(i*8)
		*memmap.PtrPtr(0x587000, off) = nil
		*memmap.PtrUint32(0x587000, off+4) = 0xdeadface
		if unsafe.Sizeof(uintptr(0)) == 8 {
			*memmap.PtrUint32(0x587000, off) = 0
		}
	}
	for i, row := range c.Table {
		var nameHex string
		var offset uint32
		if err := json.Unmarshal(row[0], &nameHex); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(row[1], &offset); err != nil {
			t.Fatal(err)
		}
		name := npcVoiceUnhex(t, nameHex)
		start := v.start + i*256
		copy(v.names[start:], name)
		v.names[start+len(name)] = 0
		off := soundSetFieldTable + uintptr(i*8)
		*memmap.PtrPtr(0x587000, off) = unsafe.Pointer(&v.names[start])
		*memmap.PtrUint32(0x587000, off+4) = offset
		// Native name sidecars contradict the old packed pointer load.
		if unsafe.Sizeof(uintptr(0)) == 8 {
			*memmap.PtrUint32(0x587000, off) = 0
		}
	}
	soundSetHead = nil
	soundSetByName = make(map[string]*soundSetRecord)
	*soundSetOldHeadC() = 0xdeadface
	*memmap.PtrUint32(0x5d4594, 588120) = 0xc0ffee
}

func (v *soundSetParserOwners) input(s []byte) unsafe.Pointer {
	copy(v.query[v.start:], s)
	v.query[v.start+len(s)] = 0
	return unsafe.Pointer(&v.query[v.start])
}

func (v *soundSetParserOwners) preserveInputs(t *testing.T) func() {
	t.Helper()
	packed := append([]byte(nil), memmap.Slice(0x587000, soundSetFieldTable-16)[:len(v.oldPacked)]...)
	names, query := append([]byte(nil), v.names...), append([]byte(nil), v.query...)
	var ptrs [19]unsafe.Pointer
	for i := range ptrs {
		ptrs[i] = *memmap.PtrPtr(0x587000, soundSetFieldTable+uintptr(i*8))
	}
	return func() {
		t.Helper()
		if !bytes.Equal(packed, memmap.Slice(0x587000, soundSetFieldTable-16)[:len(packed)]) || !bytes.Equal(names, v.names) || !bytes.Equal(query, v.query) {
			t.Fatal("public C changed table, name/query inputs or canaries")
		}
		for i, want := range ptrs {
			if *memmap.PtrPtr(0x587000, soundSetFieldTable+uintptr(i*8)) != want {
				t.Fatal("public C changed native name sidecar")
			}
		}
		if *soundSetOldHeadC() != 0xdeadface || *memmap.PtrUint32(0x5d4594, 588120) != 0xc0ffee {
			t.Fatal("native metadata overwrote an unrelated old DWORD head")
		}
	}
}

func (v *soundSetParserOwners) allocator(t *testing.T, ptrs *[]unsafe.Pointer) func(uintptr) unsafe.Pointer {
	t.Helper()
	return func(size uintptr) unsafe.Pointer {
		if size != 84 || len(*ptrs) >= 16 {
			t.Fatal("unexpected foreign allocation", size, len(*ptrs))
		}
		start := v.start + 128*len(*ptrs)
		clear(v.data[start : start+84])
		p := unsafe.Pointer(&v.data[start])
		*ptrs = append(*ptrs, p)
		return p
	}
}

func (v *soundSetParserOwners) checkNodes(t *testing.T, want []soundSetParserOriginalNode, ptrs []unsafe.Pointer) {
	t.Helper()
	if len(want) != len(ptrs) {
		t.Fatal("record count", len(ptrs), len(want))
	}
	expected := bytes.Repeat([]byte{0xa5}, len(v.data))
	seen := make(map[string]bool)
	rec := soundSetHead
	for i, node := range want {
		index := len(ptrs) - 1 - i
		start := v.start + 128*index
		clear(expected[start : start+84])
		if len(node.Sounds) != 18 {
			t.Fatal("bad original sound-word count")
		}
		for j, word := range node.Sounds {
			binary.LittleEndian.PutUint32(expected[start+4+j*4:], word)
		}
		binary.LittleEndian.PutUint32(expected[start+80:], node.Reserved)
		name := string(npcVoiceUnhex(t, node.NameHex))
		if rec == nil || rec.name != name || rec.data != ptrs[index] || soundSetName(rec.data) != name {
			t.Fatal("native owned name/order/full pointer differs", i)
		}
		if !seen[name] {
			if getDefaultSoundSet(name) != rec.data {
				t.Fatal("latest duplicate lookup")
			}
			seen[name] = true
		}
		rec = rec.next
	}
	if rec != nil || len(soundSetByName) != len(seen) || !bytes.Equal(v.data, expected) {
		t.Fatal("numeric data, reserved words, Go links or whole-owner canaries differ")
	}
}

func soundSetParserNormalizeTrace(in []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	node := 0
	for _, r := range in {
		v := make(map[string]any, len(r))
		for k, x := range r {
			v[k] = x
		}
		switch v["service"] {
		case "calloc":
			v = map[string]any{"service": "foreign_record", "bytes": 84, "index": node}
			node++
		case "malloc":
			v = map[string]any{"service": "owned_name", "bytes": v["bytes"], "index": node - 1}
		}
		out = append(out, v)
	}
	return out
}

func TestSoundSetParserNativePackedCABI(t *testing.T) {
	if soundSetPackedSizeC() != 84 || soundSetRecordSize != 84 || soundSetFieldTable != 64704 {
		t.Fatal("independent packed record/table ABI")
	}
}

func TestSoundSetParserNativeOriginalCABI(t *testing.T) {
	cases := soundSetParserCases(t)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := soundSetParserOwnersNew(t, placement)
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					v.reset(t, c)
					if c.Kind == "field_setter" {
						input := v.input(npcVoiceUnhex(t, c.QueryHex))
						copy(v.data[v.start:], npcVoiceUnhex(t, c.Before))
						check := v.preserveInputs(t)
						want := append([]byte(nil), v.data...)
						copy(want[v.start:], npcVoiceUnhex(t, c.After))
						if got := soundSetFieldC(unsafe.Pointer(&v.data[v.start]), input, c.Value); got != c.Result {
							t.Fatal("actual public C field result", got, c.Result)
						}
						if !bytes.Equal(want, v.data) {
							t.Fatal("whole84-byte field output/owner differs from original")
						}
						check()
						return
					}
					path := v.input([]byte("soundset.bin"))
					check := v.preserveInputs(t)
					file := unsafe.Pointer(&v.names[7000+v.start])
					var ptrs []unsafe.Pointer
					trace := make([]map[string]any, 0)
					cursor := 0
					allocate := v.allocator(t, &ptrs)
					appendTrace := func(service string, fields map[string]any) {
						fields["service"] = service
						trace = append(trace, fields)
					}
					o := &soundSetParserObserver424170{
						Open: func(p unsafe.Pointer, mode int32) unsafe.Pointer {
							if p != path || mode != 0 {
								t.Fatal("full native path/mode")
							}
							appendTrace("open", map[string]any{"path_hex": hex.EncodeToString([]byte(alloc.GoString((*byte)(p)))), "mode": mode, "result": c.Open})
							if c.Open {
								return file
							}
							return nil
						},
						Key: func(p unsafe.Pointer, key int32) int32 {
							if p != file || key != 5 {
								t.Fatal("full native file/key")
							}
							appendTrace("key", map[string]any{"key": key, "result": c.Key})
							return c.Key
						},
						Token: func(p, out unsafe.Pointer) int32 {
							if p != file {
								t.Fatal("truncated file handle")
							}
							if cursor == len(c.Tokens) {
								appendTrace("token", map[string]any{"token_hex": nil, "result": 0})
								return 0
							}
							token := npcVoiceUnhex(t, c.Tokens[cursor])
							cursor++
							copy(unsafe.Slice((*byte)(out), 256), token)
							*(*byte)(unsafe.Add(out, len(token))) = 0
							appendTrace("token", map[string]any{"token_hex": hex.EncodeToString(token), "result": 1})
							return 1
						},
						Sound: func(p unsafe.Pointer) uint32 {
							name := hex.EncodeToString([]byte(alloc.GoString((*byte)(p))))
							value := c.Sounds[name]
							appendTrace("sound", map[string]any{"name_hex": name, "result": value})
							return value
						},
						Allocate: func(n uintptr) unsafe.Pointer {
							p := allocate(n)
							appendTrace("foreign_record", map[string]any{"bytes": n, "index": len(ptrs) - 1})
							return p
						},
						NameStored: func(name string) {
							appendTrace("owned_name", map[string]any{"bytes": len(name) + 1, "index": len(ptrs) - 1})
						},
						Close: func(p unsafe.Pointer) int32 {
							if p != file {
								t.Fatal("truncated close handle")
							}
							appendTrace("close", map[string]any{"result": 0})
							return 0
						},
					}
					restore := soundSetParserObserve424170(o)
					defer restore()
					if got := soundSetParseC(path); got != c.Result || cursor != c.Consumed {
						t.Fatal("actual public C parser result/cursor", got, c.Result, cursor, c.Consumed)
					}
					got, _ := json.Marshal(trace)
					want, _ := json.Marshal(soundSetParserNormalizeTrace(c.Trace))
					if !bytes.Equal(got, want) {
						t.Fatalf("original service order/results differ; only declared name/link ownership is normalized\ngot %s\nwant%s", got, want)
					}
					v.checkNodes(t, c.Nodes, ptrs)
					check()
				})
			}
		})
	}
	t.Log("unchanged original sound-set cases=1998, instruction bytes=439; field setter services=0; parser declared token services=7; native name/link ownership normalized explicitly")
}

func soundSetParserReleaseFiles(t *testing.T) func() {
	t.Helper()
	files.RLock()
	before := make(map[unsafe.Pointer]bool, len(files.byHandle))
	for p := range files.byHandle {
		before[p] = true
	}
	files.RUnlock()
	return func() {
		files.Lock()
		defer files.Unlock()
		for p, f := range files.byHandle {
			if !before[p] {
				_ = f.Close()
				delete(files.byHandle, p)
			}
		}
	}
}

func soundSetParserFileRegistry(t *testing.T) {
	t.Helper()
	files.Lock()
	old := files.byHandle
	files.byHandle = make(map[unsafe.Pointer]*binfile.File)
	files.Unlock()
	handles.InitHigh()
	t.Cleanup(func() {
		files.Lock()
		owned := files.byHandle
		files.byHandle = old
		files.Unlock()
		for _, f := range owned {
			_ = f.Close()
		}
		handles.Release()
	})
}

func TestSoundSetParserNativePublicEncryptedFileCABI(t *testing.T) {
	soundSetParserFileRegistry(t)
	cases := soundSetParserCases(t)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := soundSetParserOwnersNew(t, placement)
			for i, c := range cases {
				if c.Kind != "parser" || !c.Open || c.Key != 1 {
					continue
				}
				t.Run(fmt.Sprintf("%04d", i), func(t *testing.T) {
					v.reset(t, c)
					tokens := make([]string, len(c.Tokens))
					for j, s := range c.Tokens {
						tokens[j] = string(npcVoiceUnhex(t, s))
					}
					path := filepath.Join(t.TempDir(), "soundset.bin")
					f, err := binfile.BinfileOpen(path, binfile.WriteOnly)
					if err != nil {
						t.Fatal(err)
					}
					if err = f.SetKey(5); err != nil {
						t.Fatal(err)
					}
					if _, err = f.Write([]byte(strings.Join(tokens, " \t\r\n") + "\n")); err != nil {
						t.Fatal(err)
					}
					if err = f.Close(); err != nil {
						t.Fatal(err)
					}
					wire, err := os.ReadFile(path)
					if err != nil || len(wire) == 0 {
						t.Fatal("real encrypted file", err)
					}
					var ptrs []unsafe.Pointer
					o := &soundSetParserObserver424170{Allocate: v.allocator(t, &ptrs), Sound: func(p unsafe.Pointer) uint32 { return c.Sounds[hex.EncodeToString([]byte(alloc.GoString((*byte)(p))))] }}
					restore := soundSetParserObserve424170(o)
					defer restore()
					defer soundSetParserReleaseFiles(t)()
					input := v.input([]byte(path))
					check := v.preserveInputs(t)
					if got := soundSetParseC(input); got != c.Result {
						t.Fatal("real encrypted public C parser result", got, c.Result)
					}
					v.checkNodes(t, c.Nodes, ptrs)
					check()
				})
			}
		})
	}
}

func TestSoundSetParserNativeOwnedNamesGC(t *testing.T) {
	cases := soundSetParserCases(t)
	var selected soundSetParserOriginalCase
	for _, c := range cases {
		if c.Kind == "parser" && c.Open && c.Key == 1 && len(c.Nodes) == 2 && c.Nodes[0].NameHex == c.Nodes[1].NameHex {
			selected = c
			break
		}
	}
	if len(selected.Nodes) != 2 {
		t.Fatal("missing sealed duplicate-name case")
	}
	v := soundSetParserOwnersNew(t, 24)
	v.reset(t, selected)
	path := v.input([]byte("soundset.bin"))
	var ptrs []unsafe.Pointer
	cursor := 0
	o := &soundSetParserObserver424170{Open: func(unsafe.Pointer, int32) unsafe.Pointer { return path }, Key: func(unsafe.Pointer, int32) int32 { return 1 }, Token: func(_, out unsafe.Pointer) int32 {
		if cursor == len(selected.Tokens) {
			return 0
		}
		s := npcVoiceUnhex(t, selected.Tokens[cursor])
		cursor++
		copy(unsafe.Slice((*byte)(out), 256), s)
		*(*byte)(unsafe.Add(out, len(s))) = 0
		return 1
	}, Allocate: v.allocator(t, &ptrs), Sound: func(p unsafe.Pointer) uint32 {
		return selected.Sounds[hex.EncodeToString([]byte(alloc.GoString((*byte)(p))))]
	}, Close: func(unsafe.Pointer) int32 { return 0 }}
	restore := soundSetParserObserve424170(o)
	defer restore()
	if soundSetParseC(path) != selected.Result {
		t.Fatal("duplicate parser result")
	}
	runtime.GC()
	v.checkNodes(t, selected.Nodes, ptrs)
	query := v.input([]byte(soundSetHead.name))
	if npcVoiceLookupC(query) != ptrs[len(ptrs)-1] || soundSetName(ptrs[0]) != soundSetHead.name {
		t.Fatal("public C lookup or hidden duplicate lost its owned name after GC")
	}
}

func TestSoundSetParserNativeDefaultServices(t *testing.T) {
	soundSetParserFileRegistry(t)
	cases := soundSetParserCases(t)
	for _, mode := range []string{"C", "Go"} {
		for _, bad := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/error=%t", mode, bad), func(t *testing.T) {
				v := soundSetParserOwnersNew(t, 24)
				v.reset(t, cases[1470]) // sealed stock field table
				defer soundSetParserReleaseFiles(t)()
				text := "Voice LISTEN_SOUND ConfuseCast END Voice MOVE_SOUND UnknownSoundName END\n"
				wantNames := []string{"Voice", "Voice"}
				if bad {
					text = "Partial LISTEN_SOUND ConfuseCast UnknownField ConfuseCast\n"
					wantNames = []string{"Partial"}
				}
				path := filepath.Join(t.TempDir(), "real-soundset.bin")
				f, err := binfile.BinfileOpen(path, binfile.WriteOnly)
				if err != nil {
					t.Fatal(err)
				}
				if err = f.SetKey(5); err != nil {
					t.Fatal(err)
				}
				if _, err = f.Write([]byte(text)); err != nil {
					t.Fatal(err)
				}
				if err = f.Close(); err != nil {
					t.Fatal(err)
				}
				input := v.input([]byte(path))
				check := v.preserveInputs(t)
				var failed bool
				if mode == "C" {
					failed = soundSetParseC(input) == 0
				} else {
					failed = parseSoundSetBin(path) != nil
				}
				if failed != bad {
					t.Fatal("default parser result", mode, failed, bad)
				}
				check()
				var allocated []unsafe.Pointer
				defer func() {
					for _, p := range allocated {
						alloc.FreePtr(p)
					}
				}()
				rec := soundSetHead
				for i, name := range wantNames {
					if rec == nil || rec.name != name {
						t.Fatal("default owned record order", i)
					}
					allocated = append(allocated, rec.data)
					want := make([]byte, 84)
					if bad || i == 1 {
						binary.LittleEndian.PutUint32(want[4:], 17)
					} // original ConfuseCast ID
					if !bytes.Equal(unsafe.Slice((*byte)(rec.data), 84), want) {
						t.Fatal("default sound service or packed numeric data", mode, i)
					}
					if i == 0 && npcVoiceLookupC(v.input([]byte(name))) != rec.data {
						t.Fatal("public C lookup of default parsed data")
					}
					rec = rec.next
				}
				if rec != nil {
					t.Fatal("extra default parsed records")
				}
			})
		}
	}
}
