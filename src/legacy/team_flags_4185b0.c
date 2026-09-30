#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"

//----- (004185B0) --------------------------------------------------------
int nox_xxx_wndGuiTeamCreate_4185B0(void) {
	nox_xxx_SetGameplayFlag_417D50(4);
	nox_xxx_teamCreate_4186D0(0);
	nox_xxx_teamCreate_4186D0(0);
	for (nox_object_t* obj = nox_server_getFirstObject_4DA790(); obj;
		 obj = nox_server_getNextObject_4DA7A0(obj)) {
		if (!(obj->obj_class & 0x10000000)) {
			continue;
		}
		nox_object_team_t* value = (nox_object_team_t*)&obj->field_12;
		nox_team_t* team = nox_xxx_getTeamByID_418AB0(value->id);
		if (!team) {
			continue;
		}
		int index = sub_4ECBD0(obj);
		wchar2_t* title = nox_server_teamTitle_418C20(index);
		if (title) {
			sub_418800((wchar2_t*)team, title, 1);
		}
		team->def_ind = (uint8_t)index;
		sub_4184D0(team);
		team->field_72 = obj;
	}
	// The original result was the exhausted object iterator, always NULL.
	return 0;
}

//----- (00418640) --------------------------------------------------------
int nox_xxx_teamAssignFlags_418640(void) {
	for (nox_object_t* obj = nox_server_getFirstObject_4DA790(); obj;
		 obj = nox_server_getNextObject_4DA7A0(obj)) {
		if (!(obj->obj_class & 0x10000000)) {
			continue;
		}
		uint8_t index = (uint8_t)sub_4ECBD0(obj);
		nox_object_team_t* value = (nox_object_team_t*)&obj->field_12;
		nox_team_t* team = nox_xxx_getTeamByID_418AB0(value->id);
		if (team) {
			team->def_ind = index;
			team->field_72 = obj;
		}
	}
	return 0;
}

//----- (00418690) --------------------------------------------------------
char* nox_xxx_toggleAllTeamFlags_418690(int on) {
	for (nox_team_t* team = nox_server_teamFirst_418B10(); team;
		 team = nox_server_teamNext_418B60(team)) {
		nox_object_t* obj = team->field_72;
		if (!obj) {
			continue;
		}
		if (on) {
			nox_xxx_objectSetOn_4E75B0(obj);
		} else {
			nox_xxx_objectSetOff_4E7600(obj);
		}
	}
	return NULL;
}
