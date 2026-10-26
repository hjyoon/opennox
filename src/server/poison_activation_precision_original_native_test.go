package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

type poisonPrecisionServiceR83 struct {
	Service string `json:"service"`
	Index   int32  `json:"index"`
	Minimum int32  `json:"minimum"`
	Maximum int32  `json:"maximum"`
	Value   int32  `json:"value"`
}

type poisonPrecisionCaseR83 struct {
	Input struct {
		Name         string  `json:"name"`
		Player       bool    `json:"player"`
		ModifierBits *uint32 `json:"modifier_bits"`
		Buff         bool    `json:"buff"`
		Power        byte    `json:"power"`
		BalanceBits  uint32  `json:"balance_bits"`
		Poison       byte    `json:"poison"`
		Roll         int32   `json:"roll"`
	} `json:"input"`
	ScaledBits  uint32                      `json:"scaled_bits"`
	Threshold   int32                       `json:"threshold"`
	Result      int32                       `json:"result"`
	Services    []poisonPrecisionServiceR83 `json:"services"`
	HealthFrame uint32                      `json:"health_frame"`
	Preserved   bool                        `json:"all_nonoutput_frame_bytes_preserved"`
}

func loadPoisonPrecisionOriginalR83(t *testing.T) struct {
	Original   string                   `json:"original_sha256"`
	Cases      []poisonPrecisionCaseR83 `json:"cases"`
	Historical []struct {
		Mode      string `json:"mode"`
		InputBits uint32 `json:"input_bits"`
		Result    int32  `json:"result"`
	} `json:"historical_conversions"`
} {
	t.Helper()
	var fixture struct {
		Original   string                   `json:"original_sha256"`
		Cases      []poisonPrecisionCaseR83 `json:"cases"`
		Historical []struct {
			Mode      string `json:"mode"`
			InputBits uint32 `json:"input_bits"`
			Result    int32  `json:"result"`
		} `json:"historical_conversions"`
	}
	b, err := os.ReadFile("testdata/poison_activation_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Cases) != 196 || len(fixture.Historical) != 10 {
		t.Fatal("original poison precision fixture identity/count")
	}
	return fixture
}

