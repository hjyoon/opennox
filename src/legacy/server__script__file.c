#include <stdio.h>
#include <string.h>

#include "GAME3.h"
#include "GAME4_3.h"
#include "common/fs/nox_fs.h"
#include "server__script__builtin.h"
#include "server__script__file.h"

extern unsigned int dword_5d4594_3835312;

unsigned int dword_5d4594_2489420 = 0;
unsigned int dword_5d4594_2489424 = 0;
unsigned int dword_5d4594_2489428 = 0;

//----- (00543110) --------------------------------------------------------
int sub_543110(const char* lpExistingFileName, int* a2) {
	int v2;                 // edx
	int v3;                 // eax
	char* v4;               // edi
	FILE* v6;               // ebx
	FILE* v7;               // ebp
	unsigned int v8;        // kr08_4
	int i;                  // esi
	int v10;                // esi
	int v11;                // edi
	size_t v12;             // [esp+10h] [ebp-101Ch]
	int v13;                // [esp+14h] [ebp-1018h]
	int v14;                // [esp+18h] [ebp-1014h]
	int v15;                // [esp+1Ch] [ebp-1010h]
	int j;                  // [esp+20h] [ebp-100Ch]
	int v17;                // [esp+24h] [ebp-1008h]
	int v18;                // [esp+28h] [ebp-1004h]
	char v19[2048];         // [esp+2Ch] [ebp-1000h]
	char NewFileName[2048]; // [esp+82Ch] [ebp-800h]

	sub_542BF0(*(int*)&dword_5d4594_3835312, a2[0], a2[1]);
	v2 = *getMemU32Ptr(0x587000, 282612);
	strcpy(NewFileName, (const char*)getMemAt(0x973F18, 42152));
	v3 = *getMemU32Ptr(0x587000, 282616);
	v4 = &NewFileName[strlen(NewFileName)];
	*(uint32_t*)v4 = *getMemU32Ptr(0x587000, 282608);
	*((uint32_t*)v4 + 1) = v2;
	*((uint32_t*)v4 + 2) = v3;
	nox_fs_copy(lpExistingFileName, NewFileName);
	v6 = nox_fs_open(NewFileName);
	if (!v6) {
		return 0;
	}
	v7 = nox_fs_open_rw(lpExistingFileName);
	if (!v7) {
		nox_fs_close(v6);
		return 0;
	}
	strcpy((char*)v19, "SCRIPT03");
	v8 = strlen((const char*)v19) + 1;
	nox_fs_fread(v6, v19, v8 - 1);
	nox_fs_fwrite(v7, v19, v8 - 1);
	strcpy((char*)v19, "STRG");
	v12 = strlen((const char*)v19);
	nox_fs_fread(v6, v19, v12);
	nox_fs_fwrite(v7, v19, v12);
	nox_fs_fread(v6, &v14, 4u);
	nox_fs_fwrite(v7, &v14, 4);
	for (i = 0; i < v14; ++i) {
		nox_fs_fread(v6, &v12, 4u);
		nox_fs_fread(v6, v19, v12);
		v19[v12] = 0;
		nox_fs_fwrite(v7, &v12, 4);
		nox_fs_fwrite(v7, v19, v12);
	}
	strcpy((char*)v19, "CODE");
	v12 = strlen((const char*)v19);
	nox_fs_fread(v6, v19, v12);
	nox_fs_fwrite(v7, v19, v12);
	nox_fs_fread(v6, &v15, 4u);
	nox_fs_fwrite(v7, &v15, 4);
	v10 = 0;
	for (j = 0; v10 < v15; j = v10) {
		v12 = 4;
		nox_fs_fread(v6, v19, 4u);
		nox_fs_fwrite(v7, v19, v12);
		nox_fs_fread(v6, &v12, 4u);
		nox_fs_fread(v6, v19, v12);
		v19[v12] = 0;
		if (v10 > 1) {
			sub_5435C0(v19, *(int*)&dword_5d4594_3835312, a2[0], a2[1]);
			strcpy((char*)&v19, (const char*)getMemAt(0x5D4594, 2489164));
			v10 = j;
			v12 = strlen((const char*)&v19);
		}
		nox_fs_fwrite(v7, &v12, 4);
		nox_fs_fwrite(v7, v19, v12);
		nox_fs_fread(v6, &v18, 4u);
		nox_fs_fwrite(v7, &v18, 4);
		nox_fs_fread(v6, &v17, 4u);
		nox_fs_fwrite(v7, &v17, 4);
		nox_script_readWriteIii_541D80(v6, v7);
		v12 = 4;
		nox_fs_fread(v6, v19, 4u);
		nox_fs_fwrite(v7, v19, v12);
		v11 = nox_fs_ftell(v7);
		nox_fs_fread(v6, &v13, 4u);
		nox_fs_fwrite(v7, &v13, 4);
		nox_script_readWriteYyy_542380(v6, v7, 0);
		v13 = -4 - v11 + nox_fs_ftell(v7);
		nox_fs_fseek_start(v7, v11);
		nox_fs_fwrite(v7, &v13, 4);
		nox_fs_fseek_cur(v7, v13);
		++v10;
	}
	strcpy((char*)&v19, "DONE");
	v12 = strlen((const char*)&v19);
	nox_fs_fread(v6, v19, v12);
	nox_fs_fwrite(v7, v19, v12);
	nox_fs_close(v6);
	nox_fs_close(v7);
	return nox_fs_remove(NewFileName);
}
