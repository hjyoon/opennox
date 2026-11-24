// Floor-brush compatibility bodies from 3febadeaa, kept for the R304 C fixture.

int nox_floorBrushApplyNative_51DA70(int x, int y, nox_floor_brush_t* brush, int layer, int merge) {
	nox_floor_tile_storage_t* floor_tiles = nox_floor_tile_storage();
    int tile;
    int variation;
    if (!brush) {
        tile =255;
        variation =0;
    } else if (brush->pattern && !brush->kind) {
        tile =brush->tile_id;
        variation =0;
        brush->anchor_x =x;
        brush->anchor_y =y;
        brush->kind =(uint8_t)layer;
    } else {
        int relative_x =x;
        int relative_y =y;
        int parity =layer !=1;
        if (brush->pattern) {
            relative_x -=brush->anchor_x;
            relative_y -=brush->anchor_y;
            parity =brush->kind ==layer ?0 : (brush->kind ==1 ?1 :-1);
        }
        tile =brush->tile_id;
        int width =floor_tiles->definitions[tile].field_52;
        int height =floor_tiles->definitions[tile].field_53;
        int column =(relative_y +relative_x) %width;
        int row =(parity +relative_y -relative_x) %height;
        if (column <0) column +=width;
        if (row <0) row +=height;
        variation =*getMemU32Ptr(0x973F18,35916) ==1
            ? dword_5d4594_3835348 :column +row *width;
    }
    if (layer !=1 && layer !=2) return 0;

    obj_5D4594_2650668_t* cell =&ptr_5D4594_2650668[x][y];
    int* base_tile =layer ==1 ?&cell->field_1 :&cell->field_6;
    int* base_variation =layer ==1 ?&cell->field_2 :&cell->field_7;
    nox_tile_list_node_t** head =layer ==1 ?&cell->field_5 :&cell->field_10;
    // The left scalar fields and native head are separated by alignment
    // padding; they cannot be overlaid by a standalone nox_tile_layer_t.
    if (dword_5d4594_3835352 ==1) {
        return nox_xxx_tile_543C50(*base_tile,head,tile,variation,
            brush ?brush->edge_id :255,brush ?brush->edge_variation :0,merge);
    }
    if (*base_tile ==tile && *base_variation ==variation) return 0;
    *base_tile =tile;
    *base_variation =variation;
    if (tile ==255) nox_xxx_tileFreeTile_422200(head);
    uint8_t bit =layer ==1 ?1 :2;
    if (brush) cell->field_0 |=bit;
    else cell->field_0 &= (uint8_t)~bit;
    return 1;
}

int sub_51DA70(int x, int y, int brush_token, int layer, int merge) {
    return nox_floorBrushApplyNative_51DA70(x, y,
        nox_mapgenLegacyPtrResolve((uint32_t)brush_token), layer, merge);
}

int nox_xxx_tile_543C50(uint32_t base_tile, nox_tile_list_node_t** head,
                      int tile, int variation, int edge, int edge_variation, int merge) {
    if (tile ==255 || edge ==255) {
        nox_tile_list_node_t** tail =head;
        while (*tail && (*tail)->next) tail =&(*tail)->next;
        if (*tail) {
            nox_xxx_tileFreeTileOne_4221E0(*tail);
            *tail =NULL;
        }
        return 1;
    }
    if (base_tile ==255) return 1;
    nox_tile_list_node_t** link =head;
    if (!merge) {
        while (*link) {
            nox_tile_list_node_t* node =*link;
            if (node->field_0 ==tile && node->field_1 ==variation &&
                node->field_2 ==edge && node->field_3 ==edge_variation) return 0;
            link =&node->next;
        }
        *link =nox_xxx_tileListAddNewSubtile_422160(tile,variation,edge,edge_variation);
        return 1;
    }
    if (!*head) {
        int mapped =nox_xxx_mapGenEdge_543EB0(edge,edge_variation);
        *head =nox_xxx_tileListAddNewSubtile_422160(tile,variation,edge,mapped);
        return 1;
    }
    int matched =0;
    while (*link) {
        nox_tile_list_node_t* node =*link;
        if (node->field_0 ==tile && node->field_1 ==variation && node->field_2 ==edge &&
            sub_543E60(node,edge_variation)) matched =1;
        link =&node->next;
    }
    if (!matched) {
        int mapped =nox_xxx_mapGenEdge_543EB0(edge,edge_variation);
        *link =nox_xxx_tileListAddNewSubtile_422160(tile,variation,edge,mapped);
        return 1;
    }
    int removed;
    do {
        removed =0;
        for (nox_tile_list_node_t* node =*head; node; node =node->next) {
            if (node->field_0 !=tile || node->field_1 !=variation || node->field_2 !=edge) continue;
            link =&node->next;
            while (*link) {
                nox_tile_list_node_t* other =*link;
                if (other->field_0 ==tile && other->field_1 ==variation && other->field_2 ==edge &&
                    sub_543E60(node,sub_411490(other->field_2,other->field_3))) {
                    removed =1;
                    *link =other->next;
                    nox_xxx_tileFreeTileOne_4221E0(other);
                } else link =&other->next;
            }
        }
    } while (removed);
    return 1;
}

