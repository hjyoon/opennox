// Source-bound native floor brush checks. Compile one controlled callback
// group at a time; Go inserts complete actual production bodies at the marker.
// Original13 vectors retain their inputs and assertions with explicit native
// API adapters. The20 additional vectors follow specified original branches.
// The2 DETERMINISM vectors define previously uninitialized543680 state and
// are intentionally separate from original-equivalence assertions.
#ifndef R304_FIXTURE_COMMON_H
#define R304_FIXTURE_COMMON_H
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME1.h"
#include "GAME3_2.h"
#include "GAME1_1.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "memmap.h"

extern obj_5D4594_2650668_t** ptr_5D4594_2650668;
extern uint32_t dword_5d4594_3835348, dword_5d4594_3835352;
extern nox_tileDef_t nox_tile_defs_arr[176];

_Static_assert(sizeof(void*) == 8, "this fixture targets native 64-bit pointers");
_Static_assert(sizeof(obj_5D4594_2650668_t) == 56, "actual native cell size");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_1) == 4, "left scalar fields");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_5) == 24, "left head after native alignment padding");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_6) == 32, "right scalar fields");
_Static_assert(offsetof(obj_5D4594_2650668_t, field_10) == 48, "right native head");
_Static_assert(sizeof(nox_tile_layer_t) == 24, "standalone native layer");
_Static_assert(offsetof(nox_tile_layer_t, subtiles) == 16, "standalone layer head");
_Static_assert(sizeof(nox_tile_list_node_t) == 24, "native linked node size");
_Static_assert(offsetof(nox_tile_list_node_t, next) == 16, "native linked node next");

static void need(int ok, const char* why) {
    if (!ok) { fprintf(stderr, "R304 CONTRACT_FAIL %s\n", why); exit(86); }
}
static void native_address(const void* p, const char* what) {
    printf("R304 ADDRESS %s=%p\n", what, p);
    need((uintptr_t)p > UINT32_MAX, "controlled native address must exceed 4GiB");
}
#endif

extern uint32_t dword_5d4594_3835356,dword_5d4594_3835360,dword_5d4594_2487248;

#if defined(R304_CHAIN)

enum { CAP = 128, X = 5, Y = 7 };
struct row_box {
    unsigned char before[16];
    obj_5D4594_2650668_t cells[CAP];
    unsigned char after[16];
};
// API adapter only: same32-byte scalar record and positional input values.
#define brush32 nox_floor_brush_t
_Static_assert(sizeof(struct brush32) == 32, "original brush is pointer-free 32 bytes");
_Static_assert(offsetof(struct brush32, kind) == 20, "original brush layer byte");
_Static_assert(offsetof(struct brush32, edge_id) == 24, "original brush edge DWORD");
obj_5D4594_2650668_t** ptr_5D4594_2650668;
uint32_t dword_5d4594_3835348, dword_5d4594_3835352;
nox_tileDef_t nox_tile_defs_arr[176];
static uint32_t brush_id, pattern_mode;
static struct row_box* rows[CAP];
static uintptr_t seen_layer;
static int layer_calls, free_calls;

uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
    need(base == 0x973F18, "unexpected memmap base");
    if (off == 35912) return &brush_id;
    if (off == 35916) return &pattern_mode;
    need(0, "unexpected memmap offset"); return NULL;
}
void nox_xxx_tileFreeTile_422200(nox_tile_list_node_t** head) {
    ++free_calls; need(head != NULL, "native head passed to free"); *head = NULL;
}
// Candidate API observes the owner head separately. The old assertion still
// compares the same logical scalar address, after checking actual native owner.
int nox_xxx_tile_543C50(uint32_t base, nox_tile_list_node_t** head, int tile, int variant, int edge, int mask, int merge) {
    obj_5D4594_2650668_t* cell =NULL;
    int left =0;
    for (int x =0; x <CAP && !cell; ++x) for (int y =0; y <CAP; ++y) {
        if (head ==&rows[x]->cells[y].field_5 || head ==&rows[x]->cells[y].field_10) {
            cell =&rows[x]->cells[y]; left =head ==&cell->field_5; break;
        }
    }
    need(cell !=NULL, "candidate API native owner head");
    need(base ==(uint32_t)(left ?cell->field_1 :cell->field_6), "candidate API base scalar value");
    seen_layer =(uintptr_t)(left ?&cell->field_1 :&cell->field_6); ++layer_calls;
    printf("R304 CANDIDATE_DISPATCH nativehead=%p scalar=%p tile=%d variant=%d edge=%d mask=%d merge=%d\n", head,(void*)seen_layer,tile,variant,edge,mask,merge);
    return 69;
}
static void setup(void) {
    setvbuf(stdout, NULL, _IONBF, 0);
    brush_id = 7; pattern_mode = 0; dword_5d4594_3835348 = 3;
    nox_tile_defs_arr[7].field_52 = 2; nox_tile_defs_arr[7].field_53 = 3;
    ptr_5D4594_2650668 = calloc(CAP, sizeof(*ptr_5D4594_2650668));
    need(ptr_5D4594_2650668 != NULL, "row table allocation");
    for (int x = 0; x < CAP; ++x) {
        rows[x] = malloc(sizeof(*rows[x])); need(rows[x] != NULL, "row allocation");
        memset(rows[x], 0xa7, sizeof(*rows[x])); memset(rows[x]->cells, 0, sizeof(rows[x]->cells));
        ptr_5D4594_2650668[x] = rows[x]->cells;
    }
    native_address(rows[X]->cells, "native-row");
}
static void finish(void) {
    for (int x = 0; x < CAP; ++x) {
        for (int b = 0; b < 16; ++b) need(rows[x]->before[b] == 0xa7 && rows[x]->after[b] == 0xa7, "native row canaries");
        free(rows[x]);
    }
    free(ptr_5D4594_2650668);
}
int main(int argc, char** argv) {
    if (argc != 2) return 64;
    const char* name = argv[1]; setup(); printf("R304 CASE %s\n", name);
    if (!strcmp(name, "descriptor-stack-chain")) {
        // All three production functions execute unchanged. The first
        // descriptor load precedes the grid access in sub_51DA70.
        float2 point = {448.5f, 494.5f}; native_address(&point, "caller-stack-point");
        need(sub_51D8F0(&point) == 1, "native stack descriptor chain should complete");
    } else if (!strcmp(name, "descriptor-direct-stack32")) {
        struct brush32 brush = {7, 3, 0, -1, -1, 0, {0}, 255, 0};
        native_address(&brush, "pointer-free-stack-brush32");
        printf("R304 DESCRIPTOR native=%p baseline-int-diagnostic=0x%08x\n", &brush, (uint32_t)(uintptr_t)&brush);
        // API adapter passes the unchanged32-byte record by native pointer.
        // The preceding case still exercises the actual producer.
        need(nox_floorBrushApplyNative_51DA70(X, Y, &brush, 1, 0) == 1, "native descriptor ABI should complete");
    } else if (!strcmp(name, "scalar-left") || !strcmp(name, "scalar-right")) {
        int layer = !strcmp(name, "scalar-left") ? 1 : 2;
        dword_5d4594_3835352 = 0;
        need(nox_floorBrushApplyNative_51DA70(X, Y, 0, layer, 0) == 1, "native nil-descriptor scalar update should complete");
        obj_5D4594_2650668_t* cell = &rows[X]->cells[Y];
        need((layer == 1 ? cell->field_1 : cell->field_6) == 255, "correct native scalar target");
    } else if (!strcmp(name, "dispatch-left") || !strcmp(name, "dispatch-right")) {
        int layer = !strcmp(name, "dispatch-left") ? 1 : 2;
        dword_5d4594_3835352 = 1;
        int result = nox_floorBrushApplyNative_51DA70(X, Y, 0, layer, 0);
        uintptr_t expected = layer == 1 ? (uintptr_t)&rows[X]->cells[Y].field_1 : (uintptr_t)&rows[X]->cells[Y].field_6;
        uintptr_t row = (uintptr_t)rows[X]->cells;
        printf("R304 LAYER_ADDRESS actual=0x%" PRIxPTR " expected=0x%" PRIxPTR
               " observed-low-offset=%u native-offset=%zu cell-size=%zu left-head-gap=%zu\n",
               seen_layer, expected, (uint32_t)seen_layer - (uint32_t)row,
               (size_t)(expected - row), sizeof(rows[X]->cells[Y]),
               offsetof(obj_5D4594_2650668_t, field_5) - offsetof(obj_5D4594_2650668_t, field_1));
        need(result == 69 && layer_calls == 1, "subtile helper branch reached");
        need(seen_layer == expected, "native row pointer/56-byte stride/layer scalar offset");
    } else if (!strcmp(name, "boundary-control")) {
        need(sub_51D9C0(1, 10, 0, 0, 0) == 0 && layer_calls == 0 && free_calls == 0, "original rejected-boundary control");
    } else if (!strcmp(name, "null-descriptor-control")) {
        brush_id = 255; dword_5d4594_3835352 = 1;
        float2 point = {448.5f, 494.5f};
        need(sub_51D8F0(&point) == 69 && layer_calls == 1, "original null descriptor bypass reaches controlled callback");
    } else return 64;
    finish(); printf("R304 PASS %s\n", name); return 0;
}

#elif defined(R304_LAYER)

