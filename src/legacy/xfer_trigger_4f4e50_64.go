//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const triggerXferVersion4F4E50 = int16(61)

//export nox_xxx_unitTriggerXfer_native_4F4E50
func nox_xxx_unitTriggerXfer_native_4F4E50(objC *nox_object_t) int32 {
	if err := xferTriggerNative4F4E50(cryptfile.Global(), asObjectS(objC)); err != nil {
		mapLog.Printf("nox_xxx_unitTriggerXfer_4F4E50: %v", err)
		return 0
	}
	return 1
}

func xferTriggerNative4F4E50(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil || obj.UpdateData == nil {
		return fmt.Errorf("missing crypt file, object, or trigger update data")
	}
	ud := obj.UpdateDataTrigger()
	savedField34 := obj.Field34

	versionRaw, err := monsterRWU16(cf, uint16(triggerXferVersion4F4E50))
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	version := int16(versionRaw)
	if version > triggerXferVersion4F4E50 {
		return fmt.Errorf("unsupported version %d", version)
	}
	// The shared legacy base reader still narrows the object pointer for old
	// object formats. Current Nox maps use this post-41 Trigger contract.
	if version < 41 {
		return fmt.Errorf("trigger transfer version %d still requires the PE32 path", version)
	}
	if Nox_xxx_mapReadWriteObjData_4F4530(obj, int(version)) == 0 {
		return fmt.Errorf("base object transfer failed")
	}

	width, height := int32(obj.Shape.Box.W), int32(obj.Shape.Box.H)
	if width, err = monsterRWI32(cf, width); err != nil {
		return fmt.Errorf("shape width: %w", err)
	}
	if height, err = monsterRWI32(cf, height); err != nil {
		return fmt.Errorf("shape height: %w", err)
	}
	if cf.ReadOnly() {
		obj.Shape.Box.W = float32(width)
		obj.Shape.Box.H = float32(height)
		if obj.Shape.Box.W > 60 {
			obj.Shape.Box.W = 60
		}
		if obj.Shape.Box.H > 60 {
			obj.Shape.Box.H = 60
		}
	}
	obj.Shape.Box.Calc()

	if err := monsterRWBytes528DB0(cf, ud.Colors[:]); err != nil {
		return fmt.Errorf("trigger colors: %w", err)
	}
	if ud.Flags, err = monsterRWU32(cf, ud.Flags); err != nil {
		return fmt.Errorf("flags: %w", err)
	}
	for _, it := range []struct {
		name string
		cb   *server.ScriptCallback
		off  uintptr
	}{
		{"activate", &ud.ScriptActivate, 256},
		{"deactivate", &ud.ScriptDeactivate, 384},
		{"collide", &ud.ScriptCollide, 512},
	} {
		if err := monsterXferScriptCallback528DB0(obj, it.cb, it.off); err != nil {
			return fmt.Errorf("%s callback: %w", it.name, err)
		}
	}
	if ud.ClassInclude, err = monsterRWU32(cf, ud.ClassInclude); err != nil {
		return fmt.Errorf("class include: %w", err)
	}
	if ud.ClassExclude, err = monsterRWU32(cf, ud.ClassExclude); err != nil {
		return fmt.Errorf("class exclude: %w", err)
	}
	if cf.ReadOnly() {
		ud.TeamInclude = 0
		ud.TeamExclude = 0
	}
	if ud.TeamInclude, err = monsterRWU8(cf, ud.TeamInclude); err != nil {
		return fmt.Errorf("team include: %w", err)
	}
	if ud.TeamExclude, err = monsterRWU8(cf, ud.TeamExclude); err != nil {
		return fmt.Errorf("team exclude: %w", err)
	}
	if version >= 61 {
		if ud.State, err = monsterRWU8(cf, ud.State); err != nil {
			return fmt.Errorf("state: %w", err)
		}
		if ud.Field9, err = monsterRWU8(cf, ud.Field9); err != nil {
			return fmt.Errorf("field9: %w", err)
		}
		if obj.Field33, err = monsterRWU32(cf, obj.Field33); err != nil {
			return fmt.Errorf("animation frame: %w", err)
		}
		if cf.ReadOnly() {
			obj.MarkAnimFrame(obj.Field33)
		}
	}
	if obj.Field34 != 0 && cf.ReadOnly() {
		if err := monsterXferInventory4F3E30(cf, obj, int(version), obj.Field34); err != nil {
			return err
		}
	}
	obj.Field34 = savedField34
	return nil
}