int sub_543E60(nox_tile_list_node_t* node, int edge_variation) {
    int index =edge_variation +12 *sub_411490(node->field_2,node->field_3);
    if (*getMemU32Ptr(0x587000,282736 +4 *index) ==255) return 0;
    int mapped =nox_xxx_mapGenEdge_543EB0(node->field_2,*getMemU32Ptr(0x587000,282736 +4 *index));
    if (mapped !=node->field_3) node->field_3 =mapped;
    return 1;
}

void sub_51DD50(int a1, int a2, int a3, int a4) {
	int v4;            // edx
	int v5;            // edi
	unsigned char* v6; // eax
	int v7;            // eax

	if (a1 > 0 && a1 < 127 && a2 > 0 && a2 < 127) {
		if ((v4 = a3, a3 & 2) && (v4 = a3, a4 == ptr_5D4594_2650668[a1][a2].field_6) ||
			v4 & 1 && a4 == ptr_5D4594_2650668[a1][a2].field_1) {
			if ((!(v4 & 1) || a2 != 1) && (!(a3 & 2) || a1 != 1)) {
				v5 = 0;
				if (*(int*)&dword_5d4594_2487248 > 0) {
					v6 = getMemAt(0x973F18, 16204);
					while (1) {
						if (!(*((uint32_t*)v6 - 1) != a1 || *(uint32_t*)v6 != a2 || *((uint32_t*)v6 + 1) != v4)) {
							return;
						}
						++v5;
						v6 += 12;
						if (v5 >= *(int*)&dword_5d4594_2487248) {
							break;
						}
					}
				}
				if (dword_5d4594_2487248 >= 500) {
					*getMemU32Ptr(0x973F18, 22200) = 1;
				} else {
					v7 = 12 * (dword_5d4594_2487248)++;
					*getMemU32Ptr(0x973F18, 16200 + v7) = a1;
					*getMemU32Ptr(0x973F18, 16204 + v7) = a2;
					*getMemU32Ptr(0x973F18, 16208 + v7) = v4;
				}
			}
		}
	}
}

int sub_51DE30(uint32_t* a1, uint32_t* a2, uint32_t* a3) {
	int v3;     // eax
	int result; // eax

	if (*(int*)&dword_5d4594_2487248 <= 0) {
		return 0;
	}
	v3 = dword_5d4594_2487248 - 1;
	dword_5d4594_2487248 = v3;
	*a1 = *getMemU32Ptr(0x973F18, 16200 + 12 * v3);
	*a2 = *getMemU32Ptr(0x973F18, 16204 + 12 * dword_5d4594_2487248);
	result = 1;
	*a3 = *getMemU32Ptr(0x973F18, 16208 + 12 * dword_5d4594_2487248);
	return result;
}

int sub_543680(float* position) {
    if (dword_5d4594_3835356 ==255) return 1;
    float2 fixed;
    int result =nox_xxx_mapGenFixCoords_4D3D90((float2*)position,&fixed);
    if (result) {
        // The original supplied only these four scalar fields. Initialize
        // remaining transient brush state explicitly for portable C.
        nox_floor_brush_t brush ={
            .tile_id =*getMemU32Ptr(0x973F18,35912),
            .variation =dword_5d4594_3835348,
            .edge_id =dword_5d4594_3835356,
            .edge_variation =dword_5d4594_3835360,
        };
        dword_5d4594_3835352 =1;
        int2 point ={(long long)fixed.field_0,(long long)fixed.field_4};
        result =sub_5437E0(&point.field_0,&brush,46);
        dword_5d4594_3835352 =0;
    }
    return result;
}

