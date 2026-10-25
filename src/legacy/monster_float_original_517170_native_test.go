package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterFloatOriginal517170Native(t *testing.T) {
	raw, err := os.ReadFile("testdata/monster_float_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts    int `json:"contexts"`
		Definitions []struct {
			Name        string `json:"name"`
			Assignments []struct {
				Tag   string `json:"tag"`
				Field string `json:"field"`
				Token string `json:"token"`
			} `json:"float_assignments"`
		} `json:"definitions"`
		Cases []struct {
			Mode     string `json:"mode"`
			Name     string `json:"name"`
			Field    string `json:"field"`
			Expected uint32 `json:"expected_bits32"`
		} `json:"cases"`
		Fields []struct {
			Name   string  `json:"name"`
			Offset uintptr `json:"offset"`
		} `json:"float_field_table"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 1022 || len(fixture.Cases) != 1022 || len(fixture.Definitions) != 73 || len(fixture.Fields) != 7 {
		t.Fatal("incomplete independently executed original contexts")
	}
	var text strings.Builder
	for _, d := range fixture.Definitions {
		fmt.Fprintln(&text, d.Name)
		for _, a := range d.Assignments {
			if a.Tag != "GLOBAL" {
				fmt.Fprintf(&text, "%s ", a.Tag)
			}
			fmt.Fprintf(&text, "%s %s\n", a.Field, a.Token)
		}
		fmt.Fprintln(&text, "END")
	}
	offsets := map[string]uintptr{"RETREAT_RATIO": unsafe.Offsetof(server.MonsterDef{}.RetreatRatio80), "RESUME_RATIO": unsafe.Offsetof(server.MonsterDef{}.ResumeRatio84), "FLEE_RANGE": unsafe.Offsetof(server.MonsterDef{}.FleeRange88), "RUN_MULTIPLIER": unsafe.Offsetof(server.MonsterDef{}.RunMultiplier96), "MELEE_ATTACK_RANGE": unsafe.Offsetof(server.MonsterDef{}.MeleeAttackRange112), "MELEE_ATTACK_IMPACT": unsafe.Offsetof(server.MonsterDef{}.MeleeAttackImpact120), "MISSILE_ATTACK_RANGE": unsafe.Offsetof(server.MonsterDef{}.MissileAttackRange212)}
	allowed := make([]bool, unsafe.Sizeof(server.MonsterDef{}))
	for i := 0; i < 64; i++ {
		allowed[i] = true
	}
	for _, f := range fixture.Fields {
		if offsets[f.Name] != f.Offset {
			t.Fatal("native float offset", f.Name, offsets[f.Name], f.Offset)
		}
		for i := uintptr(0); i < 4; i++ {
			allowed[f.Offset+i] = true
		}
	}
	for i := uintptr(0); i < unsafe.Sizeof(server.MonsterDef{}.Next244); i++ {
		allowed[unsafe.Offsetof(server.MonsterDef{}.Next244)+i] = true
	}
	type frame struct {
		Before [4]uint64
		Def    server.MonsterDef
		After  [4]uint64
	}
	canary := [4]uint64{0x0123456789abcdef, 0xdecafbadf00dcafe, 0xaaaabbbbccccdddd, 0xfedcba9876543210}
	allocated, freed := 0, 0
	for _, mode := range []struct {
		name  string
		flags monsterBinMode
	}{{"SOLO", monsterBinMode{solo: true}}, {"ARENA", monsterBinMode{arena: true}}} {
		frames := map[*server.MonsterDef]*frame{}
		store := newMonsterBinStore(func() (*server.MonsterDef, func()) {
			f, free := alloc.New(frame{})
			f.Before = canary
			f.After = canary
			allocated++
			if uintptr(unsafe.Pointer(f)) <= 0xffffffff {
				t.Fatal("expected actual native record above4GiB")
			}
			frames[&f.Def] = f
			return &f.Def, func() {
				if f.Before != canary || f.After != canary {
					t.Fatal("native frame canary changed")
				}
				freed++
				free()
			}
		})
		t.Cleanup(store.clear)
		if err := store.load(strings.NewReader(text.String()), mode.flags); err != nil {
			t.Fatal(err)
		}
		byName := map[string]*server.MonsterDef{}
		order := []string{}
		for d := store.head; d != nil; d = d.Next244 {
			f, ok := frames[d]
			if !ok || f.Before != canary || f.After != canary {
				t.Fatal("non-native pointer or frame corruption")
			}
			name := string(bytes.TrimRight(d.Name0[:], "\x00"))
			if _, ok := byName[name]; ok {
				t.Fatal("repeated definition", name)
			}
			byName[name] = d
			order = append(order, name)
			var expectedName [64]byte
			copy(expectedName[:], name)
			if d.Name0 != expectedName {
				t.Fatal("name padding changed")
			}
			body := unsafe.Slice((*byte)(unsafe.Pointer(d)), unsafe.Sizeof(*d))
			for i, value := range body {
				if !allowed[i] && value != 0 {
					t.Fatalf("non-float native field changed: name=%s offset=%d byte=%x", name, i, value)
				}
			}
		}
		if len(order) != 73 || len(frames) != 73 {
			t.Fatal("original definition count/order")
		}
		for i, name := range order {
			if name != fixture.Definitions[72-i].Name {
				t.Fatal("original prepend order changed", i, name)
			}
		}
		t.Run(mode.name, func(t *testing.T) {
			for _, c := range fixture.Cases {
				if c.Mode != mode.name {
					continue
				}
				t.Run(c.Name+"/"+c.Field, func(t *testing.T) {
					d := byName[c.Name]
					if d == nil {
						t.Fatal("missing actual native definition", c.Name)
					}
					var got float32
					switch c.Field {
					case "RETREAT_RATIO":
						got = d.RetreatRatio80
					case "RESUME_RATIO":
						got = d.ResumeRatio84
					case "FLEE_RANGE":
						got = d.FleeRange88
					case "RUN_MULTIPLIER":
						got = d.RunMultiplier96
					case "MELEE_ATTACK_RANGE":
						got = d.MeleeAttackRange112
					case "MELEE_ATTACK_IMPACT":
						got = d.MeleeAttackImpact120
					case "MISSILE_ATTACK_RANGE":
						got = d.MissileAttackRange212
					default:
						t.Fatal("unexpected original float field", c.Field)
					}
					if bits := math.Float32bits(got); bits != c.Expected {
						t.Fatalf("actual native monster float=%08x want original=%08x", bits, c.Expected)
					}
				})
			}
		})
		store.clear()
	}
	if allocated != 146 || freed != 146 {
		t.Fatal("native allocation cleanup", allocated, freed)
	}
	after, err := os.ReadFile("testdata/monster_float_original.json")
	if err != nil || sha256.Sum256(raw) != sha256.Sum256(after) {
		t.Fatal("original fixture changed", err)
	}
}
