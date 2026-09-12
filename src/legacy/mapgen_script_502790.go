package legacy

import (
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/internal/noxscriptwire"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapScriptMetadataBase502790 = uintptr(0x587000)
	mapScriptMetadataOff502790  = uintptr(218640)
	mapScriptMetadataStride     = uintptr(268)
	mapScriptFunctionCount      = 36
)

type mapScriptMetadata502790 struct{}

func (mapScriptMetadata502790) ArgCount(function byte) byte {
	// The executable contains 36 records. Higher indices would read the
	// neighboring name strings or other data in the original PE32 image.
	if function >= mapScriptFunctionCount {
		return 0
	}
	return memmap.Uint8(mapScriptMetadataBase502790,
		mapScriptMetadataOff502790+uintptr(function)*mapScriptMetadataStride)
}

func (mapScriptMetadata502790) ArgKind(function byte, argument byte) uint32 {
	if function >= mapScriptFunctionCount || argument >= 32 {
		return 0
	}
	return memmap.Uint32(mapScriptMetadataBase502790,
		mapScriptMetadataOff502790+uintptr(function)*mapScriptMetadataStride+8+uintptr(argument)*8)
}

type mapScriptFile502790 struct{ file *binfile.File }

func (r mapScriptFile502790) Read(p []byte) (int, error) {
	n, err := r.file.Bin.Read(p)
	if err != nil {
		r.file.Err = err
	}
	return n, err
}

func (r mapScriptFile502790) Seek(off int64, whence int) (int64, error) {
	// The active caller is read-only; FileSeek retains the cryptfile cursor
	// behavior used by nox_binfile_fseek_409050.
	err := r.file.Bin.FileSeek(off, whence)
	if err != nil {
		binfile.Log.Println(err)
	}
	return 0, err
}

var _ io.ReadSeeker = mapScriptFile502790{}

func mapgenMakeScript502790(file unsafe.Pointer, handler *server.ScriptCallback) int32 {
	return noxscriptwire.Scan(
		mapScriptFile502790{file: fileByHandle((*FILE)(file))},
		mapScriptMetadata502790{},
		&handler.Flags,
	)
}
