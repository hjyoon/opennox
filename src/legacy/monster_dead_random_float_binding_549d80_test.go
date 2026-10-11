package legacy

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type monsterDeadFloatPart549D80 struct {
	object   *server.Object
	position types.Pointf
	index    int
}

// Only world placement/force are declared services. The production closure,
// native allocator/type factory, map search, both RNGs, Raise and Decay run.
type monsterDeadFloatLegacyServer549D80 struct {
	Server
	native *server.Server
	parts  []monsterDeadFloatPart549D80
}

func (s *monsterDeadFloatLegacyServer549D80) S() *server.Server { return s.native }

func (s *monsterDeadFloatLegacyServer549D80) CreateObjectAt(obj, owner server.Obj, point types.Pointf) {
	if server.ToObject(owner) != nil {
		panic("death debris unexpectedly acquired an owner")
	}
	s.parts = append(s.parts, monsterDeadFloatPart549D80{
		object: server.ToObject(obj), position: point, index: s.native.Rand.Logic.Index(),
	})
}

func (s *monsterDeadFloatLegacyServer549D80) ApplyForce(*server.Object, types.Pointf, float64) {
	panic("mechanical debris unexpectedly requested radial force")
}

// Independent x87 instruction arithmetic followed by the caller's FSTPS.
// Neither the production RNG nor its narrowing helper supplies an expectation.
func monsterDeadFloatReference549D80(min, max float32, word int) float32 {
	arithmetic := func(a, b float64, operation byte) float64 {
		x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
		y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
		switch operation {
		case '*':
			x.Mul(x, y)
		default:
			x.Add(x, y)
		}
		value, _ := x.Float64()
		return value
	}
	delta := arithmetic(float64(max), -float64(min), '+')
	scale := arithmetic(float64(word), float64(math.Float32frombits(0x38000100)), '*')
	value := arithmetic(arithmetic(delta, scale, '*'), float64(min), '+')
	narrow, _ := new(big.Float).SetPrec(24).SetMode(big.ToZero).SetFloat64(value).Float32()
	return narrow
}

func TestMonsterDeadRandomFloat549D80ActualRuntimeBinding(t *testing.T) {
	for _, start := range []int{0, 4090} {
		t.Run(fmt.Sprintf("Logic-%d", start), func(t *testing.T) {
			s := server.New(nil, nil, strman.New())
			t.Cleanup(s.Close)
			s.FreeObjectTypes()
			t.Cleanup(s.FreeObjectTypes)
			for _, name := range []string{"MechGear", "MechHead", "MechArm", "MechLeg"} {
				if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
					t.Fatal(err)
				}
			}
			if !s.Objs.Init(64) {
				t.Fatal("C-owned allocator initialization failed")
			}
			t.Cleanup(s.Objs.FreeObjects)
			s.Map.Init()
			t.Cleanup(s.Map.Free)
			s.Walls.Init()
			t.Cleanup(s.Walls.Free)
			s.Audio.Init(s)
			t.Cleanup(s.Audio.Free)
			t.Cleanup(s.DecayDestroy5117B0)
			s.SetFrame(1400)
			s.SetTickRate(30)
			s.Rand.Logic, s.Rand.Other = prand.New(start), prand.New(17)
			previousFlags := noxflags.GetGame()
			noxflags.ResetGame()
			t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(previousFlags) })
			outer := &monsterDeadFloatLegacyServer549D80{native: s}
			previousGet := GetServer
			GetServer = func() Server { return outer }
			t.Cleanup(func() { GetServer = previousGet })
			unit := s.Objs.NewObject(&server.ObjectType{})
			update, freeUpdate := alloc.New(server.MonsterUpdateData{Field67: 67, Field124: 124})
			health, freeHealth := alloc.New(server.HealthData{Cur: 19, Max: 80, Field2: 80})
			t.Cleanup(freeUpdate)
			t.Cleanup(freeHealth)
			unit.ObjClass, unit.PosVec = object.ClassMonster, types.Ptf(300, 300)
			unit.UpdateData, unit.HealthData = unsafe.Pointer(update), health
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(health)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("C-owned native pointer=%p, want >4 GiB", pointer)
				}
			}
			beforeUpdate, beforeHealth := *update, *health
			runtime := monsterActionDeadRuntime544D80()
			if !runtime.CanDeadFunc(monsterDeadFunctions["MECHGOLEMDEAD"]) {
				t.Fatal("real monster.bin DEAD_FUNCTION binding was rejected")
			}
			runtime.DeadFunc(monsterDeadFunctions["MECHGOLEMDEAD"], unit)
			if len(outer.parts) != 8 {
				t.Fatalf("production callback created %d parts, want 8", len(outer.parts))
			}
			table := prand.New(start)
			for i, part := range outer.parts {
				table.Int(0, 0x7fff) // real RandomReachablePoint consumes one angle
				if part.index != table.Index() {
					t.Fatalf("part %d float boundary index=%d want %d", i, part.index, table.Index())
				}
				height := monsterDeadFloatReference549D80(10, 70, table.Int(0, 0x7fff))
				velocity := monsterDeadFloatReference549D80(-2, 0, table.Int(0, 0x7fff))
				seconds := table.Int(10, 20)
				if math.Float32bits(part.object.ZVal) != math.Float32bits(height) ||
					math.Float32bits(part.object.Field27) != math.Float32bits(velocity) {
					t.Errorf("part %d actual production height/velocity=%08x/%08x want original FSTPS=%08x/%08x", i,
						math.Float32bits(part.object.ZVal), math.Float32bits(part.object.Field27), math.Float32bits(height), math.Float32bits(velocity))
				}
				if part.object.Field34 != 1400+30*uint32(seconds) || part.object.Field29 != math.Float32bits(2) ||
					!part.object.Flags().Has(object.FlagBouncy|object.Flags(0x00400000)) {
					t.Fatal("production debris lost its real lifetime/bounce fields")
				}
				if math.IsNaN(float64(part.position.X)) || math.IsNaN(float64(part.position.Y)) ||
					math.Hypot(float64(part.position.X-300), float64(part.position.Y-300)) > 30.001 {
					t.Fatalf("actual reachable position=%v", part.position)
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(part.object)) <= math.MaxUint32 {
					t.Fatalf("part %d native pointer=%p, want >4 GiB", i, part.object)
				}
			}
			if s.Rand.Logic.Index() != table.Index() || s.Rand.Other.Index() != 17 || s.Frame() != 1400 ||
				*health != beforeHealth || *update != beforeUpdate {
				t.Fatal("debris altered RNG consumption, Other stream, frame, HP or AI data")
			}
			t.Logf("actual DEAD_FUNCTION binding: C-owned unit=%p parts=8 Logic=%d->%d Other=17 frame=1400 HP=19", unit, start, table.Index())
		})
	}
}
