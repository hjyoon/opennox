#include <assert.h>
#include <stdarg.h>
#include <stdint.h>
#include <string.h>

#include "../network_line_message_4d9eb0.c"

enum {
	EVENT_FORMAT = 1,
	EVENT_TALK,
	EVENT_LENGTH,
	EVENT_ENCODE,
	EVENT_INDEX,
	EVENT_SEND,
};

static int events[8];
static int event_count;
static int can_talk;
static nox_object_t* active_object;
static void* cached_update_data;
static void* replacement_update_data;

static void event(int value) { events[event_count++] = value; }

int nox_vswprintf(wchar2_t* str, const wchar2_t* fmt, va_list ap) {
	(void)ap;
	assert(fmt != NULL);
	assert(active_object->data_update == cached_update_data);
	active_object->data_update = replacement_update_data;
	str[0] = 'X';
	str[1] = 0;
	event(EVENT_FORMAT);
	return 1;
}

int nox_xxx_cliCanTalkMB_4100F0(short* str) {
	assert((wchar2_t*)str != NULL);
	event(EVENT_TALK);
	return can_talk;
}

size_t nox_wcslen(const wchar2_t* str) {
	assert(str[0] == 'X');
	assert(str[1] == 0);
	event(EVENT_LENGTH);
	return 1;
}

wchar2_t* nox_wcscpy(wchar2_t* dst, const wchar2_t* src) {
	assert(!can_talk);
	dst[0] = src[0];
	dst[1] = src[1];
	event(EVENT_ENCODE);
	return dst;
}

int nox_sprintf(char* dst, const char* format, ...) {
	assert(can_talk);
	assert(strcmp(format, "%S") == 0);
	dst[0] = 'X';
	dst[1] = 0;
	event(EVENT_ENCODE);
	return 1;
}

uint8_t nox_server_playerIndexFromUpdateData_4D9EB0(void* update_data) {
	assert(update_data == cached_update_data);
	assert(active_object->data_update == replacement_update_data);
	event(EVENT_INDEX);
	return 0xfe;
}

int nox_netlist_addToMsgListCli_40EBC0(int ind, int kind, unsigned char* buf, int size) {
	assert(ind == 0xfe);
	assert(kind == 1);
	assert(buf[0] == 0xa8);
	assert(size == (can_talk ? 13 : 15));
	event(EVENT_SEND);
	return 77;
}

static void reset_events(void) {
	event_count = 0;
	memset(events, 0, sizeof(events));
}

static void test_player_path(int talk) {
	nox_object_t object = {0};
	int first_update_data;
	int second_update_data;
	wchar2_t format[] = {'%', 's', 0};
	const int expected[] = {EVENT_FORMAT, EVENT_TALK, EVENT_LENGTH, EVENT_ENCODE, EVENT_INDEX, EVENT_SEND};

	object.obj_class = 4;
	object.data_update = &first_update_data;
	active_object = &object;
	cached_update_data = &first_update_data;
	replacement_update_data = &second_update_data;
	can_talk = talk;
	reset_events();
	assert(nox_xxx_netSendLineMessage_4D9EB0(&object, format, "ignored") == 77);
	assert(event_count == (int)(sizeof(expected) / sizeof(expected[0])));
	assert(memcmp(events, expected, sizeof(expected)) == 0);
}

int main(void) {
	nox_object_t object = {0};
	wchar2_t format[] = {0};

	assert(nox_xxx_netSendLineMessage_4D9EB0(NULL, format) == 0);
	assert(nox_xxx_netSendLineMessage_4D9EB0(&object, format) == (intptr_t)&object);
	test_player_path(0);
	test_player_path(1);
	return 0;
}
