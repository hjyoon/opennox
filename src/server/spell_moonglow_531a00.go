package server

import "github.com/opennox/libs/types"

// SpellMoonglowRuntime531A00 supplies the effects owned by the outer runtime.
// Visual objects are kept outside DurSpell.Field72: that PE32 dword cannot
// hold an object pointer on a 64-bit host.
type SpellMoonglowRuntime531A00 struct {
	EnchantmentDuration func() float32
	NewObject           func(string) *Object
	CreateAt            func(*Object, *Object, types.Pointf)
	ApplyBuff           func(*Object, EnchantID, int16, int8)
	BuffOff             func(*Object, EnchantID)
	DelayedDelete       func(*Object)
	LoadVisual          func(*DurSpell) *Object
	StoreVisual         func(*DurSpell, *Object)
}

// SpellMoonglowCreate531A00 follows GAME.EXE 00531A00 using native-width
// record and object fields. A non-player target receives only the Light buff
// and does not keep a duration record; a player keeps a Moonglow visual.
//
//go:noinline
func SpellMoonglowCreate531A00(record *DurSpell, runtime SpellMoonglowRuntime531A00) int32 {
	duration := int16(int32(runtime.EnchantmentDuration()))
	target := record.Target48
	if target == nil || uint32(target.ObjFlags)&0x8020 != 0 {
		return 1
	}
	if uint8(target.ObjClass)&4 == 0 {
		runtime.ApplyBuff(target, ENCHANT_LIGHT, duration, int8(record.Level))
		return 1
	}
	visual := runtime.NewObject("Moonglow")
	runtime.StoreVisual(record, visual)
	if visual == nil {
		return 1
	}
	player := target.ControllingPlayer()
	point := types.Pointf{X: 2944, Y: 2944}
	if player.Field3680&0x10 != 0 {
		point.X = float32(int32(player.CursorVec.X))
		point.Y = float32(int32(player.CursorVec.Y))
	}
	runtime.CreateAt(visual, target, point)
	runtime.ApplyBuff(target, ENCHANT_MOONGLOW, duration, int8(record.Level))
	return 0
}

// SpellMoonglowDestroy531AF0 follows GAME.EXE 00531AF0. The caller ignores
// the legacy return value.
//
//go:noinline
func SpellMoonglowDestroy531AF0(record *DurSpell, runtime SpellMoonglowRuntime531A00) {
	target := record.Target48
	if target == nil {
		return
	}
	if uint8(target.ObjClass)&4 == 0 {
		runtime.BuffOff(target, ENCHANT_LIGHT)
		return
	}
	if visual := runtime.LoadVisual(record); visual != nil {
		runtime.DelayedDelete(visual)
	}
	runtime.StoreVisual(record, nil)
	runtime.BuffOff(target, ENCHANT_MOONGLOW)
}
