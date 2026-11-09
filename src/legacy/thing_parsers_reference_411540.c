/* Exact private-renamed old C bodies. Normal-complete inputs ONLY.
 * The public production definitions are separately retained under #if 0.
 * Never call reference on >=32-byte 411540/411850 names or truncated inputs. */
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME2.h"
#include "client__video__draw_common.h"
#include "operators.h"
extern uint32_t nox_tile_def_cnt;
extern nox_tileDef_t nox_tile_defs_arr[176];
extern uint32_t dword_5d4594_251572;
static int thing_reference_nox_thing_read_FLOR_411540(nox_memfile* f, uint8_t* a2) {
	if (nox_tile_def_cnt >= 176) {
		return 0;
	}
	int ind = nox_tile_def_cnt;
	nox_tileDef_t* p = &nox_tile_defs_arr[ind];

	nox_memfile_skip(f, 4);

	int sz = nox_memfile_read_u8(f);
	char buf[32];
	nox_memfile_read(buf, 1, sz, f);
	buf[sz] = 0;
	strncpy(&p->name[0], buf, 31);

	char v7 = nox_memfile_read_i8(f);
	char v8 = nox_memfile_read_i8(f);
	char v9 = nox_memfile_read_i8(f);
	uint8_t cbuf[6];    // [esp+0h] [ebp-28h]
	cbuf[0] = v7;
	cbuf[1] = v8;
	cbuf[2] = v9;
	if (v7 != -1 || cbuf[1] || v9 != -1) {
		// TODO: int addressing here doesn't look right!
		p->color_48 = nox_color_rgb_4344A0(*(int*)&cbuf[0], *(int*)&cbuf[1], *(int*)&cbuf[2]);
	} else {
		p->color_48 = 0x80000000;
	}
	p->field_58 = 0;
	if (nox_xxx_checkFacade_4117E0(p) == 1) {
		p->field_58 |= 1u;
	}

	p->field_36 = nox_memfile_read_i32(f);
	p->field_40 = nox_memfile_read_i32(f);

	uint8_t v14 = nox_memfile_read_u8(f);
	p->field_57 = v14;
	v14 = nox_memfile_read_u8(f);
	p->field_53 = v14;
	v14 = nox_memfile_read_u8(f);
	p->field_52 = v14;
	v14 = nox_memfile_read_u8(f);
	p->field_54 = v14;
	p->field_46 = 0;
	v14 = nox_memfile_read_u8(f);
	p->field_55 = v14;
	p->field_56 = v14;
	p->field_44 = (int)(p->field_52) * (int)(p->field_53);
	int v22 = p->field_53 * p->field_54;
	int v23 = p->field_52;
	p->data_32 = 0;
	int v24 = v23 * v22;
	if (v24 > 0) {
		for (int i = 0; i < v24; i++) {
			int v26 = nox_memfile_read_i32(f);
			*a2 = getMemByte(0x5D4594, 251576);
			if (v26 == -1) {
				nox_memfile_skip(f, 1);
				unsigned char v33 = nox_memfile_read_u8(f);
				nox_memfile_read(a2, 1, v33, f);
				a2[v33] = 0;
			}
		}
	}
	++nox_tile_def_cnt;
	return 1;
}

