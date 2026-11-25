#include "common__system__team.h"

// These entries return the Go-owned membership references through the C ABI.
extern uintptr_t nox_server_team_first_member_native(nox_team_t* team);
nox_object_team_t* nox_xxx_teamCheckSmth_418C60(nox_team_t* team) {
	return (nox_object_team_t*)nox_server_team_first_member_native(team);
}

extern uintptr_t nox_server_team_next_member_native(nox_object_team_t* value);
nox_object_team_t* sub_418C70(nox_object_team_t* value) {
	return (nox_object_team_t*)nox_server_team_next_member_native(value);
}

uint32_t nox_color_white_2523948 = 0;
uint32_t nox_color_red_2589776 = 0;
uint32_t nox_color_blue_2650684 = 0;
uint32_t nox_color_green_2614268 = 0;
uint32_t nox_color_cyan_2649820 = 0;
uint32_t nox_color_yellow_2589772 = 0;
uint32_t nox_color_violet_2598268 = 0;
uint32_t nox_color_black_2650656 = 0;
uint32_t nox_color_orange_2614256 = 0;
