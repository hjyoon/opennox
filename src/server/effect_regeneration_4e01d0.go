package server

type regenerationRuntime4E01D0 struct {
	frame, fps func() uint32
	maxHP, hp  func(*Object) uint16
	armorFlags func(*Object) uint32
	adjustHP   func(*Object, int32)
}

// EffectRegeneration4E01D0 restores the equipped item's update callback.
// GAME.EXE reads the item's inventory holder, not the callback item as a unit.
// Frame/rate arithmetic remains unsigned DWORD with the original reloads.
func (s *Server) EffectRegeneration4E01D0(effect *ModifierEff, item *Object, adjustHP func(*Object, int32)) {
	effectRegeneration4E01D0(effect, item, regenerationRuntime4E01D0{
		frame: s.Frame, fps: s.TickRate,
		maxHP: UnitGetMaxHP4EE7A0, hp: UnitGetHP4EE780,
		armorFlags: s.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70,
		adjustHP:   adjustHP,
	})
}

func effectRegeneration4E01D0(effect *ModifierEff, item *Object, r regenerationRuntime4E01D0) {
	if item == nil {
		return
	}
	owner := item.InvHolder
	if owner == nil || owner.HealthData == nil {
		return
	}
	frame := r.frame()
	last := owner.Frame134
	fps := r.fps()
	if frame-last < fps || uint32(owner.ObjFlags)&0x8020 != 0 {
		return
	}
	maximum := r.maxHP(owner)
	if r.hp(owner) >= maximum {
		return
	}
	class := uint32(item.ObjClass)
	rate := uint32(effect.Update100.Val)
	if class&0x02000000 != 0 && r.armorFlags(item)&0x4000 != 0 {
		rate /= 3
	}
	maximum = r.maxHP(owner)
	fps = r.fps()
	interval := rate * fps / uint32(maximum)
	if r.frame()%interval == 0 {
		r.adjustHP(owner, 1)
	}
}
