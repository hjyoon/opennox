# Native UDP map-transfer validation

[client-native-map-transfer.yaml](../scripts/e2e/client-native-map-transfer.yaml)
drives an actual remote client through a missing So_Beach download, movement,
and a connected transition to a missing Estate map.
[client-native-map-reconnect.yaml](../scripts/e2e/client-native-map-reconnect.yaml)
then connects a fresh client process to the same running server and downloads
Estate again. This is fresh-process reconnection, not restoration of the same
application session. The existing `client-live-peer.yaml` remains unchanged.

Use separate owned copies of the stock data for the dedicated server and
both clients. Preserve copied Save/config files in owned backups before
creating empty runtime state; withhold map directories only from these copies.
Client 1 must lack `maps/so_beach` and `maps/estate`; client 2 must lack
`maps/estate`. Retain the server's original map files. Bind each client's copied
`game_ip.txt` to the owned server's `127.0.0.1:port` and use isolated localhost
discovery. Abort if discovery or selection identifies another endpoint.

The dedicated server uses ordinary time with every `NOX_E2E*` variable removed.
Set its `NOX_MAPS_HTTP=false` before startup to exercise the native UDP path.
Both clients select the stock Warrior class. Clients use
`NOX_E2E_SEAT=headless`, `NOX_E2E_REAL_SLEEP=true`, fail-fast enabled
and golden override disabled. These runs disable audio. Initial server state
is So_Beach with zero players. When client 1 logs `STATE: request estate
transition`, post `estate` to the owned server's `/api/v0/game/map` with its
ephemeral `X-Token`; keep that server process and endpoint for both clients.

Native `MSG_REQUEST_MAP` currently calls `PlayerGoObserver` before transfer.
Closing the stock chat-help/MOTD windows does not itself join play. The
scenarios therefore use normal Action input after each native download.
The initial stock chat-help OK center is `(512,486)` and MOTD OK is `(463,330)`
in the tested viewport. The separate read-only widget diagnostic confirmed
IDs 4103 and 4202 and real mouse dismissal. MOTD's root also has ID 4202, so a
generic root-ID match does not identify the button. The first diagnostic used
the generic dialog's ID 4001 for the separate chat-help window; the actual
chat-help OK ID is 4103. That lookup failure was a diagnostic error.

The scenario files drive input and capture observations; their completion
alone is insufficient. An independent controller must require all of:

- Exact native UDP starts: client 1 receives So_Beach (32,079 compressed bytes)
  and Estate (50,176); client 2 receives Estate (50,176).
- The downloaded `.nxz` and decoded `.map` files match the preserved stock
  SHA-256 values. This does not assert transfer of the server's `.rul` file.
- The same live server process observes player counts `0,1,0,1,0`, retaining
  player identity across the connected map transition.
- Actual `MSG_USE_MAP`, invalidation, Estate loading and reconnection precede
  the second movement phase; the complete handoff and movement take at most
  20 real seconds after the map request.
- Received Action packets and observer-to-normal status reports correspond
  to both client 1 downloads and the fresh client 2 download. Movement-phase
  units have normal collision flags.
- Each of the three named walk phases moves both the authoritative unit and
  its matching client drawable more than eight world units, with replication
  distance at most eight. Exclude disconnected/loading samples from this
  movement assertion and retain those raw samples.
- Both clients disconnect naturally, the server exits through authenticated
  quit, and all exit codes are zero without termination cleanup.

The retained bounds are 30 seconds for server startup, 90 per client, 20 for
the connected map stage, 10 per disconnect and 15 for server quit. REA uses a
270-second outer bound, 60-second idle limit and five-second heartbeats.
The original ready500/wait500/walk120/settle30 budgets are retained. Keep
failed runs and their inputs; verify complete terminal/filesystem capture,
owned-process settlement, binary digests and original personal-file hashes.

Clean Go 1.26.9 ARM64 products from `e6673a55be5ef3357e4391482cdc925dbeb391f8`
passed both normal and highres profiles. Each completed three actual native
transfers, reproduced both file formats exactly, made all three walks and
exited naturally. The recorded handoff durations are 11.248 and 11.459 seconds;
the maximum movement-phase replication distance is 1.647 units. Five earlier
negative captures remain, including missing participation after a successful
transfer. No native game-code correction was needed in this checkpoint.

REA receipts are
`ev_a3f864b620b4c5939bba78ea76fda90ad3e30e401ec8a8086dc9044bf8d91165`
and `ev_8b16ddf0ad9dcd980f08e4f082b6fd479fd00bdeab7d23211fc1a451e8492339`.
This evidence covers owned localhost peers with virtual-tick clients plus
real Sleep. Original Windows-peer interoperability, WAN traversal, combat,
ordinary-time client behavior and SDL compositor parity remain unverified.
