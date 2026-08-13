#include <assert.h>
#include <stdarg.h>
#include <stdint.h>
#include <string.h>

#include "../server__object__die__die.c"

enum {
	EVENT_LANGUAGE = 1,
	EVENT_DEFINITION,
	EVENT_LENGTH,
	EVENT_LOAD_STRING,
	EVENT_ITEM_NAME,
	EVENT_FALLBACK_NAME,
	EVENT_MESSAGE,
	EVENT_AUDIO,
	EVENT_DELETE,
};

static int events[16];
static int event_count;
static int language;
static int language_calls;
static int definition_enabled;
static int reload_description;
static int mutate_after_length;
static int mutate_on_name;
static uint16_t mutated_material;
static nox_object_t* active_object;
static nox_object_t* mutation_holder;
static nox_object_t* expected_holder;
static float2* expected_pos;
static int expected_sound;
static int expected_audio;
static const char* expected_key;
static int expected_line;
static wchar2_t* expected_name;
static obj_412ae0_t definition;
static wchar2_t initial_description[] = {'x', 'x', 0};
static wchar2_t reloaded_description[] = {'x', 'S', 0};
static wchar2_t format_value[] = {'f', 0};
static wchar2_t item_name[] = {'i', 0};
static wchar2_t fallback_name[] = {'b', 0};

static void event(int value) { events[event_count++] = value; }

int nox_strman_get_lang_code(void) {
	language_calls++;
	event(EVENT_LANGUAGE);
	return language;
}

void* nox_xxx_equipClothFindDefByTT_413270(int type_index) {
	assert(type_index == active_object->typ_ind);
	event(EVENT_DEFINITION);
	return definition_enabled ? &definition : NULL;
}

size_t nox_wcslen(const wchar2_t* str) {
	assert(str == initial_description);
	event(EVENT_LENGTH);
	if (reload_description) {
		definition.field_2 = reloaded_description;
	}
	if (mutate_after_length) {
		active_object->inv_holder = mutation_holder;
		active_object->material = mutated_material;
	}
	return 2;
}

