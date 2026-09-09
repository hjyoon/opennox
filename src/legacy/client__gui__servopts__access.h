#ifndef NOX_PORT_CLIENT_GUI_SERVOPTS_ACCESS
#define NOX_PORT_CLIENT_GUI_SERVOPTS_ACCESS

#include <stdint.h>

typedef struct nox_window nox_window;
typedef struct nox_window_data nox_window_data;

typedef struct nox_gui_server_access_state {
	nox_window* root;
	nox_window* wnd_10102;
	nox_window* wnd_10103;
	nox_window* wnd_10104;
	nox_window* wnd_10105;
	nox_window* wnd_10109;
	nox_window* wnd_10111;
	nox_window* wnd_10112;
	nox_window* wnd_10113;
	nox_window* wnd_10123;
	nox_window* wnd_10125;
	nox_window* wnd_10126;
	nox_window* wnd_10127;
	nox_window* wnd_10128;
	nox_window* wnd_10129;
	nox_window* wnd_10130;
	nox_window* wnd_10131;
	nox_window* wnd_10132;
	nox_window* wnd_10133;
	nox_window* wnd_10200;
	nox_window* active_list;
} nox_gui_server_access_state;

extern nox_gui_server_access_state nox_gui_server_access;

nox_window* nox_xxx_guiServerAccessLoad_4541D0(nox_window* parent);
int sub_454A90(nox_window* win, nox_window_data* draw);
int nox_xxx_windowAccessProc_454BA0(nox_window* win, int event, uintptr_t arg1, uintptr_t arg2);
int sub_454640(void);
int* sub_454740(void);

#endif // NOX_PORT_CLIENT_GUI_SERVOPTS_ACCESS
