package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Both actual C entries use the production runtime, random selector, native
// marker array/charge bytes, wake cache, movement/index and delayed deletion.
// The simple map bodies and wake definition are isolated synthetic controls.
// OtherToMark preserves the existing native correction; it is not claimed
// equivalent to GAME.EXE's unrelated PE32 fields for those four IDs.
func TestSpellTeleportMarkersPublicCNativeComposition5305D0(t *testing.T) {
	const childEnv = "NOX_TELEPORT_MARKERS_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, family := range []string{"pop", "mark", "other"} {
			for slot := 0; slot < 4; slot++ {
				for _, charges := range []int{1, 2} {
					mode := fmt.Sprintf("%s:%d:%d", family, slot, charges)
					t.Run(mode, func(t *testing.T) {
						cmd := exec.Command(binary, "-test.run=^TestSpellTeleportMarkersPublicCNativeComposition5305D0$", "-test.v")
						cmd.Env = append(os.Environ(), childEnv+"="+mode)
						if out, err := cmd.CombinedOutput(); err != nil {
							t.Fatalf("public C marker composition: %v\n%s", err, out)
						} else {
							t.Logf("%s", out)
						}
					})
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 3 {
		t.Fatal("child mode")
	}
	family := parts[0]
	slot, err := strconv.Atoi(parts[1])
	if err != nil || slot < 0 || slot > 3 {
		t.Fatal("slot")
	}
	charges, err := strconv.Atoi(parts[2])
	if err != nil || charges < 1 || charges > 2 {
		t.Fatal("charges")
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.spells.duration.Init(s)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.SetFrame(142)
	s.SetTickRate(30)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "TeleportWake", Class: []things.Class{"SIMPLE"}, OnCollide: &things.ProcFunc{Name: "TeleportWakeCollide"}}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(8) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	target, marker := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	target.ObjClass, marker.ObjClass = object.ClassSimple, object.ClassSimple
	origin, destination := types.Ptf(100, 200), types.Ptf(float32(300+60*slot), float32(400+40*slot))
	s.CreateObjectAt(target, nil, origin)
	s.CreateObjectAt(marker, nil, destination)
	marker.Field34 = 0x12345678
	type sourceFrame struct {
		Before [32]byte
		Object server.Object
		After  [32]byte
	}
	type dataFrame struct {
		Before [32]byte
		Data   server.PlayerUpdateData
		After  [32]byte
	}
	type durationFrame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	source, freeSource := alloc.New(sourceFrame{})
	t.Cleanup(freeSource)
	*source = sourceFrame{}
	data, freeData := alloc.New(dataFrame{})
	t.Cleanup(freeData)
	*data = dataFrame{}
	f, freeRecord := alloc.New(durationFrame{})
	t.Cleanup(freeRecord)
	*f = durationFrame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
		source.Before[i], source.After[i] = byte(0x31+i), byte(0xc7-i)
		data.Before[i], data.After[i] = byte(0x71+i), byte(0xa7-i)
	}
	source.Object = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&data.Data), PosVec: types.Ptf(110, 210)}
	data.Data.Field29[slot] = marker
	shift := uint(slot * 8)
	data.Data.Field39 = uint32(0xfedcba98)&^(uint32(255)<<shift) | uint32(charges)<<shift
	id := spell.SPELL_TELEPORT_POP
	create, update := legacy.Get_nox_xxx_castTele_530820(), legacy.Get_sub_530880()
	if family == "mark" || family == "other" {
		id = spell.SPELL_TELEPORT_TO_MARK_1 + spell.ID(slot)
		if family == "other" {
			id = spell.SPELL_TELEPORT_OTHER_TO_MARK_1 + spell.ID(slot)
		}
		create, update = legacy.Get_sub_5305D0(), legacy.Get_sub_530650()
	} else if family != "pop" {
		t.Fatal("family")
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(id), Level: 3, Obj12: &source.Object, Caster16: &source.Object, Target48: target,
		Flag20: 1, Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Field36: 0x13579bdf, Flags88: 0x77556601}
	raw := func(p unsafe.Pointer, size uintptr) []byte { return unsafe.Slice((*byte)(p), size) }
	sourceBefore := append([]byte(nil), raw(unsafe.Pointer(source), unsafe.Sizeof(*source))...)
	expectedData := append([]byte(nil), raw(unsafe.Pointer(data), unsafe.Sizeof(*data))...)
	chargeOffset := unsafe.Offsetof(dataFrame{}.Data) + unsafe.Offsetof(server.PlayerUpdateData{}.Field39)
	binary.NativeEndian.PutUint32(expectedData[chargeOffset:], data.Data.Field39&^(uint32(255)<<shift)|uint32(charges-1)<<shift)
	if charges == 1 {
		offset := unsafe.Offsetof(dataFrame{}.Data) + unsafe.Offsetof(server.PlayerUpdateData{}.Field29) + uintptr(slot)*unsafe.Sizeof(uintptr(0))
		for i := uintptr(0); i < unsafe.Sizeof(uintptr(0)); i++ {
			expectedData[offset+i] = 0
		}
	}
	for _, p := range []unsafe.Pointer{f.Record.C(), source.Object.CObj(), unsafe.Pointer(&data.Data), target.CObj(), marker.CObj()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native owned pointer below4GiB: %p", p)
		}
	}
	want := *f
	want.Record.Frame68 = 143
	if got := ccall.CallIntPtr(create, f.Record.C()); got != 0 || !bytes.Equal(raw(unsafe.Pointer(f), unsafe.Sizeof(*f)), raw(unsafe.Pointer(&want), unsafe.Sizeof(want))) {
		t.Fatalf("public create return=%d or record/canary mismatch", got)
	}
	cache := memmap.PtrUint32(0x5D4594, 2487916)
	previousType := *cache
	*cache = 0
	t.Cleanup(func() { *cache = previousType })
	noxflags.SetGame(noxflags.GameModeCoop)
	previousPending := s.Objs.Pending
	if got := ccall.CallIntPtr(update, f.Record.C()); got != 1 || !bytes.Equal(raw(unsafe.Pointer(f), unsafe.Sizeof(*f)), raw(unsafe.Pointer(&want), unsafe.Sizeof(want))) {
		t.Fatalf("public update return=%d or record/canary mismatch", got)
	}
	wake := s.Objs.Pending
	kind := uint32(s.Types.IndByID("TeleportWake"))
	if kind == 0 || *cache != kind || wake == nil || wake == previousPending || wake.CollideData == nil || uint32(wake.TypeInd) != kind || wake.ObjOwner != target || wake.PosVec != origin || (*server.TeleportWakeCollideData)(wake.CollideData).Destination != destination || wake.Field34 != 172 {
		t.Fatalf("native wake/cache mismatch: kind=%d cache=%d wake=%p", kind, *cache, wake)
	}
	if target.PosVec != destination || target.NewPos != destination || target.PrevPos != destination || !target.Flags().Has(object.FlagPartitioned) || target.Flags().Has(object.FlagDestroyed) {
		t.Fatalf("native movement/index mismatch: %v/%v/%v flags=%v", target.PosVec, target.NewPos, target.PrevPos, target.Flags())
	}
	if !bytes.Equal(sourceBefore, raw(unsafe.Pointer(source), unsafe.Sizeof(*source))) || !bytes.Equal(expectedData, raw(unsafe.Pointer(data), unsafe.Sizeof(*data))) {
		t.Fatal("source/marker data/canary mismatch")
	}
	if family != "pop" && marker.Field34 != 142 || family == "pop" && charges == 2 && marker.Field34 != 0x12345678 {
		t.Fatalf("marker timer=%#x", marker.Field34)
	}
	if charges == 1 && (!marker.Flags().Has(object.FlagDestroyed) || s.Objs.DeletedList != marker) || charges == 2 && marker.Flags().Has(object.FlagDestroyed) {
		t.Fatal("marker delayed deletion mismatch")
	}
	for _, p := range []unsafe.Pointer{wake.CObj(), wake.CollideData} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native wake link below4GiB")
		}
	}
	t.Logf("public C markers mode=%s record=%p source=%p data=%p target=%p marker=%p wake=%p charge=%d->%d expiry=%#x native-canaries=intact spatial-index=true", mode, &f.Record, &source.Object, &data.Data, target, marker, wake, charges, charges-1, marker.Field34)
}
