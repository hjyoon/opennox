package server

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
	"github.com/opennox/libs/types"
)

type attackTraceStockInput538330 struct {
	Name, Entry, Mutation            string
	Actor, Attack, Health, Ray       bool
	Candidate, Owner, Weapon, Winner int
	AreaTargets                      []int        `json:"area_targets"`
	OwnerClass                       object.Class `json:"owner_class"`
	TargetClass                      object.Class `json:"target_class"`
	FireDir                          uint8
	DirectionMask                    uint8        `json:"direction_mask"`
	TargetFlags                      object.Flags `json:"target_flags"`
	TargetMaterial                   uint16       `json:"target_material"`
	AllowDisabled                    uint32       `json:"allow_disabled"`
	MaxHealth                        uint16       `json:"max_health"`
	GameFlag                         int          `json:"game_flag"`
	Animation                        uint8
	Damage, Radius, Balance          uint32
	DamageType                       uint8 `json:"damage_type"`
	Center                           [2]uint32
}

type attackTraceStockOutput538330 struct {
	Trace             json.RawMessage
	Distance, Damaged uint32
	Nearest           int
	Return            *int
}

func TestPlayerAttackTrace538330FrozenOriginal(t *testing.T) {
	b, err := os.ReadFile("testdata/player_attack_trace_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if sha := fmt.Sprintf("%x", sha256.Sum256(b)); sha != "5decbb212050b615d84a450bccdf124ff4264598ba92fa4e75afe2a4122cadaf" {
		t.Fatalf("unexpected original trace corpus hash %s", sha)
	}
	reader, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var data struct {
		BodySizes   map[string]int `json:"executed_bytes"`
		Counts      map[string]int
		Guards      bool   `json:"guards_verified"`
		ControlWord string `json:"startup_control_word"`
		Corpus      []struct {
			Input    attackTraceStockInput538330
			Observed attackTraceStockOutput538330
		}
	}
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if !data.Guards || data.ControlWord != "0xe3f" || data.BodySizes["trace"] != 473 || data.BodySizes["damage-candidate"] != 386 ||
		data.BodySizes["float-to-int32"] != 17 || data.BodySizes["truncate-x87-qword"] != 39 ||
		data.Counts["trace"] != 42 || data.Counts["damage-candidate"] != 44 || len(data.Corpus) != 86 {
		t.Fatal("unexpected original attack trace scope")
	}
	for _, row := range data.Corpus {
		t.Run(row.Input.Name, func(t *testing.T) { checkPlayerAttackTrace538330(t, row.Input, row.Observed) })
	}
}

