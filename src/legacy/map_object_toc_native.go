package legacy

/*
#include "GAME1_2.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func readMapObjectTOC(cf *cryptfile.CryptFile) error {
	count, err := cf.ReadU16()
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	var typeIndex func(string) int
	if noxflags.HasGame(noxflags.GameClient) && !noxflags.HasGame(noxflags.GameHost) {
		typeIndex = GetClient().Cli().Things.IndByID
	} else {
		typeIndex = GetServer().S().Types.IndByID
	}
	var name [255]byte
	for range count {
		code, err := cf.ReadU16()
		if err != nil {
			return err
		}
		size, err := cf.ReadU8()
		if err != nil {
			return err
		}
		if err := mapSectionTransfer(cf, name[:size]); err != nil {
			return err
		}
		ind := uint16(typeIndex(alloc.GoStringS(name[:size])))
		Sub_42C310(int(ind), code)
	}
	return nil
}

func writeMapObjectTOC(cf *cryptfile.CryptFile) error {
	Sub_42BFE0()
	if err := cf.WriteU16(Sub_42C300()); err != nil {
		return err
	}
	types := &GetServer().S().Types
	for ind := range types.Count() {
		code := Sub_42C2E0(ind)
		if code == 0 {
			continue
		}
		if err := cf.WriteU16(code); err != nil {
			return err
		}
		name := []byte(types.ByInd(ind).ID())
		size := byte(len(name))
		if err := cf.WriteU8(size); err != nil {
			return err
		}
		if err := mapSectionTransfer(cf, name[:size]); err != nil {
			return err
		}
	}
	return nil
}

func mapRWObjectTOC(cf *cryptfile.CryptFile) error {
	if cf == nil {
		cf = cryptfile.Global()
		if cf == nil {
			return fmt.Errorf("object table of contents requires an open crypt file")
		}
	}
	version, err := cf.ReadWriteU16(1)
	if err != nil {
		return err
	}
	if int16(version) > 1 {
		return fmt.Errorf("unsupported object table of contents version %d", version)
	}
	Sub_42BFB0()
	if cf.ReadOnly() {
		return readMapObjectTOC(cf)
	}
	return writeMapObjectTOC(cf)
}

//export nox_server_mapRWObjectTOC_428B30
func nox_server_mapRWObjectTOC_428B30() C.int {
	return C.int(bool2int(mapRWObjectTOC(nil) == nil))
}