struct node_box { unsigned char before[16]; nox_tile_list_node_t node; unsigned char after[16]; };
struct cell_box { unsigned char before[16]; obj_5D4594_2650668_t cell; unsigned char after[16]; };
static struct node_box* nodes[8];
static size_t node_count, free_count;
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a, int b, int c, int d) {
    need(node_count < 8, "node allocation count");
    struct node_box* box = malloc(sizeof(*box)); need(box != NULL, "node allocation");
    memset(box, 0xa7, sizeof(*box)); box->node = (nox_tile_list_node_t){a, b, c, d, NULL};
    nodes[node_count++] = box; native_address(&box->node, "native-node"); return &box->node;
}
nox_tile_list_node_t* nox_xxx_tileFreeTileOne_4221E0(void* node) {
    ++free_count; printf("R304 FREE pointer=%p\n", node); return NULL;
}
int nox_xxx_mapGenEdge_543EB0(int a, int b) { (void)a; return b; }
int sub_543E60(nox_tile_list_node_t* node, int mask) { (void)node; (void)mask; need(0, "merge helper outside a6-zero scope"); return 0; }
int sub_411490(int a, int b) { (void)a; (void)b; need(0, "edge conversion outside a6-zero scope"); return 0; }
static void canaries(void) {
    for (size_t i = 0; i < node_count; ++i) for (int j = 0; j < 16; ++j)
        need(nodes[i]->before[j] == 0xa7 && nodes[i]->after[j] == 0xa7, "native node allocation canaries");
}
int main(int argc, char** argv) {
    if (argc != 2) return 64;
    setvbuf(stdout, NULL, _IONBF, 0); const char* name = argv[1]; printf("R304 CASE %s\n", name);
    if (!strcmp(name, "left-head-alignment") || !strcmp(name, "right-head-width")) {
        struct cell_box* box = malloc(sizeof(*box)); need(box != NULL, "cell allocation");
        memset(box, 0xa7, sizeof(*box)); memset(&box->cell, 0, sizeof(box->cell));
        box->cell.field_1 = box->cell.field_6 = 7; native_address(&box->cell, "native-cell");
        int left = !strcmp(name, "left-head-alignment");
        uint32_t* fields = left ? (uint32_t*)&box->cell.field_1 : (uint32_t*)&box->cell.field_6;
        int result = nox_xxx_tile_543C50(*fields, left ?&box->cell.field_5 :&box->cell.field_10, 20, 21, 22, 23, 0);
        nox_tile_list_node_t* actual = left ? box->cell.field_5 : box->cell.field_10;
        uint32_t padding; memcpy(&padding, (unsigned char*)&box->cell + 20, sizeof(padding));
        printf("R304 HEAD layer=%s result=%d actual=%p expected=%p padding20=0x%08x head-byte-offset=%zu scalar-byte-offset=%zu\n",
               left ? "left" : "right", result, actual, node_count ? &nodes[0]->node : NULL,
               padding, left ? offsetof(obj_5D4594_2650668_t, field_5) : offsetof(obj_5D4594_2650668_t, field_10),
               left ? offsetof(obj_5D4594_2650668_t, field_1) : offsetof(obj_5D4594_2650668_t, field_6));
        need(result == 1 && node_count == 1, "production allocated one native node"); canaries();
        for (int b = 0; b < 16; ++b) need(box->before[b] == 0xa7 && box->after[b] == 0xa7, "cell allocation canaries");
        need(actual == &nodes[0]->node, "native cell layer receives full node pointer");
        free(box);
    } else if (!strcmp(name, "node-next-append-width") || !strcmp(name, "tail-remove-width") || !strcmp(name, "duplicate-control")) {
        nox_tile_layer_t layer = {7, 8, 9, 10, NULL};
        layer.subtiles = nox_xxx_tileListAddNewSubtile_422160(10, 11, 12, 13);
        struct node_box before = *nodes[0]; nox_tile_list_node_t* first = layer.subtiles;
        if (!strcmp(name, "tail-remove-width")) {
            first->next = nox_xxx_tileListAddNewSubtile_422160(20, 21, 22, 23);
            native_address(&layer, "standalone-native-layer");
            need(nox_xxx_tile_543C50(layer.field_0, &layer.subtiles, 255, 0, 255, 0, 0) == 1, "native tail removal should complete");
            need(first->next == NULL && free_count == 1, "one tail removed without losing first node");
        } else if (!strcmp(name, "node-next-append-width")) {
            int result = nox_xxx_tile_543C50(layer.field_0, &layer.subtiles, 20, 21, 22, 23, 0);
            printf("R304 NEXT result=%d actual=%p expected=%p native-node-bytes=%zu next-byte-offset=%zu head-preserved=%s\n",
                   result, first->next, node_count > 1 ? &nodes[1]->node : NULL, sizeof(*first), offsetof(nox_tile_list_node_t, next), layer.subtiles == first ? "true" : "false");
            need(result == 1 && node_count == 2 && layer.subtiles == first, "existing full-width head used to append one node"); canaries();
            need(first->next == &nodes[1]->node, "native next link receives full pointer");
        } else {
            need(nox_xxx_tile_543C50(layer.field_0, &layer.subtiles, 10, 11, 12, 13, 0) == 0, "original exact-duplicate return");
            need(node_count == 1 && layer.subtiles == first && memcmp(&before, nodes[0], sizeof(before)) == 0, "duplicate control preserves native head/node/allocation");
            canaries();
        }
    } else return 64;
    for (size_t n = 0; n < node_count; ++n) free(nodes[n]);
    printf("R304 PASS %s\n", name); return 0;
}

