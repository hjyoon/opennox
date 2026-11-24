#include "client__draw__fx.h"

void noxClientSetOrbPayload(nox_drawable* dr, uint16_t* destination, uint8_t radius, uint8_t fade, uint8_t mode);
void noxClientSetManaBombOrbPayload(nox_drawable* dr, int16_t* path, uint16_t angle, uint8_t reverse, uint8_t fade,
								   uint8_t mode);
void noxClientSetFallingSparkPayload(nox_drawable* dr, int2* origin, uint16_t z, int8_t velocity, uint8_t fade);

// These entries preserve the C declarations for read-only payload inputs.
void nox_client_orb_set_payload_499490(nox_drawable* dr, const uint16_t* destination, uint8_t radius, uint8_t fade,
									  uint8_t mode) {
	noxClientSetOrbPayload(dr, (uint16_t*)destination, radius, fade, mode);
}

void nox_client_mana_bomb_orb_set_payload_499520(nox_drawable* dr, const int16_t* path, uint16_t angle, uint8_t reverse,
											  uint8_t fade, uint8_t mode) {
	noxClientSetManaBombOrbPayload(dr, (int16_t*)path, angle, reverse, fade, mode);
}

void nox_client_falling_spark_set_payload_499950(nox_drawable* dr, const int2* origin, uint16_t z, int8_t velocity,
											  uint8_t fade) {
	noxClientSetFallingSparkPayload(dr, (int2*)origin, z, velocity, fade);
}
