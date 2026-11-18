package legacy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterHealNativeCABI5411A0(t *testing.T) {
	s := playerVitalReportWorld(t)
	noxflags.ResetGame()
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.InitHigh() == 0 || !s.Objs.InitHigh(8) {
		t.Fatal("native world fixture allocation")
	}
	t.Cleanup(s.Walls.Free)
	t.Cleanup(s.Objs.FreeObjects)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	const name = "NativeHealCasterFixture"
	if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "spells.yml"), []byte("- name: SPELL_LESSER_HEAL\n  mana_cost: 17\n"), 0600); err != nil {
		t.Fatal(err)
	}
	oldData, oldSelection := datapath.Data(""), monsterHealSelection5411A0
	t.Cleanup(func() {
		datapath.SetData(oldData)
		monsterHealSelection5411A0 = oldSelection
	})
	datapath.SetData(root)
	mem, _ := alloc.Make([]byte{}, 4)
	file := binfile.NewMemFile(unsafe.Pointer(&mem[0]), len(mem))
	err := s.Spells.Read(file, nil, false)
	file.Free()
	if err != nil {
		t.Fatal(err)
	}
	s.SetFrame(32)

	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			caster, ally := s.NewObjectByTypeID(name), s.NewObjectByTypeID(name)
			if caster == nil || ally == nil {
				t.Fatal("native heal objects")
			}
			defer func() {
				caster.UpdateData, ally.UpdateData = nil, nil
				caster.HealthData, ally.HealthData = nil, nil
				s.Objs.FreeObject(ally)
				s.Objs.FreeObject(caster)
			}()
			data := game4ConsumerValue(t, server.MonsterUpdateData{StatusFlags: object.MonStatusCanCastSpells | object.MonStatusCanHealOthers})
			allyState := game4ConsumerValue(t, server.MonsterUpdateData{})
			health := game4ConsumerValue(t, server.HealthData{Cur: 4, Max: 11})
			caster.ObjClass, ally.ObjClass = object.ClassMonster, object.ClassMonster
			caster.ObjFlags |= object.FlagEnabled
			caster.PosVec, ally.PosVec = types.Ptf(1450, 1720), types.Ptf(1455, 1725)
			caster.UpdateData, ally.UpdateData = unsafe.Pointer(data), unsafe.Pointer(allyState)
			caster.HealthData, ally.HealthData = health, health
			data.AIStack[0].Action = uint32(ai.ACTION_HUNT)

			monsterHealSelection5411A0.Target = ally
			if got := callMonsterHealNative5411A0(route, caster); got != 0 || monsterHealSelection5411A0.Target != nil {
				t.Fatal("empty native scan must clear its target and avoid casting")
			}
			callMonsterHealCandidateNative5412A0(route, ally, caster)
			if monsterHealSelection5411A0.Target != ally {
				t.Fatalf("eligible visible ally was not stored, target=%p ally=%p", monsterHealSelection5411A0.Target, ally)
			}
			data.StatusFlags |= object.MonStatusCanHealSelf
			data.AIStackInd, data.Field132 = -1, uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			if got := callMonsterHealNative5411A0(route, caster); got != 0 || data.AIStackInd != -1 || monsterHealSelection5411A0.Target != ally {
				t.Fatal("empty-head casting sentinel must block the self heal")
			}
			data.AIStackInd, data.Field132 = 0, 0
			if got := callMonsterHealNative5411A0(route, caster); got != 0 || monsterHealSelection5411A0.Target != ally {
				t.Fatal("self heal must return zero and retain the previous ally scratch")
			}
			head := data.AIStackHead()
			if head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgU32(0) != 41 || head.Args[2] != uintptr(unsafe.Pointer(caster)) {
				t.Fatalf("self heal did not schedule spell 41 with its native target, action=%+v", head)
			}
		})
	}
}
