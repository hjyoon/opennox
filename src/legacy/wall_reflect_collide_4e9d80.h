#ifndef NOX_WALL_REFLECT_COLLIDE_4E9D80_H
#define NOX_WALL_REFLECT_COLLIDE_4E9D80_H

typedef struct nox_object_t nox_object_t;

void nox_xxx_collideSulphurShot2_4E9D80(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
void nox_xxx_collideSulphurShot_4E9E50(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

// Shared by the six registered eight-byte projectile collide records.
int sub_536D80(char* args, void* data);

#endif // NOX_WALL_REFLECT_COLLIDE_4E9D80_H
