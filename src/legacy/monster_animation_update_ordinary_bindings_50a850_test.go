package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
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
	"github.com/opennox/opennox/v1/server"
)

func TestAnimationUpdateOrdinaryBindings50A850(t *testing.T) {
	path := os.Getenv("NOX_ANIMATION_UPDATE_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_animation_update_original_50A850.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "cfc6b738e4886bb2a83046bceb301d9e625780517ede7624845d43c5d6f3ba17" {
		t.Fatal("original update seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Rows []animationRow50A850 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Rows) != 21896 {
		t.Fatal("original update dimensions")
	}
	byName := make(map[string]*animationRow50A850)
	for i := range ref.Rows {
		byName[ref.Rows[i].Name] = &ref.Rows[i]
	}
	modes := []struct {
		Name, Row string
		NoServer  bool
	}{
		{"completed nonNPC no server", "21043/completed-and-NPC-attack-short-circuits", true},
		{"completed nil table no server", "21042/completed-and-NPC-attack-short-circuits", true},
		{"NPC melee clears completed", "21138/completed-and-NPC-attack-short-circuits", true},
		{"NPC missile clears completed", "21150/completed-and-NPC-attack-short-circuits", true},
		{"completed NPC idle no server", "21126/completed-and-NPC-attack-short-circuits", true},
		{"initial nil table no server", "03440/actions", true},
		{"cached Mimic native record", "03441/actions", true},
		{"NPC idle zero frames no server", "16044/NPC-actions", true},
		{"NPC cast", "16548/NPC-actions", false},
		{"NPC dying", "16884/NPC-actions", false},
		{"NPC dead", "16912/NPC-actions", false},
		{"NPC shield block", "16632/NPC-actions", false},
		{"NPC weapon block", "16700/NPC-actions", false},
		{"zero record", "19668/counter-boundaries", false},
		{"advance frame", "20268/counter-boundaries", false},
		{"finish record", "20278/counter-boundaries", false},
		{"loop record", "20303/counter-boundaries", false},
		{"delay255 wraps", "20507/counter-boundaries", false},
		{"frame255 wraps", "20288/counter-boundaries", false},
		{"entry sprite record alias", "21420/sprite-record-alias", true},
		{"other sprite record alias", "21804/sprite-record-alias", true},
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for i := 1; i <= 44; i++ {
		name := fmt.Sprintf("UpdateAnimationBindingType%d", i)
		for j, v := range []int{11, 22, 33, 44} {
			if i == v {
				name = animationNames50A850[j]
			}
		}
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	if srv.Types.MimicID() != 11 {
		t.Fatal("real Go Mimic cache")
	}
	view := srv.TypePredicateRuntime534A10()
	view.Write(0, 22)
	view.Write(1, 33)
	view.Write(2, 44)
	oldMimic := statePredicateOriginalSwapMimicCache534840(11)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldMimic) })
	scratch := memmap.PtrOff(0x5D4594, 2487964)
	savedScratch := animationSnapshot50A850(unsafe.Add(scratch, -32), 80)
	t.Cleanup(func() { copy(unsafe.Slice((*byte)(unsafe.Add(scratch, -32)), 80), savedScratch) })
	unit, freeU := alloc.New(server.Object{})
	update, freeD := alloc.New(server.MonsterUpdateData{})
	other, freeO := alloc.New(server.MonsterUpdateData{})
	enemy, freeE := alloc.New(server.Object{})
	a, freeA := alloc.New([16]server.MonsterAnim{})
	b, freeB := alloc.New([16]server.MonsterAnim{})
	for _, free := range []func(){freeU, freeD, freeO, freeE, freeA, freeB} {
		t.Cleanup(free)
	}
	arrayPointer := func(role string) *[16]server.MonsterAnim {
		switch role {
		case "entry":
			return a
		case "other":
			return b
		case "nil":
			return nil
		}
		for i, name := range []string{"entry", "other"} {
			for delta := 0; delta < 4; delta++ {
				if role == fmt.Sprintf("alias-%s-%d", name, delta) {
					d := update
					if i == 1 {
						d = other
					}
					return (*[16]server.MonsterAnim)(unsafe.Add(unsafe.Pointer(&d.Field120_0), delta-9))
				}
			}
		}
		t.Fatal("unknown ordinary array role", role)
		return nil
	}
	expectedPointer := func(value NPCAnimationReturn50A850) uintptr {
		switch value.Role {
		case "scratch":
			return uintptr(scratch) + value.Offset
		case "nil":
			return value.Offset
		default:
			return uintptr(unsafe.Pointer(arrayPointer(value.Role))) + value.Offset
		}
	}
	framesByIndex := map[int]player.AnimType{1: player.AnimDie, 2: player.AnimDead, 21: player.AnimCastSpell, 30: player.AnimStaffBlock, 40: player.AnimRaiseShield, 47: player.AnimGreatSwordBlockLeft}
	calls := 0
	for index, mode := range modes {
		t.Run(mode.Name, func(t *testing.T) {
			row := byName[mode.Row]
			if row == nil {
				t.Fatal("ordinary row missing from original")
			}
			c := row.Input
			if c.Mutation != nil || c.Caches != [4]uint32{11, 22, 33, 44} {
				t.Fatal("ordinary mode is not cached unmodified input")
			}
			playerDraw := things.PlayerDraw{Anims: make(map[player.AnimType]*things.PlayerAnim)}
			for i, name := range framesByIndex {
				frames, duration := uint32(i*257+3), uint32(i*259+5)
				if c.TableProfile != 0 {
					frames = []uint32{0xffffffff, 0x80000000, 0, 256}[i%4]
					duration = []uint32{0x80000001, 0xffffffff, 511, 1}[i%4]
				}
				playerDraw.Anims[name] = &things.PlayerAnim{FramesPerDir: byte(frames), Field8: byte(duration)}
			}
			if err := srv.Types.ReadObjectType(&things.Thing{Name: fmt.Sprintf("UpdatePlayerFrames%d", index), Draw: playerDraw}); err != nil {
				t.Fatal(err)
			}
			owner := GetServer
			if mode.NoServer {
				GetServer = func() Server { panic("ordinary updater unexpectedly resolved server") }
			}
			defer func() { GetServer = owner }()
			for path := 0; path < 3; path++ {
				*unit = server.Object{TypeInd: c.Type, ObjClass: object.Class(c.Class), ObjSubClass: object.SubClass(c.Subclass), ObjFlags: object.Flags(c.Flags), UpdateData: unsafe.Pointer(update)}
				*enemy = server.Object{}
				for i, d := range []*server.MonsterUpdateData{update, other} {
					status, hasEnemy, state := c.Status, c.Enemy, c.State
					if i == 1 {
						status, hasEnemy, state = c.OtherStatus, c.OtherEnemy, c.OtherState
					}
					*d = server.MonsterUpdateData{AIStackInd: c.Head, StatusFlags: object.MonsterStatus(status), WeaponEquipFlags: c.Equip, Field517: uint32(c.Attack)}
					if hasEnemy {
						d.CurrentEnemy = enemy
					}
					raw := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(&d.Field120_0), -9)), 20)
					for n := range raw {
						raw[n] = 0xa5
					}
					d.Field119 = arrayPointer(c.Animation)
					*(*[4]uint8)(unsafe.Pointer(&d.Field120_0)) = state
					item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&d.AIStack[0]), int(c.Head)*int(unsafe.Sizeof(d.AIStack[0]))))
					item.Action = c.Action
				}
				for i, array := range []*[16]server.MonsterAnim{a, b} {
					raw := unsafe.Slice((*byte)(unsafe.Pointer(array)), 256)
					for n := range raw {
						raw[n] = byte(n + 7*i)
					}
					if c.ArrayRecord != nil {
						for n := 0; n < 16; n++ {
							raw[n*16+9] = byte(c.ArrayRecord[0])
							raw[n*16+10] = byte(c.ArrayRecord[1])
							binary.LittleEndian.PutUint32(raw[n*16+12:], c.ArrayRecord[2])
						}
					}
				}
				for n := range unsafe.Slice((*byte)(scratch), 16) {
					unsafe.Slice((*byte)(scratch), 16)[n] = 0x6d
				}
				beforeU, beforeD, beforeO, beforeA, beforeB := *unit, *update, *other, *a, *b
				scratchExpected := animationSnapshot50A850(unsafe.Add(scratch, -32), 80)
				record, err := hex.DecodeString(row.Output.ScratchAfter)
				if err != nil || len(record) != 16 {
					t.Fatal("original scratch record")
				}
				copy(scratchExpected[32:48], record)
				for _, want := range row.Output.Returns {
					var got uint8
					switch path {
					case 0:
						got = animationUpdatePublicC50A850(unit)
					case 1:
						got = srv.MonsterAnimationUpdate50A850(unit)
					case 2:
						Nox_xxx_updateNPCAnimData_50A850(unit)
					}
					expected := want.AL
					if want.Kind == "child_pointer" {
						if want.Pointer == nil {
							t.Fatal("original pointer return")
						}
						expected = uint8(expectedPointer(*want.Pointer))
					}
					if path != 2 && got != expected {
						t.Errorf("ordinary path%d char=%#02x original relocated=%#02x", path, got, expected)
					}
					if path == 0 {
						calls++
					}
				}
				*(*[4]uint8)(unsafe.Pointer(&beforeD.Field120_0)) = row.Output.EntryStateAfter
				*(*[4]uint8)(unsafe.Pointer(&beforeO.Field120_0)) = row.Output.OtherStateAfter
				if *unit != beforeU || *update != beforeD || *other != beforeO || *a != beforeA || *b != beforeB {
					t.Error("ordinary native inputs differ beyond original sprite outputs")
				}
				if !bytes.Equal(scratchExpected, animationSnapshot50A850(unsafe.Add(scratch, -32), 80)) {
					t.Error("ordinary scratch bytes or neighbors differ")
				}
				if view.Read(0) != 22 || view.Read(1) != 33 || view.Read(2) != 44 || *(*uint32)(statePredicateOriginalMimicCachePointer534840()) != 11 {
					t.Fatal("ordinary cache inputs changed")
				}
			}
		})
	}
	if calls != 42 {
		t.Fatal("ordinary original call count", calls)
	}
	t.Log("ordinary animation-update bindings modes=21 actual-C-calls=42 public-Go-calls=42 dispatcher-calls=42 original-char-and-memory=true actual-default-factories-and-PlayerDraw=true no-provider-substitution=true")
}
