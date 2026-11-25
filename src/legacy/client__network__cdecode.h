#ifndef NOX_PORT_CLIENT_NETWORK_CDECODE
#define NOX_PORT_CLIENT_NETWORK_CDECODE

#include "defs.h"

int nox_xxx_netOnPacketRecvCli_48EA70(int a1, unsigned char* data, int sz);
int nox_client_decodeProgressPacket(int op, unsigned char* data, int size);
int nox_client_decodeControlPacket(int ind, int op, unsigned char* data, int size);
int nox_client_decodeSettingsPacket(int op, unsigned char* data, int size);
int nox_client_decodeQuestPacket(unsigned char* data, int size);
int nox_client_decodeInventoryReport(int op, unsigned char* data, int size);

#endif // NOX_PORT_CLIENT_NETWORK_CDECODE
