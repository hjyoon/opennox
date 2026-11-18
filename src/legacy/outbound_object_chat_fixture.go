package legacy

/*
#include "GAME4_2.h"

typedef int (*visibility_fn)(int, nox_object_t*);
typedef uintptr_t (*monster_command_fn)(nox_object_t*, nox_object_t*, const char*, short);
typedef int (*clear_chat_fn)(void);
_Static_assert(_Generic(&nox_xxx_netObjectOutOfSight_528A60, visibility_fn:1, default:0), "out-of-sight signature");
_Static_assert(_Generic(&nox_xxx_netObjectInShadows_528A90, visibility_fn:1, default:0), "in-shadows signature");
_Static_assert(_Generic(&nox_xxx_monsterCmdSend_528BD0, monster_command_fn:1, default:0), "monster command signature");
_Static_assert(_Generic(&nox_xxx_destroyEveryChatMB_528D60, clear_chat_fn:1, default:0), "clear chat signature");

static uintptr_t outbound_object_chat_call_C(int entry, int indirect, int client, nox_object_t* unit,
                                             nox_object_t* source, const char* command, short duration) {
    visibility_fn volatile out_of_sight = nox_xxx_netObjectOutOfSight_528A60;
    visibility_fn volatile in_shadows = nox_xxx_netObjectInShadows_528A90;
    monster_command_fn volatile monster_command = nox_xxx_monsterCmdSend_528BD0;
    clear_chat_fn volatile clear_chat = nox_xxx_destroyEveryChatMB_528D60;
    switch (entry) {
    case 0:
        return (uint32_t)(indirect ? out_of_sight(client, unit) : nox_xxx_netObjectOutOfSight_528A60(client, unit));
    case 1:
        return (uint32_t)(indirect ? in_shadows(client, unit) : nox_xxx_netObjectInShadows_528A90(client, unit));
    case 2:
        return indirect ? monster_command(unit, source, command, duration) : nox_xxx_monsterCmdSend_528BD0(unit, source, command, duration);
    case 3:
        return (uint32_t)(indirect ? clear_chat() : nox_xxx_destroyEveryChatMB_528D60());
    default:
        return UINTPTR_MAX;
    }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callOutboundObjectChatC(entry, route, client int, unit, source *server.Object, command *byte, duration uint16) uintptr {
	return uintptr(C.outbound_object_chat_call_C(C.int(entry), C.int(route), C.int(client), asObjectC(unit), asObjectC(source),
		(*C.char)(unsafe.Pointer(command)), C.short(duration)))
}