#elif defined(R304_FUNCTIONAL)

enum { CAP=128, X=5, Y=7, MAX_NODES=32, MAX_EVENTS=128 };
obj_5D4594_2650668_t** ptr_5D4594_2650668;
uint32_t dword_5d4594_3835348, dword_5d4594_3835352;
uint32_t dword_5d4594_3835356, dword_5d4594_3835360;
nox_tileDef_t nox_tile_defs_arr[176];
static uint32_t selected_tile, override_variation;
static uint32_t orientations[256];
static nox_floor_brush_t legacy_brush;
static uintptr_t resolved_token;
static int resolve_calls;
void* nox_mapgenLegacyPtrResolve(uintptr_t value) {
    ++resolve_calls;
    resolved_token = value;
    need(value ==0 || value ==UINT32_C(0xf1234567), "legacy token zero extension");
    return value ? &legacy_brush :NULL;
}

struct event { char kind; int a,b,c,d; const void* ptr; };
static struct event events[MAX_EVENTS];
static int event_count;
static nox_tile_list_node_t* allocated[MAX_NODES];
static int allocated_count, freed_count;

static void event(char kind,int a,int b,int c,int d,const void* ptr) {
    need(event_count <MAX_EVENTS,"event capacity"); events[event_count++] =(struct event){kind,a,b,c,d,ptr};
}
uint32_t* mem_getU32Ptr(uintptr_t base,uintptr_t off) {
    if (base ==0x973F18) {
        if (off ==35912) return &selected_tile;
        if (off ==35916) return &override_variation;
    }
    if (base ==0x587000 && off >=282736 && off <282736 +sizeof(orientations))
        return &orientations[(off -282736)/4];
    need(0,"functional unexpected memmap access"); return NULL;
}
nox_tile_list_node_t* nox_xxx_tileListAddNewSubtile_422160(int a,int b,int c,int d) {
    need(allocated_count <MAX_NODES,"allocation capacity");
    nox_tile_list_node_t* node =calloc(1,sizeof(*node)); need(node !=NULL,"native node allocation");
    *node =(nox_tile_list_node_t){a,b,c,d,NULL}; allocated[allocated_count++] =node;
    event('A',a,b,c,d,node); return node;
}
nox_tile_list_node_t* nox_xxx_tileFreeTileOne_4221E0(void* p) { ++freed_count;event('F',0,0,0,0,p);return p; }
void nox_xxx_tileFreeTile_422200(nox_tile_list_node_t** head) {
    for (nox_tile_list_node_t* n =*head; n;) { nox_tile_list_node_t* next =n->next;nox_xxx_tileFreeTileOne_4221E0(n);n =next; }
    *head =NULL;
}
int sub_411490(int edge,int variation) { event('I',edge,variation,0,0,NULL);return variation -100; }
int nox_xxx_mapGenEdge_543EB0(int edge,int variation) { event('E',edge,variation,0,0,NULL);return variation +100; }
static void setup(void) {
    selected_tile =7;override_variation =0;dword_5d4594_3835348 =11;
    nox_tile_defs_arr[7].field_52 =2;nox_tile_defs_arr[7].field_53 =3;
    for (int i =0;i <256;++i) orientations[i] =255;
    ptr_5D4594_2650668 =calloc(CAP,sizeof(*ptr_5D4594_2650668));need(ptr_5D4594_2650668 !=NULL,"row table");
    for (int x =0;x <CAP;++x) { ptr_5D4594_2650668[x] =calloc(CAP,sizeof(**ptr_5D4594_2650668));need(ptr_5D4594_2650668[x] !=NULL,"row"); }
}
static nox_floor_brush_t brush(void) { return (nox_floor_brush_t){7,3,0,-1,-1,0,{0},9,2}; }
static void reset_events(void) { event_count =0;freed_count =0; }
static void expect_event(int i,char kind,int a,int b) { need(i <event_count && events[i].kind ==kind && events[i].a ==a && events[i].b ==b,"original callback event order"); }
static void changed_cell(int layer) {
    obj_5D4594_2650668_t* cell =&ptr_5D4594_2650668[X][Y];
    memset(cell,0xa7,sizeof(*cell));cell->field_0 =0xa4;cell->field_0_1 =0xd3;cell->field_0_2 =0xb612;cell->field_5 =NULL;cell->field_10 =NULL;
    obj_5D4594_2650668_t expected =*cell;
    nox_floor_brush_t b =brush();
    int variation =layer ==1 ?4 :0;
    if (layer ==1) {expected.field_1 =7;expected.field_2 =variation;}else {expected.field_6 =7;expected.field_7 =variation;}
    expected.field_0 |=(layer ==1 ?1 :2);
    obj_5D4594_2650668_t before_neighbor =ptr_5D4594_2650668[X][Y+1];
    need(nox_floorBrushApplyNative_51DA70(X,Y,&b,layer,0) ==1,"changed native cell return");
    need(memcmp(cell,&expected,sizeof(expected)) ==0,"only selected scalar pair and flag bit change; all padding and neighbor fields preserved");
    need(memcmp(&ptr_5D4594_2650668[X][Y+1],&before_neighbor,sizeof(before_neighbor)) ==0,"adjacent cell preserved");
    need(event_count ==0,"base scalar write has no list callback");
    cell->field_0 =0xa4;
    need(nox_floorBrushApplyNative_51DA70(X,Y,&b,layer,0) ==0 && cell->field_0 ==0xa4,"unchanged scalar does not change flag");
}
int main(int argc,char** argv) {
    if (argc !=2) return 64;setup();const char* name =argv[1];
    obj_5D4594_2650668_t* cell =&ptr_5D4594_2650668[X][Y];nox_floor_brush_t b =brush();
    if (!strcmp(name,"legacy-null-token")) {
        cell->field_0 =0xad;cell->field_1 =7;cell->field_2 =4;
        need(sub_51DA70(X,Y,0,1,0) ==1,"legacy NULL token reaches native empty-brush path");
        need(resolve_calls ==1 && resolved_token ==0,"NULL token resolved exactly once");
        need(cell->field_1 ==255 && cell->field_2 ==0 && cell->field_0 ==0xac,"legacy NULL token preserves native sentinel and BYTE flag semantics");
    } else if (!strcmp(name,"legacy-high-bit-token")) {
        legacy_brush =brush();native_address(&legacy_brush,"resolved-native-brush");
        need(sub_51DA70(X,Y,(int32_t)UINT32_C(0xf1234567),1,0) ==1,"legacy highbit token reaches native brush");
        need(resolve_calls ==1 && resolved_token ==UINT32_C(0xf1234567),"highbit token is zero-extended before native resolver");
        need(cell->field_1 ==7 && cell->field_2 ==4 && cell->field_0 ==1,"resolved native brush keeps selected layer output");
    } else if (!strcmp(name,"flag-left") || !strcmp(name,"flag-right")) changed_cell(!strcmp(name,"flag-left") ?1 :2);
    else if (!strcmp(name,"pattern-negative")) {
        b.pattern =1;b.kind =1;b.anchor_x =10;b.anchor_y =20;
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,1,0) ==1 && cell->field_1 ==7 && cell->field_2 ==2,"negative modulo normalization");
        need(b.anchor_x ==10 && b.anchor_y ==20 && b.kind ==1,"established pattern anchor preserved");
    } else if (!strcmp(name,"anchor-and-kind")) {
        b.pattern =1;
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,2,0) ==1 && cell->field_6 ==7 && cell->field_7 ==0,"first anchored brush variant zero");
        need(b.anchor_x ==X && b.anchor_y ==Y && b.kind ==2,"first anchor and kind set");
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,2,0) ==0,"same anchor leaves same layer unchanged");
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,1,0) ==1 && cell->field_2 ==4,"opposite kind uses negative parity");
    } else if (!strcmp(name,"variation-override")) {
        override_variation =1;
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,1,0) ==1 && cell->field_2 ==11,"selected variation override");
        b.pattern =1;b.kind =0;
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,2,0) ==1 && cell->field_7 ==0,"unanchored pattern path precedes override");
    } else if (!strcmp(name,"nil-owned-head")) {
        cell->field_0 =0xad;cell->field_1 =7;cell->field_2 =4;
        nox_tile_list_node_t* first =nox_xxx_tileListAddNewSubtile_422160(1,2,3,4);
        nox_tile_list_node_t* second =nox_xxx_tileListAddNewSubtile_422160(5,6,7,8);
        nox_tile_list_node_t* right =nox_xxx_tileListAddNewSubtile_422160(9,10,11,12);
        first->next =second;cell->field_5 =first;cell->field_10 =right;reset_events();
        need(nox_floorBrushApplyNative_51DA70(X,Y,NULL,1,0) ==1,"nil brush updates base");
        need(cell->field_1 ==255 && cell->field_2 ==0 && cell->field_0 ==0xac,"nil clears one flag bit and sentinel pair");
        need(cell->field_5 ==NULL && cell->field_10 ==right && freed_count ==2,"nil releases only owned native head");
        need(event_count ==2 && events[0].ptr ==first && events[1].ptr ==second,"list release order preserved");
    } else if (!strcmp(name,"invalid-layer")) {
        b.pattern =1;
        need(nox_floorBrushApplyNative_51DA70(X,Y,&b,3,0) ==0 && b.anchor_x ==X && b.anchor_y ==Y && b.kind ==3,"original brush update precedes invalid layer rejection");
        need(cell->field_0 ==0 && cell->field_1 ==0 && cell->field_6 ==0,"invalid layer leaves cell unchanged");
    } else if (!strcmp(name,"list-sentinels")) {
        nox_tile_list_node_t* head =NULL;
        need(nox_xxx_tile_543C50(255,&head,7,1,9,2,0) ==1 && !head && allocated_count ==0,"empty base sentinel skips list insert");
        need(nox_xxx_tile_543C50(7,&head,255,0,9,2,0) ==1 && !head,"empty sentinel remove succeeds");
        head =nox_xxx_tileListAddNewSubtile_422160(7,1,9,2);reset_events();
        need(nox_xxx_tile_543C50(7,&head,7,1,255,2,0) ==1 && !head && freed_count ==1,"edge sentinel removes sole native head");
    } else if (!strcmp(name,"merge-append")) {
        nox_tile_list_node_t* head =NULL;
        need(nox_xxx_tile_543C50(7,&head,20,21,9,3,1) ==1 && head && head->field_3 ==103,"empty merge maps edge before append");
        need(event_count ==2,"empty merge callbacks");expect_event(0,'E',9,3);expect_event(1,'A',20,21);
    } else if (!strcmp(name,"merge-map-existing")) {
        nox_tile_list_node_t* head =nox_xxx_tileListAddNewSubtile_422160(20,21,9,102);reset_events();orientations[3+12*2] =6;
        need(nox_xxx_tile_543C50(7,&head,20,21,9,3,1) ==1 && head->field_3 ==106,"merge maps existing native node");
        need(allocated_count ==1 && event_count ==2,"merge avoids append when mapping available");expect_event(0,'I',9,102);expect_event(1,'E',9,6);
    } else if (!strcmp(name,"merge-unavailable")) {
        nox_tile_list_node_t* head =nox_xxx_tileListAddNewSubtile_422160(20,21,9,102);reset_events();
        need(nox_xxx_tile_543C50(7,&head,20,21,9,3,1) ==1 && head->next && head->next->field_3 ==103,"unavailable merge appends mapped edge");
        need(allocated_count ==2 && event_count ==3,"unavailable merge callbacks");expect_event(0,'I',9,102);expect_event(1,'E',9,3);expect_event(2,'A',20,21);
    } else if (!strcmp(name,"merge-coalesce-order")) {
        nox_tile_list_node_t* head =nox_xxx_tileListAddNewSubtile_422160(20,21,9,102);
        nox_tile_list_node_t* other =nox_xxx_tileListAddNewSubtile_422160(20,21,9,104);head->next =other;reset_events();
        orientations[3+12*2] =6;orientations[3+12*4] =6;orientations[6+12*6] =6;
        need(nox_xxx_tile_543C50(7,&head,20,21,9,3,1) ==1 && head->field_3 ==106 && head->next ==NULL,"merge coalesces matching later node");
        need(allocated_count ==2 && freed_count ==1 && event_count ==8,"merge coalesce callback count");
        expect_event(0,'I',9,102);expect_event(1,'E',9,6);expect_event(2,'I',9,104);expect_event(3,'E',9,6);
        expect_event(4,'I',9,106);expect_event(5,'I',9,106);expect_event(6,'E',9,6);need(events[7].kind =='F' && events[7].ptr ==other,"coalesce unlinks before one free callback");
    } else if (!strcmp(name,"neighbor-writer")) {
        cell->field_1 =cell->field_6 =7;dword_5d4594_3835352 =1;
        sub_543BC0(0,Y,1,0,&b,6);sub_543BC0(X,1,1,0,&b,6);sub_543BC0(1,Y,2,0,&b,6);
        sub_543BC0(X,Y,1,7,&b,6);sub_543BC0(X,Y,2,7,&b,6);
        need(event_count ==0 && b.edge_variation ==2,"neighbor guards precede descriptor mutation and callbacks");
        sub_543BC0(X,Y,1,0,&b,6);
        need(b.edge_variation ==6 && cell->field_5 && cell->field_5->field_0 ==7 && cell->field_5->field_1 ==4 && cell->field_5->field_2 ==9 && cell->field_5->field_3 ==106,"left neighbor brush uses native owner and mapped edge");
        need(cell->field_10 ==NULL,"left neighbor write preserves right head");
        expect_event(0,'E',9,6);expect_event(1,'A',7,4);
        sub_543BC0(X,Y,2,0,&b,3);
        need(b.edge_variation ==3 && cell->field_10 && cell->field_10->field_0 ==7 && cell->field_10->field_1 ==0 && cell->field_10->field_3 ==103,"right neighbor brush uses native owner and mapped edge");
        expect_event(2,'E',9,3);expect_event(3,'A',7,0);
    } else if (!strcmp(name,"subtile-producer")) {
        dword_5d4594_3835356 =9;dword_5d4594_3835360 =2;
        for (int x =0;x <CAP;++x) for (int y =0;y <CAP;++y) ptr_5D4594_2650668[x][y].field_1 =ptr_5D4594_2650668[x][y].field_6 =7;
        float2 point ={448.5f,494.5f};
        need(nox_xxx_tileSubtile_544310(&point) ==1 && dword_5d4594_3835352 ==0,"subtile producer resets mode after actual list path");
        need(allocated_count ==1 && allocated[0]->field_0 ==7 && allocated[0]->field_2 ==9 && allocated[0]->field_3 ==2,"subtile contiguous descriptor fields reach node");
    } else return 64;
    for (int i =0;i <allocated_count;++i) free(allocated[i]);for (int x =0;x <CAP;++x) free(ptr_5D4594_2650668[x]);free(ptr_5D4594_2650668);
    printf("R304 FUNCTIONAL_PASS %s\n",name);return 0;
}

