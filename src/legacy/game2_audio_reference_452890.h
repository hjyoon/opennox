/* Name-only copies of active GAME2.c bodies. Service boundaries below are
 * independent fixture declarations. Only complete defined input is executed. */
#include <stdint.h>
#include <string.h>
#include "memfile.h"
extern int sub_4862E0(void*, int32_t);
static uint32_t ga_ref_words[50];
static char ga_ref_name[256];
static int32_t ga_ref_bank;
static int ga_ref_find(const char *s){return strcmp(s,ga_ref_name)==0 ? 1:0;}
static char *ga_ref_def(int i){return i==1 ? (char*)ga_ref_words:0;}
static unsigned ga_ref_sample(int unused,void *p){const char *s=p;return strcmp(s,"sample0")==0?0:strcmp(s,"sample1")==0?1:(unsigned)-1;}
#define nox_xxx_utilFindSound_40AF50 ga_ref_find
#define nox_xxx_draw_452270 ga_ref_def
#define sub_486A10 ga_ref_sample
#define dword_5d4594_1045420 ga_ref_bank
static int ga_ref_skip_inner(nox_memfile*);
#define nox_thing_skip_AVNT_inner_452B30 ga_ref_skip_inner
int ga_ref_avnt(nox_memfile* f, void* bufp) {
	char* buf = bufp;
	uint8_t name_len = nox_memfile_read_u8(f);
	nox_memfile_read(buf, 1, name_len, f);
	buf[name_len] = 0;
	int sound = nox_xxx_utilFindSound_40AF50(buf);
	char* def = sound ? nox_xxx_draw_452270(sound) : NULL;
	if (!def) {
		nox_thing_skip_AVNT_inner_452B30(f);
		return 1;
	}

	for (;;) {
		switch (nox_memfile_read_i8(f)) {
		case 0:
			*(uint32_t*)def = 1;
			return 1;
		case 1:
			((uint32_t*)def)[12] = nox_memfile_read_u8(f);
			break;
		case 2:
			((uint32_t*)def)[1] = nox_memfile_read_u8(f);
			break;
		case 3:
			sub_4862E0(def + 16, 163 * nox_memfile_read_u8(f));
			break;
		case 4:
			((uint32_t*)def)[14] = nox_memfile_read_u8(f);
			break;
		case 5:
			((uint32_t*)def)[15] = nox_memfile_read_u8(f);
			break;
		case 6:
			((uint32_t*)def)[19] = (uint32_t)nox_memfile_read_i8(f);
			((uint32_t*)def)[20] = (uint32_t)nox_memfile_read_i8(f);
			break;
		case 7: {
			int samples = 0;
			uint16_t* ids = (uint16_t*)(def + 128);
			for (;;) {
				uint8_t sample_len = nox_memfile_read_u8(f);
				if (!sample_len) {
					break;
				}
				nox_memfile_read(buf, 1, sample_len, f);
				buf[sample_len] = 0;
				if (samples < 32) {
					int16_t id = (int16_t)sub_486A10(*(int*)&dword_5d4594_1045420, buf);
					*ids = (uint16_t)id;
					if (id != -1) {
						samples++;
						ids++;
					}
				}
			}
			((uint32_t*)def)[48] = (uint32_t)samples;
			break;
		}
		case 8:
			((uint32_t*)def)[17] = nox_memfile_read_u32(f);
			((uint32_t*)def)[18] = nox_memfile_read_u32(f);
			break;
		case 9: {
			int16_t duration = nox_memfile_read_i16(f);
			if (duration > 0) {
				((uint32_t*)def)[16] = 15 * duration;
			}
			break;
		}
		case 0xA:
			((uint32_t*)def)[2] = (uint32_t)nox_memfile_read_i16(f);
			break;
		default:
			return 0;
		}
	}
}

int ga_ref_skip(nox_memfile* f) {
	int sz = nox_memfile_read_u8(f);
	nox_memfile_skip(f, sz);
	return nox_thing_skip_AVNT_inner_452B30(f);
}

static int ga_ref_skip_inner(nox_memfile* f) {
	char v3;          // bl
	unsigned char v5; // dl
	int result;       // eax
	char v8;          // [esp+Ch] [ebp+4h]

	while (1) {
		v3 = nox_memfile_read_i8(f);
		v8 = v3;
		switch (v8) {
		case 0:
			result = v3 == 0;
			break;
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
			nox_memfile_skip(f, 1);
			continue;
		case 6:
		case 9:
		case 0xA:
			nox_memfile_skip(f, 2);
			continue;
		case 7:
			while (1) {
				v5 = nox_memfile_read_u8(f);
				if (!v5) {
					break;
				}
				nox_memfile_skip(f, v5);
			}
			continue;
		case 8:
			nox_memfile_skip(f, 8);
			continue;
		default:
			result = 0;
			break;
		}
		break;
	}
	return result;
}

int ga_ref_aud(nox_memfile* f, char* buf) {
	uint8_t name_len = nox_memfile_read_u8(f);
	nox_memfile_read(buf, 1, name_len, f);
	buf[name_len] = 0;
	int sound = nox_xxx_utilFindSound_40AF50(buf);
	char* def = sound ? nox_xxx_draw_452270(sound) : NULL;
	if (!def) {
		nox_memfile_skip(f, 9);
		for (;;) {
			uint8_t sample_len = nox_memfile_read_u8(f);
			if (!sample_len) {
				return 1;
			}
			nox_memfile_skip(f, sample_len);
		}
	}

	((uint32_t*)def)[1] = 2;
	((uint32_t*)def)[2] = (uint32_t)nox_memfile_read_i16(f);
	sub_4862E0(def + 16, 163 * nox_memfile_read_u8(f));
	int16_t duration = nox_memfile_read_i16(f);
	if (duration > 0) {
		((uint32_t*)def)[16] = 15 * duration;
	}
	((uint32_t*)def)[14] = (uint32_t)nox_memfile_read_i8(f);
	((uint32_t*)def)[19] = (uint32_t)nox_memfile_read_i8(f);
	((uint32_t*)def)[20] = (uint32_t)nox_memfile_read_i8(f);
	int8_t mode = nox_memfile_read_i8(f);
	((uint32_t*)def)[12] = (uint32_t)mode;
	if (mode >= 3) {
		return 0;
	}

	int samples = 0;
	uint16_t* ids = (uint16_t*)(def + 128);
	for (;;) {
		uint8_t sample_len = nox_memfile_read_u8(f);
		if (!sample_len) {
			break;
		}
		nox_memfile_read(buf, 1, sample_len, f);
		buf[sample_len] = 0;
		char* ext = strrchr(buf, '.');
		if (ext) {
			*ext = 0;
		}
		int16_t id = (int16_t)sub_486A10(*(int*)&dword_5d4594_1045420, buf);
		*ids = (uint16_t)id;
		if (id != -1) {
			samples++;
			ids++;
		}
	}
	*(uint32_t*)def = 1;
	((uint32_t*)def)[48] = (uint32_t)samples;
	return 1;
}

int ga_ref_pan(int a1) {
	int v1; // eax

	v1 = a1;
	if (a1 <= 50) {
		if (a1 < -50) {
			v1 = -50;
		}
	} else {
		v1 = 50;
	}
	return (v1 * 8192) / 50 + 8192;
}
#undef nox_xxx_utilFindSound_40AF50
#undef nox_xxx_draw_452270
#undef sub_486A10
#undef dword_5d4594_1045420
#undef nox_thing_skip_AVNT_inner_452B30
