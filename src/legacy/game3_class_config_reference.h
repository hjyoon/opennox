#ifndef NOX_GAME3_CLASS_CONFIG_REFERENCE_H
#define NOX_GAME3_CLASS_CONFIG_REFERENCE_H
#include "compat.h"
#include "GAME1.h"
#include "GAME3.h"
#include "client__shell__selcolor.h"
#include "operators.h"
extern nox_window* dword_5d4594_1308084;

static int game3_reference_sub_4A6B50(wchar2_t* a1) {
	wchar2_t* v1;    // esi
	int v2;         // ebp
	int v3;         // ebx
	signed int v4;  // eax
	short* v5;      // edi
	wchar2_t v6;     // ax
	signed int v7;  // esi
	wchar2_t* i;     // edi
	int v10;        // [esp+10h] [ebp-3Ch]
	signed int v11; // [esp+14h] [ebp-38h]
	short v12[26];  // [esp+18h] [ebp-34h]

	v1 = a1;
	v2 = 0;
	v3 = 1;
	v10 = 0;
	v4 = nox_wcslen(a1);
	if (v4 >= 1) {
		v5 = v12;
		v11 = v4;
		do {
			if (iswspace(*v1)) {
				if (!v3) {
					*v5 = *v1;
					++v5;
					++v10;
				}
			} else {
				if (v3 == 1) {
					v6 = *v1;
					if (*v1 == 42 || v6 == 63 || v6 == 60 || v6 == 62 || v6 == 92 || v6 == 47 || v6 == 58 || v6 == 34 ||
						v6 == 124) {
						*v1 = 45;
					}
				}
				*v5 = *v1;
				++v5;
				++v10;
				v3 = 0;
				v2 = 1;
			}
			++v1;
			--v11;
		} while (v11);
		v12[v10] = 0;
		if (v2) {
			nox_wcscpy(a1, (const wchar2_t*)v12);
			v7 = nox_wcslen((const wchar2_t*)v12) - 1;
			if ((int)v7 >= 0) {
				for (i = &a1[v7]; iswspace(*i); --i) {
					if (--v7 < 0) {
						return v2;
					}
				}
				a1[v7 + 1] = 0;
			}
		}
	}
	return v2;
}

static unsigned char* game3_reference_sub_4A61E0(nox_window* a1, int a2, unsigned char* a3) {
	unsigned int v3;       // esi
	unsigned char* result; // eax
	nox_window* v5;        // eax
	nox_window* v6;        // eax

	v3 = 0;
	result = getMemAt(0x5D4594, 1307797 + 96 * a2);
	while (*(result - 1) != *a3 || *result != a3[1] || result[1] != a3[2]) {
		++v3;
		result += 3;
		if (v3 >= 0x20) {
			goto LABEL_9;
		}
	}
	if (a2 == 1) {
		int id = nox_xxx_wndGetID_46B0A0(a1);
		v5 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, id - 10);
		nox_xxx_wnd_46ABB0(v5, 1);
		v6 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, id + 10);
		if (v6) {
			v6->draw_data.field_0 |= 6u;
		}
		result = (unsigned char*)v6;
	}
LABEL_9:
	if (v3 == 32 && a2 == 1) {
		v6 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1308084, nox_xxx_wndGetID_46B0A0(a1) - 10);
		result = (unsigned char*)(uintptr_t)nox_xxx_wnd_46ABB0(v6, 0);
		LOWORD(v3) = 9;
	}
	nox_selcolor_set_value(a1, (unsigned short)a2 | ((unsigned short)v3 << 16));
	return result;
}

static int game3_reference_sub_4A7A80(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	v2 = strlen(a1) + 1;
	v3 = v2;
	v2 >>= 2;
	memcpy(getMemAt(0x5D4594, 1308644), a1, 4 * v2);
	v5 = &a1[4 * v2];
	v4 = getMemAt(0x5D4594, 1308644 + 4 * v2);
	LOBYTE(v2) = v3;
	result = 1;
	memcpy(v4, v5, v2 & 3);
	return result;
}

