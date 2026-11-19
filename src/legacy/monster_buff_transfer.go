package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_XFer_ReadMonsterBuffs_52AAB0
func nox_xxx_XFer_ReadMonsterBuffs_52AAB0(objC *nox_object_t) C.int {
	cf := cryptfile.Global()
	obj := asObjectS(objC)
	if cf == nil || obj == nil {
		return 0
	}
	if err := monsterXferBuffs52AAB0(cf, obj); err != nil {
		mapLog.Printf("nox_xxx_XFer_ReadMonsterBuffs_52AAB0: %v", err)
		return 0
	}
	return 1
}

func monsterXferBuffs52AAB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	version, err := monsterRWU16(cf, 2)
	if err != nil {
		return fmt.Errorf("buff version: %w", err)
	}
	if version == 0 || version > 2 {
		return fmt.Errorf("unsupported buff version %d", version)
	}
	count := buffIterationCountNative424CB0(obj)
	if count, err = monsterRWU8(cf, count); err != nil {
		return fmt.Errorf("buff count: %w", err)
	}
	if count > 32 {
		return fmt.Errorf("buff count %d exceeds 32", count)
	}
	if cf.ReadOnly() {
		for i := byte(0); i < count; i++ {
			name, e := monsterRWString8(cf, "")
			if e != nil {
				return fmt.Errorf("buff[%d] name: %w", i, e)
			}
			enc, ok := server.ParseEnchant(name)
			if !ok {
				return fmt.Errorf("buff[%d] has unknown name %q", i, name)
			}
			power, e := monsterRWU8(cf, 0)
			if e != nil {
				return fmt.Errorf("buff[%d] power: %w", i, e)
			}
			duration, e := monsterRWU32(cf, 0)
			if e != nil {
				return fmt.Errorf("buff[%d] duration: %w", i, e)
			}
			arg := server.SpellAcceptArg{Obj: obj, Pos: obj.Pos()}
			GetServer().SpellAccept4FD400(enc.Spell(), obj, obj, obj, &arg, int32(power))
			obj.BuffsDur[enc] = uint16(duration)
			if enc == server.ENCHANT_SHIELD && version >= 2 {
				extra, e := monsterRWU32(cf, 100)
				if e != nil {
					return fmt.Errorf("buff[%d] shield duration: %w", i, e)
				}
				if dur := obj.Server().Spells.Dur.SpellDurationFindActiveTarget4FF2D0(51, obj); dur != nil {
					dur.Field72 = int32(extra)
				}
			}
		}
		return nil
	}
	for id := buffIterationFirstNative424D00(); id != -1; id = buffIterationNextNative424D20(id) {
		if unitBuffTestLegacy4FF350(obj, id) == 0 {
			continue
		}
		enc := server.EnchantID(id)
		if _, err = monsterRWString8(cf, enc.String()); err != nil {
			return err
		}
		if _, err = monsterRWU8(cf, byte(obj.EnchantPower(enc))); err != nil {
			return err
		}
		if _, err = monsterRWU32(cf, uint32(obj.EnchantDur(enc))); err != nil {
			return err
		}
		if enc == server.ENCHANT_SHIELD {
			extra := uint32(100)
			if dur := obj.Server().Spells.Dur.SpellDurationFindActiveTarget4FF2D0(51, obj); dur != nil {
				extra = uint32(dur.Field72)
			}
			if _, err = monsterRWU32(cf, extra); err != nil {
				return err
			}
		}
	}
	return nil
}
