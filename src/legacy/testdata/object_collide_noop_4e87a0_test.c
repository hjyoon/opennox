#include "../object_collide_noop_4e87a0.h"

typedef void (*collide_noop_func)(
	struct nox_object_t* object,
	struct nox_object_t* other,
	float* collision);

static int check_noop(collide_noop_func fn) {
	unsigned char object[32];
	unsigned char other[32];
	float collision[2] = {-3.0f, 9.0f};
	unsigned char object_before[sizeof(object)];
	unsigned char other_before[sizeof(other)];
	float collision_before[2];

	for (unsigned int i = 0; i < sizeof(object); ++i) {
		object[i] = (unsigned char)(0x80u + i);
		other[i] = (unsigned char)(0x40u + i);
		object_before[i] = object[i];
		other_before[i] = other[i];
	}
	collision_before[0] = collision[0];
	collision_before[1] = collision[1];

	fn((struct nox_object_t*)object, (struct nox_object_t*)other, collision);
	for (unsigned int i = 0; i < sizeof(object); ++i) {
		if (object[i] != object_before[i]) {
			return 1;
		}
		if (other[i] != other_before[i]) {
			return 2;
		}
	}
	if (collision[0] != collision_before[0] || collision[1] != collision_before[1]) {
		return 3;
	}
	fn(0, 0, 0);
	return 0;
}

int main(void) {
	if (check_noop(nox_xxx_collideDefault_4E87A0) != 0) {
		return 1;
	}
	if (check_noop(nox_xxx_collideTelekinesis_4EADE0) != 0) {
		return 2;
	}
	if ((collide_noop_func)nox_xxx_collideDefault_4E87A0 ==
		(collide_noop_func)nox_xxx_collideTelekinesis_4EADE0) {
		return 3;
	}
	return 0;
}
