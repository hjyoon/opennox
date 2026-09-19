#include <assert.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "../GAME4_3.h"

typedef char* (*script_callback_qualify_fn)(int32_t, int32_t, int32_t);
typedef char* (*script_callback_name_fn)(const char*, int32_t, int32_t, int32_t);
typedef char* (*script_object_name_fn)(const char*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "qualifiers must remain signed dwords");
_Static_assert(
	_Generic(&sub_542BF0, script_callback_qualify_fn: 1, default: 0),
	"00542BF0 must preserve three signed dword qualifiers");
_Static_assert(
	_Generic(&sub_5435C0, script_callback_name_fn: 1, default: 0),
	"005435C0 must preserve a native string pointer and three signed dwords");
_Static_assert(
	_Generic(&sub_543620, script_object_name_fn: 1, default: 0),
	"00543620 must preserve a native string pointer and one signed dword");

static int32_t observed_qualifiers[3];
static char callback_result[256];
static char object_result[256];

char* sub_542BF0(int32_t a1, int32_t a2, int32_t a3) {
	observed_qualifiers[0] = a1;
	observed_qualifiers[1] = a2;
	observed_qualifiers[2] = a3;
	return NULL;
}

char* sub_5435C0(const char* name, int32_t a2, int32_t a3, int32_t a4) {
	snprintf(callback_result, sizeof(callback_result), "%s%%%d%%%d%%%d", name, a2, a3, a4);
	return callback_result;
}

char* sub_543620(const char* name, int32_t value) {
	snprintf(object_result, sizeof(object_result), "%s%%%d", name, value);
	return object_result;
}

int main(void) {
	script_callback_qualify_fn const qualify = sub_542BF0;
	script_callback_name_fn const callback_name = sub_5435C0;
	script_object_name_fn const object_name = sub_543620;

	assert(qualify(INT32_MIN, -1, INT32_MAX) == NULL);
	assert(observed_qualifiers[0] == INT32_MIN);
	assert(observed_qualifiers[1] == -1);
	assert(observed_qualifiers[2] == INT32_MAX);
	assert(strcmp(callback_name("Callback", INT32_MIN, -1, INT32_MAX),
		"Callback%-2147483648%-1%2147483647") == 0);
	assert(strcmp(object_name("Object", INT32_MIN), "Object%-2147483648") == 0);
	return 0;
}
