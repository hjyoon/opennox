package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type barrelProbeObject4E7470 struct {
	id int
}

type barrelProbePoint4E7470 struct {
	x int
	y int
}

func expectedBarrelDrop4E7470(name string, roll int32) (string, int32) {
	table := otherBarrelDrops4E7470[:]
	if barrelUsesPrefixTable4E7470(name) {
		table = barrelPrefixDrops4E7470[:]
	}
	entry := table[selectBarrelDrop4E7470(table, roll)]
	return entry.typeID, entry.count
}

func TestBarrelDropTables4E7470AllOriginalRolls(t *testing.T) {
	tests := []struct {
		name  string
		unit  string
		rolls []struct {
			from, to int32
			typeID   string
			count    int32
		}
	}{
		{
			name: "Barrel prefix table",
			unit: "Barrel",
			rolls: []struct {
				from, to int32
				typeID   string
				count    int32
			}{
				{0, 1, "RedApple", 1},
				{2, 3, "RedApple", 3},
				{4, 5, "Bread", 1},
				{6, 7, "Bread", 3},
				{8, 9, "Corn", 1},
				{10, 11, "Corn", 3},
				{12, 13, "Meat", 1},
				{14, 14, "Meat", 3},
				{15, 15, "Mushroom", 1},
				{16, 17, "SmallSpider", 2},
				{18, 18, "Wasp", 3},
				{19, 19, "ToxicCloud", 1},
				{20, 99, "", 0},
			},
		},
		{
			name: "non-Barrel table",
			unit: "MetalBarrel",
			rolls: []struct {
				from, to int32
				typeID   string
				count    int32
			}{
				{0, 9, "RedApple", 3},
				{10, 19, "Bread", 1},
				{20, 29, "Bread", 3},
				{30, 39, "Corn", 1},
				{40, 49, "Corn", 3},
				{50, 59, "Meat", 1},
				{60, 62, "CurePoisonPotion", 1},
				{63, 65, "BluePotion", 1},
				{66, 68, "RedPotion", 1},
				{69, 71, "LeatherArmor", 1},
				{72, 74, "LeatherHelm", 1},
				{75, 77, "EnchantedMorningStar", 1},
				{78, 80, "EnchantedSword", 1},
				{81, 83, "EnchantedWoodenShield", 1},
				{84, 99, "", 0},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seen := 0
			for _, span := range tc.rolls {
				for roll := span.from; roll <= span.to; roll++ {
					gotID, gotCount := expectedBarrelDrop4E7470(tc.unit, roll)
					if gotID != span.typeID || gotCount != span.count {
						t.Fatalf("roll %d = (%q, %d), want (%q, %d)", roll, gotID, gotCount, span.typeID, span.count)
					}
					seen++
				}
			}
			if seen != 100 {
				t.Fatalf("covered %d rolls, want 100", seen)
			}
		})
	}
}