static int thing_reference_nox_thing_read_EDGE_411850(nox_memfile* f, uint8_t* a2) {
	int v5;            // eax
	int v7;            // ecx
	int v9;            // ecx
	int v12;           // eax
	int v14;           // eax
	unsigned char v17; // dl
	unsigned char v19; // bl
	int v20;           // ecx
	int v21;           // ebx
	int v23;           // ecx
	int v26;           // ecx
	unsigned char v28; // [esp+14h] [ebp-24h]
	char v29[32];      // [esp+18h] [ebp-20h]
	unsigned char v30; // [esp+3Ch] [ebp+4h]
	unsigned char v31; // [esp+3Ch] [ebp+4h]

	if (dword_5d4594_251572 >= 64) {
		return 0;
	}
	nox_memfile_skip(f, 4);
	v30 = nox_memfile_read_u8(f);
	nox_memfile_read(v29, 1u, v30, f);
	v5 = 3 * dword_5d4594_251572;
	v29[v30] = 0;
	strcpy((char*)getMemAt(0x85B3FC, 28644 + 20 * v5), v29);
	v7 = nox_memfile_read_i32(f);
	*getMemU32Ptr(0x85B3FC, 28680 + 60 * dword_5d4594_251572) = v7;
	v9 = nox_memfile_read_i32(f);
	*getMemU32Ptr(0x85B3FC, 28684 + 60 * dword_5d4594_251572) = v9;
	LOBYTE(v9) = nox_memfile_read_u8(f);
	*getMemU8Ptr(0x85B3FC, 28701 + 60 * dword_5d4594_251572) = v9;
	LOBYTE(v9) = nox_memfile_read_u8(f);
	v28 = v9;
	v12 = 60 * dword_5d4594_251572;
	*getMemU8Ptr(0x85B3FC, 28698 + v12) = v9;
	*getMemU16Ptr(0x85B3FC, 28690 + v12) = 0;
	LOBYTE(v9) = nox_memfile_read_u8(f);
	v14 = 60 * dword_5d4594_251572;
	*getMemU8Ptr(0x85B3FC, 28699 + v14) = v9;
	*getMemU8Ptr(0x85B3FC, 28700 + v14) = v9;
	LOBYTE(v9) = nox_memfile_read_u8(f);
	if ((uint8_t)v9 == 1) {
		return 0;
	}
	v17 = nox_memfile_read_u8(f);
	*getMemU8Ptr(0x85B3FC, 28697 + 60 * dword_5d4594_251572) = v17;
	v19 = nox_memfile_read_u8(f);
	v20 = 60 * dword_5d4594_251572;
	*getMemU8Ptr(0x85B3FC, 28696 + v20) = v19;
	*getMemU16Ptr(0x85B3FC, 28688 + v20) = 2 * (v17 + v19);
	*getMemU32Ptr(0x85B3FC, 28676 + v20) = 0;
	nox_edge_images_native[dword_5d4594_251572] = NULL;
	if (2 * v28 * (v17 + v19) > 0) {
		v21 = 2 * v28 * (v17 + v19);
		do {
			v23 = nox_memfile_read_i32(f);
			*a2 = getMemByte(0x5D4594, 251580);
			if (v23 == -1) {
				nox_memfile_skip(f, 1);
				v31 = nox_memfile_read_u8(f);
				nox_memfile_read(a2, 1u, v31, f);
				a2[v31] = 0;
			}
			--v21;
		} while (v21);
	}
	v26 = nox_memfile_read_i32(f);
	if (v26 != 0x454E4420) { // "END "
		return 0;
	}
	++dword_5d4594_251572;
	return 1;
}

static int thing_reference_nox_thing_skip_AUD_414D40(nox_memfile* f) {
	int v2;           // edx
	int v3;           // esi
	unsigned char v5; // dl

	v2 = nox_memfile_read_i32(f);
	if (v2 <= 0) {
		return 1;
	}
	v3 = v2;
	do {
		v5 = nox_memfile_read_u8(f);
		nox_memfile_skip(f, v5 + 9);
		while (1) {
			v5 = nox_memfile_read_u8(f);
			if (!v5) {
				break;
			}
			nox_memfile_skip(f, v5);
		}
		--v3;
	} while (v3);
	return 1;
}

