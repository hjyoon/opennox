package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

type bowStockInput539BD0 struct {
	Name, Root, Mutation      string
	Flags                     uint32
	FirstFlags                uint32 `json:"first_flags"`
	SecondFlags               uint32 `json:"second_flags"`
	Actor                     object.Class
	Quest, Visible, Allocated bool
	Reload, Quiver            int
	WeaponCharge              uint8 `json:"weapon_charge"`
	First, Second             [3]uint8
	Equipped                  [2]bool
	Effects                   [2]string
	Radius                    uint32
	InitialSpeed              uint32 `json:"initial_speed"`
	EffectSpeed               uint32 `json:"effect_speed"`
	Pos                       [2]uint32
	Direction                 int16
}

type bowStockArrow539BD0 struct {
	Owner     int
	Velocity  [2]uint32
	Direction uint32
	Speed     uint32
	Recoil    int
}

type bowStockOutput539BD0 struct {
	Events json.RawMessage
	EAX    uint32
	Ammo   [4][3]uint8
	Arrow  bowStockArrow539BD0
}

func TestPlayerAttackBowFrozenOriginal539BD0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_attack_bow_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "9fe9f034686f10734469e5b44fcaae3531f47d7104c920517d2df79a3fc556ed" {
		t.Fatalf("stock bow corpus hash = %s", got)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var record struct {
		Startup string `json:"startup_control_word"`
		Guards  bool   `json:"world_and_register_guards"`
		Windows map[string]struct{ Bytes int }
		Entered map[string]int
		Rows    []struct {
			Input    bowStockInput539BD0
			Observed bowStockOutput539BD0
		}
	}
	if err := json.NewDecoder(reader).Decode(&record); err != nil {
		t.Fatal(err)
	}
	if !record.Guards || record.Startup != "0xe3f" || len(record.Rows) != 87 ||
		record.Windows["shoot"].Bytes != 423 || record.Windows["projectile"].Bytes != 433 || record.Windows["effects"].Bytes != 108 ||
		record.Entered["shoot"] != 37 || record.Entered["projectile"] != 51 || record.Entered["effects"] != 50 {
		t.Fatal("incomplete original bow execution")
	}
	for _, row := range record.Rows {
		t.Run(row.Input.Name, func(t *testing.T) { checkBowStock539BD0(t, row.Input, row.Observed) })
	}
}