#elif defined(R304_FLOOD)
enum { CAP=128,X=5,Y=7 };
obj_5D4594_2650668_t** ptr_5D4594_2650668;
uint32_t dword_5d4594_2487248;
uint32_t dword_5d4594_3835348,dword_5d4594_3835352;
nox_tileDef_t nox_tile_defs_arr[176];
static uint32_t error_flag,queue[1500];
struct callback {int x,y,layer,tile; nox_floor_brush_t* brush;int edge;};
static struct callback callbacks[16];static int calls;
uint32_t* mem_getU32Ptr(uintptr_t base,uintptr_t off) {
    need(base ==0x973F18,"flood memmap base");
    if (off ==22200) return &error_flag;
    if (off >=16200 && off <16200 +sizeof(queue) && off %4 ==0) return &queue[3*(0)+(0)] +(off -16200)/4;
    need(0,"flood memmap range");return NULL;
}
void* mem_getPtr(uintptr_t base,uintptr_t off) { return mem_getU32Ptr(base,off); }
void sub_543BC0(int x,int y,int layer,int tile,nox_floor_brush_t* brush,int edge) {
    need(calls <16,"callback capacity");callbacks[calls++] =(struct callback){x,y,layer,tile,brush,edge};
}
static void setup(void) {
    ptr_5D4594_2650668 =calloc(CAP,sizeof(*ptr_5D4594_2650668));need(ptr_5D4594_2650668 !=NULL,"table");
    for (int x =0;x <CAP;++x) { ptr_5D4594_2650668[x] =calloc(CAP,sizeof(**ptr_5D4594_2650668));need(ptr_5D4594_2650668[x] !=NULL,"row");
        for (int y =0;y <CAP;++y) ptr_5D4594_2650668[x][y].field_1 =ptr_5D4594_2650668[x][y].field_6 =255; }
}
int main(int argc,char** argv) {
    if (argc !=2) return 64;setup();const char* name =argv[1];
    if (!strcmp(name,"queue-capacity")) {
        ptr_5D4594_2650668[X][Y].field_6 =7;dword_5d4594_2487248 =500;
        sub_51DD50(X,Y,2,7);
        need(error_flag ==1 && dword_5d4594_2487248 ==500,"original500-entry queue limit and error flag");
        for (int i =0;i <500;++i) need(queue[3*(i)+(0)] ==0 && queue[3*(i)+(1)] ==0 && queue[3*(i)+(2)] ==0,"full queue bytes preserved");
    } else if (!strcmp(name,"queue-dedup-and-bounds")) {
        ptr_5D4594_2650668[X][Y].field_6 =7;sub_51DD50(X,Y,2,7);sub_51DD50(X,Y,2,7);
        need(dword_5d4594_2487248 ==1 && error_flag ==0,"matching native tile queued once");
        need(queue[3*(0)+(0)] ==X && queue[3*(0)+(1)] ==Y && queue[3*(0)+(2)] ==2,"queue remains3x32-bit scalars");
        sub_51DD50(0,Y,2,7);sub_51DD50(127,Y,2,7);sub_51DD50(X,0,2,7);sub_51DD50(X,127,2,7);
        ptr_5D4594_2650668[1][Y].field_6 =7;ptr_5D4594_2650668[X][1].field_1 =7;
        sub_51DD50(1,Y,2,7);sub_51DD50(X,1,1,7);sub_51DD50(X,Y,2,8);
        need(dword_5d4594_2487248 ==1,"original layer edge bounds and tile mismatch reject");
    } else if (!strcmp(name,"queue-pop-width-order")) {
        struct slot {uint32_t value,canary;} x={0,0xa5a5a5a5},y={0,0xb6b6b6b6},kind={0,0xc7c7c7c7};
        queue[3*(0)+(0)]=1;queue[3*(0)+(1)]=2;queue[3*(0)+(2)]=3;queue[3*(1)+(0)]=0x12345678;queue[3*(1)+(1)]=0xffffffff;queue[3*(1)+(2)]=2;dword_5d4594_2487248=2;
        need(sub_51DE30(&x.value,&y.value,&kind.value) ==1 && dword_5d4594_2487248 ==1,"LIFO pop return and count");
        need(x.value ==0x12345678 && y.value ==0xffffffff && kind.value ==2,"full32-bit queue output values");
        need(x.canary ==0xa5a5a5a5 && y.canary ==0xb6b6b6b6 && kind.canary ==0xc7c7c7c7,"output store width remains4bytes");
        need(sub_51DE30(&x.value,&y.value,&kind.value) ==1 && x.value ==1 && y.value ==2 && kind.value ==3,"second LIFO record");
        need(sub_51DE30(&x.value,&y.value,&kind.value) ==0 && x.value ==1 && y.value ==2 && kind.value ==3,"empty queue does not store");
    } else if (!strcmp(name,"flood-layer1-order") || !strcmp(name,"flood-layer2-order")) {
        int layer =!strcmp(name,"flood-layer1-order") ?1 :2;
        if (layer ==1) ptr_5D4594_2650668[X][Y].field_1=7;else ptr_5D4594_2650668[X][Y].field_6=7;
        int position[2] ={layer ==1 ?249 :239,layer ==1 ?331 :341};nox_floor_brush_t brush={7,3,0,-1,-1,0,{0},9,2};
        need(sub_5437E0(position,&brush,46) ==1 && calls ==8 && dword_5d4594_2487248 ==0,"single triangle flood drained with8callbacks");
        const int one[8][4]={{6,7,2,4},{6,6,2,1},{5,7,2,6},{5,6,2,3},{5,6,1,0},{6,7,1,2},{4,7,1,5},{5,8,1,7}};
        const int two[8][4]={{5,7,1,1},{5,8,1,4},{4,7,1,3},{4,8,1,6},{5,6,2,0},{6,7,2,2},{4,7,2,5},{5,8,2,7}};
        const int (*expected)[4] =layer ==1 ?one :two;
        for (int i=0;i<8;++i) need(callbacks[i].x ==expected[i][0] && callbacks[i].y ==expected[i][1] && callbacks[i].layer ==expected[i][2] && callbacks[i].edge ==expected[i][3] && callbacks[i].tile ==7 && callbacks[i].brush ==&brush,"original flood neighbor/callback sequence and native brush identity");
    } else if (!strcmp(name,"flood-outside-grid")) {
        int position[2]={0,0};nox_floor_brush_t brush={0};error_flag=1;
        need(sub_5437E0(position,&brush,46) ==1 && calls ==0 && error_flag ==0 && dword_5d4594_2487248 ==0,"outside seed resets state without callbacks");
    } else return 64;
    for (int x=0;x<CAP;++x) free(ptr_5D4594_2650668[x]);free(ptr_5D4594_2650668);printf("R304 FLOOD_PASS %s\n",name);return 0;
}

