package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/maps"
	"github.com/opennox/noxcrypt"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_xxx_mapValidateMB_4CF470
func nox_xxx_mapValidateMB_4CF470(name *C.char, expected C.int) C.int {
	if expected == 0 {
		return 6
	}
	if name == nil {
		return 0
	}
	return C.int(validateMapFile4CF470(GoString(name), uint32(expected)))
}

func validateMapFile4CF470(name string, expected uint32) uint32 {
	if expected == 0 {
		return 6
	}
	if end := strings.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}
	path := name
	if !strings.ContainsAny(name, "/\\") {
		prefix := "maps\\" + name[:min(len(name), 1018)]
		separator := memmap.Slice(0x587000, 191672)[:2]
		path = prefix[:len(prefix)-4] + alloc.GoStringS(separator) + name
	}
	if fsAccess(path) == -1 {
		return 0
	}
	var status uint32
	if fsAccess(path) == -1 {
		status = 1
	}
	if err := cryptfile.OpenGlobal(path, cryptfile.ReadOnly, crypt.MapKey); err != nil {
		if !os.IsNotExist(err) {
			binfile.Log.Println(err)
		}
		return status
	}
	defer cryptfile.Close()
	status |= 2
	if magic, err := cryptfile.Global().ReadU32(); err == nil && magic == maps.Magic {
		if crc, err := cryptfile.Global().ReadAlignedU32(); err == nil && crc == expected {
			status |= 4
		}
	}
	return status
}

//export sub_4D0670
func sub_4D0670(name *C.char) C.int {
	if !readMapRuleFile4D0670(GoString(name)) {
		return 0
	}
	return 1
}

func readMapRuleFile4D0670(path string) bool {
	in, err := ifs.Open(path)
	if err != nil {
		return false
	}
	file := binfile.NewTextFile(in)
	defer file.Close()
	mode := noxflags.GameModeMask
	ctx := context.Background()
	for {
		line, readErr := file.ReadString()
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return true
		}
		line = line[:min(len(line), 254)]
		if end := bytes.IndexAny(line, "\x00\n"); end >= 0 {
			line = line[:end]
		}
		if len(line) != 0 {
			if section := mapRuleModeByName(string(line)); section != 0 {
				mode = section
			} else if noxflags.HasGame(mode) {
				// Legacy wide printf %S widens each byte before passing the
				// command to the Go console.
				command := make([]rune, len(line))
				for i, v := range line {
					command[i] = rune(v)
				}
				ExecConsoleCmd(ctx, string(command))
			}
		}
		if readErr != nil {
			return true
		}
	}
}