func checkBowStock539BD0(t *testing.T, in bowStockInput539BD0, want bowStockOutput539BD0) {
	t.Helper()
	f32 := math.Float32frombits
	bits := math.Float32bits
	ammo := game4ConsumerValue(t, [4]server.AmmoUseData{
		{Charge0: in.WeaponCharge},
		{Charge0: in.First[0], Charge1: in.First[1], Field2: in.First[2]},
		{Charge0: in.Second[0], Charge1: in.Second[1], Field2: in.Second[2]},
		{Charge0: 13, Charge1: 21},
	})
	player := game4ConsumerValue(t, server.Player{PlayerInd: 197})
	update := game4ConsumerValue(t, server.PlayerUpdateData{Player: player})
	owner := game4ConsumerValue(t, server.Object{ObjClass: in.Actor, PosVec: types.Ptf(f32(in.Pos[0]), f32(in.Pos[1])),
		Direction1: server.Dir16(uint16(in.Direction)), UpdateData: unsafe.Pointer(update)})
	owner.Shape.Circle.R = f32(in.Radius)
	weapon := game4ConsumerValue(t, server.Object{UseData: server.UseDataPtr{Ptr: ammo[0].UseDataPtr()}})
	q1 := game4ConsumerValue(t, server.Object{UseData: server.UseDataPtr{Ptr: ammo[1].UseDataPtr()}})
	q2 := game4ConsumerValue(t, server.Object{UseData: server.UseDataPtr{Ptr: ammo[2].UseDataPtr()}})
	owner.InvFirstItem, q1.InvNextItem = q1, q2
	quivers := [3]*server.Object{nil, q1, q2}
	for i, equipped := range in.Equipped {
		if equipped {
			quivers[i+1].ObjFlags |= object.FlagEquipped
		}
	}
	arrow := game4ConsumerValue(t, server.ArrowCollideData{})
	projectile := game4ConsumerValue(t, server.Object{SpeedCur: f32(in.InitialSpeed), CollideData: unsafe.Pointer(arrow)})
	rp, sp := RecoilEffectPointer4E0640(), ProjectileSpeedEffectPointer4E09B0()
	recoil := game4ConsumerValue(t, server.ModifierEff{AttackPreHit52: server.ModifierEffFnc{Fnc: rp}})
	speed := game4ConsumerValue(t, server.ModifierEff{Attack40: server.ModifierEffFnc{Fnc: sp}})
	both := game4ConsumerValue(t, server.ModifierEff{AttackPreHit52: server.ModifierEffFnc{Fnc: rp}, Attack40: server.ModifierEffFnc{Fnc: sp}})
	effects := map[string]*server.ModifierEff{"none": nil, "recoil": recoil, "speed": speed, "both": both}
	weaponAttrs := game4ConsumerValue(t, server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, effects[in.Effects[0]], effects[in.Effects[1]]}})
	quiverAttrs := game4ConsumerValue(t, server.ModifierInitData{})
	arrowAttrs := game4ConsumerValue(t, server.ModifierInitData{})
	weapon.InitData, q1.InitData, q2.InitData, projectile.InitData =
		unsafe.Pointer(weaponAttrs), unsafe.Pointer(quiverAttrs), unsafe.Pointer(quiverAttrs), unsafe.Pointer(arrowAttrs)
	objRefs := []*server.Object{nil, owner, weapon, q1, q2, projectile}
	modRefs := []*server.ModifierEff{nil, nil, nil, nil, nil, nil, nil, nil, nil, recoil, speed, both}
	objRole := func(obj *server.Object) int {
		role := slices.Index(objRefs, obj)
		if role < 0 {
			t.Fatalf("unexpected object %p", obj)
		}
		return role
	}
	events := make([][]any, 0)
	r := playerAttackBowRuntime538960{
		playerAttackProjectileRuntime538960: playerAttackProjectileRuntime538960{
			WeaponInventoryFlags: func(o *server.Object) uint32 {
				index := objRole(o)
				if index < 2 || index > 4 {
					t.Fatalf("unexpected inventory item %p", o)
				}
				flag := [3]uint32{in.Flags, in.FirstFlags, in.SecondFlags}[index-2]
				events = append(events, []any{"equipment", index, flag})
				return flag
			},
			QuestMode: func() bool { events = append(events, []any{"quest"}); return in.Quest },
			Trace: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
				if flags != 5 {
					t.Fatalf("ray flags %d", flags)
				}
				events = append(events, []any{"ray", bits(from.X), bits(from.Y), bits(to.X), bits(to.Y)})
				if in.Mutation == "trace-replaces-ammo" {
					quivers[in.Quiver].UseData.Ptr = ammo[3].UseDataPtr()
				}
				return in.Visible
			},
			NewObject: func(name string) *server.Object {
				events = append(events, []any{"new", name})
				if in.Allocated {
					return projectile
				}
				return nil
			},
			CreateAt: func(o, source *server.Object, position types.Pointf) {
				if o != projectile || source != owner {
					t.Fatalf("create projectile=%p owner=%p", o, source)
				}
				events = append(events, []any{"create", bits(position.X), bits(position.Y), objRole(arrow.Owner)})
				if in.Mutation == "create-changes-facing" {
					owner.Direction1 = 0xfffe
				}
			},
			ApplyModifierAttrs: func(o *server.Object, attrs *server.ModifierInitData) {
				if o != projectile || attrs != quiverAttrs {
					t.Fatalf("apply attributes projectile=%p attrs=%p", o, attrs)
				}
				events = append(events, []any{"attributes", 7})
			},
			RecoilEffect: rp, ProjectileSpeedEffect: sp,
			ApplyProjectileSpeed: func(effect *server.ModifierEff, o *server.Object) {
				if o != projectile {
					t.Fatal("speed modifier got a different projectile")
				}
				events = append(events, []any{"speed", slices.Index(modRefs, effect), bits(o.SpeedCur)})
				o.SpeedCur = f32(in.EffectSpeed)
				if in.Mutation == "speed-replaces-final-slot" {
					weaponAttrs.Modifiers[3] = recoil
				}
			},
			DelayedDelete: func(item *server.Object) { events = append(events, []any{"delete", objRole(item)}) },
			AudioEvent: func(id sound.ID, source *server.Object) {
				if source != owner {
					t.Fatal("shot sound source changed")
				}
				events = append(events, []any{"audio", id})
			},
		},
		Direction: func(direction server.Dir16) types.Pointf {
			if direction == 0xfffe {
				return types.Ptf(-0.87, 0.44)
			}
			if direction != 0 {
				t.Fatalf("facing %d", direction)
			}
			return types.Ptf(0.6, 0.8)
		},
		ReportCharges: func(player uint8, item *server.Object, cur, max uint8) {
			i := objRole(item)
			events = append(events, []any{"charges", player, i, cur, max})
			if in.Mutation == "charge-report-live-value" {
				ammo[i-2].Charge1 = 5
			}
		},
		ReloadQuiver: func(o *server.Object) int {
			if o != owner {
				t.Fatal("reload owner changed")
			}
			events = append(events, []any{"reload"})
			return in.Reload
		},
		Priority: func(o *server.Object, message strman.ID, mode byte) {
			if o != owner || mode != 0 {
				t.Fatal("priority message lost owner/mode")
			}
			events = append(events, []any{"priority", message})
			if in.Mutation == "priority-live-charge" {
				ammo[0].Charge0 = 9
				weapon.UseData.Ptr = ammo[3].UseDataPtr()
			}
		},
		SetState: func(o *server.Object, state server.PlayerState) bool {
			if o != owner {
				t.Fatal("state owner changed")
			}
			events = append(events, []any{"state", state})
			return true
		},
	}
	oldBow, oldProjectile := playerAttackBowRuntimeFactory538960, playerAttackProjectileRuntimeFactory538960
	t.Cleanup(func() {
		playerAttackBowRuntimeFactory538960, playerAttackProjectileRuntimeFactory538960 = oldBow, oldProjectile
	})
	playerAttackBowRuntimeFactory538960 = func() playerAttackBowRuntime538960 { return r }
	playerAttackProjectileRuntimeFactory538960 = func() playerAttackProjectileRuntime538960 {
		return r.playerAttackProjectileRuntime538960
	}
	switch in.Root {
	case "shoot":
		if got := bowShootStockFixture539BD0(owner, weapon); uint32(got) != want.EAX {
			t.Fatalf("shoot returned %d, original %d", got, want.EAX)
		}
	case "projectile":
		bowProjectileStockFixture539D80(owner, quivers[in.Quiver], weapon, in.Flags)
	case "effects":
		if got := bowEffectsStockFixture539F40(owner, weapon, projectile); uint32(got) != want.EAX {
			t.Fatalf("effects returned %d, original %d", got, want.EAX)
		}
	default:
		t.Fatalf("unexpected original root %s", in.Root)
	}
	gotTrace, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotTrace, want.Events) {
		t.Fatalf("events %s; original %s", gotTrace, want.Events)
	}
	gotArrow := bowStockArrow539BD0{Owner: objRole(arrow.Owner),
		Velocity:  [2]uint32{bits(projectile.VelVec.X), bits(projectile.VelVec.Y)},
		Direction: uint32(projectile.Direction1) | uint32(projectile.Direction2)<<16,
		Speed:     bits(projectile.SpeedCur), Recoil: slices.Index(modRefs, arrowAttrs.Modifiers[3])}
	if gotArrow != want.Arrow {
		t.Fatalf("arrow %+v; original %+v", gotArrow, want.Arrow)
	}
	for i, payload := range ammo {
		if got := [3]uint8{payload.Charge0, payload.Charge1, payload.Field2}; got != want.Ammo[i] {
			t.Fatalf("ammo %d = %v; original %v", i, got, want.Ammo[i])
		}
	}
}