// The original activation, real inventory protection walk, buff/table getters
// and signed conversion execute together. State/audio/message/RNG/hash remain
// declared service boundaries. Native objects and every referenced record are
// C-owned above 4 GiB; cleanup checks nonoutput bytes even on an old mismatch.
func TestActivatePoisonPrecisionOriginal4EE7E0Native(t *testing.T) {
	fixture := loadPoisonPrecisionOriginalR83(t)
	for _, want := range fixture.Cases {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved {
				t.Fatal("uncertified original frame")
			}
			unit := playerCloudGreatSwordAllocateR78[Object](t)
			item := playerCloudGreatSwordAllocateR78[Object](t)
			init := playerCloudGreatSwordAllocateR78[ModifierInitData](t)
			modifier := playerCloudGreatSwordAllocateR78[ModifierEff](t)
			health := playerCloudGreatSwordAllocateR78[HealthData](t)
			update := playerCloudGreatSwordAllocateR78[PlayerUpdateData](t)
			info := playerCloudGreatSwordAllocateR78[Player](t)
			token := playerCloudGreatSwordAllocateR78[byte](t)
			*unit = Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), HealthData: health, Poison540: want.Input.Poison}
			if want.Input.Player {
				unit.ObjClass = object.ClassPlayer
			}
			*health = HealthData{Cur: 20, Max: 20, Field16: 77}
			*update = PlayerUpdateData{Player: info}
			*item = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
			if want.Input.ModifierBits != nil {
				unit.InvFirstItem = item
				init.Modifiers[0] = modifier
				modifier.Engage112 = unsafe.Pointer(token)
				modifier.EngageFloat120 = math.Float32frombits(*want.Input.ModifierBits)
			}
			if want.Input.Buff {
				unit.Buffs = 1 << 18
			}
			unit.BuffsPower[18] = want.Input.Power
			type frame struct {
				p       unsafe.Pointer
				before  []byte
				allowed []bool
			}
			var frames []frame
			watch := func(p unsafe.Pointer, n uintptr) {
				if p == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("native C frame below 4 GiB")
				}
				frames = append(frames, frame{p, slices.Clone(unsafe.Slice((*byte)(p), int(n))), make([]bool, int(n))})
			}
			watch(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
			watch(unsafe.Pointer(item), unsafe.Sizeof(*item))
			watch(unsafe.Pointer(init), unsafe.Sizeof(*init))
			watch(unsafe.Pointer(modifier), unsafe.Sizeof(*modifier))
			watch(unsafe.Pointer(health), unsafe.Sizeof(*health))
			watch(unsafe.Pointer(update), unsafe.Sizeof(*update))
			watch(unsafe.Pointer(info), unsafe.Sizeof(*info))
			watch(unsafe.Pointer(token), unsafe.Sizeof(*token))
			for i := range frames {
				if frames[i].p == unsafe.Pointer(health) {
					start := unsafe.Offsetof(health.Field16)
					for j := start; j < start+4; j++ {
						frames[i].allowed[j] = true
					}
				}
			}
			t.Cleanup(func() {
				for _, f := range frames {
					for i, value := range unsafe.Slice((*byte)(f.p), len(f.before)) {
						if !f.allowed[i] && f.before[i] != value {
							t.Fatalf("native nonoutput frame changed at %d", i)
						}
					}
				}
				t.Log("all_native_nonoutput_frames_preserved:true")
			})
			var events []poisonPrecisionServiceR83
			var protection float64
			deps := activatePoisonNativeDeps4EE7E0{
				poisonProtection: func(got *Object) float64 {
					if got != unit {
						t.Fatal("protection target")
					}
					protection = poisonProtectionNative4E0040(unit, poisonProtectionNativeDeps4E0040{
						poisonProtectEngage: unsafe.Pointer(token),
						loadBalance: func(key string, index int32) float64 {
							if key != "PoisonSpellProtection" || index != int32(want.Input.Power)-1 {
								t.Fatal("original balance arguments")
							}
							events = append(events, poisonPrecisionServiceR83{Service: "balance", Index: index})
							return float64(math.Float32frombits(want.Input.BalanceBits))
						},
					})
					return protection
				},
				randomInt: func(low, high int32, path string, line int32) int32 {
					if low != 0 || high != 100 || path != activatePoisonRandomPath4EE7E0 || line != 361 {
						t.Fatal("original RNG arguments")
					}
					events = append(events, poisonPrecisionServiceR83{Service: "random", Minimum: low, Maximum: high, Value: want.Input.Roll})
					return want.Input.Roll
				},
				priorityMessage: func(got *Object, message string, value uint8) {
					if got != unit || message != "Health.c:ResistPoison" || value != 0 {
						t.Fatal("original resistance message")
					}
					events = append(events, poisonPrecisionServiceR83{Service: "message"})
				},
				setPoison: func(got *Object, value int32) {
					if got != unit {
						t.Fatal("original setter target")
					}
					events = append(events, poisonPrecisionServiceR83{Service: "set", Value: value})
				},
				audio: func(id uint32, got *Object, kind int32, code uint32) {
					if id != 100 || got != unit || kind != 0 || code != 0 {
						t.Fatal("original poison audio arguments")
					}
					events = append(events, poisonPrecisionServiceR83{Service: "audio", Value: int32(id)})
				},
				frame: func() uint32 { return 42424242 },
			}
			result := activatePoisonNative4EE7E0(unit, 1, 3, deps)
			scaled := activatePoisonScaleProtection4EE7E0(protection)
			threshold := activatePoisonRound4EE7E0(scaled)
			if math.Float32bits(scaled) != want.ScaledBits || threshold != want.Threshold || result != want.Result || health.Field16 != want.HealthFrame || !slices.Equal(events, want.Services) {
				t.Errorf("original poison mismatch scaled=%08x/%08x threshold=%d/%d result=%d/%d frame=%d/%d services=%+v want=%+v", math.Float32bits(scaled), want.ScaledBits, threshold, want.Threshold, result, want.Result, health.Field16, want.HealthFrame, events, want.Services)
			}
		})
	}
}

func TestActivatePoisonConversionOriginal419A70Native(t *testing.T) {
	for _, want := range loadPoisonPrecisionOriginalR83(t).Historical {
		t.Run(fmt.Sprintf("bits-%08x", want.InputBits), func(t *testing.T) {
			if want.Mode != "actual-startup-inherited" {
				t.Fatal("unexpected original FP mode")
			}
			if got := activatePoisonRound4EE7E0(math.Float32frombits(want.InputBits)); got != want.Result {
				t.Fatalf("original signed DWORD conversion=%d want=%d", got, want.Result)
			}
		})
	}
}
