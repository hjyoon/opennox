package opennox

import (
	"encoding/json"
	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type lesserHealOriginalEvent struct {
	Service string `json:"service"`
	Object  string `json:"object"`
	Present bool   `json:"result_present"`
	Bits    uint32 `json:"bits32"`
	Amount  int32  `json:"amount"`
	SpellID int    `json:"spell_id"`
	Level   int    `json:"level"`
	Sound   int    `json:"sound"`
}

func TestLesserHealPrecisionOriginal52DD50NativeFrames(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("64-bit full native acceptance/unit/player/update/health frames")
	}
	var f struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Name         string                    `json:"name"`
			Group        string                    `json:"group"`
			Coefficient  uint32                    `json:"coefficient_bits32"`
			Multiplier   uint32                    `json:"multiplier_bits32"`
			CasterClass  uint32                    `json:"caster_class"`
			PlayerClass  uint8                     `json:"player_class"`
			TargetNil    bool                      `json:"target_nil"`
			CasterNil    bool                      `json:"caster_nil"`
			SecondTarget bool                      `json:"second_is_target"`
			HP           uint16                    `json:"HP"`
			MaxHP        uint16                    `json:"max_HP"`
			HealthNil    bool                      `json:"health_nil"`
			Missing      bool                      `json:"lookup_missing"`
			SpellID      int                       `json:"spell_id"`
			Result       int                       `json:"expected_result"`
			Events       []lesserHealOriginalEvent `json:"expected_services"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("testdata/lesser_heal_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 406 || len(f.Cases) != 406 || unsafe.Sizeof(server.SpellAcceptArg{}) != 16 || unsafe.Sizeof(server.Object{}) != 928 || unsafe.Sizeof(server.PlayerUpdateData{}) != 656 || unsafe.Sizeof(server.Player{}) != 6160 || unsafe.Sizeof(server.HealthData{}) != 20 {
		t.Fatal("incomplete original inputs or changed full native layout")
	}
	s := &server.Server{}
	for _, group := range []string{"stock-derived", "finite-edges", "class-controls", "early-lookup-controls"} {
		t.Run(group, func(t *testing.T) {
			for _, c := range f.Cases {
				if c.Group != group {
					continue
				}
				t.Run(c.Name, func(t *testing.T) {
					sizes := []int{16, 928, 928, 656, 6160, 20}
					frames := make([][]byte, len(sizes))
					before := make([][]byte, len(sizes))
					for i, n := range sizes {
						frame, free := alloc.Make([]byte(nil), n+128)
						defer free()
						for k := range frame {
							frame[k] = 0xa5
						}
						frames[i] = frame
						if uintptr(unsafe.Pointer(&frame[64])) <= 0xffffffff {
							t.Fatal("native frame not above4GiB")
						}
					}
					arg := (*server.SpellAcceptArg)(unsafe.Pointer(&frames[0][64]))
					target := (*server.Object)(unsafe.Pointer(&frames[1][64]))
					caster := (*server.Object)(unsafe.Pointer(&frames[2][64]))
					update := (*server.PlayerUpdateData)(unsafe.Pointer(&frames[3][64]))
					pl := (*server.Player)(unsafe.Pointer(&frames[4][64]))
					hp := (*server.HealthData)(unsafe.Pointer(&frames[5][64]))
					arg.Obj = target
					if c.TargetNil {
						arg.Obj = nil
					}
					target.HealthData = hp
					if c.HealthNil {
						target.HealthData = nil
					}
					hp.Cur, hp.Max = c.HP, c.MaxHP
					caster.ObjClass = object.Class(c.CasterClass)
					caster.UpdateData = unsafe.Pointer(update)
					update.Player = pl
					pl.Info().SetPlayerClass(playerlib.Class(c.PlayerClass))
					m := math.Float32frombits(c.Multiplier)
					s.Players.Mult.Warrior.Health = m
					s.Players.Mult.Wizard.Health = m
					s.Players.Mult.Conjurer.Health = m
					for i := range frames {
						before[i] = append([]byte(nil), frames[i]...)
					}
					label := func(p *server.Object) string {
						if p == nil {
							return "nil"
						}
						if p == target {
							return "target"
						}
						if p == caster {
							return "caster"
						}
						t.Fatal("unknown native object pointer")
						return ""
					}
					events := []lesserHealOriginalEvent{}
					second, third := target, caster
					if !c.SecondTarget {
						second = caster
					}
					if c.CasterNil {
						third = nil
					}
					result := lesserHeal52DD50(spell.ID(c.SpellID), second, third, arg, lesserHealHooks52DD50[*server.Object, *server.SpellAcceptArg]{
						loadTarget: func(a *server.SpellAcceptArg) *server.Object {
							if a != arg {
								t.Fatal("acceptance identity changed")
							}
							events = append(events, lesserHealOriginalEvent{Service: "target_load", Object: label(a.Obj)})
							return a.Obj
						},
						getHP: func(p *server.Object) uint16 {
							events = append(events, lesserHealOriginalEvent{Service: "HP", Object: label(p)})
							return server.UnitGetHP4EE780(p)
						},
						getMaxHP: func(p *server.Object) uint16 {
							events = append(events, lesserHealOriginalEvent{Service: "max_HP", Object: label(p)})
							return server.UnitGetMaxHP4EE7A0(p)
						},
						manaCost: func(id spell.ID, level int) int {
							events = append(events, lesserHealOriginalEvent{Service: "mana_cost", SpellID: int(id), Level: level})
							return 17
						},
						refundMana: func(p *server.Object, amount int) {
							events = append(events, lesserHealOriginalEvent{Service: "refund", Object: label(p), Amount: int32(amount)})
						},
						balance: func(key string) float32 {
							if key != "LesserHealAmount" {
								t.Fatal("coefficient key changed")
							}
							events = append(events, lesserHealOriginalEvent{Service: "lookup", Present: !c.Missing})
							if c.Missing {
								return 0
							}
							return math.Float32frombits(c.Coefficient)
						},
						loadClassLow:  func(p *server.Object) uint8 { return uint8(p.ObjClass) },
						loadPlayerCls: func(p *server.Object) uint8 { return uint8(p.UpdateDataPlayer().Player.PlayerClass()) },
						classHealth:   func(class uint8) float32 { return s.Players.ClassStatsMult(playerlib.Class(class)).Health },
						floatToInt: func(v float32) int32 {
							events = append(events, lesserHealOriginalEvent{Service: "round", Bits: math.Float32bits(v)})
							return lesserHealFloatToInt52DD50(v)
						},
						adjustHP: func(p *server.Object, amount int32) {
							events = append(events, lesserHealOriginalEvent{Service: "adjust_HP", Object: label(p), Amount: amount})
						},
						audio: func(id spell.ID, p *server.Object) {
							events = append(events, lesserHealOriginalEvent{Service: "sound_id", SpellID: int(id), Level: 1}, lesserHealOriginalEvent{Service: "audio", Sound: 0x1234, Object: label(p)})
						},
					})
					for i := range frames {
						if !reflect.DeepEqual(frames[i], before[i]) {
							t.Fatalf("native full input/shape/nonoutput/canary frame%d changed", i)
						}
					}
					if result != c.Result {
						t.Errorf("actual native LesserHeal return=%d original=%d", result, c.Result)
					}
					numeric := func(es []lesserHealOriginalEvent) []lesserHealOriginalEvent {
						var out []lesserHealOriginalEvent
						for _, e := range es {
							if e.Service == "round" || e.Service == "adjust_HP" {
								out = append(out, e)
							}
						}
						return out
					}
					if got, want := numeric(events), numeric(c.Events); !reflect.DeepEqual(got, want) {
						t.Errorf("actual native LesserHeal numeric=%+v original=%+v", got, want)
					}
					if !reflect.DeepEqual(events, c.Events) {
						t.Errorf("actual native LesserHeal target/order/events=%+v original=%+v", events, c.Events)
					}
				})
			}
		})
	}
}
