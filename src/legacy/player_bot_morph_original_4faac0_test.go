package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerBotMorphRow4FAAC0 struct {
	Input struct {
		Name     string `json:"name"`
		Op       string `json:"op"`
		Class    uint32 `json:"class"`
		Subclass uint32 `json:"subclass"`
		Update   string `json:"update"`
		Bot      string `json:"bot"`
		Player   string `json:"player"`
		NilUnit  bool   `json:"nil_unit"`
	} `json:"input"`
	Output struct {
		Class    uint32           `json:"class"`
		Subclass uint32           `json:"subclass"`
		Update   string           `json:"update"`
		Result   byte             `json:"result_AL"`
		Record   string           `json:"whole_unit_hex"`
		Heap     bool             `json:"all_other_heap_bytes_unchanged"`
		Stack    bool             `json:"whole_caller_stack_unchanged"`
		CW       bool             `json:"CW_TOP_preserved"`
		Fault    any              `json:"exception"`
		Invalid  []map[string]any `json:"invalid"`
	} `json:"output"`
}

func playerBotMorphLoad4FAAC0(t *testing.T) ([]playerBotMorphRow4FAAC0, []playerBotMorphRow4FAAC0) {
	t.Helper()
	file := filepath.Join("..", "server", "testdata", "player_bot_morph_original_4faac0.json.gz")
	encoded, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(encoded)) != "40042d3c0a2ad643822eca593c7f097148880a072f5e2ee1b86815652479430e" {
		t.Fatal("original gzip seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "ed682e384b7169a7e275e23fc325d4f64cbcd8cd2cd43f0ed40d6d6411d9d134" {
		t.Fatal("original decoded seal")
	}
	var data struct {
		Contexts []playerBotMorphRow4FAAC0 `json:"contexts"`
		Faults   []playerBotMorphRow4FAAC0 `json:"faults"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 4352 || len(data.Faults) != 4 {
		t.Fatal("original input dimensions")
	}
	for _, row := range append(append([]playerBotMorphRow4FAAC0{}, data.Contexts...), data.Faults...) {
		record, err := hex.DecodeString(row.Output.Record)
		if err != nil || len(record) != 1024 || !row.Output.Heap || !row.Output.Stack || !row.Output.CW {
			t.Fatal("original whole-record/heap/stack/FPU observations", row.Input.Name)
		}
		if binary.LittleEndian.Uint32(record[8:]) != row.Output.Class || binary.LittleEndian.Uint32(record[12:]) != row.Output.Subclass {
			t.Fatal("original class/subclass projection", row.Input.Name)
		}
	}
	return data.Contexts, data.Faults
}

type playerBotMorphGo4FAAC0 interface {
	MorphPlayerBotFromPlayer4FAAC0() byte
	MorphPlayerBotToPlayer4FAAF0() byte
}

func playerBotMorphCall4FAAC0(path, op string, obj *server.Object) byte {
	if path == "C" {
		return playerBotMorphOriginalC4FAAC0(obj, op == "FromPlayer")
	}
	native := any(obj).(playerBotMorphGo4FAAC0)
	if op == "FromPlayer" {
		return native.MorphPlayerBotFromPlayer4FAAC0()
	}
	return native.MorphPlayerBotToPlayer4FAAF0()
}

type playerBotMorphUnitGuard4FAAC0 struct {
	Before [32]byte
	Unit   server.Object
	After  [32]byte
}
type playerBotMorphPlayerGuard4FAAC0 struct {
	Before [32]byte
	Player server.PlayerUpdateData
	After  [32]byte
}
type playerBotMorphMonsterGuard4FAAC0 struct {
	Before  [32]byte
	Monster server.MonsterUpdateData
	After   [32]byte
}

func TestPlayerBotMorphOriginalNative4FAAC0(t *testing.T) {
	rows, _ := playerBotMorphLoad4FAAC0(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					ug, free := alloc.NewHigh(playerBotMorphUnitGuard4FAAC0{})
					defer free()
					pg, free := alloc.NewHigh(playerBotMorphPlayerGuard4FAAC0{})
					defer free()
					mg, free := alloc.NewHigh(playerBotMorphMonsterGuard4FAAC0{})
					defer free()
					ub := unsafe.Slice((*byte)(unsafe.Pointer(ug)), unsafe.Sizeof(*ug))
					pb := unsafe.Slice((*byte)(unsafe.Pointer(pg)), unsafe.Sizeof(*pg))
					mb := unsafe.Slice((*byte)(unsafe.Pointer(mg)), unsafe.Sizeof(*mg))
					for _, memory := range [][]byte{ub, pb, mb} {
						for i := range memory {
							memory[i] = 0xa5
						}
					}
					unit, pud, mud := &ug.Unit, &pg.Player, &mg.Monster
					ptrs := map[string]unsafe.Pointer{"Nil": nil, "Player": unsafe.Pointer(pud), "Monster": unsafe.Pointer(mud)}
					if unsafe.Sizeof(uintptr(0)) == 8 {
						for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(pud), unsafe.Pointer(mud)} {
							if uintptr(ptr) <= math.MaxUint32 {
								t.Fatal("native record must exceed4GiB")
							}
						}
					}
					unit.ObjClass = object.Class(row.Input.Class)
					unit.ObjSubClass = object.SubClass(row.Input.Subclass)
					unit.UpdateData = ptrs[row.Input.Update]
					pud.Field73 = (*server.MonsterUpdateData)(ptrs[row.Input.Bot])
					mud.Field545 = (*server.PlayerUpdateData)(ptrs[row.Input.Player])
					wantUnit, wantPlayer, wantMonster := bytes.Clone(ub), bytes.Clone(pb), bytes.Clone(mb)
					putWord := func(off uintptr, value uint32) { binary.LittleEndian.PutUint32(wantUnit[off:], value) }
					base := unsafe.Offsetof(ug.Unit)
					putWord(base+unsafe.Offsetof(unit.ObjClass), row.Output.Class)
					putWord(base+unsafe.Offsetof(unit.ObjSubClass), row.Output.Subclass)
					ptrOffset := base + unsafe.Offsetof(unit.UpdateData)
					if unsafe.Sizeof(uintptr(0)) == 8 {
						binary.LittleEndian.PutUint64(wantUnit[ptrOffset:], uint64(uintptr(ptrs[row.Output.Update])))
					} else {
						putWord(ptrOffset, uint32(uintptr(ptrs[row.Output.Update])))
					}
					t.Logf("NATIVE RECORDS unit=%p player=%p monster=%p calling %s %s", unit, pud, mud, path, row.Input.Op)
					if got := playerBotMorphCall4FAAC0(path, row.Input.Op, unit); got != row.Output.Result {
						t.Fatalf("AL=%02x original=%02x", got, row.Output.Result)
					}
					if !bytes.Equal(ub, wantUnit) || !bytes.Equal(pb, wantPlayer) || !bytes.Equal(mb, wantMonster) {
						t.Fatal("whole native records/neighbor guards differ from original effects")
					}
				})
			}
		})
	}
}

func TestPlayerBotMorphOriginalRawFaults4FAAC0(t *testing.T) {
	_, faults := playerBotMorphLoad4FAAC0(t)
	if child := os.Getenv("NOX_BOT_MORPH_RAW_FAULT_CHILD"); child != "" {
		parts := strings.SplitN(child, "/", 2)
		for _, row := range faults {
			if row.Input.Name != parts[1] {
				continue
			}
			obj, free := alloc.NewHigh(server.Object{})
			defer free()
			obj.ObjClass = object.Class(row.Input.Class)
			obj.ObjSubClass = object.SubClass(row.Input.Subclass)
			if row.Input.NilUnit {
				obj = nil
			}
			t.Log("CALL RAW BOT MORPH FAULT " + child)
			playerBotMorphCall4FAAC0(parts[0], row.Input.Op, obj)
			t.Fatal("raw invalid input unexpectedly returned")
		}
		t.Fatal("unknown fault child")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"C", "Go"} {
		for _, row := range faults {
			t.Run(path+"/"+row.Input.Name, func(t *testing.T) {
				if row.Output.Fault == nil || len(row.Output.Invalid) == 0 || row.Output.Class != row.Input.Class || row.Output.Subclass != row.Input.Subclass || row.Output.Update != row.Input.Update {
					t.Fatal("original fault-before-store contract")
				}
				child := path + "/" + row.Input.Name
				cmd := exec.Command(exe, "-test.run=^TestPlayerBotMorphOriginalRawFaults4FAAC0$", "-test.v")
				cmd.Env = append(os.Environ(), "NOX_BOT_MORPH_RAW_FAULT_CHILD="+child)
				output, err := cmd.CombinedOutput()
				if err == nil || !strings.Contains(string(output), "CALL RAW BOT MORPH FAULT "+child) || !strings.Contains(string(output), "invalid memory address") {
					t.Fatalf("raw fault: %v\n%s", err, output)
				}
			})
		}
	}
}