func checkPlayerAttackTrace538330(t *testing.T, input attackTraceStockInput538330, want attackTraceStockOutput538330) {
	t.Helper()
	f32 := math.Float32frombits
	bits := math.Float32bits
	point := func(p types.Pointf) [2]uint32 { return [2]uint32{bits(p.X), bits(p.Y)} }
	objects := make([]*Object, 8)
	for i := 1; i < len(objects); i++ {
		objects[i] = &Object{PosVec: types.Ptf(f32(0x44010000+uint32(i)*0x1800), f32(0x44080000+uint32(i)*0x1000)),
			Direction1: 0xfffe}
	}
	for _, target := range objects[1:3] {
		target.ObjFlags, target.ObjClass, target.Material = input.TargetFlags, input.TargetClass, input.TargetMaterial
		if input.Health {
			target.HealthData = &HealthData{Max: input.MaxHealth}
		}
	}
	for _, owner := range objects[3:5] {
		owner.ObjClass = input.OwnerClass
		if owner.ObjClass.Has(object.ClassPlayer) {
			owner.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &Player{Field8: uint16(input.Animation)}})
		} else {
			owner.UpdateData = unsafe.Pointer(&MonsterUpdateData{Field517: uint32(input.Animation)})
		}
	}
	objects[7].ObjSubClass = 0x4000
	weapons := [3]*Object{nil, objects[6], objects[7]}
	state := PlayerAttackTraceState538330{Distance: f32(0x41e80000), Damaged: 0xa1b2c3d4, Closest: objects[2]}
	record := &PlayerAttackData538330{Damage: f32(input.Damage), DamageType: input.DamageType, Radius: f32(input.Radius),
		Position: types.Ptf(f32(input.Center[0]), f32(input.Center[1])), Weapon: weapons[input.Weapon],
		AllowDisabled: input.AllowDisabled, DirectionMask: input.DirectionMask}
	if input.Owner != 0 {
		record.Owner = objects[3]
	}
	attack, attacker := record, objects[5]
	if !input.Attack {
		attack = nil
	}
	if !input.Actor {
		attacker = nil
	}
	role := func(obj *Object) int {
		i := slices.Index(objects, obj)
		if i < 0 {
			t.Fatalf("unrecognized object %p", obj)
		}
		return i
	}
	events := make([][]any, 0)
	runtime := PlayerAttackTraceRuntime538330{
		Scan: func(center *types.Pointf, radius float32, hit func(*Object)) {
			if center == &record.Position {
				events = append(events, []any{"scan-damage", point(*center), bits(radius), input.AreaTargets})
				for _, i := range input.AreaTargets {
					hit(objects[i])
				}
				return
			}
			if center != &attacker.PosVec || state.Closest != nil || bits(state.Distance) != bits(radius) {
				t.Fatal("selection must initialize native scratch and scan around the live attacker")
			}
			events = append(events, []any{"scan-select", point(*center), bits(radius), input.Winner})
			state.Closest = objects[input.Winner]
			if state.Closest != nil {
				state.Distance = f32(0x41340000)
			}
		},
		Direction: func(from types.Pointf, dir int32, to types.Pointf) uint8 {
			events = append(events, []any{"direction", point(from), dir, point(to)})
			if input.Mutation == "direction-owner" {
				record.Owner = objects[4]
			}
			return input.FireDir
		},
		Trace: func(from, to types.Pointf) bool {
			events = append(events, []any{"ray", [4]uint32{bits(from.X), bits(from.Y), bits(to.X), bits(to.Y)}, 5})
			if input.Mutation == "ray-damage" {
				record.Damage = f32(0x41e9999a)
			}
			return input.Ray
		},
		PreHit: func(target, owner, weapon *Object, info *PlayerAttackData538330) {
			if info != record {
				t.Fatal("modifier got a copied attack record")
			}
			events = append(events, []any{"pre-hit", role(target), role(owner), role(weapon), bits(info.Damage)})
			if input.Mutation == "pre-hit-replace-payload" {
				info.Damage, info.Owner, info.Weapon = f32(0x41f7ffff), objects[4], objects[6]
			}
		},
		Damage: func(target, owner, weapon *Object, damage int32, typ uint8) {
			events = append(events, []any{"damage", role(target), role(owner), role(weapon), damage, typ})
			if input.Mutation == "damage-position" {
				target.PosVec = types.Ptf(f32(0x44800000), f32(0x44900000))
			}
		},
		ReportPlayerFX: func() bool {
			events = append(events, []any{"game-flag", 2048})
			return input.GameFlag == 1
		},
		PointFX: func(pos *types.Pointf) { events = append(events, []any{"point-fx", 0x8b, point(*pos)}) },
		Force: func(origin *types.Pointf, target *Object, strength float32) {
			events = append(events, []any{"force", point(*origin), role(target), bits(strength)})
		},
		Walls: func(bounds [4]int32, center *types.Pointf, radius float32, damage int32, typ uint8, item *Object) {
			if center != &record.Position {
				t.Fatal("wall damage lost its live native position")
			}
			events = append(events, []any{"walls", bounds, point(*center), bits(radius), damage, typ, role(item)})
			if input.Mutation == "walls-weapon" {
				record.Weapon = objects[6]
			}
		},
		ItemDamage: func() float64 {
			events = append(events, []any{"item-damage-balance"})
			return float64(f32(input.Balance))
		},
		WeaponDamage: func(weapon, owner, source, effective *Object, damage float32, typ uint8) {
			events = append(events, []any{"weapon-damage", role(weapon), role(owner), role(source), role(effective), bits(damage), typ})
		},
	}
	if input.Entry == "trace" {
		if result := state.Trace(attacker, attack, runtime); want.Return == nil || result != *want.Return {
			t.Fatalf("trace returned %d, expected %v", result, want.Return)
		}
	} else {
		state.Hit(objects[input.Candidate], attack, runtime)
	}
	got, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Trace) || bits(state.Distance) != want.Distance || state.Damaged != want.Damaged || role(state.Closest) != want.Nearest {
		t.Fatalf("trace %s; original %s; state distance=%08x damaged=%08x closest=%d; original %08x/%08x/%d",
			got, want.Trace, bits(state.Distance), state.Damaged, role(state.Closest), want.Distance, want.Damaged, want.Nearest)
	}
}
