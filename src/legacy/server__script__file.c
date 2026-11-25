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

//----- (005418C0) --------------------------------------------------------
int nox_script_readWriteJjj_5418C0(FILE* a1, FILE* a2, FILE* a3) {
	int i;          // ebx
	int result;     // eax
	int j;          // edi
	size_t v6;      // [esp+10h] [ebp-1010h]
	int v7;         // [esp+14h] [ebp-100Ch]
	int v8;         // [esp+18h] [ebp-1008h]
	int v9;         // [esp+1Ch] [ebp-1004h]
	char v10[4096]; // [esp+20h] [ebp-1000h]

	strcpy((char*)v10, "STRG");
	v6 = strlen((const char*)v10);
	nox_fs_fread(a1, v10, v6);
	nox_fs_fread(a2, v10, v6);
	nox_fs_fwrite(a3, v10, v6);
	nox_fs_fread(a1, &v8, 4);
	nox_fs_fread(a2, &v7, 4);
	dword_5d4594_2489424 = v8;
	v9 = v8 + v7;
	nox_fs_fwrite(a3, &v9, 4);
	for (i = 0; i < v8; ++i) {
		nox_fs_fread(a1, &v6, 4);
		nox_fs_fread(a1, v10, v6);
		v10[v6] = 0;
		nox_fs_fwrite(a3, &v6, 4);
		nox_fs_fwrite(a3, v10, v6);
	}
	result = v7;
	for (j = 0; j < v7; ++j) {
		nox_fs_fread(a2, &v6, 4);
		nox_fs_fread(a2, v10, v6);
		v10[v6] = 0;
		nox_fs_fwrite(a3, &v6, 4);
		nox_fs_fwrite(a3, v10, v6);
		result = v7;
	}
	return result;
}

//----- (00541E40) --------------------------------------------------------
int nox_script_readWriteVvv_541E40(FILE* a1, FILE* a2, FILE* a3) {
	size_t v4; // [esp+10h] [ebp-1034h]
	int v5;    // [esp+14h] [ebp-1030h]
	int i;     // [esp+18h] [ebp-102Ch]
	int v7;    // [esp+1Ch] [ebp-1028h]
	int v8;    // [esp+20h] [ebp-1024h]
	int v9;    // [esp+24h] [ebp-1020h]
	int v10;   // [esp+28h] [ebp-101Ch]
	int v11;   // [esp+2Ch] [ebp-1018h]
	int v12;   // [esp+30h] [ebp-1014h]
	int v13;   // [esp+34h] [ebp-1010h]
	int v14;   // [esp+38h] [ebp-100Ch]
	int v15;   // [esp+3Ch] [ebp-1008h]
	int v16;   // [esp+40h] [ebp-1004h]
	unsigned char v17[4096]; // original 4-KiB scratch at [ebp-1000h]

	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, 4u);
	nox_fs_fwrite(a3, &v17, 4);
	nox_fs_fread(a1, &v4, 4u);
	nox_fs_fread(a1, &v17, v4);
	nox_fs_fread(a2, &v4, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v4, 4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v9, 4u);
	nox_fs_fread(a2, &v9, 4u);
	nox_fs_fwrite(a3, &v9, 4);
	nox_fs_fread(a1, &v8, 4u);
	nox_fs_fread(a2, &v8, 4u);
	nox_fs_fwrite(a3, &v8, 4);
	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v7, 4u);
	nox_fs_fread(a2, &v11, 4u);
	v12 = v7 + v11;
	nox_fs_fwrite(a3, &v12, 4);
	nox_fs_fread(a1, &v10, 4u);
	nox_fs_fread(a2, &v10, 4u);
	nox_fs_fwrite(a3, &v10, 4);
	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v13, 4u);
	nox_fs_fread(a2, &v14, 4u);
	nox_fs_fwrite(a3, &v13, 4);
	nox_fs_fread(a1, &v15, 4u);
	nox_fs_fread(a2, &v15, 4u);
	nox_fs_fwrite(a3, &v15, 4);
	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v4, 4u);
	nox_fs_fread(a1, &v17, v4);
	nox_fs_fread(a2, &v4, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v4, 4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v9, 4u);
	nox_fs_fread(a2, &v9, 4u);
	nox_fs_fwrite(a3, &v9, 4);
	nox_fs_fread(a1, &v8, 4u);
	nox_fs_fread(a2, &v8, 4u);
	nox_fs_fwrite(a3, &v8, 4);
	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v7, 4u);
	nox_fs_fread(a2, &v11, 4u);
	dword_5d4594_2489420 = v7;
	v12 = v11 + v7 - 4;
	nox_fs_fwrite(a3, &v12, 4);
	nox_fs_fread(a1, &v10, 4u);
	nox_fs_fread(a2, &v10, 4u);
	nox_fs_fwrite(a3, &v10, 4);
	for (i = 0; i < v7; ++i) {
		nox_fs_fread(a1, &v5, 4u);
		nox_fs_fwrite(a3, &v5, 4);
	}
	nox_fs_fread(a2, &v5, 4u);
	nox_fs_fread(a2, &v5, 4u);
	nox_fs_fread(a2, &v5, 4u);
	nox_fs_fread(a2, &v5, 4u);
	for (i = 4; i < v11; ++i) {
		nox_fs_fread(a2, &v5, 4u);
		nox_fs_fwrite(a3, &v5, 4);
	}
	v4 = 4;
	nox_fs_fread(a1, &v17, 4u);
	nox_fs_fread(a2, &v17, v4);
	nox_fs_fwrite(a3, &v17, v4);
	nox_fs_fread(a1, &v13, 4u);
	nox_fs_fread(a2, &v14, 4u);
	v16 = v14 + v13 - 4;
	nox_fs_fwrite(a3, &v16, 4);
	nox_script_readWriteYyy_542380(a1, a3, 0);
	nox_fs_fseek_cur(a3, -4);
	return nox_script_readWriteYyy_542380(a2, a3, 1);
}

