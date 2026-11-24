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
