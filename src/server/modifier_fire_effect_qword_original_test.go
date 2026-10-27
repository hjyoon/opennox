package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// The fixture executes the complete, unchanged original FireEffect and __ftol
// bodies. Damage acceptance, FX transport and audio are declared boundaries.
// This test calls the actual server method, damage registry, FX queue and audio
// emitter, including a callback that changes the cached effect and live target.
func TestModifierFireEffectQwordOriginal4E0550(t *testing.T) {
	blob, err := os.ReadFile("testdata/fire_effect_qword_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type observation struct {
		Service   string   `json:"service"`
		Damage    uint32   `json:"damage_bits32,omitempty"`
		Type      int      `json:"damage_type,omitempty"`
		Position  []uint32 `json:"position_bits32,omitempty"`
		Intensity uint8    `json:"intensity,omitempty"`
		Sound     int      `json:"sound,omitempty"`
		Kind      int      `json:"kind,omitempty"`
		Code      uint32   `json:"code,omitempty"`
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"original_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Name      string `json:"name"`
				Value     uint32 `json:"coefficient_bits32"`
				Accepted  bool   `json:"accepted"`
				Mutate    bool   `json:"mutate"`
				NilTarget bool   `json:"nil_target"`
			} `json:"input"`
			Mode      string        `json:"mode"`
			Events    []observation `json:"events"`
			Preserved bool          `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Faults) != 0 || len(fixture.Rows) != 148 {
		t.Fatal("incomplete independent original FireEffect observations")
	}
	wasSuppressed := noxflags.HasGame(noxflags.GameFlag20)
	noxflags.UnsetGame(noxflags.GameFlag20)
	t.Cleanup(func() {
		if wasSuppressed {
			noxflags.SetGame(noxflags.GameFlag20)
		}
	})
	callback := unsafe.Pointer(selfWeaponElectricAllocOriginal(t, byte(0)))
	var handler DamageFunc
	RegisterObjectDamageGo(t.Name(), callback, func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
		if handler == nil {
			return false
		}
		return handler(target, source, weapon, damage, typ)
	})
	t.Cleanup(func() { handler = nil; delete(damageFuncs, t.Name()) })
	for _, row := range fixture.Rows {
		t.Run(row.Mode+"/"+row.Input.Name, func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original changed bytes outside the declared services")
			}
			effect := selfWeaponElectricAllocOriginal(t, ModifierEff{AttackPreHit52: ModifierEffFnc{Valf: math.Float32frombits(row.Input.Value), Val: 0x12345678}})
			source := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 713, ObjClass: object.ClassPlayer})
			weapon := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 500, ObjClass: object.ClassWeapon})
			target := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 1343, ObjClass: object.ClassMonster, Damage: callback, PosVec: types.Ptf(12.75, 19.25)})
			viewer := selfWeaponElectricAllocOriginal(t, Object{ObjClass: object.ClassPlayer})
			player := selfWeaponElectricAllocOriginal(t, Player{Active: 1, PlayerInd: 0, Field10: 1000, Field12: 1000, PlayerUnit: viewer})
			update := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: player})
			viewer.UpdateData = unsafe.Pointer(update)
			if unsafe.Sizeof(uintptr(0)) > 4 {
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(effect), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(target), unsafe.Pointer(viewer), unsafe.Pointer(player), unsafe.Pointer(update), callback} {
					if uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("expected C-owned native pointer above 4 GiB: %p", pointer)
					}
				}
			}
			s := &Server{NetList: netlist.New()}
			s.NetList.Init()
			t.Cleanup(s.NetList.Free)
			s.Players.list = []Player{*player}
			s.Audio.Init(s)
			t.Cleanup(s.Audio.Free)
			s.Audio.bySound[sound.SoundWeaponEffectFire].Field12 = 1
			s.Audio.Reset()
			effectBytes := unsafe.Slice((*byte)(unsafe.Pointer(effect)), int(unsafe.Sizeof(*effect)))
			wantEffect := bytes.Clone(effectBytes)
			wantTarget, wantSource, wantWeapon, wantViewer, wantPlayer, wantUpdate := *target, *source, *weapon, *viewer, *player, *update
			if row.Input.Mutate {
				offset := unsafe.Offsetof(effect.AttackPreHit52) + unsafe.Offsetof(effect.AttackPreHit52.Valf)
				binary.LittleEndian.PutUint32(wantEffect[offset:], math.Float32bits(1))
				wantTarget.PosVec = types.Ptf(77.5, 88.25)
			}
			got := make([]observation, 0)
			handler = func(gotTarget, gotSource, gotWeapon *Object, damage int32, typ object.DamageType) bool {
				if gotTarget != target || gotSource != source || gotWeapon != weapon {
					t.Fatal("damage callback lost native argument identity")
				}
				if s.NetList.ByInd(0, netlist.Kind1).Count() != 0 || s.Audio.head != nil {
					t.Fatal("FX or audio emitted before damage")
				}
				got = append(got, observation{Service: "damage", Damage: uint32(damage), Type: int(typ)})
				if row.Input.Mutate {
					effect.AttackPreHit52.Valf = 1
					target.PosVec = types.Ptf(77.5, 88.25)
				}
				return row.Input.Accepted
			}
			s.Audio.OnSound(func(id sound.ID, kind int, obj *Object, pos types.Pointf) {
				if obj != target || pos != target.PosVec || s.Audio.head == nil {
					t.Fatal("object audio identity, live position or emission changed")
				}
				packet := s.NetList.ByInd(0, netlist.Kind1).Get()
				if len(packet) != 6 || packet[0] != byte(netmsg.MSG_FX_SPARK_EXPLOSION) || binary.LittleEndian.Uint16(packet[1:]) != uint16(pos.X) || binary.LittleEndian.Uint16(packet[3:]) != uint16(pos.Y) {
					t.Fatalf("spark packet before audio = %x", packet)
				}
				got = append(got, observation{Service: "spark", Position: []uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}, Intensity: packet[5]})
				got = append(got, observation{Service: "audio", Sound: int(id), Kind: kind})
			})
			argument := target
			if row.Input.NilTarget {
				argument = nil
			}
			s.Nox_xxx_fireEffect_4E0550(unsafe.Pointer(effect), weapon, source, argument)
			if !reflect.DeepEqual(got, row.Events) {
				t.Errorf("native observations = %+v, original = %+v", got, row.Events)
			}
			audioCount := 0
			s.Audio.EachEvent(func(event *AudioEvent) {
				audioCount++
				if event.Sound != sound.SoundWeaponEffectFire || event.Obj != target || event.Pos != target.PosVec || event.Kind != 0 || event.Code != 0 {
					t.Fatalf("audio event = %+v", event)
				}
			})
			wantAudioCount := 1
			if row.Input.NilTarget {
				wantAudioCount = 0
			}
			if audioCount != wantAudioCount || s.NetList.ByInd(0, netlist.Kind1).Count() != 0 {
				t.Fatal("unexpected extra or missing audio/FX event")
			}
			if !bytes.Equal(effectBytes, wantEffect) || *target != wantTarget || *source != wantSource || *weapon != wantWeapon || *viewer != wantViewer || *player != wantPlayer || *update != wantUpdate {
				t.Fatal("native fixture bytes changed outside the declared damage service")
			}
			handler = nil
		})
	}
}
