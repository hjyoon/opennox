package server

import (
	"bytes"
	"encoding/json"
	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

type greaterHealOriginalEvent struct {
	Service  string `json:"service"`
	Fraction uint32 `json:"fraction_bits32"`
	Amount   int32  `json:"amount"`
	Result   int32  `json:"result"`
}
type greaterHealOriginalStep struct {
	Tick     int                        `json:"tick"`
	Result   int32                      `json:"result"`
	Fraction uint32                     `json:"fraction_bits32"`
	Events   []greaterHealOriginalEvent `json:"services"`
}

func TestGreaterHealPrecisionOriginal52F2E0Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("64-bit native duration/unit/update/health/player frames")
	}
	var f struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Name        string                    `json:"name"`
			Group       string                    `json:"group"`
			CasterClass uint32                    `json:"caster_class"`
			PlayerClass uint8                     `json:"player_class"`
			Level       uint32                    `json:"level"`
			Coefficient uint32                    `json:"coefficient_bits32"`
			Fraction    uint32                    `json:"fraction_bits32"`
			Multiplier  uint32                    `json:"multiplier_bits32"`
			Ticks       int                       `json:"ticks"`
			TargetNil   bool                      `json:"target_nil"`
			TargetFlags uint32                    `json:"target_flags"`
			Buffs       uint32                    `json:"caster_buffs"`
			Mana        uint16                    `json:"mana"`
			Moved       bool                      `json:"moved"`
			Damaged     bool                      `json:"damaged"`
			FullHealth  bool                      `json:"full_health"`
			CanInteract bool                      `json:"can_interact"`
			Steps       []greaterHealOriginalStep `json:"expected_steps"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("testdata/greater_heal_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 392 || len(f.Cases) != 392 || unsafe.Sizeof(DurSpell{}) != 184 || unsafe.Sizeof(Object{}) != 928 || unsafe.Sizeof(PlayerUpdateData{}) != 656 || unsafe.Sizeof(Player{}) != 6160 || unsafe.Sizeof(HealthData{}) != 20 {
		t.Fatal("incomplete original contexts or changed native layout")
	}
	// Standalone server tests do not load the legacy C coefficient blob.
	const blobBase, blobNeed = 0x587000, 260360 + 4*5 + 4
	if blob := memmap.BlobByAddr(blobBase); blob == nil {
		memmap.RegisterBlobData(blobBase, "greater_heal_original_test", make([]byte, blobNeed))
	} else if len(blob.Data) < blobNeed {
		previous := *blob
		data := make([]byte, blobNeed)
		copy(data, blob.Data)
		blob.Data, blob.Size = data, uintptr(len(data))
		defer func() { *blob = previous }()
	}
	s := &Server{}
	s.Map.Init()
	defer s.Map.Free()
	s.Walls.byPos = make([]*Wall, wallsPerBucket*WallGridSize)
	oldFrame := s.Frame()
	s.SetFrame(100)
	defer s.SetFrame(oldFrame)
	oldMult := s.Players.Mult
	defer func() { s.Players.Mult = oldMult }()
	for _, group := range []string{"stock-one-tick", "stock-sequence", "class-controls", "finite-edges", "early-gates"} {
		t.Run(group, func(t *testing.T) {
			for _, c := range f.Cases {
				if c.Group != group {
					continue
				}
				t.Run(c.Name, func(t *testing.T) {
					sizes := []int{184, 928, 928, 656, 6160, 20, 20}
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
							t.Fatal("frame not above4GiB")
						}
					}
					record := (*DurSpell)(unsafe.Pointer(&frames[0][64]))
					caster := (*Object)(unsafe.Pointer(&frames[1][64]))
					target := (*Object)(unsafe.Pointer(&frames[2][64]))
					update := (*PlayerUpdateData)(unsafe.Pointer(&frames[3][64]))
					pl := (*Player)(unsafe.Pointer(&frames[4][64]))
					ch := (*HealthData)(unsafe.Pointer(&frames[5][64]))
					th := (*HealthData)(unsafe.Pointer(&frames[6][64]))
					caster.ObjClass = object.Class(c.CasterClass)
					caster.ObjSubClass = 0
					caster.ObjFlags = 0
					caster.PosVec = types.Ptf(300, 300)
					caster.NewPos = caster.PosVec
					caster.VelVec = types.Ptf(0, 0)
					caster.Buffs = c.Buffs
					caster.HealthData = ch
					caster.Frame134 = 0
					caster.UpdateData = unsafe.Pointer(update)
					caster.ObjOwner = nil
					if c.Damaged {
						caster.Frame134 = 100
					}
					target.ObjClass = object.ClassPlayer
					target.ObjSubClass = 0
					target.ObjFlags = object.Flags(c.TargetFlags)
					target.PosVec = types.Ptf(310, 310)
					target.NewPos = target.PosVec
					target.VelVec = types.Ptf(0, 0)
					target.Buffs = 0
					target.HealthData = th
					target.ObjOwner = nil
					ch.Cur, ch.Max = 1000, 50000
					th.Cur, th.Max = 1000, 50000
					if c.FullHealth {
						th.Cur = 50000
					}
					update.Player = pl
					update.ManaCur = c.Mana
					pl.Info().SetPlayerClass(playerlib.Class(c.PlayerClass))
					record.Caster16 = caster
					record.Target48 = target
					if c.TargetNil {
						record.Target48 = nil
					}
					record.Level = c.Level
					record.Field72 = int32(c.Fraction)
					record.Pos = types.Ptf(300, 300)
					if c.Moved {
						record.Pos.X = 320
					}
					coeff := memmap.PtrUint32(0x587000, 260360+uintptr(4*c.Level))
					oldCoeff := *coeff
					*coeff = c.Coefficient
					defer func() { *coeff = oldCoeff }()
					m := math.Float32frombits(c.Multiplier)
					s.Players.Mult.Warrior.Health = m
					s.Players.Mult.Wizard.Health = m
					s.Players.Mult.Conjurer.Health = m
					if !c.TargetNil && s.CanInteract(caster, target, 0) != c.CanInteract {
						t.Fatal("declared empty-map visibility boundary differs")
					}
					for i := range frames {
						before[i] = append([]byte(nil), frames[i]...)
					}
					if len(c.Steps) != c.Ticks {
						t.Fatal("incomplete tick sequence")
					}
					got := make([]greaterHealOriginalStep, 0, c.Ticks)
					want := make([]greaterHealOriginalStep, 0, c.Ticks)
					for tick, e := range c.Steps {
						events := []greaterHealOriginalEvent{}
						result := s.SpellGreaterHealUpdate52F2E0(record, SpellGreaterHealRuntime52F2E0{
							AdjustHP: func(p *Object, amount int32) {
								if p != target {
									t.Fatal("native target pointer changed")
								}
								events = append(events, greaterHealOriginalEvent{Service: "adjust_HP", Fraction: uint32(record.Field72), Amount: amount})
							},
							ManaSub: func(p *Object, amount int32) {
								if p != caster || amount != 1 {
									t.Fatal("native caster/mana argument changed")
								}
								events = append(events, greaterHealOriginalEvent{Service: "mana_sub", Fraction: uint32(record.Field72), Amount: amount})
							},
						})
						got = append(got, greaterHealOriginalStep{Tick: tick, Result: result, Fraction: uint32(record.Field72), Events: events})
						filtered := []greaterHealOriginalEvent{}
						for _, event := range e.Events {
							if event.Service != "interact" {
								filtered = append(filtered, event)
							}
						}
						want = append(want, greaterHealOriginalStep{Tick: e.Tick, Result: e.Result, Fraction: e.Fraction, Events: filtered})
					}
					off := 64 + int(unsafe.Offsetof(DurSpell{}.Field72))
					for i, frame := range frames {
						for k := range frame {
							if i == 0 && off <= k && k < off+4 {
								continue
							}
							if frame[k] != before[i][k] {
								t.Fatalf("native frame%d nonoutput/canary byte%d changed", i, k)
							}
						}
					}
					if !bytes.Equal(raw, mustReadGreaterHealFixture(t)) {
						t.Fatal("original fixture changed")
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("actual native GreaterHeal=%+v original=%+v", got, want)
					}
				})
			}
		})
	}
}
func mustReadGreaterHealFixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/greater_heal_precision_original.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
