package legacy

import "github.com/opennox/opennox/v1/server"

func monsterResolveXferRefs528DB0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	if refs := obj.TakeMonsterActionRefs(); refs != nil {
		applyMonsterXferRefs528DB0(obj, refs)
		return
	}
	resolveLegacyMonsterRefs528DB0(obj)
}

func applyMonsterXferRefs528DB0(obj *server.Object, refs *server.MonsterActionRefs) {
	if obj == nil || obj.UpdateData == nil || refs == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	srv := obj.Server()
	for i := 0; i < int(ud.Field282_1); i++ {
		ud.SeenEnemies[i] = srv.ObjectByScriptID4ECF10(refs.Seen[i])
		if ud.SeenEnemies[i] == nil {
			ud.Field282_1 = 0
		}
	}
	ud.CurrentEnemy = srv.ObjectByScriptID4ECF10(refs.Current)
	if int32(obj.ObjFlags) >= 0 {
		resolveMonsterActionRefs52BAF0(obj, refs)
	}
	if target := srv.ObjectByScriptID4ECF10(refs.Field98); target != nil {
		ud.Field98 = target.NetCode
	} else {
		ud.Field98 = 0
	}
	if target := srv.ObjectByScriptID4ECF10(refs.Field300); target != nil {
		ud.Field300 = target.NetCode
	} else {
		ud.Field300 = 0
	}
	ud.PreferredEnemy = srv.ObjectByScriptID4ECF10(refs.Preferred)
	for i := 0; i < int(ud.Field543_0); i++ {
		if target := srv.ObjectByScriptID4ECF10(refs.Field535[i]); target != nil {
			ud.Field535[i] = target.NetCode
		} else {
			ud.Field543_0 = 0
		}
	}
}
