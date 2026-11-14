package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/soundset_original_424320_424350.json.gz
var soundSetOriginalNative []byte

//go:embed testdata/npc_voice_wire_original_52ad10.json.gz
var npcVoiceWireOriginalNative []byte

type soundSetOriginalCase struct {
	Kind          string   `json:"kind"`
	Present       bool     `json:"present"`
	Class         uint32   `json:"cls"`
	UpdatePresent bool     `json:"update_present"`
	NewRole       *string  `json:"new_role"`
	OldRole       *string  `json:"old_role"`
	NamesHex      []string `json:"names_hex"`
	QueryHex      *string  `json:"query_hex"`
	ResultIndex   *int     `json:"result_index"`
	Result        int32    `json:"result"`
	Before        string   `json:"before"`
	After         string   `json:"after"`
}

type npcVoiceOriginalTrace struct {
	Length int     `json:"length"`
	Data   string  `json:"data_hex"`
	Result uintptr `json:"result"`
}

type npcVoiceWireOriginalCase struct {
	Mode           uint32                  `json:"mode"`
	Present        bool                    `json:"present"`
	Class          uint32                  `json:"cls"`
	NamesHex       []string                `json:"names_hex"`
	VoiceIndex     *int                    `json:"voice_index"`
	InputHex       string                  `json:"input_hex"`
	ServiceReturns []uintptr               `json:"service_returns"`
	Result         uintptr                 `json:"result"`
	AfterIndex     *int                    `json:"after_index"`
	Trace          []npcVoiceOriginalTrace `json:"trace"`
	OutputHex      string                  `json:"output_hex"`
}

func npcVoiceOriginalDecode(t *testing.T, compressed []byte, zippedSHA, rawSHA string, count int, dest any) {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(compressed)); got != zippedSHA {
		t.Fatal("sealed original compressed fixture changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != rawSHA {
		t.Fatal("sealed original raw fixture changed", got)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatal(err)
	}
	if got := reflect.ValueOf(dest).Elem().Len(); got != count {
		t.Fatal("sealed original case count", got, count)
	}
}

func npcVoiceUnhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type npcVoiceTestOwners struct {
	object, update, data []byte
	start                int
	obj                  *server.Object
	ud                   *server.MonsterUpdateData
}

func npcVoiceOwners(t *testing.T, placement int) *npcVoiceTestOwners {
	t.Helper()
	v := &npcVoiceTestOwners{start: 32 + placement}
	for _, entry := range []struct {
		out *[]byte
		n   int
	}{{&v.object, int(unsafe.Sizeof(server.Object{})) + 128}, {&v.update, int(unsafe.Sizeof(server.MonsterUpdateData{})) + 128}, {&v.data, 8192}} {
		b, free := alloc.MakeHigh([]byte{}, entry.n)
		t.Cleanup(free)
		*entry.out = b
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
			t.Fatal("real C input must be foreign memory above 4GiB")
		}
	}
	v.obj = (*server.Object)(unsafe.Pointer(&v.object[v.start]))
	v.ud = (*server.MonsterUpdateData)(unsafe.Pointer(&v.update[v.start]))
	return v
}

