package server

// playerDamagePrefix4E17B0 is the entry state retained by 004E184A..004E18F0.
// ObserveClear runs after the marker reset and equipment/absorption reads,
// before any defense. Later marker/state accesses use this update base even
// when a callback replaces the unit's live update or controlling player.
// A non-nil context means the prefix has already run, not that the live
// player still has (or no longer has) an observation target.
type playerDamagePrefix4E17B0 struct {
	update                  *PlayerUpdateData
	armorFlags, weaponFlags uint32
}
