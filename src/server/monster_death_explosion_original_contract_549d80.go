package server

// MonsterDeathExplosionRuntime549D80 names the seven external operations used
// by the original 549D80, 549E00, 549E70 and 54A250 bodies. A point owner names
// its live PosVec allocation, never a copied point. Native C service adapters
// derive that field address using the native C layout at the call boundary.
// This also leaves original nil/unmapped addresses to the declared world
// operation instead of adding an earlier Go field dereference.
type MonsterDeathExplosionRuntime549D80 struct {
	GameFlag func(mask uint32) uint32
	Push     func(pointOwner *Object, outer, inner, force float32, source *Object, callback, callbackArg uintptr) uint32
	Damage   func(pointOwner *Object, outer, inner float32, damage, damageType uint32, source, exclude *Object) uint32
	Spark    func(pointOwner *Object, code uint32) uint32
	Audio    func(id uint32, source *Object, kind, arg uint32) uint32
	Delete   func(*Object) uint32
	Point    func(code uint32, pointOwner *Object) uint32
}