func (v *npcVoiceTestOwners) reset(class uint32, updatePresent bool) {
	for _, b := range [][]byte{v.object, v.update, v.data} {
		for i := range b {
			b[i] = 0xa5
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		// The old PE32 class and update offsets are unrelated native fields.
		binary.LittleEndian.PutUint32(v.object[v.start+8:], class^2)
		binary.LittleEndian.PutUint32(v.object[v.start+748:], 0xdeadface)
	}
	v.obj.ObjClass = object.Class(class)
	v.obj.UpdateData = nil
	if updatePresent {
		v.obj.UpdateData = unsafe.Pointer(v.ud)
	}
}

func (v *npcVoiceTestOwners) sets(t *testing.T, names []string) []unsafe.Pointer {
	t.Helper()
	soundSetHead = nil
	soundSetByName = make(map[string]*soundSetRecord)
	ptrs := make([]unsafe.Pointer, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		ptrs[i] = unsafe.Pointer(&v.data[v.start+i*128])
		rec := &soundSetRecord{name: string(npcVoiceUnhex(t, names[i])), data: ptrs[i], next: soundSetHead}
		soundSetHead = rec
		soundSetByName[rec.name] = rec
	}
	return ptrs
}

func npcVoicePreserveRegistry(t *testing.T) {
	t.Helper()
	head, names := soundSetHead, soundSetByName
	t.Cleanup(func() { soundSetHead, soundSetByName = head, names })
}

func npcVoiceRole(role *string, v *npcVoiceTestOwners) unsafe.Pointer {
	if role == nil {
		return nil
	}
	switch *role {
	case "node0":
		return unsafe.Pointer(&v.data[v.start])
	case "node1":
		return unsafe.Pointer(&v.data[v.start+128])
	case "object":
		return unsafe.Pointer(v.obj)
	case "update":
		return unsafe.Pointer(v.ud)
	case "raw_deadface":
		return unsafe.Pointer(uintptr(0xdeadface))
	case "raw_ffffffff":
		return unsafe.Pointer(uintptr(0xffffffff))
	default:
		panic("unknown sealed pointer role")
	}
}

func npcVoiceIndex(index *int, pointers []unsafe.Pointer) unsafe.Pointer {
	if index == nil {
		return nil
	}
	return pointers[*index]
}

func TestNPCVoiceNativeLayoutCABI(t *testing.T) {
	size, class, update := npcVoiceLayoutC()
	if size != unsafe.Offsetof(server.Object{}.ScriptPickup)+unsafe.Sizeof(server.Object{}.ScriptPickup) || class != unsafe.Offsetof(server.Object{}.ObjClass) || update != unsafe.Offsetof(server.Object{}.UpdateData) {
		t.Fatal("C and Go native object prefix disagree", size, class, update)
	}
	want := [4]uintptr{772, 8, 748, 488}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = [4]uintptr{912, 12, 872, 568}
	}
	got := [4]uintptr{size, class, update, unsafe.Offsetof(server.MonsterUpdateData{}.SoundSet122)}
	if got != want {
		t.Fatal("independent native layout expectations disagree", got, want)
	}
}

func TestNPCVoiceNativeOriginalCABI(t *testing.T) {
	npcVoicePreserveRegistry(t)
	var cases []soundSetOriginalCase
	npcVoiceOriginalDecode(t, soundSetOriginalNative, "de77f54c7e86f5f64bec2009f1076c6171d6d451ddeaa432d45596041f37e2d0", "8c6987d94a36b3801c95dccffe3b9a9e6f84a2cc4bd9e1b2cba6c7aa07371112", 3467, &cases)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := npcVoiceOwners(t, placement)
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					v.reset(c.Class, c.UpdatePresent)
					if c.Kind == "setter" {
						v.ud.SoundSet122 = npcVoiceRole(c.OldRole, v)
						before, after := npcVoiceUnhex(t, c.Before), npcVoiceUnhex(t, c.After)
						if len(before) != 32 || len(after) != 32 || !bytes.Equal(before[:4], after[:4]) || !bytes.Equal(before[8:], after[8:]) {
							t.Fatal("original setter modified outside its pointer slot")
						}
						wantUpdate := append([]byte(nil), v.update...)
						newData := npcVoiceRole(c.NewRole, v)
						if c.Result == 1 {
							*(*unsafe.Pointer)(unsafe.Pointer(&wantUpdate[v.start+int(unsafe.Offsetof(v.ud.SoundSet122))])) = newData
						}
						wantObj, wantData := append([]byte(nil), v.object...), append([]byte(nil), v.data...)
						var obj unsafe.Pointer
						if c.Present {
							obj = unsafe.Pointer(v.obj)
						}
						if got := npcVoiceSetC(obj, newData); got != c.Result {
							t.Fatal("actual public C setter result", got, c.Result)
						}
						if !bytes.Equal(v.update, wantUpdate) || !bytes.Equal(v.object, wantObj) || !bytes.Equal(v.data, wantData) {
							t.Fatal("C setter lost pointer bits, used old offsets or changed nonoutput bytes")
						}
						want := unsafe.Pointer(nil)
						if c.Present && c.Class&2 != 0 {
							want = newData
						}
						if got := npcVoiceGetC(obj); got != want {
							t.Fatalf("existing public C getter returned %p, want %p", got, want)
						}
					} else if c.Kind == "lookup" {
						ptrs := v.sets(t, c.NamesHex)
						var input unsafe.Pointer
						if c.QueryHex != nil {
							query := npcVoiceUnhex(t, *c.QueryHex)
							copy(v.data[4096+v.start:], query)
							v.data[4096+v.start+len(query)] = 0
							input = unsafe.Pointer(&v.data[4096+v.start])
						}
						want := append([]byte(nil), v.data...)
						if got := npcVoiceLookupC(input); got != npcVoiceIndex(c.ResultIndex, ptrs) {
							t.Fatalf("actual const C lookup=%p, expected index %v", got, c.ResultIndex)
						}
						if !bytes.Equal(want, v.data) {
							t.Fatal("C lookup changed name, packed sound IDs, links or canaries")
						}
					} else {
						t.Fatal("unknown original leaf", c.Kind)
					}
				})
			}
		})
	}
}

