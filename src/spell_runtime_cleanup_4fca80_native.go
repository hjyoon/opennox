package opennox

import (
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellRuntimeCleanupNativeDeps4FCA80 struct {
	freeDurations        func()
	loadMagicClass       func() alloc.ClassT[server.MagicEntityClass]
	freeMagicClass       func(alloc.ClassT[server.MagicEntityClass])
	loadImaginaryCaster  func() *server.Object
	clearMagicEntityHead func()
	delayedDelete        func(*server.Object)
	clearImaginaryCaster func()
}

func spellRuntimeCleanupNative4FCA80(deps spellRuntimeCleanupNativeDeps4FCA80) int32 {
	return spellRuntimeCleanup4FCA80(
		spellRuntimeCleanupHooks4FCA80[alloc.ClassT[server.MagicEntityClass], *server.Object]{
			freeDurations:        deps.freeDurations,
			loadMagicClass:       deps.loadMagicClass,
			freeMagicClass:       deps.freeMagicClass,
			loadImaginaryCaster:  deps.loadImaginaryCaster,
			clearMagicEntityHead: deps.clearMagicEntityHead,
			delayedDelete:        deps.delayedDelete,
			clearImaginaryCaster: deps.clearImaginaryCaster,
		},
	)
}

// nox_xxx_freeSpellRelated_4FCA80 binds the original cleanup to native-width
// allocator and Object handles. Its sole decoded caller is Go-owned session
// teardown, so no public C entrypoint is needed.
func nox_xxx_freeSpellRelated_4FCA80() int32 {
	s := noxServer
	return spellRuntimeCleanupNative4FCA80(spellRuntimeCleanupNativeDeps4FCA80{
		freeDurations: s.Spells.Free,
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] {
			return magicEntityAlloc
		},
		freeMagicClass: func(value alloc.ClassT[server.MagicEntityClass]) {
			value.Free()
		},
		loadImaginaryCaster: func() *server.Object {
			return nox_xxx_imagCasterUnit_1569664
		},
		clearMagicEntityHead: func() {
			magicEntityHead = nil
		},
		delayedDelete: s.DelayedDelete,
		clearImaginaryCaster: func() {
			nox_xxx_imagCasterUnit_1569664 = nil
		},
	})
}
