#ifndef OPENNOX_GAME4_TILES_CURRENT_C_REFERENCE_H
#define OPENNOX_GAME4_TILES_CURRENT_C_REFERENCE_H
#define nox_xxx_tileAllocTileInCoordList_5040A0 game4_tiles_ref_alloc
#define nox_xxx_tileInit_504150 game4_tiles_ref_init
#define sub_4D3C70 game4_tiles_test_snapshot
#define nox_xxx_tileInitdataClear_4D3C50 game4_tiles_test_restore
#define nox_xxx_tileCheckImage_51D540 game4_tiles_test_image
#define nox_xxx_tileCheckImageVari_51D570 game4_tiles_test_variation
#define nox_xxx_tile_51D5C0 game4_tiles_test_mode
#define sub_51D8F0 game4_tiles_test_base
#define nox_xxx_tileCheckByte3_544070 game4_tiles_test_byte3
#define nox_xxx_tileCheckByte4_5440A0 game4_tiles_test_byte4
#define nox_xxx_tileSubtile_544310 game4_tiles_test_subtile
static nox_tile_coord_entry_t* nox_xxx_tileAllocTileInCoordList_5040A0(int a1, int a2, float a3) {
	nox_tile_coord_entry_t* entry = calloc(1, sizeof(*entry));
	if (!entry) {
		return NULL;
	}
	entry->layer = calloc(1, sizeof(*entry->layer));
	if (!entry->layer) {
		free(entry);
		return NULL;
	}
	entry->prev = dword_5d4594_1599556;
	if (entry->prev) {
		entry->prev->next = entry;
	}
	dword_5d4594_1599556 = entry;
	++*getMemU32Ptr(0x5D4594, 1599560);
	uint32_t kind_bits;
	memcpy(&kind_bits, &a3, sizeof(kind_bits));
	entry->kind = (uint8_t)kind_bits;
	entry->x = (float)a1 * 46.0f;
	entry->y = (float)a2 * 46.0f;
	if (entry->kind == 1) {
		entry->x += 23.0f;
	} else {
		entry->y += 23.0f;
	}
	return entry;
}

static int nox_xxx_tileInit_504150(int a1, int a2) {
	nox_tile_coord_entry_t* v5; // edi
	nox_tile_list_node_t* i;    // esi
	float2 a1a;     // [esp+Ch] [ebp-50h]
	char v8[72];    // [esp+14h] [ebp-48h]
	float v9;       // [esp+60h] [ebp+4h]
	float v10;      // [esp+64h] [ebp+8h]

	if (*getMemIntPtr(0x587000, 229704) == -1) {
		if (nox_tile_def_cnt > 0) {
			int v2 = 0;
			for (int i = 0; i < nox_tile_def_cnt; i++) {
				nox_tileDef_t* p = &nox_tile_defs_arr[i];
				if (strcmp(&p->name[0], "TransparentFloor") == 0) {
					*getMemU32Ptr(0x587000, 229704) = i;
					v2 = i;
					break;
				}
			}
			if (v2 == -1) {
				return 0;
			}
		}
	}
	memcpy(v8, sub_4D3C70(), sizeof(v8));
	v5 = dword_5d4594_1599556;
	if (dword_5d4594_1599556) {
		v9 = (double)a1;
		v10 = (double)a2;
		do {
			a1a.field_0 = v9 + v5->x;
			a1a.field_4 = v10 + v5->y;
			nox_xxx_tileCheckImage_51D540(v5->layer->field_0);
			nox_xxx_tileCheckImageVari_51D570(v5->layer->field_1);
			nox_xxx_tile_51D5C0(1);
			if (v5->layer->field_0 != *getMemU32Ptr(0x587000, 229704)) {
				sub_51D8F0(&a1a);
			}
			for (i = v5->layer->subtiles; i; i = i->next) {
				nox_xxx_tileCheckByte3_544070(i->field_2);
				nox_xxx_tileCheckByte4_5440A0(i->field_3);
				nox_xxx_tileCheckImage_51D540(i->field_0);
				nox_xxx_tileCheckImageVari_51D570(i->field_1);
				nox_xxx_tileSubtile_544310(&a1a);
			}
			v5 = v5->prev;
		} while (v5);
	}
	nox_xxx_tileInitdataClear_4D3C50(v8);
	return 1;
}

#undef nox_xxx_tileAllocTileInCoordList_5040A0
#undef nox_xxx_tileInit_504150
#undef sub_4D3C70
#undef nox_xxx_tileInitdataClear_4D3C50
#undef nox_xxx_tileCheckImage_51D540
#undef nox_xxx_tileCheckImageVari_51D570
#undef nox_xxx_tile_51D5C0
#undef sub_51D8F0
#undef nox_xxx_tileCheckByte3_544070
#undef nox_xxx_tileCheckByte4_5440A0
#undef nox_xxx_tileSubtile_544310
#endif
