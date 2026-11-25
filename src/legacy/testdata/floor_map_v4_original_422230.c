int nox_server_mapRWFloorMap_422230(void* region) {
	int row_delta;
	int v2;             // ecx
	char v3;            // bl
	int v4;             // esi
	int l;              // edx
	uint32_t* v6;       // eax
	char v7;            // bl
	int v8;             // esi
	int m;              // edx
	int v10;            // ecx
	uint32_t* v11;      // eax
	char v12;           // bl
	int v13;            // edx
	uint8_t** v14;      // esi
	uint8_t* v15;       // ecx
	int v16;            // eax
	uint8_t** v17;      // esi
	int v18;            // ebp
	char v19;           // bl
	int n;              // edx
	uint8_t* v21;       // ecx
	int v22;            // eax
	int v23;            // edi
	int v24;            // eax
	int v25;            // esi
	int ii;             // ebx
	int v27;            // ecx
	unsigned short v28; // dx
	int v29;            // ecx
	int v30;            // eax
	uint8_t* v31;       // ecx
	uint8_t* v32;       // ecx
	int v33;            // ebx
	int v34;            // ebp
	char v35;           // al
	int v36;            // esi
	uint8_t* v37;       // eax
	uint8_t* v38;       // edi
	uint8_t* v39;       // eax
	unsigned short v40; // cx
	int v41;            // ebp
	int v42;            // esi
	unsigned short v43; // cx
	unsigned short k;   // cx
	short v47;          // bx
	int v48;            // esi
	int v49;            // edi
	int v50;            // eax
	int v51;            // ebx
	int v52;            // edi
	int v53;            // esi
	uint8_t* v54;       // eax
	uint8_t* v55;       // ecx
	int v56;            // edi
	int v58;            // esi
	int v59;            // [esp+10h] [ebp-40h]
	int v60;            // [esp+14h] [ebp-3Ch]
	int v61;            // [esp+18h] [ebp-38h]
	int v62;            // [esp+1Ch] [ebp-34h]
	int v63 = 0;        // [esp+20h] [ebp-30h]
	int v64 = 0;        // [esp+24h] [ebp-2Ch]
	int v65 = 0;        // [esp+28h] [ebp-28h]
	int v66;            // [esp+2Ch] [ebp-24h]
	int v67;            // [esp+30h] [ebp-20h]
	int v68;            // [esp+34h] [ebp-1Ch]
	int2 v69;           // [esp+38h] [ebp-18h]
	int4 v70;           // [esp+40h] [ebp-10h]

	v62 = 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v62, 2u);
	if ((short)v62 > 4) {
		return 0;
	}
	if ((short)v62 <= 3) {
		return nox_xxx_tile_422C10(v62, region);
	}
	if (nox_crypt_IsReadOnly()) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v60, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v61, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v66, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v67, 4u);
		if (region) {
			sub_428170(region, &v70);
			v41 = v70.field_0 / 23 - v60;
			v42 = v70.field_4 / 23 - v61;
			row_delta = v70.field_4 / 23 - v61;
			v61 = v70.field_4 / 23;
			v60 = v70.field_0 / 23;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
			v43 = v59;
			if ((uint16_t)v59 != (uint16_t)-1) {
				while (1) {
					uint32_t* tile_fields = NULL;
					nox_tile_list_node_t** tile_head = NULL;
					v50 = v41 + (v43 >> 8);
					v51 = v42 + (unsigned char)v43;
					if (((uint8_t)v42 + (uint8_t)v43) & 1) {
						v56 = (23 * v50 - 23) / 46;
						v58 = (23 * v51 + 23) / 46;
						if (nox_common_gameFlags_check_40A5C0(0x400000)) {
							nox_tile_coord_entry_t* entry =
								nox_xxx_tileAllocTileInCoordList_5040A0(v56, v58, COERCE_FLOAT(1));
							if (entry) {
								tile_fields = &entry->layer->field_0;
								tile_head = &entry->layer->subtiles;
							}
						} else {
							obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v56][v58];
							tile->field_0 |= 0x1;
							tile_fields = (uint32_t*)&tile->field_1;
							tile_head = &tile->field_5;
						}
					} else {
						v52 = 23 * v50 / 46;
						v53 = 23 * v51 / 46;
						if (nox_common_gameFlags_check_40A5C0(0x400000)) {
							nox_tile_coord_entry_t* entry =
								nox_xxx_tileAllocTileInCoordList_5040A0(v52, 23 * v51 / 46, COERCE_FLOAT(2));
							if (entry) {
								tile_fields = &entry->layer->field_0;
								tile_head = &entry->layer->subtiles;
							}
						} else {
							obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v52][v53];
							tile->field_0 |= 0x2;
							tile_fields = (uint32_t*)&tile->field_6;
							tile_head = &tile->field_10;
						}
					}
					if (!tile_fields || !nox_xxx_tileReadOne_422A40(v62, tile_fields, tile_head)) {
						return 0;
					}
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
					v43 = v59;
					if ((uint16_t)v59 == (uint16_t)-1) {
						break;
					}
					v42 = row_delta;
				}
				return 1;
			}
		} else {
			for (int i = 0; i < ptr_5D4594_2650668_cap; i++) {
				for (int j = 0; j < ptr_5D4594_2650668_cap; j++) {
					*(uint8_t*)(&ptr_5D4594_2650668[j][i].field_0) = 0;
				}
			}
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
			for (k = v59; (uint16_t)v59 != (uint16_t)-1; k = v59) {
				v47 = 0;
				v48 = (k >> 8) & 0x7F;
				v49 = k & 0x7F;
				if ((k & 0x8000u) != 0) {
					v47 = -32768;
				}
				if ((k & 0x80u) != 0) {
					LOBYTE(v47) = v47 | 0x80;
				}
				if (v47 < 0) {
					obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v48][v49];
					tile->field_0 |= 0x1;
					if (!nox_xxx_tileReadOne_422A40(v62, (uint32_t*)&tile->field_1, &tile->field_5)) {
						return 0;
					}
				}
				if (v47 & 0x80) {
					obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v48][v49];
					tile->field_0 |= 0x2;
					if (!nox_xxx_tileReadOne_422A40(v62, (uint32_t*)&tile->field_6, &tile->field_10)) {
						return 0;
					}
				}
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
			}
		}
		return 1;
	}
	if (region) {
		sub_428170(region, &v70);
		v60 = v70.field_0 / 23;
		v18 = v70.field_8 / 23;
		v64 = v70.field_8 / 23;
		v63 = v70.field_C / 23;
		v61 = v70.field_4 / 23;
		v66 = v70.field_8 / 23 - v70.field_0 / 23 + 1;
		v67 = v70.field_C / 23 - v70.field_4 / 23 + 1;
		goto LABEL_43;
	}
	v60 = ptr_5D4594_2650668_cap;
	v61 = ptr_5D4594_2650668_cap;
	v64 = -1;
	v63 = -1;
	for (int x = 0; x < ptr_5D4594_2650668_cap; ++x) {
		for (int y = 0; y < ptr_5D4594_2650668_cap; ++y) {
			if (ptr_5D4594_2650668[x][y].field_0 != 0) {
				if (x < v60) v60 = x;
				if (x > v64) v64 = x;
				if (y < v61) v61 = y;
				if (y > v63) v63 = y;
			}
		}
	}
	if (v64 < 0 || v63 < 0) {
		return 1;
	}
	v18 = v64;
	v66 = v18 - v60 + 1;
	v67 = v63 - v61 + 1;