static int thing_reference_nox_thing_read_FLOR_414DB0(nox_memfile* f) {
	unsigned char* v2;  // ecx
	unsigned char* v3;  // ecx
	unsigned char* v4;  // esi
	int v5;             // edx
	int v6;             // ecx
	int* v7;            // edx
	int v8;             // esi
	unsigned char* v9;  // edx
	unsigned char* v10; // edx
	int* v11;           // ecx
	int v12;            // edx
	unsigned char v14;  // [esp+8h] [ebp+4h]
	unsigned char v15;  // [esp+8h] [ebp+4h]

	v2 = (unsigned char*)(f->cur + 4);
	f->cur = (char*)v2;
	v3 = &v2[*v2 + 13];
	f->cur = (char*)v3;
	v4 = (unsigned char*)(f->cur + 1);
	v14 = *v3;
	f->cur = (char*)v4;
	v15 = *v4;
	f->cur = (char*)(v4 + 1);
	v4 += 3;
	v5 = v14 * v15;
	v14 = *(v4 - 2);
	f->cur = (char*)v4;
	v6 = v5 * v14;
	if (v6 > 0) {
		do {
			v7 = f->cur;
			v8 = *v7;
			v9 = (unsigned char*)(v7 + 1);
			f->cur = (char*)v9;
			if (v8 == -1) {
				v10 = (unsigned char*)(v9 + 1);
				f->cur = (char*)v10;
				f->cur += *v10 + 1;
			}
			--v6;
		} while (v6);
	}
	v11 = f->cur;
	v12 = *v11;
	f->cur = v11 + 1;
	return v12 == 1162757152;
}

static int thing_reference_nox_thing_read_EDGE_414E70(nox_memfile* f, void* a2) {
	nox_memfile_skip(f, 4);
	unsigned char v20 = nox_memfile_read_u8(f);
	nox_memfile_read(a2, 1u, v20, f);
	nox_memfile_skip(f, 9);
	unsigned char v19 = nox_memfile_read_u8(f);
	nox_memfile_skip(f, 1);
	unsigned char v6 = nox_memfile_read_u8(f);
	if (v6 == 1) {
		return 0;
	}
	unsigned char v9 = nox_memfile_read_u8(f);
	unsigned char v21 = nox_memfile_read_u8(f);
	int v11 = 2 * v19 * (v9 + v21);
	if (v11 > 0) {
		do {
			int v13 = nox_memfile_read_i32(f);
			if (v13 == -1) {
				nox_memfile_skip(f, 1);
				unsigned char n = nox_memfile_read_u8(f);
				nox_memfile_skip(f, n);
			}
			--v11;
		} while (v11);
	}
	return nox_memfile_read_i32(f) == 1162757152;
}

static int thing_reference_nox_thing_read_WALL_414F60(nox_memfile* f, void* a2) {
	uint8_t* buf = a2;
	nox_memfile_skip(f, 4);
	unsigned char n = nox_memfile_read_u8(f);
	nox_memfile_read(buf, 1u, n, f);
	buf[n] = 0;
	nox_memfile_skip(f, 14);

	unsigned char group_count = 0;
	nox_memfile_read64align_40AD60((char*)&group_count, 1, 1, f);
	for (int i = 0; i < group_count; ++i) {
		if (i >= 8) {
			return 0;
		}
		n = nox_memfile_read_u8(f);
		nox_memfile_read(buf, 1u, n, f);
		buf[n] = 0;
	}

	for (int i = 0; i < 3; ++i) {
		n = nox_memfile_read_u8(f);
		nox_memfile_read(buf, 1u, n, f);
		buf[n] = 0;
	}
	nox_memfile_skip(f, 1);

	for (int i = 0; i < 15; ++i) {
		unsigned char record_count = 0;
		nox_memfile_read64align_40AD60((char*)&record_count, 1, 1, f);
		for (int j = 0; j < record_count; ++j) {
			for (int k = 0; k < 4; ++k) {
				nox_memfile_skip(f, 8);
				if (nox_memfile_read_i32(f) == -1) {
					nox_memfile_skip(f, 1);
					n = nox_memfile_read_u8(f);
					nox_memfile_skip(f, n);
				}
			}
		}
	}
	return nox_memfile_read_i32(f) == 1162757152;
}

static int thing_reference_nox_thing_skip_spells_415100(nox_memfile* f) {
	int count = nox_memfile_read_i32(f);
	if (count <= 0) {
		return 1;
	}
	for (int i = 0; i < count; ++i) {
		unsigned char n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n + 3);

		n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);
		for (int j = 0; j < 2; ++j) {
			if (nox_memfile_read_i32(f) == -1) {
				nox_memfile_skip(f, 1);
				n = nox_memfile_read_u8(f);
				nox_memfile_skip(f, n);
			}
		}

		nox_memfile_skip(f, 4);
		n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);
		short block_size = nox_memfile_read_i16(f);
		nox_memfile_skip(f, block_size);
		for (int j = 0; j < 3; ++j) {
			n = nox_memfile_read_u8(f);
			nox_memfile_skip(f, n);
		}
	}
	return 1;
}

