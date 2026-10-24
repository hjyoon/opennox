package legacy

/*
#include <stdint.h>
#include <string.h>
#include "client__draw__fx.h"
#include "GAME3_1.h"

typedef struct {
 uint32_t effect[5];
 uint32_t field127;
 uintptr_t address;
 int canary;
 int callback;
} nox_mana_bomb_probe_result;

static nox_mana_bomb_probe_result nox_mana_bomb_probe_499520(
 const int16_t* path, uint16_t angle, uint8_t reverse, uint8_t fade, uint8_t mode, uint8_t fill) {
 nox_drawable dr;
 memset(&dr, fill, sizeof(dr));
 dr.field_92 = &dr;
 nox_drawable before;
 memcpy(&before, &dr, sizeof(dr));
 nox_client_mana_bomb_orb_set_payload_499520(&dr, path, angle, reverse, fade, mode);
 nox_mana_bomb_probe_result out = {0};
 memcpy(out.effect, &dr.union_u32[0], sizeof(out.effect));
 out.field127 = dr.field_127;
 out.address = (uintptr_t)&dr;
 out.callback = dr.field_116 == (void*)sub_4CA720;
 out.canary = 1;
 const unsigned char* a = (const unsigned char*)&before;
 const unsigned char* b = (const unsigned char*)&dr;
 size_t payload = (const unsigned char*)&dr.union_u32[0] - b;
 size_t callback = (const unsigned char*)&dr.field_116 - b;
 size_t marker = (const unsigned char*)&dr.field_127 - b;
 for (size_t i = 0; i < sizeof(dr); ++i) {
  if ((i >= payload && i < payload + 15) ||
      (i >= callback && i < callback + sizeof(dr.field_116)) ||
      (i >= marker && i < marker + 2)) continue;
  if (a[i] != b[i]) out.canary = 0;
 }
 return out;
}
*/
import "C"

func manaBombPayloadProbe499520(path [4]uint16, angle uint16, reverse, fade, mode, fill byte) (effect [5]uint32, marker uint32, address uintptr, canary, callback bool) {
	var input [4]C.int16_t
	for i, v := range path {
		input[i] = C.int16_t(v)
	}
	out := C.nox_mana_bomb_probe_499520(&input[0], C.uint16_t(angle), C.uint8_t(reverse), C.uint8_t(fade), C.uint8_t(mode), C.uint8_t(fill))
	for i := range effect {
		effect[i] = uint32(out.effect[i])
	}
	return effect, uint32(out.field127), uintptr(out.address), out.canary != 0, out.callback != 0
}