LABEL_43:
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v60, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v61, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v66, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v67, 4u);
	if (region) {
		v33 = v61;
		if (v61 > v63) {
			v59 = 0xFFFF;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
			return 1;
		}
		v34 = v60;
		while (1) {
			v35 = v34;
			v65 = v34;
			if (v34 <= v64) {
				v36 = 23 * v34 + 11;
				while (1) {
					if (!(((uint8_t)v33 + v35) & 1)) {
						if (v33 & 1) {
							obj_5D4594_2650668_t* tile =
								&ptr_5D4594_2650668[(v36 - 34) / 46][23 * (v33 + 1) / 46];
							if (tile->field_0 & 1) {
								v69.field_0 = v36;
								v69.field_4 = 23 * v33 + 34;
								if (nox_xxx_wallMath_427F30(&v69, region)) {
									LOBYTE(v40) = 0;
									HIBYTE(v40) = v65;
									v68 = v33 | v40;
									nox_xxx_fileReadWrite_426AC0_file3_fread(&v68, 2u);
									nox_xxx_tileReadOne_422A40(
										v62, (uint32_t*)&tile->field_1, &tile->field_5);
								}
								v34 = v60;
								goto LABEL_73;
							}
						} else {
							obj_5D4594_2650668_t* tile =
								&ptr_5D4594_2650668[(v36 - 11) / 46][23 * v33 / 46];
							if (tile->field_0 & 2) {
								v69.field_0 = v36;
								v69.field_4 = 23 * v33 + 34;
								if (nox_xxx_wallMath_427F30(&v69, region)) {
									LOBYTE(v40) = 0;
									HIBYTE(v40) = v65;
									v68 = v33 | v40;
									nox_xxx_fileReadWrite_426AC0_file3_fread(&v68, 2u);
									nox_xxx_tileReadOne_422A40(
										v62, (uint32_t*)&tile->field_6, &tile->field_10);
								}
								v34 = v60;
								goto LABEL_73;
							}
						}
					}
				LABEL_73:
					v35 = v65 + 1;
					v36 += 23;
					if (++v65 > v64) {
						break;
					}
				}
			}
			++v33;
			if (v33 > v63) {
				v59 = 0xFFFF;
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
				return 1;
			}
		}
	}
	v23 = v61;
	if (v61 <= v63) {
		do {
			for (ii = v60; ii <= v18; ++ii) {
				obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[ii][v23];
				if (tile->field_0 & 3) {
					uint16_t encoded = (uint16_t)(((uint16_t)ii << 8) | (uint8_t)v23);
					if (tile->field_0 & 1) encoded |= 0x8000u;
					if (tile->field_0 & 2) encoded |= 0x0080u;
					nox_xxx_fileReadWrite_426AC0_file3_fread(&encoded, 2u);
				}
				if (tile->field_0 & 1) {
					nox_xxx_tileReadOne_422A40(v62, (uint32_t*)&tile->field_1, &tile->field_5);
				}
				if (tile->field_0 & 2) {
					nox_xxx_tileReadOne_422A40(v62, (uint32_t*)&tile->field_6, &tile->field_10);
				}
			}
			++v23;
		} while (v23 <= v63);
	}
	v59 = 0xFFFF;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v59, 2u);
	return 1;
}