static int thing_reference_nox_thing_read_image_415240(nox_memfile* f) {
	int count = nox_memfile_read_i32(f);
	if (count <= 0) {
		return 1;
	}
	for (int i = 0; i < count; ++i) {
		unsigned char n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);
		unsigned char kind = nox_memfile_read_u8(f);
		unsigned char refs = 1;
		if (kind == 2) {
			refs = nox_memfile_read_u8(f);
			nox_memfile_skip(f, 1);
			n = nox_memfile_read_u8(f);
			nox_memfile_skip(f, n);
			if (!refs) {
				continue;
			}
		}
		for (int j = 0; j < refs; ++j) {
			if (nox_memfile_read_i32(f) == -1) {
				nox_memfile_skip(f, 1);
				n = nox_memfile_read_u8(f);
				nox_memfile_skip(f, n);
			}
		}
	}
	return 1;
}

static int thing_reference_nox_thing_read_ability_415320(nox_memfile* f) {
	int count = nox_memfile_read_i32(f);
	if (count <= 0) {
		return 1;
	}
	for (int i = 0; i < count; ++i) {
		unsigned char n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n + 1);
		for (int j = 0; j < 3; ++j) {
			if (nox_memfile_read_i32(f) == -1) {
				nox_memfile_skip(f, 1);
				n = nox_memfile_read_u8(f);
				nox_memfile_skip(f, n);
			}
		}
		n = nox_memfile_read_u8(f);
		nox_memfile_skip(f, n);
		short block_size = nox_memfile_read_i16(f);
		nox_memfile_skip(f, block_size);
		for (int j = 0; j < 3; ++j) {
			n = nox_memfile_read_u8(f);
			nox_memfile_skip(f, n);
		}
	}
	return 1;
}

static int thing_reference_nox_thing_read_audio_415660(nox_memfile* f, char* buf) {
	int count = nox_memfile_read_i32(f);
	if (count <= 0) {
		return 1;
	}
	for (int i = 0; i < count; i++) {
		if (!sub_452BD0(f, buf)) {
			return 0;
		}
		if (i + 1 >= count) {
			return 1;
		}
	}
	return 1;
}

