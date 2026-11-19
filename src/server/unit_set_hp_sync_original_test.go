package server

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
)

//go:embed testdata/original_sync_hp_and_state_v2.json.gz
var unitSyncStock4E4560 []byte

type unitSyncStockInput struct {
	Entry        string
	Name         string
	Class        uint32 `json:"cls"`
	Subclass     uint32
	Flags        uint32
	Z            uint32
	Frame        uint32
	Field37      uint32 `json:"field37"`
	TypeDiffWord int    `json:"type_diff"`
	Arg          int64
	HasHealth    bool `json:"has_health"`
	Holder       string
	HolderPlayer bool    `json:"holder_player"`
	NextHolder   *string `json:"next_holder"`
	AfterClass   *uint32 `json:"after_class"`
	Status       int32
	ReloadHealth bool `json:"reload_health"`
}

type unitSyncStockCall struct {
	Name   string
	Token  uint32
	Health uint16
	Amount uint16
}

type unitSyncStockOutput struct {
	Result struct {
		Status *int32 `json:"signed_status"`
		Holder string
	}
	Calls         []unitSyncStockCall
	Flags         uint32
	Z             uint32
	Frame         uint32
	SyncPending   uint32 `json:"sync_pending"`
	Masks         [32]uint32
	Health        [2]uint32
	CurrentHolder string `json:"current_holder"`
	CurrentHealth string `json:"current_health"`
	CurrentClass  uint32 `json:"current_class"`
}

func TestUnitSyncSetterFrozenOriginal4E4560(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(unitSyncStock4E4560))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var stock struct {
		Initial [32]uint32 `json:"initial_sync_fields"`
		Faults  []json.RawMessage
		Rows    []struct {
			Input    unitSyncStockInput
			Observed unitSyncStockOutput
		}
	}
	if err = json.NewDecoder(reader).Decode(&stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 1494 || len(stock.Faults) != 4 {
		t.Fatalf("original records = %d with %d faults", len(stock.Rows), len(stock.Faults))
	}
	s := NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.Types.Free()
	baseHP := 3
	const typeName = "UnitSyncStock4E4560"
	if err := s.Types.ReadObjectType(&things.Thing{Name: typeName, Health: &baseHP}); err != nil {
		t.Fatal(err)
	}
	typ := s.Types.ByID(typeName)
	target, holder, replacement := new(Object), new(Object), new(Object)
	health, newHealth := new(HealthData), new(HealthData)
	for _, row := range stock.Rows {
		in, want := row.Input, row.Observed
		t.Run(in.Name, func(t *testing.T) {
			selfInfo := &Player{ProtUnitHPCur: 0xa0000001}
			holderInfo := &Player{Prot4632: 0xa0000012}
			replaceInfo := &Player{Prot4632: 0xa0000013}
			holderClass := object.ClassMonster
			if in.HolderPlayer {
				holderClass = object.ClassPlayer
			}
			*holder = Object{ObjClass: holderClass, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: holderInfo})}
			*replacement = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: replaceInfo})}
			holders := map[string]*Object{"nil": nil, "holder": holder, "replacement": replacement}
			healthRecords := map[string]*HealthData{"nil": nil, "health": health, "replacement": newHealth}
			*health = HealthData{Cur: 0x1234, Field2: 0xa6a6, Max: 1000}
			*newHealth = HealthData{Cur: 0x4321, Field2: 0xb5b5, Max: 2000}
			*target = Object{
				TypeInd: uint16(typ.Ind()), ObjClass: object.Class(in.Class), ObjSubClass: object.SubClass(in.Subclass),
				ObjFlags: object.Flags(in.Flags), UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: selfInfo}),
				InvHolder: holders[in.Holder], ZVal: math.Float32frombits(in.Z), Field33: in.Frame, Field37: in.Field37,
				Field38: 0x77667766, Field140: stock.Initial, serverHandle: s.handle,
			}
			if in.HasHealth {
				target.HealthData = health
			}
			typ.flags = object.Flags(want.Flags)
			typ.health.Cur = uint16(in.Arg)
			if in.TypeDiffWord != 0 {
				typ.flags ^= object.FlagEnabled
				typ.health.Cur ^= 1
			}

			switch in.Entry {
			case "unit-set-hp":
				var calls []unitSyncStockCall
				got := target.SetHP4E4560(uint16(in.Arg), UnitSetHPRuntime4E4560{
					CheckHeld: func(token uint32, item *Object) {
						if item != target {
							t.Fatal("check received a different object")
						}
						calls = append(calls, unitSyncStockCall{Name: "protect-before", Token: token, Health: item.HealthData.Cur})
						if in.NextHolder != nil {
							item.InvHolder = holders[*in.NextHolder]
						}
						if in.ReloadHealth {
							item.HealthData = newHealth
						}
						if in.AfterClass != nil {
							item.ObjClass = object.Class(*in.AfterClass)
						}
					},
					SetPlayerHP: func(token uint32, amount uint16) {
						calls = append(calls, unitSyncStockCall{Name: "protect-player-hp", Token: token, Amount: amount})
					},
					ProtectHeld: func(token uint32, item *Object) int32 {
						if item != target {
							t.Fatal("protection received a different object")
						}
						calls = append(calls, unitSyncStockCall{Name: "protect-after", Token: token, Health: item.HealthData.Cur})
						return in.Status
					},
				})
				result := uintptr(unsafe.Pointer(holders[want.Result.Holder]))
				if want.Result.Status != nil {
					result = uintptr(int64(*want.Result.Status))
				}
				if got != result {
					t.Fatalf("setter result = %#x, want %#x", got, result)
				}
				var protection []unitSyncStockCall
				for _, call := range want.Calls {
					switch call.Name {
					case "protect-before", "protect-player-hp", "protect-after":
						protection = append(protection, call)
					}
				}
				if !slices.Equal(calls, protection) {
					t.Fatalf("protection calls = %+v, want %+v", calls, protection)
				}
			case "unit-on-off":
				target.SetOnOff(in.Arg != 0)
			case "unit-raise":
				target.Raise(math.Float32frombits(uint32(in.Arg)))
			case "unit-anim-frame":
				target.MarkAnimFrame(uint32(in.Arg))
			default:
				t.Fatalf("unexpected original entry %q", in.Entry)
			}
			words := [2]uint32{uint32(health.Field2)<<16 | uint32(health.Cur), uint32(newHealth.Field2)<<16 | uint32(newHealth.Cur)}
			if uint32(target.ObjFlags) != want.Flags || math.Float32bits(target.ZVal) != want.Z ||
				target.Field33 != want.Frame || target.Field38 != want.SyncPending ||
				target.Field140 != want.Masks || words != want.Health {
				t.Fatalf("original state differs: flags=%08x z=%08x frame=%08x sync=%08x health=%08x masks=%08x",
					target.ObjFlags, math.Float32bits(target.ZVal), target.Field33, target.Field38, words, target.Field140)
			}
			if target.InvHolder != holders[want.CurrentHolder] || target.HealthData != healthRecords[want.CurrentHealth] ||
				uint32(target.ObjClass) != want.CurrentClass || target.Field37 != in.Field37 || health.Max != 1000 || newHealth.Max != 2000 {
				t.Fatal("original pointer, class or neighboring fields changed")
			}
		})
	}
}