wchar2_t* nox_strman_loadString_40F1D0(char* key, char** str_out, char* source, int line) {
	assert(strcmp(key, expected_key) == 0);
	assert(str_out == NULL);
	assert(strcmp(source, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c") == 0);
	assert(line == expected_line);
	event(EVENT_LOAD_STRING);
	return format_value;
}

wchar2_t* nox_xxx_itemGetName_4E77E0_obj_util(nox_object_t* obj) {
	assert(obj == active_object);
	event(EVENT_ITEM_NAME);
	if (mutate_on_name) {
		obj->inv_holder = mutation_holder;
		obj->material = mutated_material;
	}
	return item_name;
}

wchar2_t* sub_415B60(nox_object_t* obj) {
	assert(obj == active_object);
	event(EVENT_FALLBACK_NAME);
	if (mutate_on_name) {
		obj->inv_holder = mutation_holder;
		obj->material = mutated_material;
	}
	return fallback_name;
}

intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* holder, wchar2_t* format, ...) {
	va_list va;
	va_start(va, format);
	wchar2_t* name = va_arg(va, wchar2_t*);
	va_end(va);
	assert(holder == expected_holder);
	assert(format == format_value);
	assert(name == expected_name);
	event(EVENT_MESSAGE);
	return 1;
}

void nox_xxx_audCreate_501A30(int sound, float2* pos, int a3, int a4) {
	assert(expected_audio);
	assert(sound == expected_sound);
	assert(pos == expected_pos);
	assert(a3 == 0);
	assert(a4 == 0);
	event(EVENT_AUDIO);
}

void nox_xxx_delayedDeleteObject_4E5CC0(nox_object_t* obj) {
	assert(obj == active_object);
	event(EVENT_DELETE);
}

static void reset_state(nox_object_t* obj) {
	memset(events, 0, sizeof(events));
	event_count = 0;
	language_calls = 0;
	definition_enabled = 0;
	reload_description = 0;
	mutate_after_length = 0;
	mutate_on_name = 0;
	mutated_material = 0;
	active_object = obj;
	mutation_holder = NULL;
	expected_holder = NULL;
	expected_pos = NULL;
	expected_sound = 0;
	expected_audio = 0;
	expected_key = NULL;
	expected_line = 0;
	expected_name = NULL;
	memset(&definition, 0, sizeof(definition));
	definition.field_2 = initial_description;
}

static void assert_events(const int* expected, int count) {
	assert(event_count == count);
	assert(memcmp(events, expected, (size_t)count * sizeof(expected[0])) == 0);
}

static void test_armor_materials(void) {
	struct armor_case {
		uint16_t material;
		const char* key;
		int line;
		int sound;
		int use_holder;
	} tests[] = {
		{0x18, "ArmorDieMetal", 1538, 806, 1},
		{0x08, "ArmorDieWood", 1549, 812, 1},
		{0x04, "ArmorDieHide", 1560, 809, 1},
		{0x02, "ArmorDieCloth", 1571, 815, 1},
		{0x00, "ArmorDieGeneric", 1579, 0, 0},
	};
	const int expected[] = {
		EVENT_LANGUAGE, EVENT_LOAD_STRING, EVENT_ITEM_NAME, EVENT_MESSAGE, EVENT_AUDIO, EVENT_DELETE,
	};
	for (size_t i = 0; i < sizeof(tests) / sizeof(tests[0]); i++) {
		nox_object_t object = {0};
		nox_object_t holder = {0};
		reset_state(&object);
		language = 2;
		object.material = tests[i].material;
		object.inv_holder = tests[i].use_holder ? &holder : NULL;
		expected_holder = object.inv_holder;
		expected_pos = expected_holder ? (float2*)&expected_holder->x : (float2*)&object.x;
		expected_key = tests[i].key;
		expected_line = tests[i].line;
		expected_sound = tests[i].material ? tests[i].sound : (int)(uint32_t)(uintptr_t)expected_pos;
		expected_audio = 1;
		expected_name = item_name;
		nox_xxx_dieArmor_54E170_obj_die(&object);
		assert(language_calls == 1);
		assert_events(expected, (int)(sizeof(expected) / sizeof(expected[0])));
	}
}

static void test_armor_description_reload(void) {
	struct suffix_case {
		int lang;
		wchar2_t suffix;
		const char* key;
		int line;
	} tests[] = {
		{0, 'S', "ArmorDieMetalPlural", 1536},
		{1, 's', "ArmorDieMetalPlural", 1536},
		{1, 'x', "ArmorDieMetal", 1538},
	};
	const int expected[] = {
		EVENT_LANGUAGE, EVENT_DEFINITION, EVENT_LENGTH, EVENT_LOAD_STRING,
		EVENT_ITEM_NAME, EVENT_MESSAGE, EVENT_AUDIO, EVENT_DELETE,
	};
	for (size_t i = 0; i < sizeof(tests) / sizeof(tests[0]); i++) {
		nox_object_t object = {0};
		nox_object_t initial_holder = {0};
		nox_object_t reloaded_holder = {0};
		reset_state(&object);
		language = tests[i].lang;
		definition_enabled = 1;
		reload_description = 1;
		mutate_after_length = 1;
		mutated_material = 0x10;
		mutation_holder = &reloaded_holder;
		reloaded_description[1] = tests[i].suffix;
		object.typ_ind = 0xabcd;
		object.material = 0x08;
		object.inv_holder = &initial_holder;
		expected_holder = &reloaded_holder;
		expected_pos = (float2*)&reloaded_holder.x;
		expected_key = tests[i].key;
		expected_line = tests[i].line;
		expected_sound = 806;
		expected_audio = 1;
		expected_name = item_name;
		nox_xxx_dieArmor_54E170_obj_die(&object);
		assert(language_calls == 1);
		assert_events(expected, (int)(sizeof(expected) / sizeof(expected[0])));
	}
}

static void test_weapons(void) {
	struct weapon_case {
		uint16_t material;
		const char* key;
		int line;
		int sound;
		int audio;
		int fallback;
		int use_holder;
	} tests[] = {
		{0x18, "WeaponDieMetal", 1626, 818, 1, 0, 1},
		{0x08, "WeaponDieWood", 1633, 819, 1, 0, 0},
		{0x00, "WeaponDieGeneric", 1640, 0, 0, 1, 1},
	};
	const int item_events[] = {EVENT_ITEM_NAME, EVENT_LOAD_STRING, EVENT_MESSAGE, EVENT_AUDIO, EVENT_DELETE};
	const int fallback_events[] = {EVENT_FALLBACK_NAME, EVENT_LOAD_STRING, EVENT_MESSAGE, EVENT_DELETE};
	for (size_t i = 0; i < sizeof(tests) / sizeof(tests[0]); i++) {
		nox_object_t object = {0};
		nox_object_t holder = {0};
		nox_object_t replacement_holder = {0};
		reset_state(&object);
		object.material = tests[i].material;
		object.inv_holder = tests[i].use_holder ? &holder : NULL;
		expected_holder = object.inv_holder;
		expected_pos = expected_holder ? (float2*)&expected_holder->x : (float2*)&object.x;
		expected_key = tests[i].key;
		expected_line = tests[i].line;
		expected_sound = tests[i].sound;
		expected_audio = tests[i].audio;
		expected_name = tests[i].fallback ? fallback_name : item_name;
		mutate_on_name = 1;
		mutation_holder = &replacement_holder;
		mutated_material = 0;
		nox_xxx_dieWeapon_54E370_obj_die(&object);
		assert(language_calls == 0);
		if (tests[i].fallback) {
			assert_events(fallback_events, (int)(sizeof(fallback_events) / sizeof(fallback_events[0])));
		} else {
			assert_events(item_events, (int)(sizeof(item_events) / sizeof(item_events[0])));
		}
	}
}

int main(void) {
	test_armor_materials();
	test_armor_description_reload();
	test_weapons();
	return 0;
}
