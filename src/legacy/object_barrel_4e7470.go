package legacy

const barrelSpawnRadius4E7470 float32 = 35

type barrelDropEntry4E7470 struct {
	typeID    string
	count     int32
	threshold int32
}

// These records are the native-width form of the two 12-byte tables at
// 005B8948 and 005B89E8. GAME.EXE chooses the first record whose threshold is
// strictly greater than a random integer in [0, 99]. A zero type terminates
// the table and represents the no-drop result.
var barrelPrefixDrops4E7470 = [...]barrelDropEntry4E7470{
	{typeID: "RedApple", count: 1, threshold: 2},
	{typeID: "RedApple", count: 3, threshold: 4},
	{typeID: "Bread", count: 1, threshold: 6},
	{typeID: "Bread", count: 3, threshold: 8},
	{typeID: "Corn", count: 1, threshold: 10},
	{typeID: "Corn", count: 3, threshold: 12},
	{typeID: "Meat", count: 1, threshold: 14},
	{typeID: "Meat", count: 3, threshold: 15},
	{typeID: "Mushroom", count: 1, threshold: 16},
	{typeID: "SmallSpider", count: 2, threshold: 18},
	{typeID: "Wasp", count: 3, threshold: 19},
	{typeID: "ToxicCloud", count: 1, threshold: 20},
	{},
}

var otherBarrelDrops4E7470 = [...]barrelDropEntry4E7470{
	{typeID: "RedApple", count: 3, threshold: 10},
	{typeID: "Bread", count: 1, threshold: 20},
	{typeID: "Bread", count: 3, threshold: 30},
	{typeID: "Corn", count: 1, threshold: 40},
	{typeID: "Corn", count: 3, threshold: 50},
	{typeID: "Meat", count: 1, threshold: 60},
	{typeID: "CurePoisonPotion", count: 1, threshold: 63},
	{typeID: "BluePotion", count: 1, threshold: 66},
	{typeID: "RedPotion", count: 1, threshold: 69},
	{typeID: "LeatherArmor", count: 1, threshold: 72},
	{typeID: "LeatherHelm", count: 1, threshold: 75},
	{typeID: "EnchantedMorningStar", count: 1, threshold: 78},
	{typeID: "EnchantedSword", count: 1, threshold: 81},
	{typeID: "EnchantedWoodenShield", count: 1, threshold: 84},
	{},
}

type barrelSpawnHooks4E7470[S any, O comparable, P any] struct {
	unitName    func(S) string
	randomInt   func(min, max int32) int32
	newObject   func(typeID string) O
	randomPoint func(radius float32) P
	createAt    func(obj, owner O, pos P)
}

func barrelUsesPrefixTable4E7470(name string) bool {
	// Equivalent to strncmp(name, "Barrel", 6) == 0 for a decoded C string.
	return len(name) >= len("Barrel") && name[:len("Barrel")] == "Barrel"
}

func selectBarrelDrop4E7470(table []barrelDropEntry4E7470, roll int32) int {
	index := 0
	if table[index].typeID != "" {
		for {
			if table[index].threshold > roll {
				break
			}
			index++
			if table[index].typeID == "" {
				break
			}
		}
	}
	return index
}

func spawnSomeBarrelWithTables4E7470[S any, O comparable, P any](
	source S,
	hooks barrelSpawnHooks4E7470[S, O, P],
	prefixTable, otherTable []barrelDropEntry4E7470,
) {
	name := hooks.unitName(source)
	table := otherTable
	if barrelUsesPrefixTable4E7470(name) {
		table = prefixTable
	}
	roll := hooks.randomInt(0, 99)
	selected := &table[selectBarrelDrop4E7470(table, roll)]
	if selected.typeID == "" || selected.count <= 0 {
		return
	}

	var nilObject O
	for i := int32(0); i < selected.count; i++ {
		// GAME.EXE reloads both fields from the selected record: type before
		// each allocation and count after every iteration.
		obj := hooks.newObject(selected.typeID)
		if obj != nilObject {
			pos := hooks.randomPoint(barrelSpawnRadius4E7470)
			hooks.createAt(obj, nilObject, pos)
		}
	}
}

func spawnSomeBarrel4E7470[S any, O comparable, P any](source S, hooks barrelSpawnHooks4E7470[S, O, P]) {
	spawnSomeBarrelWithTables4E7470(
		source,
		hooks,
		barrelPrefixDrops4E7470[:],
		otherBarrelDrops4E7470[:],
	)
}
