package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/script_raise_zombie_original_v1.json.gz
var scriptRaiseZombieOriginal []byte

type scriptRaiseZombieState struct {
	Class           uint32 `json:"cls"`
	Flags           uint32
	Update          int
	Status          uint32
	AlternateStatus uint32 `json:"alternate_status"`
}

type scriptRaiseZombieEvent struct {
	Service string
	Unit    int
	Before  scriptRaiseZombieState
	After   scriptRaiseZombieState
	Result  uint32
}

type scriptRaiseZombieCase struct {
	Input struct {
		Name              string
		Present           bool
		Class             uint32 `json:"class"`
		Flags             uint32
		Status            uint32
		AlternateStatus   uint32 `json:"alternate_status"`
		Update            int
		ChildReturn       uint32 `json:"child_return"`
		Mutation          string
		ChildClass        uint32 `json:"child_class"`
		ChildFlags        uint32 `json:"child_flags"`
		ChildStatusXor    uint32 `json:"child_status_xor"`
		ChildAlternateXor uint32 `json:"child_alternate_xor"`
		Seed              byte
	}
	OriginalFault bool                     `json:"original_fault"`
	ReturnKind    string                   `json:"return_kind"`
	Return        *uint32                  `json:"return_u32"`
	Events        []scriptRaiseZombieEvent `json:"events"`
	Final         scriptRaiseZombieState
	Preserved     bool `json:"all_nonoutput_owner_bytes_preserved"`
}

func scriptRaiseZombieCases(t *testing.T) []scriptRaiseZombieCase {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(scriptRaiseZombieOriginal)); got != "f25393f311f5e1e3d7fb91517b3b79c325ed7314f7d7c80b95846be09e2e957e" {
		t.Fatal("sealed original compressed fixture", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(scriptRaiseZombieOriginal))
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
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "f60ec6bf1d74e2d344b942df0b3334adbc0e68c5acccc1e43b0c7f9f04fb5f9f" {
		t.Fatal("sealed original uncompressed fixture", got)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		RootBytes   int    `json:"root_bytes"`
		Cases       int
		Rows        []scriptRaiseZombieCase
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.RootBytes != 54 || f.Cases != 296 || len(f.Rows) != f.Cases {
		t.Fatal("original entry provenance")
	}
	return f.Rows
}

func TestScriptRaiseZombieNativeLayout(t *testing.T) {
	var unit server.Object
	prefix := unsafe.Offsetof(unit.ScriptPickup) + unsafe.Sizeof(unit.ScriptPickup)
	want := [6]uintptr{unsafe.Sizeof(uintptr(0)), prefix, unsafe.Offsetof(unit.ObjClass), unsafe.Offsetof(unit.ObjFlags), unsafe.Offsetof(unit.UpdateData), unsafe.Sizeof(uintptr(0))}
	if got := scriptRaiseZombieNativeLayout(); got != want || unsafe.Sizeof(unit) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("native C/Go Object prefix and mixed return layout", got, want)
	}
}

