package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "game2_audio_native.h"
int sub_4133D0(nox_object_t* a1);
*/
import "C"
import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_soundPlayerDamageSound_5328B0  func(obj, obj2 *server.Object) int
	Nox_xxx_soundDefaultDamageSound_532E20 func(obj, obj2 *server.Object) int
	ClientPlaySoundSpecial                 func(sound.ID, int)
)

//export nox_xxx_clientPlaySoundSpecial_452D80
func nox_xxx_clientPlaySoundSpecial_452D80(id, volume C.int) {
	if ClientPlaySoundSpecial != nil {
		ClientPlaySoundSpecial(sound.ID(id), int(volume))
	}
}

func Nox_xxx_clientPlaySoundSpecial_452D80(id sound.ID, volume int) {
	nox_xxx_clientPlaySoundSpecial_452D80(C.int(id), C.int(volume))
}

func Sub_4133D0(a1 *server.Object) int {
	return int(C.sub_4133D0(asObjectC(a1)))
}

//export nox_xxx_soundPlayerDamageSound_5328B0
func nox_xxx_soundPlayerDamageSound_5328B0(a1, a2 *nox_object_t) int32 {
	return int32(Nox_xxx_soundPlayerDamageSound_5328B0(asObjectS(a1), asObjectS(a2)))
}

//export nox_xxx_soundDefaultDamageSound_532E20
func nox_xxx_soundDefaultDamageSound_532E20(a1, a2 *nox_object_t) int32 {
	return int32(Nox_xxx_soundDefaultDamageSound_532E20(asObjectS(a1), asObjectS(a2)))
}
