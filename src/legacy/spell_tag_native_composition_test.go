package legacy

import (
	"bytes"
	"encoding/hex"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func tagNativeCompositionAlloc530160[T comparable](t *testing.T, value T) *T {
	t.Helper()
	type frame struct {
		Before [32]byte
		Value  T
		After  [32]byte
	}
	p, free := alloc.New(frame{})
	*p = frame{Value: value}
	t.Cleanup(free)
	for i := range p.Before {
		p.Before[i], p.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	t.Cleanup(func() {
		for i := range p.Before {
			if p.Before[i] != byte(0x51+i) || p.After[i] != byte(0xe7-i) {
				t.Error("native Tag composition canary changed")
			}
		}
	})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatal("native Tag composition below4GiB")
	}
	return &p.Value
}

// Only the final packet transport is observed. The actual C exports, native
// Go binding, balance reader, minimap ring and unit-code helper all execute.
func TestSpellTagNativeCRealMinimapAndPacket530160(t *testing.T) {
	const childEnv = "NOX_TAG_NATIVE_COMPOSITION_CHILD"
	if os.Getenv(childEnv) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestSpellTagNativeCRealMinimapAndPacket530160$", "-test.v")
		cmd.Env = append(os.Environ(), childEnv+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual Tag native composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	fake := &tagNativeEntryServer530160{base: s}
	old := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = old })
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("TagDurationPerLevel: 1.999999999999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	datapath.SetData(dir)
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(s.Balance.Float("TagDurationPerLevel")) != math.Float64bits(1.999999999999) {
		t.Fatal("real binary64 balance load")
	}
	info := s.Players.ByIndRaw(7)
	info.PlayerInd = 7
	caster := tagNativeCompositionAlloc530160(t, server.Object{TypeInd: 713, ObjClass: object.ClassPlayer})
	ud := tagNativeCompositionAlloc530160(t, server.PlayerUpdateData{Player: info})
	caster.UpdateData = unsafe.Pointer(ud)
	record := tagNativeCompositionAlloc530160(t, server.DurSpell{ID: 0xabcd, Spell: 116, Level: 19, Caster16: caster, Field36: 0x13579bdf, Flags88: 0x77556601})
	type sent struct {
		index            int
		packet           []byte
		related          *server.Object
		remove, sequence int
	}
	var packets []sent
	s.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		packets = append(packets, sent{index, append([]byte(nil), packet...), related, remove, sequence})
		return 77
	}
	for _, test := range []struct {
		name            string
		class           object.Class
		kind            uint16
		code            uint32
		initial         uint32
		create, destroy string
	}{
		{"monster-new", object.ClassMonster, 1339, 0x2345, 0, "d245233b050101", "d245233b050201"},
		{"player-retained", object.ClassPlayer, 713, 0x2346, 0, "d24623c9020101", "d24623c9020201"},
		{"monster-existing-marker", object.ClassMonster, 1339, 0x2345, 2, "d245233b050101", "d245233b050201"},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := tagNativeCompositionAlloc530160(t, server.Object{TypeInd: test.kind, ObjClass: test.class, NetCode: test.code, Field5: 0x80})
			if test.class == object.ClassPlayer {
				other := s.Players.ByIndRaw(8)
				other.PlayerInd = 8
				target.UpdateData = unsafe.Pointer(tagNativeCompositionAlloc530160(t, server.PlayerUpdateData{Player: other}))
			}
			if info.Field4580 != nil {
				t.Fatal("prior native minimap ring not released")
			}
			if test.initial != 0 {
				if s.Players.Nox_xxx_netMarkMinimapObject_417190(7, target, test.initial) != 1 {
					t.Fatal("real existing marker")
				}
			}
			record.Target48 = target
			record.Frame68 = 0x87654321
			s.SetFrame(0xfffffff0)
			before := *record
			before.Frame68 = 3
			start := len(packets)
			if got := ccall.CallIntPtr(Get_nox_xxx_spellTagCreature_530160(), record.C()); got != 0 {
				t.Fatal("real C Tag create", got)
			}
			if *record != before {
				t.Fatal("original directed spill/FISTP frame wrap or native record changed", record.Frame68)
			}
			item := info.Field4580
			if item == nil || item.Field4 != target || item.Field0 != test.initial|1 || item.Field8 != item || item.Field12 != item || target.Field5 != 0x81 {
				t.Fatal("actual native minimap mark/ring")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= 0xffffffff {
				t.Fatal("actual minimap item below4GiB")
			}
			if got := ccall.CallIntPtr(Get_sub_530250(), record.C()); got != 0 {
				t.Fatal("real living update", got)
			}
			target.ObjFlags = object.Flags(0x20)
			if got := ccall.CallIntPtr(Get_sub_530250(), record.C()); got != 1 {
				t.Fatal("real dead update", got)
			}
			target.ObjFlags = 0
			ccall.CallVoidPtr(Get_sub_530270(), record.C())
			if *record != before {
				t.Fatal("update/destroy changed native duration bytes")
			}
			if test.class == object.ClassPlayer {
				if info.Field4580 != item || item.Field0 != 1 || target.Field5 != 0x81 {
					t.Fatal("original player-target marker must remain")
				}
				s.Players.Nox_xxx_netUnmarkMinimapObj_417300(7, target, 1)
			} else if test.initial != 0 {
				if info.Field4580 != item || item.Field0 != 2 || target.Field5 != 0x81 {
					t.Fatal("unrelated marker must survive Tag destroy")
				}
				s.Players.Nox_xxx_netUnmarkMinimapObj_417300(7, target, 2)
			} else if info.Field4580 != nil || target.Field5 != 0x80 {
				t.Fatal("real Tag minimap unmark/release")
			}
			if info.Field4580 != nil || target.Field5 != 0x80 {
				t.Fatal("actual minimap cleanup")
			}
			if len(packets)-start != 2 {
				t.Fatal("actual native packet count", len(packets)-start)
			}
			for i, text := range []string{test.create, test.destroy} {
				want, _ := hex.DecodeString(text)
				p := packets[start+i]
				if p.index != 7 || p.related != nil || p.remove != 1 || p.sequence != 0 || !bytes.Equal(p.packet, want) {
					t.Fatalf("native Tag transport packet%v want %x", p, want)
				}
			}
			t.Logf("TAG_NATIVE_COMPOSITION_OK name=%s timer=%d packets=%d marker_flags=%d", test.name, record.Frame68, len(packets)-start, test.initial)
		})
	}
}
