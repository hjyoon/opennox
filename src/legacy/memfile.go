package legacy

/*
#include "memfile.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
)

var _ = [1]struct{}{}[unsafe.Sizeof(C.nox_memfile{})-unsafe.Sizeof(binfile.MemFile{})]

type nox_memfile = C.nox_memfile

func asMemfile(p *nox_memfile) *binfile.MemFile {
	return asMemfileP(unsafe.Pointer(p))
}

func asMemfileP(p unsafe.Pointer) *binfile.MemFile {
	return (*binfile.MemFile)(p)
}
