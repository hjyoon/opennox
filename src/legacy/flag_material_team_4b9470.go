//go:build !server

package legacy

/*
#include "GAME3_1.h"

static int nox_client_flag_material_team_call_4b9470(const char* name) {
	const char* slot = name;
	return sub_4B9470(&slot);
}

static int nox_client_flag_material_drawable_call_4b94e0(const char* name, uint32_t flags) {
	const char* material = name;
	nox_drawable drawable = {0};
	drawable.flags28 = flags;
	drawable.item_modifiers[1] = &material;
	return sub_4B94E0(&drawable);
}
*/
import "C"

import "unsafe"

func flagMaterialTeamCall4B9470(name *byte) int {
	return int(C.nox_client_flag_material_team_call_4b9470((*C.char)(unsafe.Pointer(name))))
}

func flagMaterialTeamNilCall4B9470() int {
	return int(C.sub_4B9470(nil))
}

func flagMaterialDrawableTeamCall4B94E0(name *byte, flags uint32) int {
	return int(C.nox_client_flag_material_drawable_call_4b94e0((*C.char)(unsafe.Pointer(name)), C.uint32_t(flags)))
}
