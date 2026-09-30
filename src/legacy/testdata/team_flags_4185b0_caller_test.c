#include <assert.h>
#include <stdint.h>
#include <string.h>

#include "../team_flags_4185b0.c"

typedef int (*create_flags_fn)(void);
typedef char* (*toggle_flags_fn)(int);
_Static_assert(_Generic(&nox_xxx_wndGuiTeamCreate_4185B0, create_flags_fn: 1, default: 0), "team creation ABI");
_Static_assert(_Generic(&nox_xxx_teamAssignFlags_418640, create_flags_fn: 1, default: 0), "flag assignment ABI");
_Static_assert(_Generic(&nox_xxx_toggleAllTeamFlags_418690, toggle_flags_fn: 1, default: 0), "flag toggle ABI");

static nox_object_t* objects;
static nox_team_t* teams;
static int object_count;
static int team_count;
static int created;
static int gameplay_flag;
static int material_calls;
static int renamed;
static int notified;
static int toggled;
static wchar2_t title[] = {'R', 'e', 'd', 0};

void nox_xxx_SetGameplayFlag_417D50(int flag) { gameplay_flag = flag; }

nox_team_t* nox_xxx_teamCreate_4186D0(char arg) {
	assert(arg == 0);
	assert(gameplay_flag == 4);
	assert(created < 2);
	return &teams[created++];
}

nox_object_t* nox_server_getFirstObject_4DA790(void) {
	return object_count ? objects : NULL;
}

nox_object_t* nox_server_getNextObject_4DA7A0(nox_object_t* obj) {
	assert(obj >= objects && obj < objects + object_count);
	return obj + 1 < objects + object_count ? obj + 1 : NULL;
}

nox_team_t* nox_xxx_getTeamByID_418AB0(int id) {
	for (int i = 0; i < team_count; ++i) {
		if (teams[i].field_57 == id) {
			return &teams[i];
		}
	}
	return NULL;
}

int sub_4ECBD0(nox_object_t* obj) {
	assert(obj >= objects && obj < objects + object_count);
	assert(obj->obj_class & 0x10000000);
	++material_calls;
	return obj->material;
}

wchar2_t* nox_server_teamTitle_418C20(int index) {
	assert(created == 2);
	assert(index == 1 || index == 2);
	return index == 1 ? title : NULL;
}

void sub_418800(wchar2_t* team, wchar2_t* name, int arg) {
	assert(team == (wchar2_t*)&teams[0]);
	assert(name == title && arg == 1);
	memcpy(team, name, sizeof(title));
	++renamed;
}

void sub_4184D0(nox_team_t* team) {
	assert(created == 2);
	assert(team == &teams[notified]);
	assert(team->def_ind == notified + 1);
	assert(team->field_72 == NULL); // Notification precedes flag assignment.
	++notified;
}

nox_team_t* nox_server_teamFirst_418B10(void) { return team_count ? teams : NULL; }

nox_team_t* nox_server_teamNext_418B60(nox_team_t* team) {
	assert(team >= teams && team < teams + team_count);
	return team + 1 < teams + team_count ? team + 1 : NULL;
}

char nox_xxx_objectSetOn_4E75B0(nox_object_t* obj) {
	assert(obj == &objects[toggled++ * 2]);
	obj->obj_flags |= 0x01000000;
	return 1;
}

int nox_xxx_objectSetOff_4E7600(nox_object_t* obj) {
	assert(obj == &objects[toggled++ * 2]);
	obj->obj_flags &= ~0x01000000;
	return 1;
}

int main(void) {
	nox_object_t local_objects[4] = {0};
	nox_team_t local_teams[2] = {0};
	objects = local_objects;
	teams = local_teams;
	object_count = 4;
	team_count = 2;
	if (sizeof(void*) > 4) {
		assert((uintptr_t)objects > UINT32_MAX);
		assert((uintptr_t)teams > UINT32_MAX);
	}
	objects[0].obj_class = objects[2].obj_class = objects[3].obj_class = 0x10000000;
	objects[0].field_13 = 1;
	objects[2].field_13 = 2;
	objects[3].field_13 = 99; // Orphan flag must not dereference a missing team.
	objects[0].material = 1;
	objects[2].material = 2;
	objects[3].material = 3;
	for (int i = 0; i < 2; ++i) {
		teams[i].field_57 = i + 1;
		teams[i].field_68 = 0x11223344;
		teams[i].field_76 = 0x55667788;
	}
	teams[1].name[0] = 'B';
	assert(nox_xxx_wndGuiTeamCreate_4185B0() == 0);
	assert(created == 2 && gameplay_flag == 4);
	assert(material_calls == 2 && renamed == 1 && notified == 2);
	assert(memcmp(teams[0].name, title, sizeof(title)) == 0);
	assert(teams[1].name[0] == 'B'); // A missing material title preserves the name.
	assert(teams[0].field_72 == &objects[0] && teams[1].field_72 == &objects[2]);

	objects[0].material = 0x107;
	objects[2].material = 9;
	teams[0].field_72 = teams[1].field_72 = &objects[1];
	material_calls = 0;
	assert(nox_xxx_teamAssignFlags_418640() == 0);
	assert(material_calls == 3); // Assignment scans even orphan flag materials.
	assert(renamed == 1 && notified == 2);
	assert(teams[0].def_ind == 7 && teams[1].def_ind == 9);
	assert(teams[0].field_72 == &objects[0] && teams[1].field_72 == &objects[2]);
	for (int i = 0; i < 2; ++i) {
		assert(teams[i].field_68 == 0x11223344 && teams[i].field_76 == 0x55667788);
	}

	assert(nox_xxx_toggleAllTeamFlags_418690(7) == NULL && toggled == 2);
	assert((objects[0].obj_flags & 0x01000000) && (objects[2].obj_flags & 0x01000000));
	toggled = 0;
	assert(nox_xxx_toggleAllTeamFlags_418690(0) == NULL && toggled == 2);
	assert(!(objects[0].obj_flags & 0x01000000) && !(objects[2].obj_flags & 0x01000000));
	toggled = 0;
	teams[1].field_72 = NULL;
	assert(nox_xxx_toggleAllTeamFlags_418690(1) == NULL && toggled == 1);
	team_count = object_count = 0;
	assert(nox_xxx_teamAssignFlags_418640() == 0);
	assert(nox_xxx_toggleAllTeamFlags_418690(1) == NULL && toggled == 1);
	return 0;
}
