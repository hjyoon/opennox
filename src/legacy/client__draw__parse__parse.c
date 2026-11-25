#include "draw_allocation.h"
#include "client__draw__parse__parse.h"
#include "GAME1_2.h"
#include "operators.h"

//----- (0044B8B0) --------------------------------------------------------
int nox_xxx_loadVectorAnimated_44B8B0(int a1, nox_memfile* f) {
	*(uint16_t*)(a1 + 40) = nox_memfile_read_u8(f);

	*(uint16_t*)(a1 + 42) = nox_memfile_read_u8(f);

	const uint8_t anim_kind_length = nox_memfile_read_u8(f);

	char animation_kind[256];
	nox_memfile_read(animation_kind, 1u, anim_kind_length, f);
	animation_kind[anim_kind_length] = 0;

	*(uint32_t*)(a1 + 44) = get_animation_kind_id_44B4C0(animation_kind);

	return 1;
}

//----- (0044BC50) --------------------------------------------------------
int nox_xxx_loadVectorAnimated_44BC50(int a1, nox_memfile* f) {
	int v2;          // ebp
	int v3;          // esi
	void* v4;        // eax
	int v5;          // ebx
	int v7;          // ebp
	char v9;         // dl
	int v11;         // esi
	int v13;         // [esp+10h] [ebp-90h]
	int v14;         // [esp+14h] [ebp-8Ch]
	const char* v15; // [esp+1Ch] [ebp-84h]
	char v16[128];   // [esp+20h] [ebp-80h]

	v2 = 0;
	v14 = 0;
	while (1) {
		v13 = v2 >= 16 ? v2 + 4 : v2;
		v3 = a1;
		v4 = nox_draw_memory_calloc(*(short*)(a1 + 40), 4);
		*(uint32_t*)(v13 + a1 + 4) = v4;
		if (!v4) {
			break;
		}
		v5 = 0;
		if (*(uint16_t*)(a1 + 40) > 0) {
			do {
				v7 = nox_memfile_read_i32(f);
				v16[0] = getMemByte(0x5D4594, 830844);
				if (v7 == -1) {
					v9 = nox_memfile_read_i8(f);
					LOBYTE(v15) = v9;
					v11 = nox_memfile_read_u8(f);
					nox_memfile_read(v16, 1u, v11, f);
					v16[v11] = 0;
					v3 = a1;
				}
				*(uint32_t*)(*(uint32_t*)(v13 + v3 + 4) + 4 * ++v5 - 4) = nox_xxx_readImgMB_42FAA0(v7, v15, v16);
			} while (v5 < *(short*)(v3 + 40));
			v2 = v14;
		}
		v2 += 4;
		v14 = v2;
		if (v2 >= 32) {
			return 1;
		}
	}
	return 0;
}