func scriptRaiseZombieRun(t *testing.T, c scriptRaiseZombieCase, route int, unit *server.Object, updates [2]*server.MonsterUpdateData, whole []byte, unitOffset uintptr, updateOffsets [2]uintptr) {
	t.Helper()
	i := c.Input
	if !c.Preserved || i.ChildReturn == 31 {
		t.Fatal("original canaries or declared early-return child boundary")
	}
	*unit = server.Object{ObjClass: object.Class(i.Class), ObjFlags: object.Flags(i.Flags)}
	*updates[0] = server.MonsterUpdateData{StatusFlags: object.MonsterStatus(i.Status)}
	*updates[1] = server.MonsterUpdateData{StatusFlags: object.MonsterStatus(i.AlternateStatus)}
	dataPointer := func(kind int) unsafe.Pointer {
		switch kind {
		case 0:
			return nil
		case 1, 2:
			return unsafe.Pointer(updates[kind-1])
		default:
			t.Fatal("fixture update owner", kind)
			return nil
		}
	}
	unit.UpdateData = dataPointer(i.Update)
	state := func() scriptRaiseZombieState {
		kind := -1
		for k := 0; k < 3; k++ {
			if unit.UpdateData == dataPointer(k) {
				kind = k
			}
		}
		if kind < 0 {
			t.Fatal("foreign update owner escaped the declared domain")
		}
		return scriptRaiseZombieState{uint32(unit.ObjClass), uint32(unit.ObjFlags), kind, uint32(updates[0].StatusFlags), uint32(updates[1].StatusFlags)}
	}
	events := []scriptRaiseZombieEvent{}
	old := zombieEntryRaise534AB0
	defer func() { zombieEntryRaise534AB0 = old }()
	currentActions := 0
	zombieEntryRaise534AB0 = func() server.MonsterRaiseZombieRawRuntime534AB0 {
		return server.MonsterRaiseZombieRawRuntime534AB0{
			IsZombie: func(got *server.Object) uint32 {
				if got != unit {
					t.Fatal("native child received a different Object pointer")
				}
				before := state()
				if i.Mutation == "flags" || i.Mutation == "both" {
					unit.ObjClass = object.Class(i.ChildClass)
					unit.ObjFlags = object.Flags(i.ChildFlags)
				}
				if i.Mutation == "update" || i.Mutation == "both" {
					unit.UpdateData = unsafe.Pointer(updates[1])
					updates[0].StatusFlags ^= object.MonsterStatus(i.ChildStatusXor)
					updates[1].StatusFlags ^= object.MonsterStatus(i.ChildAlternateXor)
				}
				events = append(events, scriptRaiseZombieEvent{"raise534AB0", 1, before, state(), i.ChildReturn})
				if i.ChildReturn == 0 {
					return 0
				}
				return 1
			},
			CurrentAction: func(got *server.Object) uint32 {
				if got != unit {
					t.Fatal("native child action received a different Object pointer")
				}
				currentActions++
				return i.ChildReturn
			},
		}
	}
	before := bytes.Clone(whole)
	arg := unit
	if !i.Present {
		arg = nil
	}
	result := scriptRaiseZombieNativeCall(route, arg)
	if !reflect.DeepEqual(events, c.Events) || state() != c.Final {
		t.Fatalf("original service states: got %+v / %+v; want %+v / %+v", events, state(), c.Events, c.Final)
	}
	wantActions := 0
	if len(c.Events) != 0 && i.ChildReturn != 0 {
		wantActions = 1
	}
	if currentActions != wantActions {
		t.Fatal("already-native child early-return route", currentActions, wantActions)
	}
	var want uintptr
	switch c.ReturnKind {
	case "fault":
		if !c.OriginalFault || i.Present || c.Return != nil {
			t.Fatal("original eager-null fault diagnostic")
		}
		// The incoming C entry's explicit nil guard returns0; keep that policy.
	case "pointer":
		if c.Return == nil || *c.Return != 0x21001800 {
			t.Fatal("original pointer provenance")
		}
		want = uintptr(unsafe.Pointer(unit))
	case "scalar":
		if c.Return == nil {
			t.Fatal("missing original scalar result")
		}
		want = uintptr(*c.Return)
	default:
		t.Fatal("unknown original return kind", c.ReturnKind)
	}
	if result != want {
		t.Fatalf("native mixed return %#x, want %#x (%s)", result, want, c.ReturnKind)
	}
	if got := scriptRaiseZombieNativeData(unit); got != uintptr(dataPointer(c.Final.Update)) {
		t.Fatal("C observed a truncated or stale UpdateData pointer", got)
	}
	for n, update := range updates {
		want := c.Final.Status
		if n == 1 {
			want = c.Final.AlternateStatus
		}
		if got := scriptRaiseZombieNativeStatus(update); got != want {
			t.Fatal("C observed incorrect status through native update owner", n, got, want)
		}
	}
	put32 := func(at uintptr, value uint32) { binary.LittleEndian.PutUint32(before[at:at+4], value) }
	put32(unitOffset+unsafe.Offsetof(unit.ObjClass), c.Final.Class)
	put32(unitOffset+unsafe.Offsetof(unit.ObjFlags), c.Final.Flags)
	dataOffset := unitOffset + unsafe.Offsetof(unit.UpdateData)
	value := uintptr(dataPointer(c.Final.Update))
	if unsafe.Sizeof(value) == 8 {
		binary.LittleEndian.PutUint64(before[dataOffset:dataOffset+8], uint64(value))
	} else {
		binary.LittleEndian.PutUint32(before[dataOffset:dataOffset+4], uint32(value))
	}
	put32(updateOffsets[0]+unsafe.Offsetof(updates[0].StatusFlags), c.Final.Status)
	put32(updateOffsets[1]+unsafe.Offsetof(updates[1].StatusFlags), c.Final.AlternateStatus)
	if !bytes.Equal(whole, before) {
		t.Fatal("native entry changed unrelated owner or guard bytes")
	}
}

