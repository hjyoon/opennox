int nox_xxx_tile_422C10(int a1, void* region) {
	int column_index;
	int v44, v45, v46, v47, v48, v49, v51;
	uint32_t* v52;
	nox_tile_list_node_t** v52_head;
	int v64, v66, v67;
	bool v70;
	int v75;
	int v76, v77, v78, v79;
	int v81;
	int4 v83;

	if ((short)a1 < 3) {
		return 0;
	}
	if (!nox_crypt_IsReadOnly()) {
		int max_x = 0;
		int max_y = 0;
		if (region) {
			sub_428170(region, &v83);
			v76 = v83.field_0 / 23;
			v77 = v83.field_4 / 23;
			max_x = v83.field_8 / 23;
			max_y = v83.field_C / 23;
		} else {
			bool found = false;
			for (int y = 0; y < ptr_5D4594_2650668_cap && !found; ++y) {
				for (int x = 0; x < ptr_5D4594_2650668_cap; ++x) {
					if (ptr_5D4594_2650668[x][y].field_0) {
						v77 = y;
						found = true;
						break;
					}
				}
			}
			if (!found) {
				return 1;
			}
			// v3 scans the original 128-row map when finding the upper edge.
			found = false;
			for (int y = 127; y >= 0 && !found; --y) {
				for (int x = 0; x < ptr_5D4594_2650668_cap; ++x) {
					if (ptr_5D4594_2650668[x][y].field_0) {
						max_y = y;
						found = true;
						break;
					}
				}
			}
			found = false;
			for (int x = 0; x < 128 && !found; ++x) {
				for (int y = 0; y < ptr_5D4594_2650668_cap; ++y) {
					if (ptr_5D4594_2650668[x][y].field_0) {
						v76 = x;
						found = true;
						break;
					}
				}
			}
			found = false;
			for (int x = ptr_5D4594_2650668_cap - 1; x >= 0 && !found; --x) {
				for (int y = 0; y < ptr_5D4594_2650668_cap; ++y) {
					if (ptr_5D4594_2650668[x][y].field_0) {
						max_x = x;
						found = true;
						break;
					}
				}
			}
		}
		v78 = max_x - v76 + 1;
		v79 = max_y - v77 + 1;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v76, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v77, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v78, 4u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v79, 4u);
		if (!region) {
			for (int y = v77; y <= max_y; ++y) {
				for (int x = v76; x <= max_x; ++x) {
					obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[x][y];
					nox_xxx_fileReadWrite_426AC0_file3_fread(&tile->field_0, 1u);
					int* fields[2] = {&tile->field_1, &tile->field_6};
					nox_tile_list_node_t* heads[2] = {tile->field_5, tile->field_10};
					for (int layer = 0; layer < 2; ++layer) {
						if (!(tile->field_0 & (1u << layer))) {
							continue;
						}
						nox_xxx_fileReadWrite_426AC0_file3_fread(fields[layer], 16u);
						uint8_t count = 0;
						for (nox_tile_list_node_t* node = heads[layer]; node; node = node->next) {
							++count;
						}
						nox_xxx_fileReadWrite_426AC0_file3_fread(&count, 1u);
						for (nox_tile_list_node_t* node = heads[layer]; node; node = node->next) {
							nox_xxx_fileReadWrite_426AC0_file3_fread(node, 16u);
						}
					}
				}
			}
			return 1;
		}
		for (int y = v77; y <= max_y; ++y) {
			for (int x = v76; x <= max_x; ++x) {
				int* fields = NULL;
				nox_tile_list_node_t* head = NULL;
				int2 point = {23 * x + 11, 23 * y + 34};
				if (!(((uint8_t)y + (uint8_t)x) & 1)) {
					if (y & 1) {
						obj_5D4594_2650668_t* tile =
							&ptr_5D4594_2650668[(point.field_0 - 34) / 46][23 * (y + 1) / 46];
						if (tile->field_0 & 1) {
							fields = &tile->field_1;
							head = tile->field_5;
						}
					} else {
						obj_5D4594_2650668_t* tile =
							&ptr_5D4594_2650668[(point.field_0 - 11) / 46][23 * y / 46];
						if (tile->field_0 & 2) {
							fields = &tile->field_6;
							head = tile->field_10;
						}
					}
				}
				uint8_t present = fields && nox_xxx_wallMath_427F30(&point, region);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&present, 1u);
				if (present) {
					nox_xxx_fileReadWrite_426AC0_file3_fread(fields, 16u);
					uint8_t count = 0;
					for (nox_tile_list_node_t* node = head; node; node = node->next) {
						++count;
					}
					nox_xxx_fileReadWrite_426AC0_file3_fread(&count, 1u);
					for (nox_tile_list_node_t* node = head; node; node = node->next) {
						nox_xxx_fileReadWrite_426AC0_file3_fread(node, 16u);
					}
				}
			}
		}
		return 1;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v76, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v77, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v78, 4u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v79, 4u);
	if (!region) {
		for (int y = 0; y < ptr_5D4594_2650668_cap; ++y) {
			for (int x = 0; x < ptr_5D4594_2650668_cap; ++x) {
				ptr_5D4594_2650668[x][y].field_0 = 0;
			}
		}
		for (int y = v77; y < v77 + v79; ++y) {
			for (int x = v76; x < v76 + v78; ++x) {
				obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[x][y];
				nox_xxx_fileReadWrite_426AC0_file3_fread(&tile->field_0, 1u);
				int* fields[2] = {&tile->field_1, &tile->field_6};
				nox_tile_list_node_t** heads[2] = {&tile->field_5, &tile->field_10};
				for (int layer = 0; layer < 2; ++layer) {
					if (!(tile->field_0 & (1u << layer))) {
						continue;
					}
					nox_xxx_fileReadWrite_426AC0_file3_fread(fields[layer], 16u);
					uint8_t count = 0;
					nox_xxx_fileReadWrite_426AC0_file3_fread(&count, 1u);
					nox_tile_list_node_t** tail = heads[layer];
					*tail = NULL;
					for (unsigned int i = 0; i < count; ++i) {
						nox_tile_list_node_t* node = nox_xxx_tileListAddNewSubtile_422160(0, 0, 0, 0);
						if (!node) {
							return 0;
						}
						nox_xxx_fileReadWrite_426AC0_file3_fread(node, 16u);
						*tail = node;
						tail = &node->next;
					}
				}
			}
		}
		return 1;
	}
	sub_428170(region, &v83);
	v44 = v83.field_0 / 23;
	v76 = v83.field_0 / 23;
	v45 = v83.field_4 / 23;
	v77 = v83.field_4 / 23;
	v46 = v83.field_4 / 23;
	if (v83.field_4 / 23 >= v83.field_4 / 23 + v79) {
		return 1;
	}
	v47 = v78;
	do {
		column_index = v44;
		if (v44 < v47 + v44) {
			v81 = 23 * v44;
			v48 = 23 * v44;
			do {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 1u);
				if ((uint8_t)a1) {
					v52 = NULL;
					v52_head = NULL;
					if (v46 & 1) {
						v64 = (v48 - 23) / 46;
						v66 = 23 * (v46 + 1) / 46;
						if (nox_common_gameFlags_check_40A5C0(0x400000)) {
							nox_tile_coord_entry_t* entry =
								nox_xxx_tileAllocTileInCoordList_5040A0(v64, v66, COERCE_FLOAT(1));
							if (entry) {
								v52 = &entry->layer->field_0;
								v52_head = &entry->layer->subtiles;
							}
						} else {
							obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v64][v66];
							tile->field_0 |= 1u;
							v52 = (uint32_t*)&tile->field_1;
							v52_head = &tile->field_5;
						}
					} else {
						v49 = v48 / 46;
						v51 = 23 * v46 / 46;
						if (nox_common_gameFlags_check_40A5C0(0x400000)) {
							nox_tile_coord_entry_t* entry =
								nox_xxx_tileAllocTileInCoordList_5040A0(v49, v51, COERCE_FLOAT(2));
							if (entry) {
								v52 = &entry->layer->field_0;
								v52_head = &entry->layer->subtiles;
							}
						} else {
							obj_5D4594_2650668_t* tile = &ptr_5D4594_2650668[v49][v51];
							tile->field_0 |= 2u;
							v52 = (uint32_t*)&tile->field_6;
							v52_head = &tile->field_10;
						}
					}
					if (!v52 || !v52_head) {
						return 0;
					}
					nox_xxx_fileReadWrite_426AC0_file3_fread(v52, 0x10u);
					nox_xxx_fileReadWrite_426AC0_file3_fread(&v75, 1u);
					*v52_head = NULL;
					if ((uint8_t)v75) {
						v67 = 0;
						nox_tile_list_node_t** tail = v52_head;
						if ((uint8_t)v75) {
							do {
								nox_tile_list_node_t* node = nox_xxx_tileListAddNewSubtile_422160(0, 0, 0, 0);
								if (!node) {
									return 0;
								}
								nox_xxx_fileReadWrite_426AC0_file3_fread(node, 0x10u);
								*tail = node;
								tail = &node->next;
								++v67;
							} while (v67 < (unsigned char)v75);
						}
					}
				}
				v47 = v78;
				v44 = v76;
				v48 = v81 + 23;
				v70 = ++column_index < v78 + v76;
				v81 += 23;
			} while (v70);
			v45 = v77;
		}
		++v46;
	} while (v46 < v45 + v79);
	return 1;
}
