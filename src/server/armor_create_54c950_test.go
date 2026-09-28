package server

import (
	"math"
	"reflect"
	"testing"
)

type armorCreateTestObject54C950 struct {
	typeInd uint16
	health  *armorCreateTestHealth54C950
}

type armorCreateTestDefinition54C950 struct {
	durability uint16
}

type armorCreateTestHealth54C950 struct {
	id      int
	current uint16
	maximum uint16
}

func TestArmorCreate54C950DefinitionGatePrecedesHealth(t *testing.T) {
	obj := &armorCreateTestObject54C950{typeInd: 0x3456}
	var events []string
	armorCreate54C950(obj, armorCreateHooks54C950[
		*armorCreateTestObject54C950,
		*armorCreateTestDefinition54C950,
		*armorCreateTestHealth54C950,
	]{
		loadTypeInd: func(got *armorCreateTestObject54C950) uint16 {
			events = append(events, "type")
			if got != obj {
				t.Fatalf("object = %p, want %p", got, obj)
			}
			return got.typeInd
		},
		findDefinition: func(typeInd uint16) *armorCreateTestDefinition54C950 {
			events = append(events, "definition")
			if typeInd != 0x3456 {
				t.Fatalf("type = %#x, want 0x3456", typeInd)
			}
			return nil
		},
		loadHealth: func(*armorCreateTestObject54C950) *armorCreateTestHealth54C950 {
			t.Fatal("HealthData loaded after nil definition")
			return nil
		},
	})
	if want := []string{"type", "definition"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestArmorCreate54C950HealthGatePrecedesDurability(t *testing.T) {
	obj := &armorCreateTestObject54C950{typeInd: 7}
	definition := &armorCreateTestDefinition54C950{durability: 99}
	var events []string
	armorCreate54C950(obj, armorCreateHooks54C950[
		*armorCreateTestObject54C950,
		*armorCreateTestDefinition54C950,
		*armorCreateTestHealth54C950,
	]{
		loadTypeInd: func(*armorCreateTestObject54C950) uint16 {
			events = append(events, "type")
			return 7
		},
		findDefinition: func(uint16) *armorCreateTestDefinition54C950 {
			events = append(events, "definition")
			return definition
		},
		loadHealth: func(*armorCreateTestObject54C950) *armorCreateTestHealth54C950 {
			events = append(events, "health")
			return nil
		},
		loadDurability: func(*armorCreateTestDefinition54C950) uint16 {
			t.Fatal("durability loaded after nil HealthData")
			return 0
		},
	})
	if want := []string{"type", "definition", "health"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestArmorCreate54C950NonQuestUsesLiveHealthAndDefinition(t *testing.T) {
	first := &armorCreateTestHealth54C950{id: 1, current: 100, maximum: 101}
	second := &armorCreateTestHealth54C950{id: 2, current: 200, maximum: 201}
	obj := &armorCreateTestObject54C950{typeInd: 9, health: first}
	definition := &armorCreateTestDefinition54C950{durability: 11}
	var events []string
	armorCreate54C950(obj, armorCreateHooks54C950[
		*armorCreateTestObject54C950,
		*armorCreateTestDefinition54C950,
		*armorCreateTestHealth54C950,
	]{
		loadTypeInd: func(*armorCreateTestObject54C950) uint16 {
			events = append(events, "type")
			return obj.typeInd
		},
		findDefinition: func(typeInd uint16) *armorCreateTestDefinition54C950 {
			events = append(events, "definition")
			if typeInd != 9 {
				t.Fatalf("type = %d, want 9", typeInd)
			}
			return definition
		},
		loadHealth: func(*armorCreateTestObject54C950) *armorCreateTestHealth54C950 {
			events = append(events, "health")
			return obj.health
		},
		loadDurability: func(got *armorCreateTestDefinition54C950) uint16 {
			events = append(events, "durability")
			value := got.durability
			got.durability = 22
			return value
		},
		storeCurrent: func(health *armorCreateTestHealth54C950, value uint16) {
			events = append(events, "store-current")
			health.current = value
			obj.health = second
		},
		storeMaximum: func(health *armorCreateTestHealth54C950, value uint16) {
			events = append(events, "store-maximum")
			health.maximum = value
		},
		gameFlag: func(flag uint32) int32 {
			events = append(events, "flag")
			if flag != armorCreateQuestFlag54C950 {
				t.Fatalf("flag = %#x, want %#x", flag, armorCreateQuestFlag54C950)
			}
			return 0
		},
		loadBalance: func(string) float32 {
			t.Fatal("non-Quest path loaded the durability multiplier")
			return 0
		},
	})
	wantEvents := []string{
		"type", "definition", "health", "durability", "store-current",
		"health", "durability", "store-maximum", "flag",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if first.current != 11 || first.maximum != 101 {
		t.Fatalf("first health = %d/%d, want 11/101", first.current, first.maximum)
	}
	if second.current != 200 || second.maximum != 22 {
		t.Fatalf("second health = %d/%d, want 200/22", second.current, second.maximum)
	}
}

func TestArmorCreate54C950QuestReloadsHealthAtEveryAccess(t *testing.T) {
	health := []*armorCreateTestHealth54C950{
		{id: 0, current: 100, maximum: 101},
		{id: 1, current: 110, maximum: 111},
		{id: 2, current: 3, maximum: 121},
		{id: 3, current: 130, maximum: 131},
		{id: 4, current: 140, maximum: 5},
		{id: 5, current: 150, maximum: 151},
	}
	obj := &armorCreateTestObject54C950{typeInd: 10}
	definition := &armorCreateTestDefinition54C950{durability: 7}
	loadHealthIndex := 0
	convertIndex := 0
	var events []string
	armorCreate54C950(obj, armorCreateHooks54C950[
		*armorCreateTestObject54C950,
		*armorCreateTestDefinition54C950,
		*armorCreateTestHealth54C950,
	]{
		loadTypeInd: func(*armorCreateTestObject54C950) uint16 {
			events = append(events, "type")
			return obj.typeInd
		},
		findDefinition: func(uint16) *armorCreateTestDefinition54C950 {
			events = append(events, "definition")
			return definition
		},
		loadHealth: func(*armorCreateTestObject54C950) *armorCreateTestHealth54C950 {
			got := health[loadHealthIndex]
			events = append(events, "health:"+string(rune('0'+got.id)))
			loadHealthIndex++
			return got
		},
		loadDurability: func(got *armorCreateTestDefinition54C950) uint16 {
			events = append(events, "durability")
			value := got.durability
			got.durability = 9
			return value
		},
		storeCurrent: func(got *armorCreateTestHealth54C950, value uint16) {
			events = append(events, "store-current:"+string(rune('0'+got.id)))
			got.current = value
		},
		storeMaximum: func(got *armorCreateTestHealth54C950, value uint16) {
			events = append(events, "store-maximum:"+string(rune('0'+got.id)))
			got.maximum = value
		},
		gameFlag: func(flag uint32) int32 {
			events = append(events, "flag")
			if flag != 0x1000 {
				t.Fatalf("flag = %#x, want 0x1000", flag)
			}
			return -7 // every nonzero C int enables the Quest path
		},
		loadBalance: func(key string) float32 {
			events = append(events, "balance")
			if key != armorCreateQuestMultiplier54C950 {
				t.Fatalf("balance key = %q", key)
			}
			return 1.5
		},
		loadCurrent: func(got *armorCreateTestHealth54C950) uint16 {
			events = append(events, "load-current:"+string(rune('0'+got.id)))
			return got.current
		},
		loadMaximum: func(got *armorCreateTestHealth54C950) uint16 {
			events = append(events, "load-maximum:"+string(rune('0'+got.id)))
			return got.maximum
		},
		floatToInt: func(value float32) int32 {
			events = append(events, "convert")
			want := []float32{4.5, 7.5}[convertIndex]
			if value != want {
				t.Fatalf("conversion %d input = %v, want %v", convertIndex, value, want)
			}
			result := []int32{-1, 65537}[convertIndex]
			convertIndex++
			return result
		},
	})
	wantEvents := []string{
		"type", "definition", "health:0", "durability", "store-current:0",
		"health:1", "durability", "store-maximum:1", "flag", "balance",
		"health:2", "load-current:2", "convert", "health:3", "store-current:3",
		"health:4", "load-maximum:4", "convert", "health:5", "store-maximum:5",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if health[0].current != 7 || health[1].maximum != 9 {
		t.Fatalf("initial stores = %d/%d, want 7/9", health[0].current, health[1].maximum)
	}
	if health[3].current != math.MaxUint16 || health[5].maximum != 1 {
		t.Fatalf("scaled low-word stores = %#x/%#x, want 0xffff/0x1", health[3].current, health[5].maximum)
	}
}

func TestArmorCreate54C950DoesNotGuardPostGateHealthReload(t *testing.T) {
	obj := &armorCreateTestObject54C950{typeInd: 1}
	definition := &armorCreateTestDefinition54C950{durability: 2}
	first := new(armorCreateTestHealth54C950)
	loads := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil live HealthData reload did not fault")
		}
	}()
	armorCreate54C950(obj, armorCreateHooks54C950[
		*armorCreateTestObject54C950,
		*armorCreateTestDefinition54C950,
		*armorCreateTestHealth54C950,
	]{
		loadTypeInd: func(*armorCreateTestObject54C950) uint16 { return 1 },
		findDefinition: func(uint16) *armorCreateTestDefinition54C950 {
			return definition
		},
		loadHealth: func(*armorCreateTestObject54C950) *armorCreateTestHealth54C950 {
			loads++
			if loads == 1 {
				return first
			}
			return nil
		},
		loadDurability: func(got *armorCreateTestDefinition54C950) uint16 {
			return got.durability
		},
		storeCurrent: func(got *armorCreateTestHealth54C950, value uint16) {
			got.current = value
		},
		storeMaximum: func(got *armorCreateTestHealth54C950, value uint16) {
			got.maximum = value
		},
	})
}

func TestArmorCreateRoundFloat32ToInt32_54C950(t *testing.T) {
	for _, test := range []struct {
		name  string
		value float32
		want  int32
	}{
		{name: "zero", value: 0},
		{name: "positive tie even", value: 2.5, want: 2},
		{name: "positive tie odd", value: 3.5, want: 4},
		{name: "negative tie even", value: -2.5, want: -2},
		{name: "negative tie odd", value: -3.5, want: -4},
		{name: "largest finite result", value: math.Float32frombits(0x4effffff), want: 2147483520},
		{name: "positive boundary", value: math.Float32frombits(0x4f000000), want: math.MinInt32},
		{name: "negative boundary", value: math.Float32frombits(0xcf000000), want: math.MinInt32},
		{name: "negative out of range", value: math.Float32frombits(0xcf000001), want: math.MinInt32},
		{name: "positive infinity", value: float32(math.Inf(1)), want: math.MinInt32},
		{name: "negative infinity", value: float32(math.Inf(-1)), want: math.MinInt32},
		{name: "nan", value: float32(math.NaN()), want: math.MinInt32},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := armorCreateRoundFloat32ToInt32_54C950(test.value); got != test.want {
				t.Fatalf("round(%08x) = %d, want %d", math.Float32bits(test.value), got, test.want)
			}
		})
	}
}

func TestArmorCreateScale54C950SpillsBinary32(t *testing.T) {
	if got := armorCreateScale54C950(3, 1.5); got != 4.5 {
		t.Fatalf("scale = %v, want 4.5", got)
	}
	value := uint16(65535)
	multiplier := math.Float32frombits(0x3f7fffff)
	want := float32(float64(value) * float64(multiplier))
	if got := armorCreateScale54C950(value, multiplier); math.Float32bits(got) != math.Float32bits(want) {
		t.Fatalf("scale bits = %08x, want %08x", math.Float32bits(got), math.Float32bits(want))
	}
}