func npcVoiceWireCases(t *testing.T) []npcVoiceWireOriginalCase {
	t.Helper()
	var cases []npcVoiceWireOriginalCase
	npcVoiceOriginalDecode(t, npcVoiceWireOriginalNative, "5b7fa2dc5f798c24ab65cb103ecfcc0d6a9e416cfd0b1eb7913fa627ea74aeda", "c52e02a59780b98b6c70ed859d21ba63313dc3a89fc9bf917340bfcfc397a055", 1606, &cases)
	return cases
}

func TestNPCVoiceNativeOriginalWireService(t *testing.T) {
	npcVoicePreserveRegistry(t)
	cases := npcVoiceWireCases(t)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := npcVoiceOwners(t, placement)
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d", i), func(t *testing.T) {
					v.reset(c.Class, true)
					ptrs := v.sets(t, c.NamesHex)
					v.ud.SoundSet122 = npcVoiceIndex(c.VoiceIndex, ptrs)
					wantObj, wantData, wantUpdate := append([]byte(nil), v.object...), append([]byte(nil), v.data...), append([]byte(nil), v.update...)
					*(*unsafe.Pointer)(unsafe.Pointer(&wantUpdate[v.start+int(unsafe.Offsetof(v.ud.SoundSet122))])) = npcVoiceIndex(c.AfterIndex, ptrs)
					var obj *server.Object
					if c.Present {
						obj = v.obj
					}
					input := npcVoiceUnhex(t, c.InputHex)
					cursor := 0
					var trace []npcVoiceOriginalTrace
					var output []byte
					got := npcVoiceWireNative52AD10(obj, c.Mode != 0, func(buf []byte) uintptr {
						if c.Mode != 0 {
							copy(buf, input[cursor:cursor+len(buf)])
							cursor += len(buf)
						} else {
							output = append(output, buf...)
						}
						r := c.ServiceReturns[len(trace)]
						trace = append(trace, npcVoiceOriginalTrace{len(buf), hex.EncodeToString(buf), r})
						return r
					})
					if got != c.Result || !reflect.DeepEqual(trace, c.Trace) || hex.EncodeToString(output) != c.OutputHex || c.Mode != 0 && cursor != len(input) {
						t.Fatalf("wire transcript diverged: result=%d/%d trace=%+v/%+v output=%x/%s", got, c.Result, trace, c.Trace, output, c.OutputHex)
					}
					if !bytes.Equal(v.object, wantObj) || !bytes.Equal(v.data, wantData) || !bytes.Equal(v.update, wantUpdate) {
						t.Fatal("wire changed nonoutput bytes or truncated a native sound pointer")
					}
				})
			}
		})
	}
}