int sub_5437E0(int* position, nox_floor_brush_t* brush, int step) {
    double px =(double)position[0] +11.5;
    double py =(double)position[1] +11.5;
    int x =(long long)(px *0x1.642c86p-6f);
    float rounded_y =py;
    int y =(long long)(py *0x1.642c86p-6f);
    int remainder_x =(int)(long long)px %46;
    int remainder_y =(long long)rounded_y %46;
    *getMemU32Ptr(0x973F18,22200) =0;
    dword_5d4594_2487248 =0;
    int tile =step;
    if (x >0 && x <127 && y >0 && y <127) {
        int selected_x =x;
        int selected_y =y;
        int layer;
        if (remainder_x <=remainder_y) {
            if (step -remainder_x <=remainder_y) layer =2;
            else { --selected_x; layer =1; }
        } else if (step -remainder_x <=remainder_y) layer =1;
        else { --selected_y; layer =2; }
        obj_5D4594_2650668_t* cell =&ptr_5D4594_2650668[selected_x][selected_y];
        tile =layer ==1 ?cell->field_1 :cell->field_6;
        sub_51DD50(selected_x,selected_y,layer,tile);
    }
    // This worklist is3 scalar32 values per entry, independent of pointer size.
    for (int i =0; i <(int)dword_5d4594_2487248; ++i) {
        uint32_t* entry =getMemU32Ptr(0x973F18,16200 +12 *i);
        x =entry[0];
        y =entry[1];
        uint32_t layer =entry[2];
        if (layer &2) {
            sub_51DD50(x,y,1,tile);
            sub_51DD50(x,y +1,1,tile);
            sub_51DD50(x -1,y,1,tile);
            sub_51DD50(x -1,y +1,1,tile);
        } else if (layer &1) {
            sub_51DD50(x +1,y,2,tile);
            sub_51DD50(x +1,y -1,2,tile);
            sub_51DD50(x,y,2,tile);
            sub_51DD50(x,y -1,2,tile);
        }
    }
    uint32_t queue_x,queue_y,layer;
    while (sub_51DE30(&queue_x,&queue_y,&layer)) {
        x =queue_x;
        y =queue_y;
        if (layer &2) {
            sub_543BC0(x,y,1,tile,brush,1);
            sub_543BC0(x,y +1,1,tile,brush,4);
            sub_543BC0(x -1,y,1,tile,brush,3);
            sub_543BC0(x -1,y +1,1,tile,brush,6);
            sub_543BC0(x,y -1,2,tile,brush,0);
            sub_543BC0(x +1,y,2,tile,brush,2);
            sub_543BC0(x -1,y,2,tile,brush,5);
            sub_543BC0(x,y +1,2,tile,brush,7);
        } else if (layer &1) {
            sub_543BC0(x +1,y,2,tile,brush,4);
            sub_543BC0(x +1,y -1,2,tile,brush,1);
            sub_543BC0(x,y,2,tile,brush,6);
            sub_543BC0(x,y -1,2,tile,brush,3);
            sub_543BC0(x,y -1,1,tile,brush,0);
            sub_543BC0(x +1,y,1,tile,brush,2);
            sub_543BC0(x -1,y,1,tile,brush,5);
            sub_543BC0(x,y +1,1,tile,brush,7);
        }
    }
    return *getMemU32Ptr(0x973F18,22200) ==0;
}

void sub_543BC0(int x, int y, int layer, int tile, nox_floor_brush_t* brush, int edge_variation) {
    if (x >0 && x <127 && y >0 && y <127 &&
        (!(layer &1) || y !=1) && (!(layer &2) || x !=1)) {
        obj_5D4594_2650668_t* cell =&ptr_5D4594_2650668[x][y];
        if ((layer &2) && tile ==cell->field_6) return;
        if ((!(layer &1) || tile !=cell->field_1) && brush) {
            brush->edge_variation =edge_variation;
            nox_floorBrushApplyNative_51DA70(x,y,brush,layer,1);
        }
    }
}

int nox_xxx_tileSubtile_544310(float2* a1) {
	double v1;  // st7
	double v2;  // st6
	int v3;     // esi
	int v4;     // edi
	int v5;     // ebp
	int v6;     // edx
	int result; // eax
	nox_floor_brush_t brush;  // [esp+10h] [ebp-20h]
	float v9;   // [esp+34h] [ebp+4h]

	dword_5d4594_3835352 = 1;
	v1 = a1->field_0 + 11.5;
    brush = (nox_floor_brush_t){
        .tile_id =*getMemU32Ptr(0x973F18,35912),
        .variation =dword_5d4594_3835348,
        .pattern =0,
        .anchor_x =-1,
        .anchor_y =-1,
        .kind =0,
        .edge_id =dword_5d4594_3835356,
        .edge_variation =dword_5d4594_3835360,
    };
	v2 = a1->field_4 + 11.5;
	v3 = (long long)(v1 * 0x1.642c86p-6f);
	v9 = v2;
	v4 = (long long)(v2 * 0x1.642c86p-6f);
	v5 = (int)(long long)v1 % 46;
	v6 = (long long)v9 % 46;
	if (*getMemU32Ptr(0x973F18, 35912) == 255) {
		result = sub_51D9C0(v3, v4, v5, v6, 0);
	} else {
		result = sub_51D9C0(v3, v4, v5, v6, &brush);
	}
	dword_5d4594_3835352 = 0;
	return result;
}
