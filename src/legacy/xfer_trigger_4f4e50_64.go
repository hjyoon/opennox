//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_unitTriggerXfer_native_4F4E50
func nox_xxx_unitTriggerXfer_native_4F4E50(objC *nox_object_t) int32 {
	return Nox_xxx_UnitTriggerXferNative4F4E50(cryptfile.Global(), asObjectS(objC))
}