func TestScriptRaiseZombieNativeOriginal(t *testing.T) {
	cases := scriptRaiseZombieCases(t)
	for route := 0; route < 2; route++ {
		for _, placement := range []int{0x100, 0x200, 0x300} {
			t.Run(fmt.Sprintf("route%d/place%x", route, placement), func(t *testing.T) {
				buffer, free := alloc.MakeHigh([]byte{}, 65536+64)
				defer free()
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&buffer[0])) <= math.MaxUint32 {
					t.Fatal("foreign Object owner must exceed4GiB")
				}
				unitOffset := uintptr(32 + placement + 0x1800)
				updateOffsets := [2]uintptr{uintptr(32 + placement + 0x4000), uintptr(32 + placement + 0x8000)}
				unit := (*server.Object)(unsafe.Pointer(&buffer[unitOffset]))
				updates := [2]*server.MonsterUpdateData{(*server.MonsterUpdateData)(unsafe.Pointer(&buffer[updateOffsets[0]])), (*server.MonsterUpdateData)(unsafe.Pointer(&buffer[updateOffsets[1]]))}
				if unsafe.Sizeof(*unit) > 0x800 || unsafe.Sizeof(*updates[0]) > 0x3000 {
					t.Fatal("declared foreign owner ranges overlap")
				}
				for _, c := range cases {
					t.Run(c.Input.Name, func(t *testing.T) {
						for i := range buffer {
							buffer[i] = byte(int(c.Input.Seed) + 17*i + 3*(i>>8))
						}
						clearForeignFixtureRecord(unit)
						clearForeignFixtureRecord(updates[0])
						clearForeignFixtureRecord(updates[1])
						scriptRaiseZombieRun(t, c, route, unit, updates, buffer, unitOffset, updateOffsets)
					})
				}
			})
		}
	}
}

func TestScriptRaiseZombieNativeGoOwners(t *testing.T) {
	selected := map[string]bool{"nil-eager-read-diagnostic": true, "nonmonster-null-update": true, "alive-null-update": true, "child-both-ffffffff": true, "child-update-80000000": true, "child-none-00000000": true}
	for _, c := range scriptRaiseZombieCases(t) {
		if !selected[c.Input.Name] {
			continue
		}
		for route := 0; route < 2; route++ {
			t.Run(fmt.Sprintf("%s/route%d", c.Input.Name, route), func(t *testing.T) {
				// One Go allocation keeps complete typed owners and all padding visible.
				frame := new(struct {
					Before    [32]byte
					Unit      server.Object
					Middle    [32]byte
					Update    server.MonsterUpdateData
					Alternate server.MonsterUpdateData
					After     [32]byte
				})
				for i := range frame.Before {
					frame.Before[i], frame.Middle[i], frame.After[i] = 0xa5, 0x6d, 0x3c
				}
				var pin runtime.Pinner
				pin.Pin(frame)
				defer pin.Unpin()
				updates := [2]*server.MonsterUpdateData{&frame.Update, &frame.Alternate}
				unitOffset := unsafe.Offsetof(frame.Unit)
				updateOffsets := [2]uintptr{unsafe.Offsetof(frame.Update), unsafe.Offsetof(frame.Alternate)}
				whole := unsafe.Slice((*byte)(unsafe.Pointer(frame)), unsafe.Sizeof(*frame))
				scriptRaiseZombieRun(t, c, route, &frame.Unit, updates, whole, unitOffset, updateOffsets)
				runtime.GC()
				var data uintptr
				if c.Final.Update != 0 {
					data = uintptr(unsafe.Pointer(updates[c.Final.Update-1]))
				}
				if scriptRaiseZombieNativeData(&frame.Unit) != data || scriptRaiseZombieNativeStatus(updates[0]) != c.Final.Status || scriptRaiseZombieNativeStatus(updates[1]) != c.Final.AlternateStatus {
					t.Fatal("native C owner reads after GC")
				}
				runtime.KeepAlive(frame)
			})
		}
	}
}
