//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const (
	spellRewardXferVersion4F5F30 = int16(60)
	spellRewardNameLimit4F5F30   = 128
	spellRewardOldIDLimit4F5F30  = byte(0x89)
)

//export nox_xxx_XFerSpellReward_native_4F5F30
func nox_xxx_XFerSpellReward_native_4F5F30(objC *nox_object_t) int32 {
	if err := xferSpellRewardNative4F5F30(cryptfile.Global(), asObjectS(objC)); err != nil {
		mapLog.Printf("nox_xxx_XFerSpellReward_4F5F30: %v", err)
		return 0
	}
	return 1
}

// xferSpellRewardNative4F5F30 preserves GAME.EXE's fixed-width stream while
// resolving Object.UseData and inventory links at the host pointer width. The
// original PE32 implementation remains active on 32-bit builds.
func xferSpellRewardNative4F5F30(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil || obj.UseData.Ptr == nil {
		return fmt.Errorf("missing crypt file, object, or spell-reward use data")
	}
	savedField34 := obj.Field34
	defer func() { obj.Field34 = savedField34 }()

	versionRaw, err := monsterRWU16(cf, uint16(spellRewardXferVersion4F5F30))
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	version := int16(versionRaw)
	if version > spellRewardXferVersion4F5F30 {
		return fmt.Errorf("unsupported version %d", version)
	}
	if Nox_xxx_mapReadWriteObjData_4F4530(obj, int(version)) == 0 {
		return fmt.Errorf("base object transfer failed")
	}

	data := obj.UseDataSpellReward()
	spellID, err := rwSpellRewardPayload4F5F30(cf, data.Spell, version)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		data.Spell = spellID
	}
	if cf.ReadOnly() && obj.Field34 != 0 {
		if err := monsterXferInventory4F3E30(cf, obj, int(version), obj.Field34); err != nil {
			return err
		}
	}
	return nil
}

func rwSpellRewardPayload4F5F30(cf *cryptfile.CryptFile, current byte, version int16) (byte, error) {
	if cf == nil {
		return current, fmt.Errorf("missing crypt file")
	}
	if !cf.ReadOnly() {
		name := spell.ID(current).String()
		if name == "" {
			return current, fmt.Errorf("invalid spell id %d", current)
		}
		if _, err := rwSpellRewardName4F5F30(cf, name); err != nil {
			return current, fmt.Errorf("spell name: %w", err)
		}
		return current, nil
	}

	if version >= 41 {
		name, err := rwSpellRewardName4F5F30(cf, "")
		if err != nil {
			return current, fmt.Errorf("spell name: %w", err)
		}
		return spellRewardParseID4F5F30(name), nil
	}
	if version >= 31 {
		var selected byte
		for i := 0; i < 3; i++ {
			name, err := rwSpellRewardName4F5F30(cf, "")
			if err != nil {
				return current, fmt.Errorf("legacy spell name %d: %w", i, err)
			}
			id := spellRewardParseID4F5F30(name)
			switch i {
			case 1:
				selected = id
			case 2:
				if id != 0 {
					selected = id
				}
			}
		}
		return selected, nil
	}

	var raw [3]byte
	for i := range raw {
		value, err := monsterRWU8(cf, 0)
		if err != nil {
			return current, fmt.Errorf("legacy spell id %d: %w", i, err)
		}
		raw[i] = value
	}
	if raw[1] >= spellRewardOldIDLimit4F5F30 {
		raw[1] = 0
	}
	if raw[2] >= spellRewardOldIDLimit4F5F30 {
		raw[2] = 0
	}
	selected := raw[1]
	if raw[2] != 0 {
		selected = raw[2]
	}
	if version == 10 {
		if _, err := monsterRWU8(cf, 0); err != nil {
			return current, fmt.Errorf("version 10 trailing spell id: %w", err)
		}
	}
	return selected, nil
}

func rwSpellRewardName4F5F30(cf *cryptfile.CryptFile, value string) (string, error) {
	if !cf.ReadOnly() && len(value) >= spellRewardNameLimit4F5F30 {
		return "", fmt.Errorf("name length %d exceeds PE32 buffer", len(value))
	}
	size, err := monsterRWU8(cf, byte(len(value)))
	if err != nil {
		return "", err
	}
	if cf.ReadOnly() {
		if int(size) >= spellRewardNameLimit4F5F30 {
			return "", fmt.Errorf("name length %d exceeds PE32 buffer", size)
		}
		buf := make([]byte, int(size))
		if err := monsterRWBytes528DB0(cf, buf); err != nil {
			return "", err
		}
		return string(buf), nil
	}
	if err := monsterRWBytes528DB0(cf, []byte(value)); err != nil {
		return "", err
	}
	return value, nil
}

func spellRewardParseID4F5F30(name string) byte {
	id := spell.ParseID(name)
	if id <= 0 || id > 0xff {
		return 0
	}
	return byte(id)
}
