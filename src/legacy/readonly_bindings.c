// Keep const input contracts in C. cgo cannot express const pointees; its
// private Go exports only read/copy these inputs and retain the public ABI.
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "common/fs/nox_fs.h"
#include "common__strman.h"

extern nox_window* nox_new_window_from_file_go(char* cname, void* fnc);
nox_window* nox_new_window_from_file(const char* cname, void* fnc) {
	return nox_new_window_from_file_go((char*)cname, fnc);
}

extern nox_video_bag_image_t* nox_xxx_gLoadImg_42F970_go(char* name);
nox_video_bag_image_t* nox_xxx_gLoadImg_42F970(const char* name) {
	return nox_xxx_gLoadImg_42F970_go((char*)name);
}

extern int nox_xxx_getTTByNameSpriteMB_44CFC0_go(char* a1);
int nox_xxx_getTTByNameSpriteMB_44CFC0(const char* a1) {
	return nox_xxx_getTTByNameSpriteMB_44CFC0_go((char*)a1);
}

extern FILE* nox_fs_open_go(char* path);
FILE* nox_fs_open(const char* path) {
	return nox_fs_open_go((char*)path);
}

extern FILE* nox_fs_open_text_go(char* path);
FILE* nox_fs_open_text(const char* path) {
	return nox_fs_open_text_go((char*)path);
}

extern FILE* nox_fs_open_rw_go(char* path);
FILE* nox_fs_open_rw(const char* path) {
	return nox_fs_open_rw_go((char*)path);
}

extern bool nox_fs_remove_go(char* path);
bool nox_fs_remove(const char* path) {
	return nox_fs_remove_go((char*)path);
}

extern bool nox_fs_copy_go(char* src, char* dst);
bool nox_fs_copy(const char* src, const char* dst) {
	return nox_fs_copy_go((char*)src, (char*)dst);
}

extern int nox_xxx_weaponInventoryEquipFlags_415820_go(nox_object_t* item);
int nox_xxx_weaponInventoryEquipFlags_415820(const nox_object_t* item) {
	return nox_xxx_weaponInventoryEquipFlags_415820_go((nox_object_t*)item);
}

extern int nox_xxx_unitArmorInventoryEquipFlags_415C70_go(nox_object_t* item);
int nox_xxx_unitArmorInventoryEquipFlags_415C70(const nox_object_t* item) {
	return nox_xxx_unitArmorInventoryEquipFlags_415C70_go((nox_object_t*)item);
}

extern void nox_xxx_servStartCountdown_40A2A0_go(int a1, char* a2);
void nox_xxx_servStartCountdown_40A2A0(int a1, const char* a2) {
	nox_xxx_servStartCountdown_40A2A0_go(a1, (char*)a2);
}

extern int nox_xxx_modifGetIdByName_413290_go(char* a1);
int nox_xxx_modifGetIdByName_413290(const char* a1) {
	return nox_xxx_modifGetIdByName_413290_go((char*)a1);
}

extern wchar2_t* nox_strman_loadString_40F1D0_go(char* name, char** strOut, char* srcFile, int srcLine);
wchar2_t* nox_strman_loadString_40F1D0(const char* name, char** strOut, const char* srcFile, int srcLine) {
	return nox_strman_loadString_40F1D0_go((char*)name, strOut, (char*)srcFile, srcLine);
}

extern int nox_xxx_drawGetStringSize_43F840_go(void* a1, wchar2_t* a2, int* a3, int* a4, int a5);
int nox_xxx_drawGetStringSize_43F840(void* a1, const wchar2_t* a2, int* a3, int* a4, int a5) {
	return nox_xxx_drawGetStringSize_43F840_go(a1, (wchar2_t*)a2, a3, a4, a5);
}