//----- (00541D80) --------------------------------------------------------
int nox_script_readWriteIii_541D80(FILE* a1, FILE* a2) {
	int v2;     // ebx
	int result; // eax
	int v4;     // [esp+Ch] [ebp-100Ch]
	int v5;     // [esp+10h] [ebp-1008h]
	int v6;     // [esp+14h] [ebp-1004h]
	int v7;     // [esp+18h] [ebp-1000h]

	nox_fs_fread(a1, &v7, 4u);
	v2 = 0;
	nox_fs_fwrite(a2, &v7, 4);
	nox_fs_fread(a1, &v4, 4u);
	nox_fs_fwrite(a2, &v4, 4);
	nox_fs_fread(a1, &v5, 4u);
	nox_fs_fwrite(a2, &v5, 4);
	result = v4;
	if (v4 && v4 > 0) {
		do {
			nox_fs_fread(a1, &v6, 4u);
			nox_fs_fwrite(a2, &v6, 4);
			result = v4;
			++v2;
		} while (v2 < v4);
	}
	return result;
}

//----- (00541A50) --------------------------------------------------------
int nox_script_readWriteXxx_541A50(FILE* f1, FILE* f2, FILE* f3) {
	unsigned int v3; // kr04_4
	int result;      // eax
	int i;           // ebx
	int j;           // edi
	size_t v7;       // [esp+10h] [ebp-101Ch]
	int v8;          // [esp+14h] [ebp-1018h]
	int v9;          // [esp+18h] [ebp-1014h]
	int v10;         // [esp+1Ch] [ebp-1010h]
	int v11;         // [esp+20h] [ebp-100Ch]
	int v12;         // [esp+24h] [ebp-1008h]
	int v13;         // [esp+28h] [ebp-1004h]
	char v14[4096];  // [esp+2Ch] [ebp-1000h]

	strcpy((char*)v14, "CODE");
	v3 = strlen((const char*)v14) + 1;
	nox_fs_fread(f1, v14, v3 - 1);
	nox_fs_fread(f2, v14, v3 - 1);
	nox_fs_fwrite(f3, v14, v3 - 1);
	nox_fs_fread(f1, &v11, 4u);
	nox_fs_fread(f2, &v10, 4u);
	dword_5d4594_2489428 = v11;
	v13 = v10 + v11 - 2;
	nox_fs_fwrite(f3, &v13, 4);
	result = nox_script_readWriteVvv_541E40(f1, f2, f3);
	for (i = 0; i < v11 - 2; result = v11 - 2) {
		v7 = 4;
		nox_fs_fread(f1, &v14, 4u);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f1, &v7, 4u);
		nox_fs_fread(f1, v14, v7);
		v14[v7] = 0;
		nox_fs_fwrite(f3, &v7, 4);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f1, &v12, 4u);
		nox_fs_fwrite(f3, &v12, 4);
		nox_fs_fread(f1, &v8, 4u);
		nox_fs_fwrite(f3, &v8, 4);
		nox_script_readWriteIii_541D80(f1, f3);
		v7 = 4;
		nox_fs_fread(f1, v14, 4u);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f1, &v9, 4u);
		nox_fs_fwrite(f3, &v9, 4);
		nox_script_readWriteYyy_542380(f1, f3, 0);
		++i;
	}
	for (j = 0; j < v10 - 2; result = v10 - 2) {
		v7 = 4;
		nox_fs_fread(f2, v14, 4u);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f2, &v7, 4u);
		nox_fs_fread(f2, v14, v7);
		v14[v7] = 0;
		nox_fs_fwrite(f3, &v7, 4);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f2, &v12, 4u);
		nox_fs_fwrite(f3, &v12, 4);
		nox_fs_fread(f2, &v8, 4u);
		nox_fs_fwrite(f3, &v8, 4);
		nox_script_readWriteIii_541D80(f2, f3);
		v7 = 4;
		nox_fs_fread(f2, v14, 4u);
		nox_fs_fwrite(f3, v14, v7);
		nox_fs_fread(f2, &v9, 4u);
		nox_fs_fwrite(f3, &v9, 4);
		nox_script_readWriteYyy_542380(f2, f3, 1);
		++j;
	}
	return result;
}

//----- (005417C0) --------------------------------------------------------
size_t nox_script_readWriteWww_5417C0(FILE* a1, FILE* a2, FILE* a3) {
	unsigned int v3; // kr04_4
	unsigned int v4; // kr08_4
	char v6[1024];   // [esp+10h] [ebp-400h]

	dword_5d4594_2489420 = 0;
	dword_5d4594_2489424 = 0;
	dword_5d4594_2489428 = 0;
	strcpy(v6, "SCRIPT03");
	v3 = strlen(v6) + 1;
	nox_fs_fread(a1, v6, v3 - 1);
	nox_fs_fread(a2, v6, v3 - 1);
	nox_fs_fwrite(a3, v6, v3 - 1);
	nox_script_readWriteJjj_5418C0(a1, a2, a3);
	nox_script_readWriteXxx_541A50(a1, a2, a3);
	strcpy(v6, "DONE");
	v4 = strlen(v6) + 1;
	nox_fs_fread(a1, v6, v4 - 1);
	nox_fs_fread(a2, v6, v4 - 1);
	return nox_fs_fwrite(a3, v6, v4 - 1);
}

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