#elif defined(R304_DETERMINISM)
uint32_t dword_5d4594_3835348=3,dword_5d4594_3835352=0,dword_5d4594_3835356=9,dword_5d4594_3835360=2;
static uint32_t tile=7;static int fix_calls,flood_calls,fix_success=1;static nox_floor_brush_t captured;
uint32_t* mem_getU32Ptr(uintptr_t base,uintptr_t off) {need(base==0x973F18 && off==35912,"determinism selected tile access");return &tile;}
int nox_xxx_mapGenFixCoords_4D3D90(float2* point,float2* fixed) {++fix_calls;need(dword_5d4594_3835352==0,"fix callback precedes subtile mode");*fixed=*point;return fix_success;}
int sub_5437E0(int* point,nox_floor_brush_t* brush,int step) {++flood_calls;need(dword_5d4594_3835352==1 && step==46 && point[0]==100 && point[1]==200,"flood callback state and scalar coordinates");captured=*brush;return 69;}
int main(int argc,char** argv) {
    if(argc!=2)return 64;float2 point={100.75f,200.25f};
    if(!strcmp(argv[1],"flood-brush-defined-state")) {
        need(sub_543680(&point.field_0)==69 && fix_calls==1 && flood_calls==1 && dword_5d4594_3835352==0,"deterministic producer callback order and return");
        nox_floor_brush_t expected={7,3,0,0,0,0,{0},9,2};need(memcmp(&captured,&expected,sizeof(expected))==0,"explicitly defined32-byte brush including formerly unspecified state");
        memset(&captured,0xa7,sizeof(captured));need(sub_543680(&point.field_0)==69 && memcmp(&captured,&expected,sizeof(expected))==0,"repeated calls produce same brush state");
    } else if(!strcmp(argv[1],"flood-brush-skip-and-fix-failure")) {
        dword_5d4594_3835356=255;need(sub_543680(&point.field_0)==1 && fix_calls==0 && flood_calls==0,"NONE edge skips coordinate callback");
        dword_5d4594_3835356=9;fix_success=0;need(sub_543680(&point.field_0)==0 && fix_calls==1 && flood_calls==0 && dword_5d4594_3835352==0,"failed coordinate fix returns before mode and flood callback");
    }else return 64;
    printf("R304 DETERMINISM_PROPOSAL_PASS %s\n",argv[1]);return 0;
}

#else
#error "select an R304 floor brush fixture group"
#endif

// PRODUCTION_FLOOR_BRUSH_BODIES