// Independent bitwise CRC, preserving the original per-call complement.
func npcVoiceCRC(crc uint32, p []byte) uint32 {
	for _, b := range p {
		crc ^= uint32(b)
		for bit := 0; bit < 8; bit++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xedb88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}

func TestNPCVoiceNativePublicCryptFileCABI(t *testing.T) {
	npcVoicePreserveRegistry(t)
	cases := npcVoiceWireCases(t)
	// Original full-service cases span NULL, nonMonster, missing/duplicate/
	// empty names, embedded NUL, every length boundary, and nonboolean modes.
	for _, placement := range []int{0, 8, 24} {
		v := npcVoiceOwners(t, placement)
		for i, c := range cases {
			if !reflect.DeepEqual(c.ServiceReturns, []uintptr{1, 1}) {
				continue
			}
			for _, key := range []int{-1, crypt.MapKey} {
				for _, xor := range []bool{false, true} {
					t.Run(fmt.Sprintf("placement-%d/%04d/key=%d/xor=%t", placement, i, key, xor), func(t *testing.T) {
						v.reset(c.Class, true)
						ptrs := v.sets(t, c.NamesHex)
						v.ud.SoundSet122 = npcVoiceIndex(c.VoiceIndex, ptrs)
						wantObj, wantData, wantUpdate := append([]byte(nil), v.object...), append([]byte(nil), v.data...), append([]byte(nil), v.update...)
						*(*unsafe.Pointer)(unsafe.Pointer(&wantUpdate[v.start+int(unsafe.Offsetof(v.ud.SoundSet122))])) = npcVoiceIndex(c.AfterIndex, ptrs)
						var obj unsafe.Pointer
						if c.Present {
							obj = unsafe.Pointer(v.obj)
						}
						path := filepath.Join(t.TempDir(), "owned-npc-voice.bin")
						mode := cryptfile.WriteOnly
						if c.Mode != 0 {
							mode = cryptfile.ReadOnly
							if err := os.WriteFile(path, cryptFileWire236(t, npcVoiceUnhex(t, c.InputHex), key, xor), 0o600); err != nil {
								t.Fatal(err)
							}
						}
						cf := cryptFileGlobal236(t, path, mode, key)
						if c.Mode == 0 && key >= 0 {
							if _, err := cf.File.Write(make([]byte, crypt.Block)); err != nil {
								t.Fatal(err)
							}
						}
						cf.SetXOR(xor)
						wantResult := c.Result
						if c.Mode == 0 && c.Trace[len(c.Trace)-1].Length == 0 {
							wantResult = 0 // actual fread/fwrite item result for zero-size payload
						}
						if got := npcVoiceTransferC(obj); got != wantResult {
							t.Fatal("actual public C wire result", got, wantResult)
						}
						if !bytes.Equal(v.object, wantObj) || !bytes.Equal(v.data, wantData) || !bytes.Equal(v.update, wantUpdate) {
							t.Fatal("public C wire changed nonoutputs or full pointer identity")
						}
						if c.Mode == 0 {
							if key >= 0 {
								if _, err := cf.File.FileFlush(); err != nil {
									t.Fatal(err)
								}
								if err := cf.WriteChecksumAt(0); err != nil {
									t.Fatal(err)
								}
							}
							if err := cf.Close(); err != nil {
								t.Fatal(err)
							}
							wire, err := os.ReadFile(path)
							if err != nil {
								t.Fatal("persisted public C wire", err)
							}
							offset := 0
							if key >= 0 {
								offset = crypt.Block
								if len(wire) < offset {
									t.Fatal("missing persisted encrypted checksum")
								}
								wantCRC := uint32(0xffffffff)
								for _, call := range c.Trace {
									wantCRC = npcVoiceCRC(wantCRC, npcVoiceUnhex(t, call.Data))
								}
								header := append([]byte(nil), wire[:offset]...)
								cipher, err := crypt.NewCipher(key)
								if err != nil {
									t.Fatal(err)
								}
								cipher.Decrypt(header, header)
								if binary.LittleEndian.Uint32(header[:4]) != wantCRC || binary.LittleEndian.Uint32(header[4:]) != 0 {
									t.Fatalf("original-call CRC changed: header=%x want=%08x", header, wantCRC)
								}
							}
							wantWire := cryptFileWire236(t, npcVoiceUnhex(t, c.OutputHex), key, xor)
							if key >= 0 {
								// Binfile.FileFlush appends one raw empty block after
								// flushing the encrypted or XOR payload.
								wantWire = append(wantWire, make([]byte, crypt.Block)...)
							}
							if !bytes.Equal(wire[offset:], wantWire) {
								t.Fatalf("public C voice file differs from original transcript plus container footer: got=%x want=%x", wire[offset:], wantWire)
							}
						}
						if err := cf.Close(); err != nil {
							t.Fatal(err)
						}
					})
				}
			}
		}
	}
}

func TestNPCVoiceNativeNameOwnership(t *testing.T) {
	npcVoicePreserveRegistry(t)
	v := npcVoiceOwners(t, 24)
	v.reset(2, true)
	ptrs := v.sets(t, []string{hex.EncodeToString([]byte("Dup")), hex.EncodeToString([]byte("Dup"))})
	// The first definition is hidden from name lookup, but its live foreign
	// data still owns its original name through the retained Go record list.
	soundSetByName = map[string]*soundSetRecord{"Dup": soundSetHead}
	v.ud.SoundSet122 = ptrs[1]
	runtime.GC()
	var calls [][]byte
	if got := npcVoiceWireNative52AD10(v.obj, false, func(buf []byte) uintptr {
		calls = append(calls, append([]byte(nil), buf...))
		runtime.GC()
		return 1
	}); got != 1 || !reflect.DeepEqual(calls, [][]byte{{3}, []byte("Dup")}) {
		t.Fatal("Go metadata did not retain a hidden live voice name", got, calls)
	}
}
