//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const invLightXferVersion4F5AA0 = int16(60)

//export nox_xxx_XFerInvLight_native_4F5AA0
func nox_xxx_XFerInvLight_native_4F5AA0(objC *nox_object_t) int32 {
	if err := xferInvLightNative4F5AA0(cryptfile.Global(), asObjectS(objC)); err != nil {
		mapLog.Printf("nox_xxx_XFerInvLight_4F5AA0: %v", err)
		return 0
	}
	return 1
}

// xferInvLightNative4F5AA0 preserves GAME.EXE's fixed 140-byte PE32 light
// stream while resolving Object and Drawable fields at the host pointer
// width. The original implementation remains active on 32-bit builds.
func xferInvLightNative4F5AA0(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil {
		return fmt.Errorf("missing crypt file or object")
	}
	savedField34 := obj.Field34
	defer func() { obj.Field34 = savedField34 }()

	versionRaw, err := monsterRWU16(cf, uint16(invLightXferVersion4F5AA0))
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	version := int16(versionRaw)
	if version > invLightXferVersion4F5AA0 {
		return fmt.Errorf("unsupported version %d", version)
	}
	if Nox_xxx_mapReadWriteObjData_4F4530(obj, int(version)) == 0 {
		return fmt.Errorf("base object transfer failed")
	}

	var light [client.DrawableLightXferSize]byte
	preview := noxflags.HasGame(noxflags.GameFlag22 | noxflags.GameFlag23)
	if !cf.ReadOnly() {
		dr, err := invLightDrawable4F5AA0(obj, preview)
		if err != nil {
			return err
		}
		if dr != nil {
			light = dr.LightXferData()
		}
	}
	apply, err := rwInvLightPayload4F5AA0(cf, light[:], version)
	if err != nil {
		return err
	}
	if apply && preview {
		if obj.Field189 == nil {
			return fmt.Errorf("preview object has no script/light data")
		}
		copy(unsafe.Slice((*byte)(unsafe.Add(obj.Field189, 2432)), len(light)), light[:])
	}
	if cf.ReadOnly() && obj.Field34 != 0 {
		if err := monsterXferInventory4F3E30(cf, obj, int(version), obj.Field34); err != nil {
			return err
		}
	}
	return nil
}

func invLightDrawable4F5AA0(obj *server.Object, preview bool) (*client.Drawable, error) {
	if GetClient == nil {
		return nil, fmt.Errorf("client drawable registry is unavailable")
	}
	lc := GetClient()
	if lc == nil || lc.Cli() == nil {
		return nil, fmt.Errorf("client drawable registry is unavailable")
	}
	objs := &lc.Cli().Objs
	if preview {
		for dr := objs.FirstList1(); dr != nil; dr = dr.NextPtr {
			if dr.NetCode32 == obj.Extent {
				return dr, nil
			}
		}
		// GAME.EXE deliberately writes a zero light block when a preview/editor
		// drawable is absent.
		return nil, nil
	}
	if obj.Class().HasAny(0x20400000) {
		if dr := objs.ByNetCodeStatic(int(obj.Extent)); dr != nil {
			return dr, nil
		}
		return nil, fmt.Errorf("static drawable extent %d not found for %s", obj.Extent, obj.String())
	}
	if dr := objs.ByNetCodeDynamic(int(obj.NetCode)); dr != nil {
		return dr, nil
	}
	return nil, fmt.Errorf("dynamic drawable net code %d not found for %s", obj.NetCode, obj.String())
}

func rwInvLightPayload4F5AA0(cf *cryptfile.CryptFile, light []byte, version int16) (bool, error) {
	if len(light) != client.DrawableLightXferSize {
		return false, fmt.Errorf("light payload size %d, want %d", len(light), client.DrawableLightXferSize)
	}
	rw := func(off, size int, name string) error {
		if err := monsterRWBytes528DB0(cf, light[off:off+size]); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}

	if version >= 2 {
		for _, part := range []struct {
			off, size int
			name      string
		}{
			{0, 4, "flags"}, {4, 4, "intensity"}, {8, 4, "radius"}, {12, 4, "fixed intensity"},
			{16, 12, "color"}, {28, 2, "direction"}, {30, 2, "penumbra"}, {32, 4, "field 42"},
			{40, 2, "field 44 low"}, {42, 48, "fields 44..56"}, {90, 16, "fields 56..60"},
			{106, 16, "fields 60..64"}, {122, 2, "field 64 high"}, {124, 2, "field 65 low"},
			{126, 2, "field 65 high"}, {128, 4, "field 66"}, {134, 2, "field 67 high"},
			{136, 2, "field 68 low"}, {138, 1, "field 68 byte 2"},
		} {
			if err := rw(part.off, part.size, part.name); err != nil {
				return false, err
			}
		}
		if version > 40 {
			if version >= 42 {
				if err := rw(36, 4, "field 43"); err != nil {
					return false, err
				}
			} else {
				legacyField43 := []byte{light[36]}
				if err := monsterRWBytes528DB0(cf, legacyField43); err != nil {
					return false, fmt.Errorf("legacy field 43: %w", err)
				}
				if cf.ReadOnly() {
					binary.LittleEndian.PutUint32(light[36:], uint32(legacyField43[0]))
				}
			}
			return cf.ReadOnly(), nil
		}
		if cf.ReadOnly() {
			binary.LittleEndian.PutUint32(light[36:], 0)
			return true, nil
		}
		return false, nil
	}

	for _, part := range []struct {
		off, size int
		name      string
	}{
		{0, 4, "flags"}, {4, 4, "intensity"}, {8, 4, "radius"}, {12, 4, "fixed intensity"},
		{16, 12, "color"}, {28, 2, "direction"}, {30, 2, "penumbra"}, {32, 4, "field 42"},
	} {
		if err := rw(part.off, part.size, part.name); err != nil {
			return false, err
		}
	}
	for i := 40; i < 42; i++ {
		light[i] = 0
	}
	for _, part := range []struct{ off, size int }{{122, 2}, {124, 2}, {126, 2}, {128, 4}, {134, 2}} {
		clear(light[part.off : part.off+part.size])
	}
	light[138] = 0x80
	if !cf.ReadOnly() {
		return false, nil
	}
	intensity := math.Float32frombits(binary.LittleEndian.Uint32(light[4:]))
	fixedIntensity := int32(binary.LittleEndian.Uint32(light[12:]))
	if intensity > 63 || float64(fixedIntensity)/65536.0 > 63 {
		binary.LittleEndian.PutUint32(light[4:], math.Float32bits(63))
		binary.LittleEndian.PutUint32(light[8:], uint32(client.LightRadius(63)))
	}
	return true, nil
}