typedef struct {
 nox_tileDef_t tiles[176]; unsigned char edges[64*60]; void* images[64];
 uint32_t floors,edge_count; unsigned char floor_byte,edge_byte;
} thing_fixture_state;
static unsigned char thing_fixture_opaque[16];
void* thing_fixture_save_411540(void) {
 thing_fixture_state* s=calloc(1,sizeof(*s)); if(!s)return NULL;
 memcpy(s->tiles,nox_tile_defs_arr,sizeof(s->tiles));memcpy(s->edges,getMemAt(0x85B3FC,28644),sizeof(s->edges));memcpy(s->images,nox_edge_images_native,sizeof(s->images));
 s->floors=nox_tile_def_cnt;s->edge_count=dword_5d4594_251572;s->floor_byte=getMemByte(0x5D4594,251576);s->edge_byte=getMemByte(0x5D4594,251580);return s;
}
void thing_fixture_restore_411540(void* p) {
 thing_fixture_state* s=p;memcpy(nox_tile_defs_arr,s->tiles,sizeof(s->tiles));memcpy(getMemAt(0x85B3FC,28644),s->edges,sizeof(s->edges));memcpy(nox_edge_images_native,s->images,sizeof(s->images));
 nox_tile_def_cnt=s->floors;dword_5d4594_251572=s->edge_count;*getMemU8Ptr(0x5D4594,251576)=s->floor_byte;*getMemU8Ptr(0x5D4594,251580)=s->edge_byte;
}
size_t thing_fixture_state_size_411540(void){return sizeof(thing_fixture_state);}
void thing_fixture_seed_411540(uint32_t floor,uint32_t edge) {
 memset(nox_tile_defs_arr,0xB6,sizeof(nox_tile_defs_arr));
 for(int i=0;i<176;i++){nox_tile_defs_arr[i].name[31]=0;nox_tile_defs_arr[i].data_32=(nox_video_bag_image_t**)thing_fixture_opaque;}
 memset(getMemAt(0x85B3FC,28644),0xC3,64*60);
 for(int i=0;i<64;i++)nox_edge_images_native[i]=thing_fixture_opaque;
 nox_tile_def_cnt=floor;dword_5d4594_251572=edge;*getMemU8Ptr(0x5D4594,251576)=0x61;*getMemU8Ptr(0x5D4594,251580)=0x62;
}
void* thing_fixture_memfile_411540(const unsigned char* data,int n,int start) {
 nox_memfile* f=calloc(1,sizeof(*f));if(!f)return NULL;f->data=malloc((size_t)n+32);if(!f->data){free(f);return NULL;}
 memset(f->data,0xD8,(size_t)n+32);if(n)memcpy(f->data,data,n);f->size=n;f->cur=f->data+start;f->end=f->data+n;return f;
}
void thing_fixture_memfile_free_411540(void* p){nox_memfile* f=p;if(f){free(f->data);free(f);}}
intptr_t thing_fixture_offset_411540(void* p){nox_memfile* f=p;return f->cur-f->data;}
void* thing_fixture_input_411540(void* p){return ((nox_memfile*)p)->data;}
uintptr_t thing_fixture_pointer_411540(void* p){return (uintptr_t)p;}
void thing_fixture_free_411540(void* p){free(p);}
size_t thing_fixture_layout_411540(int i){
 const size_t a[]={sizeof(void*),sizeof(nox_memfile),offsetof(nox_memfile,data),offsetof(nox_memfile,size),offsetof(nox_memfile,cur),offsetof(nox_memfile,end),sizeof(nox_tileDef_t),offsetof(nox_tileDef_t,data_32),offsetof(nox_tileDef_t,field_36),offsetof(nox_tileDef_t,color_48),offsetof(nox_tileDef_t,field_58)};return a[i];
}
int thing_fixture_call_411540(uint32_t entry,int reference,nox_memfile* f,void* p) {
 switch(entry){
 case 0x411540:return reference?thing_reference_nox_thing_read_FLOR_411540(f,p):nox_thing_read_FLOR_411540(f,p);
 case 0x411850:return reference?thing_reference_nox_thing_read_EDGE_411850(f,p):nox_thing_read_EDGE_411850(f,p);
 case 0x414D40:return reference?thing_reference_nox_thing_skip_AUD_414D40(f):nox_thing_skip_AUD_414D40(f);
 case 0x414DB0:return reference?thing_reference_nox_thing_read_FLOR_414DB0(f):nox_thing_read_FLOR_414DB0(f);
 case 0x414E70:return reference?thing_reference_nox_thing_read_EDGE_414E70(f,p):nox_thing_read_EDGE_414E70(f,p);
 case 0x414F60:return reference?thing_reference_nox_thing_read_WALL_414F60(f,p):nox_thing_read_WALL_414F60(f,p);
 case 0x415100:return reference?thing_reference_nox_thing_skip_spells_415100(f):nox_thing_skip_spells_415100(f);
 case 0x415240:return reference?thing_reference_nox_thing_read_image_415240(f):nox_thing_read_image_415240(f);
 case 0x415320:return reference?thing_reference_nox_thing_read_ability_415320(f):nox_thing_read_ability_415320(f);
 case 0x415660:return reference?thing_reference_nox_thing_read_audio_415660(f,p):nox_thing_read_audio_415660(f,p);
 default:return -999;
 }
}

int thing_fixture_signed_char_411540(void){return (char)255 < 0;}
uint32_t thing_fixture_rgb_411540(uint8_t r,uint8_t g,uint8_t b){return nox_color_rgb_4344A0(r,g,b);}

int thing_fixture_audio_real_415660(nox_memfile* f,void* p){return sub_452BD0(f,p);}