static int game3_reference_sub_4A7AC0(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	v2 = strlen(a1) + 1;
	v3 = v2;
	v2 >>= 2;
	memcpy(getMemAt(0x5D4594, 1308172), a1, 4 * v2);
	v5 = &a1[4 * v2];
	v4 = getMemAt(0x5D4594, 1308172 + 4 * v2);
	LOBYTE(v2) = v3;
	result = 1;
	memcpy(v4, v5, v2 & 3);
	return result;
}

static int game3_reference_sub_4A7B00(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	v2 = strlen(a1) + 1;
	v3 = v2;
	v2 >>= 2;
	memcpy(getMemAt(0x5D4594, 1308352), a1, 4 * v2);
	v5 = &a1[4 * v2];
	v4 = getMemAt(0x5D4594, 1308352 + 4 * v2);
	LOBYTE(v2) = v3;
	result = 1;
	memcpy(v4, v5, v2 & 3);
	return result;
}

static int game3_reference_sub_4A7BA0(char* a1) {
	int result; // eax

	result = (int)a1;
	if (a1) {
		*getMemU32Ptr(0x5D4594, 1308740) = atoi(a1);
		result = 1;
	}
	return result;
}

static int game3_reference_sub_4A7BC0(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	v2 = strlen(a1) + 1;
	v3 = v2;
	v2 >>= 2;
	memcpy(getMemAt(0x5D4594, 1308324), a1, 4 * v2);
	v5 = &a1[4 * v2];
	v4 = getMemAt(0x5D4594, 1308324 + 4 * v2);
	LOBYTE(v2) = v3;
	result = 1;
	memcpy(v4, v5, v2 & 3);
	return result;
}

static int game3_reference_sub_4A7C00(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	v2 = strlen(a1) + 1;
	v3 = v2;
	v2 >>= 2;
	memcpy(getMemAt(0x5D4594, 1308364), a1, 4 * v2);
	v5 = &a1[4 * v2];
	v4 = getMemAt(0x5D4594, 1308364 + 4 * v2);
	LOBYTE(v2) = v3;
	result = 1;
	memcpy(v4, v5, v2 & 3);
	return result;
}

static int game3_reference_sub_4A7C40(char* a1) {
	int result; // eax

	result = (int)a1;
	if (a1) {
		*getMemU32Ptr(0x5D4594, 1308188) = atoi(a1);
		result = 1;
	}
	return result;
}

static int game3_reference_sub_4A7C60(char* a1) {
	char* v1;   // eax
	char* v2;   // eax

	if (!a1) {
		return 0;
	}
	*getMemU32Ptr(0x5D4594, 1308736) = 0;
	*getMemU32Ptr(0x5D4594, 1308732) = 0;
	v1 = strtok(a1, ",\t\r\n");
	if (v1) {
		*getMemU32Ptr(0x5D4594, 1308732) = atoi(v1);
	}
	v2 = strtok(0, " \t\r\n");
	if (v2) {
		*getMemU32Ptr(0x5D4594, 1308736) = atoi(v2);
	}
	if (*getMemU32Ptr(0x5D4594, 1308732) && *getMemU32Ptr(0x5D4594, 1308736)) {
		return 1;
	} else {
		return 0;
	}
}

static int game3_reference_sub_4A7CE0(char* a1) {
	int result; // eax

	result = (int)a1;
	if (a1) {
		*getMemU32Ptr(0x5D4594, 1308728) = atoi(a1);
		result = 1;
	}
	return result;
}

static int game3_reference_sub_4A7D00(const char* a1) {
	int result;        // eax
	unsigned int v2;   // ecx
	char v3;           // al
	unsigned char* v4; // edi
	const char* v5;    // esi

	if (!a1) {
		return 0;
	}
	result = 0;
	if (strlen(a1) <= 0x80) {
		v2 = strlen(a1) + 1;
		v3 = v2;
		v2 >>= 2;
		memcpy(getMemAt(0x5D4594, 1308192), a1, 4 * v2);
		v5 = &a1[4 * v2];
		v4 = getMemAt(0x5D4594, 1308192 + 4 * v2);
		LOBYTE(v2) = v3;
		result = 1;
		memcpy(v4, v5, v2 & 3);
	}
	return result;
}

static int game3_reference_sub_4A7D50(char* a1) {
	int result; // eax

	result = (int)a1;
	if (a1) {
		*getMemU32Ptr(0x5D4594, 1308348) = atoi(a1);
		result = 1;
	}
	return result;
}
#endif
