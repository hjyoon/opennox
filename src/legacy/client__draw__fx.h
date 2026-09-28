#ifndef NOX_PORT_CLIENT_DRAW_FX
#define NOX_PORT_CLIENT_DRAW_FX

#include "defs.h"

void nox_client_orb_set_payload_499490(nox_drawable* dr, const uint16_t* destination, uint8_t radius,
	uint8_t fade, uint8_t mode);
void nox_client_mana_bomb_orb_set_payload_499520(nox_drawable* dr, const int16_t* path, uint8_t angle,
	uint8_t reverse, uint8_t fade, uint8_t mode);
void nox_client_ballistic_fx_set_state_499610(nox_drawable* dr, uint32_t speed, uint32_t start_frame,
	uint32_t end_frame, uint8_t angle);
void nox_client_falling_spark_set_payload_499950(nox_drawable* dr, const int2* origin, uint16_t z,
	int8_t velocity_z, uint8_t fade);
void sub_499490(int a1, uint16_t* a2, int a3, int a4, char a5, char a6);
void sub_499520(int a1, short* a2, short a3, char a4, char a5);
int nox_xxx_makePointFxCli_499610(int a1, int a2, int a3, int a4, int a5, int a6);
int nox_xxx_drawEnergyBolt_499710(int a1, int a2, short a3, int a4);
nox_drawable* sub_499950(int a1, int2* a2, int2* a3, unsigned short a4, char a5);
int nox_xxx_makeLightningParticles_4999D0(int a1, int2* a2, int2* a3);
int nox_xxx_draw_499E70(int a1, int a2, int a3, int a4, int a5, int a6, int a7);
int sub_49A150(int2* a1, int a2, unsigned char a3);

#endif // NOX_PORT_CLIENT_DRAW_FX
