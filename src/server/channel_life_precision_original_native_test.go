package server

import (
	"encoding/json"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type channelLifeOriginalEvent struct {
	Service  string `json:"service"`
	Fraction uint32 `json:"fraction_bits32"`
	Present  bool   `json:"result_present"`
	Raw32    int32  `json:"amount_raw32"`
	Amount   int32  `json:"amount"`
	Amount16 int16  `json:"amount16"`
}
type channelLifeOriginalStep struct {
	Tick     int                        `json:"tick"`
	Result   int32                      `json:"result"`
	Fraction uint32                     `json:"fraction_bits32"`
	Events   []channelLifeOriginalEvent `json:"services"`
}

func TestChannelLifePrecisionOriginal52F460Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("64-bit native duration/unit/update/health frames")
	}
	var f struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Name        string                    `json:"name"`
			Group       string                    `json:"group"`
			CasterClass uint32                    `json:"caster_class"`
			TargetClass uint32                    `json:"target_class"`
			Level       uint32                    `json:"level"`
			Coefficient uint32                    `json:"coefficient_bits32"`
			Fraction    uint32                    `json:"fraction_bits32"`
			Ticks       int                       `json:"ticks"`
			TargetNil   bool                      `json:"target_nil"`
			CasterNil   bool                      `json:"caster_nil"`
			TargetFlags uint32                    `json:"target_flags"`
			Buffs       uint32                    `json:"caster_buffs"`
			Mana        uint16                    `json:"mana"`
			MaximumMana uint16                    `json:"maximum_mana"`
			HP          uint16                    `json:"caster_HP"`
			Mode        uint32                    `json:"mode"`
			Moved       bool                      `json:"moved"`
			Missing     bool                      `json:"lookup_missing"`
			Steps       []channelLifeOriginalStep `json:"expected_steps"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("testdata/channel_life_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 322 || len(f.Cases) != 322 || unsafe.Sizeof(DurSpell{}) != 184 || unsafe.Sizeof(Object{}) != 928 || unsafe.Sizeof(PlayerUpdateData{}) != 656 || unsafe.Sizeof(HealthData{}) != 20 {
		t.Fatal("incomplete original inputs or changed native layout")
	}
	s := &Server{}
	for _, group := range []string{"stock-one-tick", "stock-sequence", "finite-edges", "early-and-lookup-controls", "class-controls"} {
		t.Run(group, func(t *testing.T) {
			for _, c := range f.Cases {
				if c.Group != group {
					continue
				}
				t.Run(c.Name, func(t *testing.T) {
					sizes := []int{184, 928, 928, 656, 20, 20}
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
					record := (*DurSpell)(unsafe.Pointer(&frames[0][64]))
					caster := (*Object)(unsafe.Pointer(&frames[1][64]))
					target := (*Object)(unsafe.Pointer(&frames[2][64]))
					update := (*PlayerUpdateData)(unsafe.Pointer(&frames[3][64]))
					ch := (*HealthData)(unsafe.Pointer(&frames[4][64]))
					th := (*HealthData)(unsafe.Pointer(&frames[5][64]))
					caster.ObjClass = object.Class(c.CasterClass)
					caster.ObjSubClass = 0
					caster.ObjFlags = 0
					caster.PosVec = types.Ptf(300, 300)
					caster.NewPos = caster.PosVec
					caster.Buffs = c.Buffs
					caster.HealthData = ch
					caster.UpdateData = unsafe.Pointer(update)
					target.ObjClass = object.Class(c.TargetClass)
					target.ObjSubClass = 0
					target.ObjFlags = object.Flags(c.TargetFlags)
					target.PosVec = types.Ptf(300, 300)
					target.NewPos = target.PosVec
					target.Buffs = 0
					target.HealthData = th
					target.UpdateData = unsafe.Pointer(update)
					ch.Cur, ch.Max = c.HP, 50000
					th.Cur, th.Max = 1000, 50000
					update.ManaCur, update.ManaMax = c.Mana, c.MaximumMana
					update.Player = nil
					record.Level = c.Level
					record.Field72 = int32(c.Fraction)
					record.Flag20 = c.Mode
					record.Caster16 = caster
					record.Target48 = target
					record.Pos = types.Ptf(300, 300)
					if c.CasterNil {
						record.Caster16 = nil
					}
					if c.TargetNil {
						record.Target48 = nil
					}
					if c.Moved {
						record.Pos.X = 320
					}
					for i := range frames {
						before[i] = append([]byte(nil), frames[i]...)
					}
					if len(c.Steps) != c.Ticks {
						t.Fatal("incomplete original tick sequence")
					}
					got, want := make([]channelLifeOriginalStep, 0, c.Ticks), make([]channelLifeOriginalStep, 0, c.Ticks)
					for tick, e := range c.Steps {
						events := []channelLifeOriginalEvent{}
						result := s.SpellChannelLifeUpdate52F460(record, SpellChannelLifeRuntime52F460{
							AddMana: func(p *Object, amount int16) {
								if p != target {
									t.Fatal("native mana target changed")
								}
								events = append(events, channelLifeOriginalEvent{Service: "add_mana", Fraction: uint32(record.Field72), Amount16: amount})
							},
							ClearDamage: func(p *Object, amount int32) {
								expected := caster
								if c.Mode != 0 {
									expected = target
								}
								if p != expected {
									t.Fatal("native damage target changed")
								}
								events = append(events, channelLifeOriginalEvent{Service: "clear_damage", Fraction: uint32(record.Field72), Raw32: amount, Amount: amount})
							},
							Coefficient: func(index uint32) float64 {
								if index != c.Level-1 || index >= 5 {
									t.Fatal("native coefficient index changed")
								}
								events = append(events, channelLifeOriginalEvent{Service: "lookup", Fraction: uint32(record.Field72), Present: !c.Missing})
								if c.Missing {
									return 0
								}
								return float64(math.Float32frombits(c.Coefficient))
							},
						})
						got = append(got, channelLifeOriginalStep{Tick: tick, Result: result, Fraction: uint32(record.Field72), Events: events})
						expected := make([]channelLifeOriginalEvent, 0, len(e.Events))
						for _, event := range e.Events {
							if event.Service == "add_mana" {
								if int16(event.Raw32) != event.Amount16 {
									t.Fatal("original int16 effect ABI differs")
								}
								event.Raw32 = 0
							}
							expected = append(expected, event)
						}
						want = append(want, channelLifeOriginalStep{Tick: e.Tick, Result: e.Result, Fraction: e.Fraction, Events: expected})
					}
					offset := 64 + int(unsafe.Offsetof(DurSpell{}.Field72))
					for i, frame := range frames {
						for k := range frame {
							if i == 0 && offset <= k && k < offset+4 {
								continue
							}
							if frame[k] != before[i][k] {
								t.Fatalf("native frame%d nonoutput/canary byte%d changed", i, k)
							}
						}
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("actual native ChannelLife=%+v original=%+v", got, want)
					}
				})
			}
		})
	}
}
