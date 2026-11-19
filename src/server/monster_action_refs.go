package server

// MonsterActionRefs holds serialized script IDs and waypoint indices until
// map restore can resolve them to live objects.
type MonsterActionRefs struct {
	Seen        [16]int32
	Current     int32
	Preferred   int32
	Field98     int32
	Field300    int32
	Field535    [8]int32
	AIObjects   [24][2]int32
	AIWaypoints [24][2]uint32
}

func (obj *Object) StoreMonsterActionRefs(refs *MonsterActionRefs) {
	obj.Server().Objs.monsterActionRefs.Store(obj, refs)
}

func (obj *Object) TakeMonsterActionRefs() *MonsterActionRefs {
	srv := obj.Server()
	if srv == nil {
		return nil
	}
	value, _ := srv.Objs.monsterActionRefs.LoadAndDelete(obj)
	refs, _ := value.(*MonsterActionRefs)
	return refs
}