func TestBarrelPrefix4E7470MatchesSixByteStrncmp(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"Barrel", true},
		{"BarrelWooden", true},
		{"Barrel\x00suffix", true},
		{"Barre", false},
		{"barrel", false},
		{"MetalBarrel", false},
		{"Bar\x00rel", false},
	}
	for _, tc := range tests {
		if got := barrelUsesPrefixTable4E7470(tc.name); got != tc.want {
			t.Errorf("prefix(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestSpawnSomeBarrel4E7470PreservesCallOrderAndNilOwner(t *testing.T) {
	source := &barrelProbeObject4E7470{id: 99}
	created := &barrelProbeObject4E7470{id: 1}
	point := barrelProbePoint4E7470{x: 12, y: 34}
	var events []string
	hooks := barrelSpawnHooks4E7470[*barrelProbeObject4E7470, *barrelProbeObject4E7470, barrelProbePoint4E7470]{
		unitName: func(got *barrelProbeObject4E7470) string {
			if got != source {
				t.Fatalf("source = %p, want %p", got, source)
			}
			events = append(events, "name")
			return "BarrelWooden"
		},
		randomInt: func(min, max int32) int32 {
			events = append(events, fmt.Sprintf("random:%d:%d", min, max))
			return 0
		},
		newObject: func(typeID string) *barrelProbeObject4E7470 {
			events = append(events, "new:"+typeID)
			return created
		},
		randomPoint: func(radius float32) barrelProbePoint4E7470 {
			events = append(events, fmt.Sprintf("point:%g", radius))
			return point
		},
		createAt: func(obj, owner *barrelProbeObject4E7470, gotPoint barrelProbePoint4E7470) {
			if obj != created || owner != nil || gotPoint != point {
				t.Fatalf("create = (%p, %p, %v), want (%p, nil, %v)", obj, owner, gotPoint, created, point)
			}
			events = append(events, "create")
		},
	}

	spawnSomeBarrel4E7470(source, hooks)
	want := []string{"name", "random:0:99", "new:RedApple", "point:35", "create"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpawnSomeBarrel4E7470AllocationFailureDefersPositionRead(t *testing.T) {
	var events []string
	hooks := barrelSpawnHooks4E7470[int, *barrelProbeObject4E7470, barrelProbePoint4E7470]{
		unitName: func(int) string {
			events = append(events, "name")
			return "Barrel"
		},
		randomInt: func(min, max int32) int32 {
			events = append(events, "random")
			return 2 // three RedApple attempts
		},
		newObject: func(typeID string) *barrelProbeObject4E7470 {
			events = append(events, "new:"+typeID)
			return nil
		},
		randomPoint: func(float32) barrelProbePoint4E7470 {
			t.Fatal("position was read after allocation failure")
			return barrelProbePoint4E7470{}
		},
		createAt: func(*barrelProbeObject4E7470, *barrelProbeObject4E7470, barrelProbePoint4E7470) {
			t.Fatal("nil object was created")
		},
	}

	spawnSomeBarrel4E7470(1, hooks)
	want := []string{"name", "random", "new:RedApple", "new:RedApple", "new:RedApple"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpawnSomeBarrel4E7470NoDropStopsAfterRandom(t *testing.T) {
	var events []string
	hooks := barrelSpawnHooks4E7470[int, *barrelProbeObject4E7470, struct{}]{
		unitName: func(int) string {
			events = append(events, "name")
			return "Barrel"
		},
		randomInt: func(min, max int32) int32 {
			events = append(events, "random")
			return 20
		},
		newObject: func(string) *barrelProbeObject4E7470 {
			t.Fatal("no-drop sentinel allocated an object")
			return nil
		},
		randomPoint: func(float32) struct{} { t.Fatal("no-drop sentinel read position"); return struct{}{} },
		createAt: func(*barrelProbeObject4E7470, *barrelProbeObject4E7470, struct{}) {
			t.Fatal("no-drop sentinel created an object")
		},
	}

	spawnSomeBarrel4E7470(1, hooks)
	if want := []string{"name", "random"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpawnSomeBarrel4E7470ReloadsSelectedRecord(t *testing.T) {
	table := []barrelDropEntry4E7470{{typeID: "First", count: 3, threshold: 100}, {}}
	var allocated []string
	hooks := barrelSpawnHooks4E7470[int, *barrelProbeObject4E7470, struct{}]{
		unitName:  func(int) string { return "Barrel" },
		randomInt: func(int32, int32) int32 { return 0 },
		newObject: func(typeID string) *barrelProbeObject4E7470 {
			allocated = append(allocated, typeID)
			if len(allocated) == 1 {
				table[0].typeID = "Reloaded"
				table[0].count = 2
			}
			return nil
		},
		randomPoint: func(float32) struct{} { return struct{}{} },
		createAt:    func(*barrelProbeObject4E7470, *barrelProbeObject4E7470, struct{}) {},
	}

	spawnSomeBarrelWithTables4E7470(1, hooks, table, table)
	if want := []string{"First", "Reloaded"}; !reflect.DeepEqual(allocated, want) {
		t.Fatalf("allocated types = %v, want %v", allocated, want)
	}
}

func TestSpawnSomeBarrel4E7470NilSourceFaultsBeforeRandom(t *testing.T) {
	randomCalled := false
	hooks := barrelSpawnHooks4E7470[*barrelProbeObject4E7470, *barrelProbeObject4E7470, struct{}]{
		unitName:  func(source *barrelProbeObject4E7470) string { return fmt.Sprint(source.id) },
		randomInt: func(int32, int32) int32 { randomCalled = true; return 0 },
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		if randomCalled {
			t.Fatal("random generator was called after the source fault")
		}
	}()
	spawnSomeBarrel4E7470[*barrelProbeObject4E7470, *barrelProbeObject4E7470, struct{}](nil, hooks)
}
