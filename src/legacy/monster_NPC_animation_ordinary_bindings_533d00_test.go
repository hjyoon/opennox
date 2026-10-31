package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestNPCAnimationOrdinaryBindings533D00(t *testing.T) {
	path := os.Getenv("NOX_NPC_ANIMATION_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_NPC_animation_original_533D00.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "849441cc80e93c8d8b9a904b7cbcd4fe96d046e6ace6e0bd5a195ccbc3eccdab" {
		t.Fatal("unsealed original parent reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Rows []animationRow533D00 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Rows) != 19668 {
		t.Fatal("original parent dimensions")
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for index := 1; index <= 44; index++ {
		name := fmt.Sprintf("NPCAnimationBindingType%d", index)
		for i, v := range []int{11, 22, 33, 44} {
			if v == index {
				name = animationNames533D00[i]
			}
		}
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	indices := map[int]player.AnimType{0: player.AnimIdle, 1: player.AnimDie, 2: player.AnimDead, 17: player.AnimSit, 21: player.AnimCastSpell, 26: player.AnimMaceStrike, 30: player.AnimStaffBlock, 40: player.AnimRaiseShield, 47: player.AnimGreatSwordBlockLeft, 52: player.AnimRecoil}
	playerDraw := things.PlayerDraw{Anims: make(map[player.AnimType]*things.PlayerAnim)}
	for index, name := range indices {
		playerDraw.Anims[name] = &things.PlayerAnim{FramesPerDir: byte(index*257 + 3), Field8: byte(index*259 + 5)}
	}
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NPCAnimationPlayerFrames", Draw: playerDraw}); err != nil {
		t.Fatal(err)
	}
	if srv.Types.MimicID() != 11 {
		t.Fatal("ordinary Go Mimic cache")
	}
	view := srv.TypePredicateRuntime534A10()
	view.Write(0, 22)
	view.Write(1, 33)
	view.Write(2, 44)
	oldMimic := statePredicateOriginalSwapMimicCache534840(11)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldMimic) })
	weapon := memmap.PtrOff(0x587000, 215824)
	weaponBefore := animationSnapshot533D00(weapon, 108)
	t.Cleanup(func() { copy(unsafe.Slice((*byte)(weapon), 108), weaponBefore) })
	for i := 0; i < 27; i++ {
		*memmap.PtrUint32(0x587000, uintptr(215824+4*i)) = uint32((i*7 + 3) % 53)
	}
	scratch := memmap.PtrOff(0x5D4594, 2487964)
	savedScratch := animationSnapshot533D00(unsafe.Add(scratch, -32), 80)
	t.Cleanup(func() { copy(unsafe.Slice((*byte)(unsafe.Add(scratch, -32)), 80), savedScratch) })
	unit, freeU := alloc.New(server.Object{})
	update, freeD := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeU)
	t.Cleanup(freeD)
	a, freeA := alloc.New([16]server.MonsterAnim{})
	b, freeB := alloc.New([16]server.MonsterAnim{})
	t.Cleanup(freeA)
	t.Cleanup(freeB)
	for i, p := range []*[16]server.MonsterAnim{a, b} {
		v := unsafe.Slice((*byte)(unsafe.Pointer(p)), 256)
		for n := range v {
			v[n] = byte(n + 7*i)
		}
	}
	type mode struct {
		Name                    string
		Subclass, Action, Equip uint32
		Attack                  byte
		Type                    uint16
		Status                  uint32
		NilAnimation, NoServer  bool
	}
	modes := []mode{
		{Name: "idle scratch without server", Subclass: 0x10, NoServer: true},
		{Name: "cast real player frames", Subclass: 0x10, Action: 18},
		{Name: "dying real player frames", Subclass: 0x10, Action: 30},
		{Name: "dead real player frames", Subclass: 0x10, Action: 31},
		{Name: "shield block", Subclass: 0x10, Action: 21},
		{Name: "staff block", Subclass: 0x10, Action: 23, Equip: 1 << 17},
		{Name: "great sword block", Subclass: 0x10, Action: 23, Equip: 4},
		{Name: "melee weapon table and repeated scratch", Subclass: 0x10, Action: 16, Equip: 4},
		{Name: "missile final weapon bit", Subclass: 0x10, Action: 17, Equip: 1 << 26},
		{Name: "melee byte fallback and repeated scratch", Subclass: 0x10, Action: 16, Attack: 52},
		{Name: "unsupported high weapon bit", Subclass: 0x10, Action: 17, Equip: 1 << 31},
		{Name: "nil table skips selector and server", NilAnimation: true, NoServer: true},
		{Name: "cached Mimic native animation pointer", Type: 11, NoServer: true},
		{Name: "burning Zombie native animation pointer", Type: 33, Action: 30, Status: 0x80000},
	}
	calls := 0
	for _, c := range modes {
		t.Run(c.Name, func(t *testing.T) {
			typ := c.Type
			if typ == 0 {
				typ = 117
			}
			var row *animationRow533D00
			for i := range ref.Rows {
				r := &ref.Rows[i]
				x := r.Input
				if x.Head == 0 && x.Class == 2 && x.Subclass == c.Subclass && x.Type == typ && x.Action == c.Action && x.Status == c.Status && !x.Enemy && x.Equip == c.Equip && x.Attack == c.Attack && x.TableProfile == 0 && x.Mutation == nil && ((x.Animation == "nil") == c.NilAnimation) && x.Caches == [4]uint32{11, 22, 33, 44} && x.Found == [4]bool{true, true, true, true} {
					row = r
					break
				}
			}
			if row == nil {
				t.Fatal("ordinary input missing from frozen original")
			}
			var array *[16]server.MonsterAnim
			switch row.Input.Animation {
			case "entry":
				array = a
			case "other":
				array = b
			case "nil":
			default:
				t.Fatal("unknown original array role")
			}
			*update = server.MonsterUpdateData{AIStackInd: 0, WeaponEquipFlags: c.Equip, Field517: uint32(c.Attack), StatusFlags: object.MonsterStatus(c.Status), Field119: array}
			update.AIStack[0].Action = c.Action
			*unit = server.Object{TypeInd: typ, ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(c.Subclass), UpdateData: unsafe.Pointer(update)}
			beforeU, beforeD := *unit, *update
			owner := GetServer
			if c.NoServer {
				GetServer = func() Server { panic("ordinary parent unexpectedly resolved server") }
			}
			defer func() { GetServer = owner }()
			wantScratch, err := hex.DecodeString(row.Output.ScratchAfter)
			if err != nil || len(wantScratch) != 16 {
				t.Fatal("invalid original scratch")
			}
			for _, publicGo := range []bool{false, true} {
				for i := range unsafe.Slice((*byte)(scratch), 16) {
					unsafe.Slice((*byte)(scratch), 16)[i] = 0x6d
				}
				guard := animationSnapshot533D00(unsafe.Add(scratch, -32), 80)
				copy(guard[32:48], wantScratch)
				for _, want := range row.Output.Returns {
					var got unsafe.Pointer
					if publicGo {
						got = unsafe.Pointer(srv.MonsterNPCAnimationPointer533D00(unit))
					} else {
						got = ccall.CallPtrPtr(NPCAnimationPublicC533D00(), unit.CObj())
					}
					var base uintptr
					switch want.Role {
					case "scratch":
						base = uintptr(scratch)
					case "entry":
						base = uintptr(unsafe.Pointer(a))
					case "other":
						base = uintptr(unsafe.Pointer(b))
					case "nil":
					default:
						t.Fatal("unknown original pointer role")
					}
					if uintptr(got) != base+want.Offset {
						t.Errorf("ordinary publicGo=%t pointer=%p original=%s+%d", publicGo, got, want.Role, want.Offset)
					}
					if !bytes.Equal(guard, animationSnapshot533D00(unsafe.Add(scratch, -32), 80)) {
						t.Error("ordinary scratch record or neighboring memory differs")
					}
					if !publicGo {
						calls++
					}
				}
			}
			if *unit != beforeU || *update != beforeD || view.Read(0) != 22 || view.Read(1) != 33 || view.Read(2) != 44 || *(*uint32)(statePredicateOriginalMimicCachePointer534840()) != 11 {
				t.Fatal("ordinary binding changed object/update/cache inputs")
			}
		})
	}
	if calls != 16 {
		t.Fatal("ordinary original call count", calls)
	}
	t.Log("ordinary NPC-animation bindings modes=14 actual-C-calls=16 public-Go-calls=16 original-pointer-and-record=true real-PlayerDraw-and-weapon-tables=true no-factory-frame-provider-or-predicate-substitution=true")
}
