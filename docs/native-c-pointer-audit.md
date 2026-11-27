# Native C entry pointer audit

## C and Go boundary inventory, 2026-11-26

Source revision `817e81c75` selects all 54 top-level legacy C files
across the 38 audited operating-system, architecture, and tag profiles.
The standard compilation contains 135 function bodies in 53 files.
The safe compilation contains 142 in 54 files. Together they contain
144 definition sites for 142 symbols. Ten files contain declarations or
storage. The bodies are native adapters, including C variadic handling,
memory access, and reserved no-op callbacks. The audit identified zero
standalone game algorithms in this set of C files.

Compilers observed 490 of 496 top-level legacy headers and 18 inline
helper definitions. The other six headers belong to retained references
or had no observed product include. The macOS and Windows builds shared
4,592 matching generated cgo declarations. Platform `long` remains part
of four CRT-style file-position, file-size, and integer-parsing APIs.

`11912070d` gives the window constructor, briefing initializer, and
inventory initializer one typed callback ABI. The event code remains a
C `int`. The window pointer, event payloads, and result use native
width. The constructor test checks its initialization call and a
full-width pointer round trip. `817e81c75` corrects four diagnostic
declarations or stack-lifetime probes found while checking generated C.

The eight shared object, player, drawable, window, thing, and viewport
layout checks passed on macOS ARM64 and AMD64, native Linux ARM64, and
emulated Linux 386, ARMv7, and AMD64. The Linux 386 and ARM64 callback
tests passed, including the established PIE and high-address allocator
profile on ARM64. Windows 386, AMD64, and ARM64 built the server and
compiled the size assertions. The ARMv7 layout probe reported a linker
attribute warning for two-byte `wchar_t` and passed its comparisons.

Normal and HD completed six briefing and inventory end-to-end runs with
28 exact screen matches. The Skull guard-page oracle passed with
`cgocheck2` and `checkptr=2`. The same oracle observes an earlier
base-address access with `-race`, which a separate protected-page probe
reproduced. Other selected race tests and the forced-GC ownership replay
passed.

The initial window registration inventory recorded 82 call sites. The
follow-up traced the parser, widget, and forwarded registrations to 84
event procedures and two constructor procedures. They retain the original
addresses in the C-visible window fields. Go invokes each implementation
with its declared argument types. Signed C `int` results retain their
low 32 bits, while the pointer-returning edit procedure and constructor
procedures preserve the full word. Control payloads become window
pointers only for events that read them. The rule-list control sentinel
is handled as a scalar.

The supplied `GAME.EXE` has C `int` setter status and stores the two
procedure addresses at PE32 offsets 372 and 376. The shared structure
keeps those slots at 372 and 376 on 32-bit hosts and 464 and 472 on
64-bit hosts. The constructor signature supplies the result width for an
external initializer and an ordinary C event callback uses the signed
event signature.

The draw and tooltip inventory resolves 60 draw procedures and 16
tooltip procedures to Go implementations. Dispatch preserves the
procedure addresses and uses each implementation's typed arguments.
External C procedures take the window and draw-data pointers. The
tooltip's third argument is a 32-bit packed cursor position. The
original Nox setters write their addresses at PE32 offsets 380 and
384. The matching native offsets are 480 and 488 on 64-bit hosts.

### Memory map and MP3 follow-up

Go now exports the 12 memory-map accessors with native-width addresses
and typed C pointer results. Removing `memmap.c` brings the top-level C
file count to 53. Native callers and invalid-access termination passed
on macOS ARM64 and on Linux ARM64 with the safe profile. The Windows 386
test executable cross-compiled. On macOS the safe profile stalled
during AddressSanitizer startup at both the audit baseline and the
ported revision.

The MP3 Layer III decoder now runs in Go. The migration removed the
codec's C translation unit, header, and cgo calls from the AIL package.
The C comparisons matched 46,706 frame-synchronization inputs, 25,000
side-information inputs, 16,000 reservoir transitions, 768 hybrid
transforms and 768 polyphase synthesis inputs per tested profile, and
2,187 spectral decoding inputs with initialized scratch storage.
The 8 kHz mixed-block case preserves the packed spectral buffer geometry
with a zero-initialized Go workspace.

The supplied Nox installation's 2,383 MP3 files matched the
`f257d38b1` C-backed baseline on macOS ARM64 and Linux ARM64. Each run
checked the per-file PCM digest and sample count for 379,593,216
interleaved samples. The Go arithmetic preserves the two tested
compilers' rounding orders.

The window constructor and four callback registration functions now
export their C signatures directly from Go. Removing their forwarding
file brings the top-level legacy C file count to 52.

### Collision, minimap, and scoring exports

The two collision no-ops, the two minimap position iterators, and three
scoring entry points now export their existing signatures from Go. The
Nox executable returns immediately at the two no-op addresses. For
minimap iteration it returns the address of the object's X coordinate
or null. The native implementation takes the address of the same field
in the host layout. The Elimination and King of the Realm entry points
preserve their native object pointers, and Arena score tracking retains
its 32-bit argument.

Removing the five translation units brings the top-level legacy C file
count to 47. The default macOS ARM64 build selects 46 files with 111
function bodies. The safe profile selects all 47 files with 118
function bodies. Ten of the selected files provide data or declarations.

### Read-only C inputs

The 14 public window, image, sprite, file, equipment, countdown,
modifier, localization, and text-measurement entries now export
directly from Go. Named C pointee types retain the original `const`
qualifiers in the generated declarations. The original PE32 text
measurement reads two-byte units. Its native export uses the same
`wchar2_t` input and separate integer outputs.

The C file probe exercised binary and text opens, read-write opens,
copy, and remove, including missing paths. A 64-bit fixture supplied
native file handles above 4 GiB. The selected macOS E2E runs completed
chapter one, manual save and load, and advanced server options.
Removing `readonly_bindings.c` brings the top-level legacy C file count
to 46.

Eight more read-only input declarations now link to Go for equipment
removal, modifier disengagement, audio format data, and effect payloads.
The public dequip entry was exercised with an equipped item and owner
above 4 GiB. The Nox executable passes both pointers and flags 1 and 1
to the weapon routine, and invokes the armor routine after a zero
weapon result. The Go dispatch retains that order and the signed
32-bit result. The audio format and effect payload tests exercise the
public C entries and native layouts. Removing these four C files
leaves 42 top-level legacy C files.

### Transient ray callbacks

The four ray registry entries now export their C signatures from Go.
The registry keeps native drawable pointers and clears the occupied
slots at frame end. The Nox entry at `0x49BDD0` limits the registry
to 96 drawables. The C test fills all 96 slots with native pointers,
checks invalid inputs and full-width lookup, and verifies clearing.
The selected end-to-end cases completed chain lightning, plasma, and
the charm effect. Removing the forwarding file leaves 41 top-level
legacy C files.

### Fixed-argument localized text

The two bounded wide-formatting entries now pass typed numbers or text
directly to the Go formatter. Their C signatures retain `size_t`, a
signed C `int` for the numeric entry, and read-only two-byte input
units. The original Nox string table uses the numeric and string
formats for server limits, player names, and game notices. The C
comparisons exercise 40 combinations of formats and arguments at each
of 33 buffer capacities. They check every output word and the guards.
The floating-point formatting test reloads the original double bits for
the reference, native helper, and Go-backed entry. On 386, x87 can quiet a
signaling NaN and raise `FE_INVALID` before the formatter runs.

Removing `localized_text_abi.c` leaves 40 top-level legacy C files.
Clang counts 82 function bodies in the 39 files selected by the
standard macOS ARM64 build and 89 in the 40 files selected by the safe
profile. Ten files provide declarations or storage. The standard
function bodies include 18 C formatting routines and their helpers,
nine reserved modifier callbacks, two raw float adapters, and four
damage dispatch adapters. The other 49 forward arguments or results
between C and Go.

### Modifier callbacks

The original effect tables point their nine reserved entries at distinct
single-instruction returns. Go exports the nine symbols, the four
multiplier callbacks, and the enchantment tooltip. The C tests call the
reserved entries through distinct addresses and replay 128 multiplier
cases from the original executable with native objects and scalars. The
multiplier entries return the same address supplied by the C caller.
Go-owned attack and defense values use typed Go dispatch. The tooltip
retains the incoming byte and the native UTF-16 result pointer.

Removing `common__object__modifier.c` leaves 39 top-level legacy C files.

### Team member iteration

The two team entry points now return native membership pointers directly
from Go. The original PE32 head and next fields are at offsets 44 and 0.
The Go link tables preserve native pointer width while the C-visible
record sizes and offsets retain their layout. C tests reproduce the
original query order and check the foreign owner guards across link,
unlink, reset, and close. A pinned Go fixture verifies the same C
pointer returns after garbage collection and verifies that detach
clears the links.

### Read-only safe C services

The safe-build memory comparison, string length, and string comparison
entries now export directly from Go. The generated C declarations keep
the read-only inputs and the native `size_t` width. Four write routines
return their destination pointers on the C side.
A Linux ARM64 C caller verified empty inputs, comparison order, and
byte comparisons that extend past a zero byte.

A Go 1.27.3 callback probe wrote through its input pointer and returned
the same address. The runtime rejected the return when the Go caller
used only the cgo call's implicit pin. The same call succeeded with an
explicit `runtime.Pinner`.

The standard macOS ARM64 profile selects 38 top-level C files with 66
function bodies. The safe profile selects 39 files and 70 bodies.

### Damage flag queries

Damage, timers, and settings now use the Go flag predicates directly.
The two C entry points also export from Go with their original 32-bit
mask and result types. The Nox code at `0x40A5C0` and `0x417DA0`
returns one when the mask overlaps the stored flags. A C probe checked
seven masks including zero, multiple bits, and the signed gameplay
mask's high bit. Observer replies keep their complete word.

### Drawable light color

Both C color setters now export from Go and return the input address.
At `0x484BE0` the original code sets the packed light flag to 2 and
stores the three color words at offsets 16, 20, and 24. The native C
tests check signed inputs, a nil drawable, and an unaligned packed
record with guard bytes. Their Go-backed test buffers are pinned for
the C calls. The ColorLight end-to-end case completed its render and
cleanup checks.

Removing `client__drawable__drawdb.c` leaves 38 top-level legacy C files.

### Map generator pointer tokens

The Nox prefab routine at `0x526830` reads its list head and next link
as 32-bit fields at offsets 80 and 156. Registration and release now
take native pointers directly in Go. The C resolver receives an address
word so its pointer result preserves unregistered values. The C
lifecycle test checks a live owner, null, a released token, and the
unregistered word 1.

The prefab fixture supplies separate native buffers from Go's
high-address allocator. On Linux ARM64 the previous libc allocation
could put one or more buffers below the PE32 address range. The
assertions continue to compare original addresses and require high
addresses on 64-bit targets.

The standard macOS ARM64 build now selects 37 top-level C files with
60 function bodies. The safe build selects 38 files and 64 bodies.

### C data linkage

The `byte_*` blobs provide addressable data symbols to native C
fixtures and back the Go memory map. In a macOS ARM64 linker probe
with Go 1.27.3, an array named with `go:linkname` appeared as a local
data symbol in the Go object. Clang could not resolve an external C
array reference to it.

### Trade slots, quest score rows, and map list storage

Go allocates the two trade slot tables, the six quest score rows, and the
map list sentinel once in native memory. The trade accessor returns
distinct contiguous tables with 140-byte slots on 32-bit hosts and
144-byte slots on 64-bit hosts. The original `GAME.EXE` trade lookup at
`0x4C1910` uses 140-byte slots in order 0, 2, 1, 3.

The original win screen at `0x450770` steps through 16-byte score rows.
The native row is 16 bytes on 32-bit hosts and 24 bytes on 64-bit hosts.
The C comparator receives those native row addresses, and sorting
continues to swap whole rows including trailing padding. The standalone
C score-draw oracle defines its own fixture rows.

The map list iterator at `0x4D09B0` passes the permanent sentinel at
`0x748308` to the list service. Native C tests access the Go-owned
sentinel by pointer, use the public iterators, and verify the linked
records and end marker. Removing the two storage translation units
leaves 36 top-level legacy C files.

### Wall collision definitions

The 11 wall definitions are now stored in the Go type consumed by
circle-wall collision. The original routine reads 24-byte records
from `0x5D37B8` at `0x550068`. All 264 bytes in the supplied
`GAME.EXE` match the sealed records. The native test compares every
decoded field and replays both stock collision points for each wall
type through the production geometry.

Removing `game5_wall_data.c` leaves 35 top-level legacy C files.

### Radial damage and object velocity float entries

The public C entries for radial damage and object velocity now call Go
directly on AMD64 and ARM64. Their C declarations retain the native
pointer, float, and integer arguments. The radial damage probe preserves
both signs of signaling NaN as well as quiet NaN payloads, negative zero,
and a subnormal. The original radial damage code compares the radii
and also moves the encoded words through integer registers.

The 32-bit builds keep the C entries that forward the encoded floats as
integer words. This preserves the argument before a 386 cgo wrapper can
quiet a signaling NaN through x87. The Go callers for lifetime effects
and meteor damage call the shared radial damage service directly.

The Meteor tests compare the retired PE32 cache bytes with the separate
native cache through their actual owners. The fixtures read the retired
bytes from the blob so the safe profile's normal runtime checks still
apply to the gameplay calls.

The 64-bit default build selects 32 C files with 58 function bodies.
There are 38 fixed-signature adapters, 18 formatting or variadic
routines, and two indirect damage and sound callback frames. Eight of
the selected files provide storage. The safe tag adds one source file
with four memory write adapters. The 32-bit builds also select the two
float adapters.

### Shop inventory and viewport

Go allocates the 60 inventory cells and the shop viewport in native
memory. The C header keeps its array expression and cell accessor
through a typed storage view. The C tests check column-major addresses,
140-byte and 144-byte cell strides, and viewport identity across
initialization.

The original lookup at `0x4780A0` gates each 32-code search on the
cell's count at PE32 offset 4. The native lookup uses that same
admission field. The tests cover a zero-count cell with a drawable and
a nonempty cell with a matching code in the final slot. Removing
`client_shop_abi.c` leaves 34 top-level legacy C files.

### Variadic integer and string arguments

The Go formatter advances the caller's native argument list for
promoted integers and string pointers on 386, AMD64, ARMv7, and ARM64.
SysV AMD64 uses the general-register offset and overflow area. AAPCS64
uses the signed register-save offset and stack cursor. The Apple ARM64,
Windows, and 32-bit cursor layouts advance by one native word.

C initializes and copies the variadic list, and the floating-point
formatter advances its part of the same list. The mixed C caller
checks integer signs, full-width pointers, ten doubles, termination,
and guard bytes across the register and stack boundaries. Compiler
assertions check the size and alignment of the variadic state.

The 64-bit default profile selects 31 top-level C files with 54
function bodies. The safe profile selects 32 with 58 bodies.
The 32-bit profiles also select the two float adapters.

### Go calls for gold, player levels, and GUI settings

Go trade code and the C gold exports share the same player balance
workers. The original gold routines at `0x4FA590` and `0x4FA5D0`
update a 32-bit balance and return the protection service's result.
The native entry points still pass through that result word.

The level routine at `0x4EF410` passes kind 3, the ability number, and
the ability number minus one to the book service. Go now passes the
kind as a word to the book and calls the protection setters directly
for level updates and player-file loading. This removes six C helpers
from cgo preambles. The Go menu handlers also call the existing MOTD
cleanup, title loaders, and server settings services directly.

### CRT end pointer comparison

The retained wide-integer fixture captures errno directly after
`strtol`. A Linux ARM64 probe observed the ASan interceptor replace
the end cursor for an invalid base. The fixture checks zero consumed
units for an invalid base and compares the CRT cursor for valid bases.

### Legacy memory blob ownership

Go allocates the eight zero-initialized buffers using the native
allocator. The original base addresses, lookup order, and sizes total
4,014,405 bytes. The buffers live for the process lifetime and the
native lookup returns their full-width addresses.

Three C fixtures now use the raw storage accessor to obtain the byte
buffer and its size by the exact original base. They still compare
the storage with the mapped accessors. The supplied `GAME.EXE`
returns `0x5D5310` directly from `0x409B80` and reads buff state at
`0x5971D0` and `0x597244`. Those addresses resolve to the same
base and offset pairs in the native buffers.

Removing the C array definitions brings the top-level legacy C file
count to 33. macOS and Linux ARM64 passed the native access and retained
C comparisons. Linux ARM64 also passed the safe memory-access checks.
The Linux 386 probe passed the storage, native accessor, and
invalid-address diagnostic checks under QEMU.

### Class, color, and options pointer storage

Go holds the three animation references, the class Next button, and
the three clothing modifiers in native-width pointer slots. The
corresponding PE32 addresses in `GAME.EXE` are `0x7139E8`,
`0x713B50`, and `0x7141C0` for animation, `0x7139E4` for the
button, and `0x713B90` through `0x713B98` for the modifiers.

The animation allocator and release path still use the shared native
animation list. Its C and Go structure checks cover size, alignment,
state ID, and the four callback offsets. The native records measure
68 bytes on 32-bit hosts and 136 bytes on 64-bit hosts. The four
callback offsets are 48, 52, 56, and 60 on 32-bit hosts and 96, 104,
112, and 120 on 64-bit hosts.

Class and color teardown save the continuation before releasing the
animation and clear their slots before invoking it. The options apply
transition at `0x4AB0C0` also clears its slots. The ordinary options
close at `0x4AAA10` retains the original values after release.
Lifecycle checks use the active window and animation lists.

The options C entry and native GUI tests passed on macOS ARM64 with
`cgocheck2` and `checkptr=2`. The menu E2E runs passed for options
open, Escape and Back transitions, warrior chapter one, and NPC
equipment colors. Chapter one matched all six reference screenshots.
The selected GUI ABI tests also passed on Linux ARM64 and emulated
Linux 386. Windows 386 and AMD64 compiled the callback offset checks
in their server builds.

There are 32 top-level legacy C files. The audited 64-bit builds
select 29 files with 54 C function bodies by default and 30 files
with 58 bodies under the safe tag. The audited 32-bit builds select
two additional float adapters.

### Network log buffer

The supplied `GAME.EXE` passes the static buffer at `0x611CFC` to
its formatter from `0x413D30`. Go now formats each call into a
zero-initialized 512-byte buffer. The output is limited to 511 bytes
with a byte reserved for its terminator. The C variadic entry
initializes the native argument list and passes it to Go.

The C entry test exercises a message longer than the buffer, a
message that fills the 511-byte payload, and an ordinary log call
afterward. It compares all three strings delivered to the logger.
The tests passed on macOS ARM64 with strict pointer checks, Linux
ARM64, and emulated Linux 386.

### Windows 386 team membership replay

GitHub Actions run `45852354788` recorded 69 passing terminal events
from the retained Windows 386 test executable under `GOGC=10`.
The checker expected 66 because its list omitted the three repetitions
of `TestTeamMembershipNativePinnedGoNodes`.

The checker now includes the pinned-node test and derives the event
count from the expected identities and repetitions. The archived
native output matches all 69 events after verifying the output and
executable hashes against the run's records.

### Input binding panels and quantity dialog

The shell input panel uses the fields at `0x748140` through
`0x74815C` in `GAME.EXE`. Go holds the shell and in-game window
references and the shell animation pointer at native width.
After releasing the animation and windows, the shell transition
clears the panel and selection references before its continuation.
The in-game panel clears its capture selection when its windows
are released. The options lifecycle E2E opens and closes the shell
panel twice.

The quantity modal holds the root, child windows, drawable, eight
images, and both callbacks in Go. Its amounts and prices retain
32-bit arithmetic. The original accept and cancel calls pass a
position pointer followed by the item ID, thing type, amount, and
extra value. Native tests use real pointer addresses, force GC,
and check both callbacks against that argument order. Full modal
teardown clears the active flag with the released window references.

The C entry and retained draw tests passed on macOS ARM64 with
strict pointer checks, Linux ARM64, and emulated Linux 386. The
hosted quantity E2E matched all three reference frames with the
default mock audio setup. Enabling the optional mock audio handle
trace changed the frame hash at both revisions. The shop quantity
E2E confirmed two stock sale sound submissions and queued buffers
through OpenAL.

Removing `game3_1_abi.c` brings the top-level legacy C file count
to 31. The two input-binding C entries retain their pointer-shaped
return values.

### GAME4 map generator storage

Go allocates the map generator's shared state for the process
lifetime. Two 32-bit fields track the loaded index and the placement
status. Seven native pointers track the archive handle, path buffers,
and object, wall, waypoint, and tile lists. The C layout occupies 36
bytes on 32-bit hosts and 64 on 64-bit hosts.

The original executable uses `0x75ADF4` for the archive handle and
reads the four list heads during cleanup from `0x503F40`. The Go
workers and the native fixtures use the same live storage. The
retained C references still exercise archive seek and close, wire
record parsing, and list allocation, traversal, and release.

The decompiler-only charm-all flag had one console writer and zero
live readers. The Go charm worker applies the original spell
restrictions. The console command displays its configured messages.

The map archive and list tests passed on macOS ARM64 with strict
pointer checks, Linux ARM64, and emulated Linux 386. Windows 386
and AMD64 server cross-builds passed. Removing
`game4_native_data.c` brings the top-level legacy C file count to 30.

The audited 64-bit builds select 27 C files by default and 28 with
the safe tag. The 32-bit builds select 29 and 30, respectively.
The default builds compile 54 C function bodies on 64-bit hosts
and 56 on 32-bit hosts.

### Original tile coordinate rounding

The tile allocator at `0x5040A0` scales integer coordinates by 46.
For kind 1 the original x87 code adds 23 to the exact X product
before storing the float. The other branch adds 23 to Y after
storing and reloading the scaled Y coordinate as a float.
The Go allocator makes both rounding points explicit.

The instruction bytes and constants from the supplied `GAME.EXE`
ran under 386 emulation with allocator and deallocator hooks. The
native test uses the six observed coordinate pairs and checks
allocation, zero-filled layers, links, and counts through forced GC.
The suite passed on macOS ARM64 with strict pointer checking, Linux
ARM64, and emulated Linux 386. Windows 386 and AMD64 server builds
passed. The final client passed the new-warrior quest and item
quantity scenarios, including all three quantity reference frames.

### Quickbar window registry

The original selection routine at `0x45DA50` loads the row address at
offset 204 and indexes the eight-byte slots within the row. Go allocates
the separate eight-entry window registry in native memory for the
lifetime of the process. Each registry entry is 68 bytes on 32-bit
targets and 136 bytes on 64-bit targets. The registry clears each
entry after destroying its root, buttons, and nugget windows.

The C fixture populates the same registry used by the Go callbacks.
The tests check each native offset, address bits above 4 GiB on
64-bit targets, reuse of cleared slots, and the order of destruction.
The quickbar drag-out handler checks all 245 cases captured from the
original `0x45EF50` routine.

### Book, server options, and briefing storage

Go initializes the book's 12 window and image pointer slots in native
memory for the lifetime of the process. The slots occupy 48 bytes on
32-bit targets and 96 bytes on 64-bit targets. The retained C book
reference and Go use the same native slots. The local player netcode
continues to use a 32-bit unsigned value across the book, team, chat,
save, and HUD paths.

The original `0x459870` returns the address of the server rules list
head at `0x6D3B58`. The Go initializer allocates a stable head in native
memory and sets all three links to its address. The list head occupies
12 or 24 bytes, depending on pointer width. The options loader binds
the server name input as a native window pointer for child 10101.

In the original, `0x44D970` reads a 32-bit value at `0x69F408` and
sets the briefing flag when the value is nonzero. The Go audio driver
keeps the native handle and supplies the 32-bit value through the
existing callback. The dialog fallback starts at 1 and changes on the
first successful dialog read.

Removing `game2_native_data.c` leaves 29 top-level legacy C files. The
standard 64-bit build selects 26 files with 54 C function bodies. The
standard 32-bit build selects 28 files with 56 bodies. The safe tag
adds one C file and four bodies to each profile.

The remaining storage definitions are in `vardefs.c` and
`common__system__team.c`. Clang reports 676 definitions in
`vardefs.c`, including 406 `uint32_t` words, 137 window pointers and
two initialized drawing callback pointers. The team file contains nine
32-bit color cells.

| Remaining boundary | Files |
| --- | --- |
| Native variadic argument lists and host formatting | `cgo_fs.c`, `common__log.c`, `game3_2_abi.c`, `game4_3_abi.c`, `network_line_message_4d9eb0.c`, `noxstring_format_abi.c` |
| Legacy pointer, tag and scalar conversions | `client_chat_bubble_abi.c`, `client_death_feed_abi.c`, `client_render_clip_abi.c`, `common__object__armrlook.c`, `common__object__weaplook.c`, `game1_split_abi.c`, `game2_book_abi.c`, `game2_quickbar_abi.c`, `game2_server_abi.c`, `game3_class_config_options_abi.c`, `game3_map_effect_abi.c`, `game4_gold_consumers_abi.c`, `game4_map_archive_abi.c`, `game4_player_abi.c`, `gui_input_config_abi.c`, `mapgen_area_records_abi.c`, `mapgen_legacy_ptr.c` |
| Raw float argument words on 32-bit targets | `map_radial_damage_4e25b0.c`, `objects_velocity_5118a0.c` |
| Indirect damage and sound callback frames | `raw_default_damage_native_bridge_4e0b30.c` |
| AddressSanitizer memory and string wrappers | `cgo_safe.c` |

### Renderer color cells

The original color initializer at `0x4A94A0` publishes the color
pointers in the table beginning at `0x85B480`. OpenNox publishes
native-width addresses to the same table entries. The nine cells
from `common__system__team.c` now occupy 36 bytes allocated by Go
in native memory. They remain live for the process, and each
published pointer continues to refer to the cell updated through
the renderer's existing setters.

The retained C color lookup and the Go server list read each
cell after a write through its published address and a forced
garbage collection. Removing the team color storage file leaves
28 top-level C files. The standard 64-bit and 32-bit profiles
select 25 and 27 files, respectively.

### Floor render callbacks and buffer release

The original `0x481420` stores the floor tile and edge callbacks at
`0x5ACD3C` and `0x5ACD40`. Go now holds both addresses at native
width. Known callbacks dispatch directly to Go. Other callbacks use
typed C calls with a native position and image or edge pointer.
The tile callback also takes the original 16-bit tile identifier.
The edge test verifies both addresses as the loop traverses two
native nodes and checks the nil and no-edge selections.

The original `0x430EC0` clears the published floor buffer start
address before running the remaining renderer cleanup. OpenNox
clears both ends of the borrowed range when releasing the buffer.
This migration removes four storage definitions from `vardefs.c`,
leaving 672.

### Spellbook notice variants in the wizard scenario

The original CSF entry provides three phrasings for the award
notice. A rerun reached the same quickbar state and differed from
the final reference in the Force Field and Shock notices. The
wizard scenario now selects the stock notice variants visible in
the reference frame while granting those spells through the normal
network path. It restores the entries after each award and checks
the exact reference frame using the pixel digest for the English
or Korean CSF profile.

The conjurer book-default scenario selects the three stock
award notices visible in its final reference frame.

The conjurer quickbar scenario selects its award and cast notices
before entering the real input and spell paths. It checks the
self-target and other-target frames under both CSF profiles and
observes natural buff expiry.

The map download test supplies a scenario path and the default
map profile when it invokes the runner's preparation block.
The six subcases exercise the copied maps and generated ZIPs.

### Inventory window and cell storage

The original inventory constructor at `0x465E00` publishes its
root at `0x6D7BC8`. Go holds the root, the panel and its child
windows, and the font and image references. The shared root
slot, nine equipment-list heads, and 32-bit inventory mode occupy
one native allocation owned by Go.

The original lookup at `0x461EF0` advances by 148 bytes per cell.
The 21-by-4 grid keeps that stride on 32-bit targets and uses a
152-byte stride on 64-bit targets. Go allocates the grid for the
process lifetime so secondary-weapon selection and the shop still
refer to the same cells. The C test fixtures write those live cells
and verify their layouts and pointer identity.

This migration removes 17 definitions from `vardefs.c`, including
11 window pointer slots.

### Multiplayer browser windows and connection state

The browser constructor at `0x4379F0` publishes its root at
`0x69B518`. Go stores the root, server information popup,
map and region panels, and the five game-list columns at their
native pointer width. The filter and nearby-server windows use
the same typed pointers. Browser teardown at `0x43A9D0` keeps
the original order for the filter, nearby popup, detached server
popup, and root. It clears the child references after destruction.

Go keeps the connection deadline and server refresh deadline as
64-bit values and the connection error code as a 32-bit value.
The original connection loop at `0x438770` compares both words
of the deadline at `0x69B500`. The C checks cover the event payload,
window draw and listbox callback signatures. This migration removes
20 definitions from `vardefs.c`.

### Rank window storage

The original rank constructor at `0x46E870` publishes the root at
`0x6DE794`. Go holds the root and looks up the position, timer, and
lesson-limit labels in the existing rank window table. The root's
cleanup clears the table and the root slot before the window pool
can recycle their addresses.

The current and pending modes, refresh flag, and highlighted row
retain their 32-bit values and existing C entry points. Layout
calculations use the same local dimensions. This migration removes
ten definitions from `vardefs.c`.

### Character menus and options windows

The class, color, and options constructors publish their roots at
`0x7139EC`, `0x713B48`, and `0x7141AC` in the supplied executable.
Go holds those roots, the color palette and swatch controls, the
sound checkboxes, and the connection-loss windows. The class and
color screens borrow the original profile bytes through full-width
pointers. The color picker keeps the palette group and index packed
as two 16-bit values and clears its window references on close.

The retained color-matching fixture receives the same native root
as the Go entry. It compares the pointer and scalar return values,
the packed color word, and the affected controls. This migration
removes 27 definitions from `vardefs.c`.

### Room ranking distance precision

At `0x525A26` the original executable loads the two room dimensions
as floats. The products, sum, and square root stay on the x87 stack.
The accumulated distance is added at `0x525A42`, followed by the
32-bit store at `0x525A53`. The Go path uses wide intermediates on
each architecture. The C comparison checks both an empty and a
nonzero accumulator for 4,096 dimension pairs. The emulated x87
sequence matched all 8,192 stored words with 53-bit precision and
round-to-nearest.

### Server configuration windows and lists

The original constructor at `0x457500` publishes the server options
root at `0x6D3D70`. The player tab constructor at `0x456270` publishes
its root at `0x6D3A48`. Go keeps the tab controls and their two image
references and clears the borrowed controls when the options root
closes.

The options, player, filter, and objective HUD slots shared with the C
fixtures occupy 44 bytes on 32-bit targets and 88 bytes on 64-bit
targets. Go allocates those slots and the existing server access state
and player and team list heads in native memory. The fixtures use the
same live addresses as the Go handlers. This removes 26 definitions
from `vardefs.c`.

### Shop, trade, and other gameplay windows

The shop, trade, summon cage, NPC dialog, and quest game-over roots are
published at `0x6E08E4`, `0x716D80`, `0x716DDC`, `0x6E6A58`, and
`0x712930` in the supplied executable. Go stores the roots, the shop
slider, and the summon cage and command controls at native pointer
width. The trade state and summon count retain their 32-bit values.
Summon cage coordinates use signed 32-bit values through the animation.

The window pool cleanup clears borrowed controls and the selected summon
record when their owning windows are released. The quest timer reads the
same root slot as the game-over event handler. This removes 19 definitions
from `vardefs.c`.

### Advanced server and in-game options windows

The original advanced server and rule-list constructors publish their
roots at `0x715CF8` and `0x7482E4`. Go holds the roots, tab selection,
rule controls, performance window, and the in-game options and backdrop
references. The tab selection stays an unsigned 32-bit value. Window
cleanup releases the borrowed rule, object, spell, and audio controls
before the GUI allocator can recycle them. This removes 17 definitions
from `vardefs.c`.

### Quickbar controls and book details

The original quickbar constructor at `0x45E190` stores its left
control at `0x6D4938`. The book constructor at `0x45B9D0` stores its
page arrows, page icon, and reward window at `0x6D3F34` through
`0x6D3F40`. Go holds the quickbar's eight control slots and the book's
six window and animation slots in native memory. The layouts occupy
32 and 24 bytes on 32-bit targets, and 64 and 48 bytes on 64-bit
targets.

The quickbar cleanup clears each control before its address can be
reused, including a spell or trap button carrying an active drag.
The book clears its borrowed controls with the book list and clears
the separate reward window when it is released. The C fixtures use
the same live slots as the Go callbacks. This removes 14 definitions
from `vardefs.c`.

The four window and control migrations leave 522 definitions in
`vardefs.c`, including 389 unsigned 32-bit words and 22 window
pointers. The remaining window slots belong to the briefing and
escape menus, game setup, chat, vote and observer controls, status
icons, meters, and text input.

| Profile | Top-level C files selected | C function bodies |
| --- | ---: | ---: |
| AMD64 and ARM64 | 25 | 54 |
| 386 and ARM | 27 | 56 |
| AMD64 and ARM64 with `safe` | 26 | 58 |
| 386 and ARM with `safe` | 28 | 60 |

The 54 common bodies contain 38 fixed-signature conversions, 14
formatting and variadic adapters, and two indirect damage and sound
callback frames. The 386 and ARM profiles add the two float-word
adapters. The safe profiles add four memory and string adapters.

### Client status windows, meters, and chat

The original chat constructor at `0x46A730` publishes the root, entry
field, and its data at `0x6D852C` through `0x6D8534`. Go retains the two
window references and reads the data from the live entry field. The
chat active word remains 32 bits. The retained C fixture writes the
same word read by the chat callbacks.

The meter constructor at `0x4714E0` publishes the meter root at
`0x6DE878`. Go allocates the seven meter records and three potion
drawables in native memory. A meter record is 20 bytes on 32-bit
targets and 24 bytes on 64-bit targets. The draw comparisons use the
same drawable array. The three potion children are owned by the
meter window and have no independent Go callers.

Go also holds the status icons, vote controls, and PE32 IME focus
reference at native pointer width. The original vote constructor at
`0x48D000` publishes the root and two child controls at `0x6F8A94`
through `0x6F8A9C`. The window references are cleared when their
owners leave the GUI pool. This removes 17 definitions, including
13 standalone window pointers, from `vardefs.c`.

### Setup dialogs, shell windows, and briefing previews

The original help and connection constructors at `0x49C560` and
`0x49C820` store the roots at `0x7131E4` and `0x7131E8`. Their Go
slots remain addressable at native width for the C callback tests.
The shell backdrop, pause menu, message-of-the-day window and list,
and server key window clear their references through GUI pool cleanup.
The server key mode remains a 32-bit word and clears with its root.

The briefing constructor at `0x44E560` publishes the outer window and
panel at `0x69F498` and `0x69F494`. Those two slots and the preview
font occupy 12 bytes on 32-bit targets and 24 bytes on 64-bit
targets. Go holds all twelve preview drawable pointers and preserves
their existing deletion order. The C fixture and Go callbacks share
the live briefing slots.

This removes another 24 definitions from `vardefs.c`. The 481
remaining definitions include no standalone window pointer or
meter-record array.

### Renderer and meter state

The original clip save routine at `0x49F7F0` loads the active draw
data through `0x973FA8` and sets the 32-bit save flag at `0x713228`.
Go holds that flag and the native reference to the active buffer.
Video initialization still publishes both buffers in the same order.
The C clip entry retains its pointer-shaped status. Go receives and
returns the status as an integer.

Go holds the meter font reference, visibility word, animation ticks,
poison flag, colors, and item type IDs. The word and drawable fixture
calls reach the live Go-owned state. This removes 13 definitions
from `vardefs.c`, leaving 468.

### Unreferenced C storage

The C, Go, assembly, and script inventory identified 85 unused
definitions. The list includes 80 32-bit words, two 64-bit words,
and the former spawn, monster-list, and pending-owner allocator
pointers. Removing the definitions and their declarations leaves
383 variables in `vardefs.c`, including 296 `uint32_t` words.

### Numeric state accessed through Go

Go owns 182 numeric cells previously defined in `vardefs.c`.
136 are ordinary Go fields. The other 46 use native allocation so
their addresses stay valid for C calls and legacy buffer readers.
The field types preserve the previous widths and the signed connection
state read. All 22 nonzero initial values carry over, including the
32-bit all-ones sentinel read by Go as `-1`.
The ten nonzero cells with segment offsets in their names also match
the bytes at those offsets in the reference `GAME.EXE`.

There are 201 remaining definitions in `vardefs.c`.

### Spellbook, quickbar, and briefing state

Go owns another 34 cells across the spellbook, quickbar, and briefing.
The spellbook keeps the original 32-bit reward coordinates and a
native-width player reference. Quickbar drag sources and briefing
dialog names retain their native pointer types. The C comparison
fixtures access the same addresses through their storage accessors.
The live dialog service also keeps the address of the briefing
enable word.
The frozen spellbook reference and the standalone C oracles retain
their original bodies.

| Shared structure | 32-bit size | 64-bit size |
| --- | ---: | ---: |
| Spellbook state | 92 bytes | 96 bytes |
| Quickbar controls | 60 bytes | 104 bytes |
| Briefing state | 32 bytes | 48 bytes |

There are 167 remaining definitions in `vardefs.c`.

The legacy directory contains 28 production C translation units and
three comparison units outside `testdata`. The production units
contain 60 function bodies across the supported build configurations.
They implement ABI calls, C varargs and runtime services, including
four `safe` wrappers and two 32-bit adapters that preserve float bits.
The three comparison units contain 115 reference and fixture functions.

### Statistics, map generation, and startup storage

Twenty-five further numeric cells are owned by Go. Seven use native
allocation because their addresses are passed to comparison or
runtime code. The same-width initial values cover the two 64-bit
floating-point bit patterns as well as the existing 32-bit defaults.
Each standalone C comparison keeps its independent storage.

The quest statistics columns use eleven native pointers in a Go-owned
native allocation. The startup focus registration keeps two adjacent
six-byte strings with distinct addresses. Together these changes leave
140 definitions in `vardefs.c`.

### Native pointer slots

Go now allocates 41 further slots in native memory. They cover mixer
and timer references, video rows, inventory and shop references,
server list heads, and allocators for gameplay records. The C and Go
callers that retain slot addresses still share the live slots for the
duration of the process. Inventory drag identity and the WOL list heads
use typed native pointers. The respawn list and allocator keep the
existing native pointer stores used by the scheduler.

There are 99 remaining definitions in `vardefs.c`.

### UI, vote, trade, and floor editing state

Go allocates 24 more cells in native memory for UI modes and gold,
vote thresholds and allocation, trade pools, and floor editing.
The C comparisons use typed storage accessors and the existing
entry points. The frozen vote reference reads the configured
thresholds through its fixture bindings. The PE32 trade head stays
a 32-bit word beside the native pool pointers. Shop insertion uses
the existing Go allocator entry directly.

There are 75 remaining definitions in `vardefs.c`.

### Display, audio, and client item storage

Go allocates 35 slots for the display dimensions and tile geometry,
edge images, health changes, music and dialog state, and trade and
inventory. The image table holds 64 native pointers. The inventory
selection is a native-width word and the related item codes are
32 bits. Audio starts with the same enabled and idle values.
The C comparisons use the live slots through typed accessors.

The music module owns its nesting counter and the audio comparison
owns its private bank. Their two obsolete C definitions were removed,
leaving 38 definitions in `vardefs.c`.

The opening dialogue replay now dismisses the captured chapter briefing
as soon as it becomes ready, then observes the first native voice stream
and both requested replays through natural completion.

### Map generation, world, and focus storage

Go allocates the remaining 37 cells for archive records, map
generation, focus registration, world lists, and player state.
Archive and list owners retain native pointer width. Packed map
indices, tokens, counters, and flags retain their 32-bit widths.
The tile coordinate table uses a compile-time capacity of 128.
C fixtures read the live slots through their storage bindings. The
native word allocation includes space for the settings comparisons'
16-byte neighbour reads.

The storage migration removes `vardefs.c` after its last definitions.

### Remaining top-level C entry points

The 27 top-level C files have 60 function definitions across the
supported architecture and tag profiles. Default 64-bit builds select
24 files and 54 definitions. Default 32-bit builds select 26 files
and 56 definitions. The `safe` tag adds four definitions in one file
to either selection.

The union includes 38 pointer and scalar ABI adapters, 14 variadic and
host formatting entries, two native indirect-call frames, two float
bit adapters, and four safe memory and string entries. Their standalone
game algorithm count is zero. The count covers the top-level C files.
The three C comparison sources, Cgo preambles, and inline headers
have separate audit scopes.

### Fixed localized formatting in Cgo preambles

The server options, map list, inventory, quickbar, quest briefing and
game-over screens, overlays, tooltips, shop, trade, and network notices
now format their fixed arguments directly in Go. The 28 retired C helpers
covered 37 calls. The argument list retains native text pointers and
32-bit numeric words. Wide output keeps the original UTF-16 code units,
format capacity, and result length. The map-list narrowing preserves
the low byte of each code unit.

### Floating-point variadic arguments

Go reads the double slots from the SysV and AAPCS64 register
save areas and from the native stack cursors. The 32-bit ARM cursor
aligns each double before consuming its eight bytes. The inventory
damage formatter passes its promoted double directly from Go.
The fixed-signature CRT call receives the raw bits and supplies
the native decimal rounding, nonfinite spelling, and bounded output.

### Console and network notices

Go calls the console, player broadcast, and direct line-message
formatters with their typed arguments. Twelve Cgo helpers were
removed from the console, map generator, and reward paths.
The C variadic entries use the same Go formatters. Console output
keeps its 512-word buffer. Broadcast and direct line messages
retain their 256-word buffers, packet encoding, and recipient order.

The default macOS ARM64 client linked Go callers for 166 Cgo
preamble definitions at `cc5ea78dc`. The three formatting changes
bring that count to 125, with 107 in `legacy` and 18 in
`legacy/common/ccall`. This is a linked-symbol count for one product.
It includes helpers used by in-game regression probes.

### Protected player values

Health, mana, experience, and gold updates share the Go protection
record operations with the C exports. The 14 retired Cgo preambles
forwarded updates and UTF-16 name length requests. The protection
token keeps its signed 32-bit gate, mana deltas keep their signed
16-bit input, and experience uses the original x87 conversion model.
The mana refresh and subtraction results sign-extend to the native
word width. The macOS ARM64 baseline and updated paths pass the same
return-width regression, and the protected player tests pass on macOS
ARM64 with strict Cgo checking and on Linux ARM64 PIE.

### World state and settings services

The team list, settings refresh, access list, sudden-death timer, player
iteration, and wall polygon services share their Go implementations with
the C exports. Twelve Cgo preambles were removed. The access configuration
uses the same checked memory-map entry. Player and polygon pointers keep
native width, wall radius inputs retain their float32 bits, and the live
sudden-death values retain their float64 representation. The C declarations
and compile-time layout checks continue to validate the exported ABI.

### Map generation compass image keys

The progress display resolves its packed 32-bit image keys in the Go
sprite registry. The registry retains its live image owners and looks
up exact native handle values before checking low-word aliases.
The 36 compass slots retain their 32-bit layout. The 32 paired compass
frames and the local and network progress paths passed the headless
raster checks on macOS ARM64 with strict Cgo checking. The integer
key tests passed on Linux 386.

### Object scans and direction calls

Chakram retargeting and collision use the typed Go rectangle iterator.
Retargeting reads the live source position during each visit. Collision
copies the bounds before iteration and retains the registered Pentagram
callback address for filtering. The Go monster-facing call consumes the
integer result directly. The C entry continues to expose its original
pointer-shaped result.

Compared with `64c0032c1`, 30 Go-callable Cgo preamble entry points
disappear from the default macOS ARM64 client link. The comparison
matches preamble definitions to linked symbols in their Go packages.
The remaining adapters include native indirect calls, mixed pointer
and scalar results, unaligned native loads and stores, CRT integration,
and regression probes.

### Cloud and trap services

The Polyp and Troll death paths use the Go object type registry and
balance table for their poison clouds. The shooting trap reads its
damage from the same Go balance table. The five retired Cgo helpers
forwarded type requests and double-valued balance requests. The
callbacks keep their original call order and native object pointers.
The linked-entry reduction reaches 35 with this change. Production C
entry tests passed on macOS ARM64 with strict Cgo checking and on Linux
ARM64 PIE. The four gameplay replays cover Polyp death, the Troll
cloud and its damage ticks, and the shooting trap.

### Headless replay artifacts and peer movement

The headless GUI runner puts observed frames and pixel differences
in the run's `screens` directory. The Warrior and Conjurer Quest
starts have distinct screen names. Both Quest scenarios and the
Warrior chapter-one reference passed in parallel. The expected
fixtures retained the same file names and SHA-256 digests.

The independent peer replay dismisses the stock dialogs and sends
the normal Action input before walking. A loopback replay confirmed
a shared player netcode and over 194 world units of movement on the
server and client with the unit in normal collision state.

The manual-save replay uses the game's case-insensitive filesystem
lookup when writing and reading its expected player state. The Linux
ARM64 seed and fresh-process manual load both passed on a
case-sensitive filesystem.

The replays across this checkpoint passed all 203 stand-alone
scenarios and the six dedicated, peer, and saved-game scenarios.
Four campaign cases passed with verified earned checkpoints.

### Checked tile edge reads

The tile edge reader applies the memory check to its addressed byte.
It derives the blob base from that checked pointer for operand reports.
The reported native pointer and logical PE32 offset stay paired.
The Linux ARM64 `safe` build passed the tile ownership, access list,
collision, and mana layout tests.

The default macOS ARM64 client has 87 Go-callable preamble entries
matched by source definition and linked name. The count includes 18
shared and 13 typed native call adapters, 31 native memory and scalar
adapters, four platform CRT adapters, and 21 regression probes.

### Callback ABI probes

Eight callback ABI probes compute their markers with the C compiler's
`_Generic` expressions and expose the results as C constants. Go
selects the markers using the original signed 32-bit selectors. Four
callback address getters read the linked C function symbols directly.
The native call and field-layout controls still run through their C
entries. The matching tests passed on Linux ARM64 PIE and on macOS
ARM64 with `cgocheck2` and `checkptr=2`.

### Monster death services

Golem, Skeleton, Boulder, and monster explosion callbacks pass the
live origin field directly to the Go point, force, damage, and packet
services. Each service reads the field at its original call site so a
preceding operation can change the next position. Three shared debris
adapters carry a nil object's field address through C as a scalar and
leave the first dereference to the original consumer. The explosion
uses the same point effect adapter. The live paths use the existing
Go random, object, decay, sound, and game flag services.

Seven Cgo helper definitions disappear from these sources. The C flag
query for the monster edible behavior links only in the test binary.
The native entry and layout tests passed on macOS ARM64 with
`cgocheck2` and `checkptr=2`. The Golem, Skeleton, special death, and
object death spawns replays passed with the resulting gameplay paths.

### Unaligned tile and secret wall reads

The native tile-list and secret-wall readers load unaligned integers
in host byte order and copy exactly the native pointer width into a
typed local pointer. The wall iterator reaches the existing Go service
with the same record pointer. The raw wall store and freeing of C-owned
nodes continue through their C entries.

The new test exercises a misaligned pointer and both integer widths.
It passed against the previous C loads and the new Go loads. On macOS
ARM64 with `cgocheck2` and `checkptr=2`, the full native tile corpus
passed 2,098 calls with 4,196 original control-word comparisons. The
secret-wall corpus passed 1,092 calls and comparisons. The secret-wall
gameplay replay also passed.

The default macOS ARM64 client now links 63 Go-callable Cgo preamble
entries, down from 87 at `dacc21866`. The census pairs source
definitions with linked symbols in their Go packages. It finds 18
shared call adapters and 45 legacy package adapters. The latter
include 13 typed indirect calls, 13 pointer and scalar transports,
eight malformed-address adapters, three CRT integrations, and eight
regression probes. The top-level legacy directory has 27 C files.

### Native verification and earned campaign checkpoints

Linux ARM64 PIE passed the 41 packages with tests and the network
release probe. The focused native death tests also passed. Linux 386
and ARMv7 ran the unaligned field and related legacy tests with the
`server,safe` tags under QEMU. The ARMv7 linker emitted the existing
two-byte `wchar_t` diagnostic. Both Windows 386 and AMD64 server builds
completed.

The client passed 17 targeted macOS ARM64 replays and eight Linux
ARM64 replays. They include real death callbacks, raw regeneration,
inversion, secret-wall contact, and the chapter-one screen on macOS.

The Wizard3 MagicMissile replay observes historical HP33 on Load and
native HP immediately before and after Lesser Heal. When the hero
retains HP33 through combat, subsequent observations must retain 33.
An observed injury below maximum must heal to HP32 and persist through
manual Save. The new passive recovery wait provides enough time for
the existing 40 MP minimum after variable combat spending. The
revised replay passed with both the `dacc21866` client and the
client containing the three boundary changes.

The Wizard3 stat audit used a combat-cleared HP33/MP47 Save from the
archived cellar replay. Ordinary gameplay recovered two mana in 182
frames, followed by the game's manual Save. The original Swordsman
retained its DEAD flag and native HP0. Loading the resulting HP33/MP49
Save verified the exact stat words and the original maxima of 32 HP
and 98 MP.

The same current macOS client passed 12 of 29 individual campaign
cases. The two Swap stages and the Warrior crypt caster passed with
archives matched to the recorded save manifests. Wizard2 passed
Fireball acquisition and Spike from an earned GoldKey checkpoint.
An archived pit-cleared checkpoint led to Book of Oblivion
acquisition and reload, the video Save/Load smoke, the upper
Urchins, and potion and food recovery.
An ordinary replay of the upper Urchin route saved at HP26/MP2
after 105 frames of recovery at the refuge. The potion and food
replay loaded that exact state and passed with the required pickups.

## GAME4 native ownership and validation boundary, 2026-11-10

The complete 140,091-byte `GAME4.c` baseline remains byte-exact inside one
outer false branch, with its nested guards and between-definition text.
The source map records 140 literal definitions: 120 same-name Go exports,
six thin C ABI adapters, four private Go helpers and ten renamed integrations,
including earlier Go-backed entries. Frozen map SHA-256:
`01aff0e7f2e17a8bfb0a9a624daf5c6910894033b904e007ee9b9525352ba014`.
The separate C data translation unit has ten owners and no algorithms;
typed C storage, services, ABI transport and private references remain.

The 70 support definitions (15 temporary-list, 12 map-object-list and 43 vote;
57 public / 13 private) use a separate denominator. Their 35-name overlap
with the GAME4 inventory is counted once, rather than presented as extra
semantic migration or original-CPU proof.

Native Object, PlayerUpdateData, Player, TradeSession, list/tile and FILE
links retain full-width owners. Packet scalars, 76-byte wire records, float
bits and protection tokens remain fixed-width. Mixed gold/CRC return words
cross C as `uintptr_t` before Go sees them; Go never manufactures a pointer
from a low token. The eight adjacent native gold consumers close 21 old C
gold calls, while the C trade packet branch forwards the actual native formal
owner to the existing 13-subtype Go decoder. Other direct PE32 helper chains
and unobserved mixed shop return values remain outside the claim.

Vote creation caches the target before allocation/reentry and keeps ordered
type/frame/team stores. Bounded 48-byte reward names and zero-initialized
52-byte packet padding are explicit native policies; added invalid-owner or
malformed-input guards do not prove original fault behavior. A declaration-only
Cgo correction aligns the tile-head slot with `nox_tile_coord_entry_t*`,
preserving public Go accessor signatures and algorithm bytes (proof
`35143b5d60030e43bf9cf0bb81ac74a660b0f2e24c14026527d9681ec1bc0a01`).
The fixture allocator constraint correction leaves inputs and assertions
unchanged. Private reference-head accesses now use typed C getters/setters
in the same translation unit; all 70 reference definitions and the complete
header remain exact. The first three failed normal outcomes, including the
third attempt's reference-head link failure, remain in full canonical form.
Across these attempts, 1,115,947,883 verified duplicate bytes were removed;
this is a logical count, not measured physical reclamation.

Normal macOS ARM64 / Go 1.27.3 root qualification passes 1,168 selected
IDs / 253 roots, zero failures or skips, three actual mains and unchanged
source: `8aebba36c2d2807f8d4862ac6214b6c47041595e02fc0b02f63c0f05cfafd272`.
Normal lease (9,063 files):
`b02427194acba9359b1d686272f7e98eb999fe5a973006ae1718b9b882d122ac`.
Strict/highres cgocheck2/checkptr2 also passes the same 1,168 IDs / 253 roots,
with zero failures or skips and three actual mains:
`842d80fdf67b52adf0961d171fcd37b114009a9ec143609ee249c5b640009f9b`. Both profiles use that same full source lease.
The independent selected-tests reader verifies the same IDs, six actual
before-EXEC-preserved mains and canonical records:
`ff11fa89fa9493c45de0f8d00e73397dc1dbcb50695e8f0cb5f0480a60e6c6e4`.
Both clean normal/HD products build and run help with Go 1.27.3 and
`vcs.modified=false`, from implementation commit `c92ed1ede8981adaa824995024aaff31a8cf2eb5`.
Their full clean source lease is
`b397a196f0febb707e730af43e641a5891a8b816f7d6b6565daba9e176be03d1`; independent product Q is
`54c710f07796adcc328e30bf33268f492859623ad61579088a7761de4521611c`.

All eight local product GUI/E2E invocations pass: the four unchanged source
cases in both normal and HD, with natural whole-parent exit 0 in 318.760 s.
Independent GUI Q:
`139d2f732f5328b57836ebf9e6b914f326c3837652a886eaf80b87257cb9c584`. Root actually viewed all
six declared current options PNGs; visual ledger:
`8271573761b2ed6cad56ffab305aaadc1b2984da28e521351f84664cc1db69b4`. Cases without PNGs
receive their controller assertion scope only.

Exactly eight owned raw runtime case directories were removed after full
canonical SHA/CRC verification and actual frame review. A fresh canonical-only
reader passes all eight cases after cleanup:
`64a7445b2180eea8729995e947b013ea8b24e89693b83a3e68c89dbfcb3f96e7`. Original assets
(4,439 files), protected Save files (126), products, goldens and failure
archives remain preserved. This does not establish original pixel equivalence
or post-cleanup inode identity.

CI is unused. These source connections do not supply all-entry original-CPU,
all-gameplay, physical-audio, final campaign or whole-project 100% credit.

## GAME3 native pointer ownership and local validation, 2026-11-10

The exact `GAME3.c` payload is preserved byte-for-byte: 143,656 original
bytes, all 119 definitions and intervening text, including nested guards,
inside one outer false branch. Connections partition into 100 direct Go
exports, 12 thin C ABI adapters and seven private connections using six
distinct Go helpers, including earlier Go-backed entries. The separate seven
C data owners have no functions.
Typed C storage, ABI forwarding and private references remain; cgo remains.

Native window/data, text, selection and callback fields retain full pointer
width and declared owners. The two listbox draw and five radio signatures
are widened together with their native consumers. Scalar clock DWORDs,
opaque image handles and fixed wire fields are not widened as pointers.
Selection shifting uses exactly Count owned uint32 entries; the omitted original
extra read produced a value immediately overwritten by the last sentinel.
This is a native allocation-bound policy, with pinned-owner and actual
guard-boundary controls, not equivalence for an out-of-owner fault.

Source corrections cover radio cached-parent/live-XOR reentry, low-BYTE door
status, pressure-plate activation before width/height RW, signed spark Y, and lazy bounce
frame access after the Z write. Two old declarations now match their actual
zero-argument ABI. Nine older cgo scalar declarations now use the exact
`uint32_t` type of their shared C owner. The macro/function boundary preserves the full callback
and native registry. Undefined old map locals receive explicit zero values;
invalid-input and valid-owner registry limits remain separate policies.

C-owned output and explicit Pin/GC controls exercise the C-to-Go pointer
return boundary. Deferred cleanup releases C data owned by the migrated slider
constructor before its window returns to the pool, with explicit-event and reentrant-disposal
controls. Cleanup state lives in Go side storage; the C-facing window and
GUI layouts remain unchanged.

Both profiles passed with zero failures or skips, local macOS ARM64 / Go 1.27.3 only, with CI unused.
Normal: `866` IDs / `121` roots,
Q `88d1b4f37118919524f367b93faefc7c6c589146920e1f749479a0fa260926bc`. Strict/highres with cgocheck2/checkptr2:
`866` / `121`,
Q `a5d50c4da842d5b6524ae8f3205aa9f643ccee3e6da9afc7aecfd520f7491718`. Both used the exact source lease
`48aac0ba271f1defcfe9017c0fa1c0c7fe42d9802aaaa48a1aa1f10cb8a2c402` and natural terminal records.
The independent selected-tests reader Q is
`75fb624cdee4176b699b01d7f9d57100e61c85fc7513f54c2128539ff41c7036`.
Clean product code commit is
`03568afb8d3a2ffe75a985e9b93649d1b60c290d`; only these two evidence
documents change afterward. Both normal/HD products build and run help with
`vcs.modified=false`; source lease
`5ccb19f29c48bda1032c309dba3a17ec27524dc5f4936597e59e484d07eef605`,
independent product Q
`0f441718cc8fa8947152d8e7345318d9e570e695e00fe5bca49db67ff7fdb274`.

All eight current product GUI/E2E invocations pass (the four unchanged source
cases in normal and HD), with independent GUI Q
`a36f4e773d49c5c590379a7571badfc4d73bb00dd744254dd76fd545065ec3a9`.
Root viewed the six declared current options PNGs; ledger
`0948dc89cd5ca375aad2a911dbd52a83e623f9e164fce273009b5fac125cafaf`.
No visual credit is assigned to cases without PNGs. Exactly eight owned raw
runtime directories were removed after full canonical verification; a fresh
archive-only reader passes all eight,
Q `6cc1bbf9c0cca9652d7b6d60534f33b168f169f46e6115089030355fa75e9c6e`.
All 4,439 original assets and 126 protected Save files remain exact; products,
archives and original goldens are preserved. This does not claim original
pixel equivalence or post-cleanup inode identity.
No all-119 original-CPU, all-gameplay, physical-audio, final campaign or
whole-project 100% credit follows from this source inventory.

## GAME2 native ownership and evidence boundary, 2026-11-10

The source-only v3 audit verifies the complete original 264-definition
`GAME2.c` payload, every body, source order and all between-definition text
inside one outer false branch. Its 262 previously active names have unique
connections: 244 direct Go exports, 13 thin C ABI-to-Go adapters and five
private Go helpers. Two previously disabled definitions are recorded
separately. There are no active algorithm definitions left in `GAME2.c`, and
the new `game2_native_data.c` has no functions. This is a source partition,
not independent behavioral/original-CPU proof for every entry or every path.

The retained native data owners store C-allocated links, windows, strings,
contexts, events and voices at full pointer width. The server-options window
getter reads its actual `nox_window*` owner rather than a low DWORD. Book,
briefing and quickbar class decisions use the existing native player-class
accessor instead of treating the packed player-info member as a cgo field.
Fixed wire IDs/masks and the 200-byte scalar sound record remain fixed-width;
pointer-bearing event/voice structures and their native sidecars are distinct.
Thin adapters handle const or mixed return ABIs without making a Go pointer
from a serialized DWORD. Public pointer signatures were widened where needed;
this batch does not claim an unchanged `GAME2.h` or removal of cgo.

The bounded changed-ABI query checks 37 names. Its four active retained C
consumer calls use typed drawable pointers; no int/LODWORD argument or return
sink was found in that bounded set. Former packed object-message cases 47,
48 and 0xC3 are preserved inactive, while their active C-to-Go bridge checks
one complete 9/11/12-byte message, opcode, callback registration and success
before advancing. High-pointer/canary framing controls are distinct from
selected main receiver/visual-state assertions and from broader dispatcher
or peer/gameplay qualification. External C algorithms outside this query are
not declared ported or pointer-safe by this finding.

Ordinary `nativeAudioFX` remains the production FX owner. The explicit legacy
service registers a real AIL context/bank and owns separate native resources;
ordinary initialization does not automatically duplicate them. Parser taps
run only when the explicit service is initialized, and shared reset/shutdown
preserve the ordinary owner until its own lifecycle closes. New native event
tokens do not enter the old packed voice/device callback path. C-owned
storage stays alive for retained links; controls exposing Go owners pin them
for the C retention interval. Freed-slot return values remain opaque tokens.

Two failures were corrected in production without changing fixed assertions:
zero sample-list lengths now break before touching caller scratch, while
initial empty sound names still write NUL and AUD dot-NUL/tail bytes remain
intact. A stopped owner is now an invalid handle even if generation zero
matches Stop's serial zero. Raw uint32 serial allocation and valid live zero
handles remain unchanged. This stopped-owner boundary, including setting a
handle after Stop, is a native lifetime policy rather than old-C equivalence
for that stopped/invalid domain.

The completed independent normal review checks legacy 773 IDs/40 roots and
main-normal 86 IDs/16 roots, zero failures/skips, against the same 8,997-file
lease. Its canonical ZIP checks stream full member SHA/CRC and preserve
regular bytes and symlink-target bytes without extracting or following ZIP
links. Main PREEXEC/running/terminal events retain the same actual PID and
main bytes. The final all-five reader also independently checks normal versus
highres/cgocheck2/checkptr2 embedded Go 1.27.3 settings. A vanished transient
main path is optional when
the actual retained and output copies are present. The original main-normal
metadata rejection remains exit 1/false; the independent review qualifies
the same execution without a rerun or an invented original driver pass.

All selected test parents have now closed naturally: legacy normal and
strict each 773 IDs/40 roots, main normal 86/16, real-null-OpenAL main 7/7,
and strict/highres main 93/23. There are no failures/skips. The 1,732 passing
profile events include repeated IDs; legacy normal/strict IDs are identical,
and all five source leases are the same 8,997 files. The final independent
all-five review qualifies these actual runs and source/main/strict-metadata
bindings. Canonical archives retain executed mains, available compiler outputs
and profile evidence before exact-owned duplicate cleanup. Byte/SHA/CRC and
symlink-target preservation do not establish inode identity or physical disk
reclamation. Each real-audio profile also preserves the new bridge child's
own PID/main bytes; historical six-child individual PREEXEC credit is still
excluded.
Clean normal/HD products now bind code commit
`35a68f9dbac3f5d37d06bc82d21d708e621a4123`, Go 1.27.3 and
`vcs.modified=false`; both build/help processes close naturally with zero. Independent
product review SHA-256 is
`757fb4849a1ce40142f8518e2a314160bba74628924193965d9a9e113deffdc2`.
Ten unchanged product GUI scenarios (five per product, including advanced
server options) pass. Their independent GUI Q has SHA-256
`2ff2db769a0d1e1cc1fefd42354279241faba4bdb3e3871fda57c28202125f11`.
All34 recorded PNGs were actually viewed by root; advanced-options has no
recorded PNG, and the book images do not prove a complete book-page display.
Ten exact owned runtime directories are removed only after canonical SHA/CRC,
independent controller review and visual review. Post-cleanup canonical-only
verification passes, SHA-256
`7d3f1db0c7805966b31ae949590ea668fa152ad19db11d776daf0943067493ec`.
Mutable files and PNGs remain complete in canonical case archives; immutable
stock map bytes are bound to protected original-map seal references. Original
4,439 assets and 126 protected recovered Save files remain exact. Total removed
logical duplicates across tests/builds/GUI are 4,484,509,207 bytes; no physical
reclamation claim is made. The final documentation commit changes only these
two evidence documents from the code/product commit.

Source facts are in `independent-function-coverage-audit-v1/` (v3 detailed
mapping SHA-256
`d7d41820eb161495f4d8c817597cb1ccda6206adc7307ca31f2301b839057cc2`,
changed-ABI audit
`809f8723fdfb6844ae41eeb687e0f43e5aca4e439d04a67e9f2ed11e15904ad6`).
The completed subset review is
`independent-completed-normal-review-v1/independent-runtime-qualified-v1.json`,
SHA-256 `2312e859c32a8231bfd2cf8cd3474a04af1721a6537872aaf326c2c2f0f546e6`.
Root terminal strict qualification `focus-strict-v1/tests-qualified.json`
has SHA-256 `cec6fe7c357ad136cfb0aadd630e17fe8b9c600b0618f315db0f5d172dd02615`;
`main-focused-v3/main-focused-qualified.json` has SHA-256
`cbbe67065268f9fa08ad6e60f26d64aa122ab0722cbe0fd117b2ddf74f7f73ac`.
Final `independent-all-five-runtime-review-v1/independent-runtime-qualified-v1.json`
is 190,820 bytes, SHA-256
`88e172a53afa26902f5a97412aa96c7f2848cae7916c59cd760309dfffe9a566`.
These findings do not supply all-entry numeric/CPU equivalence, historical
six-audio-child individual PREEXEC credit, physical audibility, broader product
gameplay, final campaign audits or overall project 100%.


## GAME1 support implementations in Go, 2026-11-10

The five remaining access-list support implementations now run in Go: lazy
initialization of both sentinels, native head selection, and the three borrowed
name/serial field accessors. Public C function signatures and nil behavior remain
unchanged; any nonzero banned selector still selects the banned list. Go
`sync.Once` initializes both existing native C data objects before publishing
either head. Their storage, full pointer width and process lifetime remain in C;
no Go heap head or PE32 pointer cell is exposed to retained C links.

The four arithmetic operations already had Go implementations. Their old C
support bodies are now reference-only alongside the preserved algorithms.
All 155 literal C definitions retain their exact definition/body hashes and
source order, and `GAME1.h` remains byte-exact. Only six existing ABI forwarders
remain active in `GAME1.c`; their algorithm implementations are in Go. Typed C
data and ABI forwarding remain, so this does not claim removal of cgo or all C.
This supersedes the nine retained C support implementations described below.

Local Go 1.27.3 validation passes **444 selected IDs across 15 roots in each
profile**, normal and highres with `cgocheck2` plus `checkptr=2`, with zero
failures or skips. Both profiles execute identical IDs against the same 8,946
source files. The new tests exercise actual C exports, nil and high native
addresses, independent C field offsets and canaries, UTF16/serial aliasing,
nonzero selectors, repeated lookup and GC, populated-head preservation, and
existing Go/C list traversal, removal, string consumers and free order.
Existing access-list controls and frozen chat/trig references also pass.
The historical 21,720-ID suites and ten product games below remain historical;
they are not relabeled as fresh full-suite or gameplay runs for this change.

Evidence is under `artifacts/opennox-port-20261110/GAME1-support-native-port-v1`.
`tests-qualified.json` has SHA-256
`aed0fc29dbb475b2361fb06d76e948cb6b28eb786302fafe8ee1e8c04a171afd`;
`support-port-literal-inventory-audit-v1.json` has SHA-256
`6e44c77ab3f48c03aad61c77cb3ea6b778295c12c0bd9a9b552d6a436349cbe6`.
The nine-operation Go mapping is recorded separately. Full regular-file bytes,
executed test mains, standalone compiler-produced executables and available
WORK are retained in SHA/CRC-verified canonical mode archives. Exact owned
runtime/build duplicates were removed after verification. These cleanup records
make no inode, symlink preservation or physical disk reclamation claim.
Original 4,439 assets and 126 Save files match their protected input seals.

Independent focused evidence review is
`independent-tests-evidence-review-v3.json`, SHA-256
`406f6a7fa4ed3ac4ab247be26c3eeb7da5ab0b8e73035bab628731052e95123b`.
It verifies both profiles, main PREEXEC event chains, strict embedded settings,
compiler-produced executable copies, full regular-file archives and the source
checkpoint before the two documentation-only edits. Its first metadata reader
attempt incorrectly required a vanished Go temporary main path; that rejection
is preserved. The corrected reader verifies the actual retained main copies
and event chain. No suite, input, expected value or implementation was changed
or rerun for this metadata correction.

Clean code commit `a0b29529dcb0047077ff68eae2b962650cabc788` supplies the current
normal and HD macOS ARM64 products. Both embed Go 1.27.3 and that revision with
`vcs.modified=false`, and their build and help processes exit naturally with
zero. Product qualification `products-qualified.json` has SHA-256
`2026f7c888171a3bfe5b20b0944bd241a7770078cf225fcc6f9c51fae3d57f1a`.
Independent product evidence review
`independent-products-evidence-review-v1.json`, SHA-256
`038634b3dfc32b47b1726f39a877b8a5482e2232dd978ffcdcbd1a2fc13ed0fc`,
confirms complete product bytes, PREEXEC archives, independently read embedded
build settings, natural build/help termination and canonical regular-file
preservation. The later final documentation commit changes only these two
evidence documents relative to the code/product commit.

| Product | Bytes | SHA-256 |
| --- | ---: | --- |
| Normal | 68,356,594 | `dd543831240a5e2bfe15db4ef1443bb141a67eb7cf1597882ebddd4cb1727898` |
| HD/highres | 68,472,226 | `d6267a7c537e7cead7b91f1e4d1de6f0e8815e786368f05152bf4df6597d1da3` |

All four test/build modes are canonicalized and their 12 exact owned duplicate
subtrees are removed, totaling 1,082,741,210 logical bytes. Current products and
full PREEXEC product copies remain available. A final fresh check after both
products and cleanup again verifies all original 4,439 assets and 126 Save
files against the protected seals. Build/help checks add no new gameplay or E2E
coverage, and physical disk reclamation remains unmeasured.

Whole-project pointer/call-graph coverage and the existing campaign, UI and
physical-audio completion gates remain open. This closes the selected GAME1
support implementation work, without an overall project 100% claim.

## Remaining GAME1 non-sub native ownership, 2026-11-09

The additional GAME1 port comprises head/settings21, world/player23,
thing-parser10 and diagnostics2. All 155 literal C definition/body hashes and
source order remain exact, with the 56 replaced bodies retained under `#if 0`;
the public header remains unchanged. The retained15 audit classifies six existing
Go ABI forwarders, five real typed access-owner/field helpers and four numerical
support helpers. No pointer-to-DWORD reconstruction was found within those
15 definitions. This finding covers the selected definitions and consumers,
rather than every call graph or macro configuration.

The complete literal inventory partitions into **146 Go entry connections**
(137 direct public Go exports and nine C ABI forwarders) and nine C support
helpers. The earlier three map-mode Go exports and four named ABI forwarders
are counted once alongside the 83 sub_ and 56 additional entries.

Native pointer ownership and serialized scalar storage remain separate:

| Boundary | Native representation | Fixed scalar/record representation |
| --- | --- | --- |
| Players, objects, teams and modifiers | Existing typed owners and full links | Masks, IDs, coordinates and counters keep declared widths |
| Secret-wall and access lists | Actual C/native head owners, native Next, typed fields | Named scalar fields remain DWORDs; no PE token becomes a host pointer |
| Tile allocation and lookup | Full row pointers, actual immutable capacity owner; checked C/Go cell layout | Cell scalar fields remain fixed-width; missing-map transition policy stays explicit |
| Tile/edge definitions | Checked native TileDef and edge-image pointer sidecar | Variable on-disk BYTE/WORD/DWORD parsing is distinct from EDGE's packed legacy memory-record stride60 and its native image pointer sidecar |
| Settings and status | Actual C dirty/flag symbols and native memory owners | 58-byte settings records and seven-byte status packet stay fixed |
| Memory diagnostics | Native pointer side slot, two pointer roles and named node fields | Three node DWORDs; 1,024 output rows of80-byte name plus DWORD count |

Typed C callback bridges preserve full function and argument addresses. Go
allocations exposed to C are pinned individually while retained; saved C heads
are restored before unpinning. C Free applies to C-owned controls. High-address,
layout, whole-region canary, callback-order and GC assertions are source-bound
checks; actual execution and strict-mode qualification are recorded below.

The parser cursor uses native data/current/end pointers and signed32 sizes.
Zero-sized reads do not manufacture pointers. Short or inconsistent streams,
invalid fixed names and negative signed block skips fail through declared
native policies, preserving completed effects. Audio validates a whole serialized
record before entering the unchanged external C audio helper. RGB policy supplies
three actual serialized bytes to the existing color endpoint; old uninitialized
stack bytes are outside comparison scope. Unsized caller-buffer ABIs retain their
caller-capacity requirement. Map-basename bounding/nil protection is also an
explicit native policy rather than evidence for executing old C undefined behavior.

The native debug head is a memmap pointer side slot, not a DWORD reinterpretation.
The source audit found no active producer; valid native controls supply actual
linked allocations. Cycle/overcapacity/unterminated-name handling bounds output,
with DWORD count wrap retained and the actual C qsort/raw SUB32 comparator kept.
Trig owns table writes/control/conversions in Go while retaining C sin/acos.
The 49,152-byte reference tables and flag cases are finite tests, not all floating
point environments or arbitrary input equivalence.

The before-port diagnostic baseline covers20 selected IDs. Private exact-body
current-C references cover54 definitions, comprising21 settings,23 world and10
parser entries, on their defined input domains. They supply separate reference
and ownership checks, not all56 original PE CPU equivalence or fabricated
pointer-truncation crashes for bodies already using native C pointers.

Go 1.27.3 local macOS normal and strict/highres each pass **21,720 selected
IDs**, with zero failures or skips: all prior 19,555 IDs plus 2,165 additional
IDs, including the two existing trig roots. Strict compiles highres,
`cgocheck2` and `checkptr=2`. All 8,944 source files remain exact across both
profiles, and all six actual mains retain their full bytes before execution.
Compiler-produced executable bytes are also preserved before wrapper return;
these records do not establish actual C test-child PIDs or invocation counts.

Independent source-bound qualification is
`actual-nonsub56-source-bound-qualified-v2.json`, SHA-256
`e9d69f9d376cecd9408b809565b58536482e32531a742d282d873099e6e5a7e7`.
The canonical MODE-archive reread also qualifies:
`actual-nonsub56-canonical-source-bound-qualified-v2.json`, SHA-256
`3e166e0c547cb7024f1566e39a4f9e6a70495c29d6f80be19d2b810b2db988f0`.
Its scope is transport and source-bound assertions; all56 original CPU
equivalence, independent recomputation of every numeric array and
recertification of all historical raw corpora remain unclaimed. Archive
qualification alone supplies no duplicate-deletion completion claim.

Clean code commit `803a3ba5e07dbc69b8ce91199e8c0f89d659044e` supplies both
normal and HD macOS ARM64 products with Go 1.27.3. Each embeds that exact
revision and `vcs.modified=false`, and each build/help process exits naturally
with zero. Product SHA-256 values are:

| Product | Bytes | SHA-256 |
| --- | ---: | --- |
| Normal | 68,338,850 | `ac308f53a66e1dde994ff8720f08375182630f4dc33ef3e892a8e3822f81e20c` |
| HD/highres | 68,470,994 | `670f7aeb7a1ac4c139e3735c070fe788a93981e7cb8e6f1ce6f2bb25ab735706` |

The independent two-product qualification is
`actual-nonsub56-clean-two-products-independently-qualified-v1.json`, SHA-256
`3a8b86a1fdda57c6768be7037a5351c3dc6254984a5ea176ab621bb139ffdf4f`.
It verifies both complete PREEXEC product archives, embedded build settings,
actual build/help records and the clean source binding to the tested native
epoch. The 4,439 original assets and 126 original Save files are verified
unchanged. This code/product commit is separate from the later final evidence
documentation commit; build/help qualification adds no gameplay coverage.

The normal/strict qualification verifies source-bound assertions and evidence
transport; legacy all-entry raw comparison qualification remains false.

All ten local repair/rejection/quantity/equipped/hearing games have same-PID
product PREEXEC preservation and natural zero exits. Independent qualification
`actual-nonsub56-products-and-ten-games-independently-qualified-v2.json` has
SHA-256 `1b82981d4e932e3cc9a192e619fb8ab96aa60c550729653c915f79f7ecbb0cab`. It
verifies eight sale sound-307 events, twenty prepared hearing conditions, eight
native turns and twelve negative controls; default-NPC and physical-speaker
credit remain open.

All 56 current capture paths close over 34 distinct encoded PNG streams: 16
current original-resolution actual views, 39 prior exact-byte aliases and one
current exact-byte alias. Visual ledger `root-current56-visual-ledger-v1.json`
has SHA-256 `298de3f52f5bad9cd8084a4ed58dcf17b41c0760e5a33a0cecdd4699261a44e3`.
The known MISSING shopkeeper heading and right clipping remain open.

Complete native MODE/available-WORK data, executed mains, compiler outputs,
failure records, baseline evidence and both products remain preserved in
verified canonical archives. Only exact owned duplicates were removed: 12
focused paths (777,505,233 logical bytes), 12 full-native paths (2,154,842,238
bytes), four baseline paths (430,714,859 bytes), two owned product-build WORK
directories, ten E2E runtime trees and 56 temporary PNGs (1,466,323,082 bytes).
Full inode/stat/SHA lists are compressed inside ZIP manifests. Removal receipt
`actual-root-nonsub56-owned-cleanup-receipt-v1.json` has SHA-256
`7564802e8c8ec9933e26af7dce28d62c457792a64b31c2d141c12d6050f409a3`. Original
4,439 assets, 126 Save files, products and canonical archives remain exact after
removal. Physical disk reclamation is not measured.

The selected GAME1 textual inventory is completely classified: 146 Go entry
connections and nine C support helpers. Whole-project100 remains unestablished. The retained15
audit separately queues the complete active GAME3_2 `4D8D40` ABI/object-update-
player/team chain for a later whole-entry audit; no isolated two-cast patch or
current runtime traversal is inferred. Campaign, default NPC hearing, physical
speaker fidelity, arbitrary corrupt-pointer safety and old raw-corpus
recertification remain outside this source-bound closure.

## All remaining GAME1 sub_ algorithms, 2026-11-09

The final 33 algorithms in the retained `GAME1.c` `sub_` inventory now use Go:
four title lookups, twelve settings entries, eleven access-list entries and six
team entries. The same denominator is **83 Go connections, zero remaining C
algorithms**. Seventy-eight use direct public Go exports; five keep C adapters
with only typed/const ABI forwarding. All 83 old definition/body hashes, their
source order and the complete public header remain exact. This selected set
excludes other C entry names, files, preambles and macro configurations.

The four title lookups now read localized UTF-16 at +0 and preserve the char key
at +4, DWORD mask at +8 and 12-byte row stride. Native pointer side slots retain
full addresses. Both real C loaders (`4158C0`/`415D50`) delegate to the same Go
loader as startup, removing their pointer-to-DWORD writes. The old native-C
loader diagnostic deliberately transports a key pointer across an unused mask
slot; it is preserved as a controlled bug observation, not a valid production
packed-row equivalence or a fabricated crash for already-native entry pointers.

Settings use the original 58-byte raw records, actual C global owners, signed
char promotion and full DWORD masks. Access nodes use the C sentinel owners,
native-width links, exact name/serial fields and the existing truncated-32-bit
clock boundary before uint64 expiry arithmetic. Team traversal uses typed
native object/player/update/team links. C object size is the shared prefix;
two Go-only handles follow it. Nearest-team math retains precision53/chop,
unordered comparisons and ties; team packets preserve all 18 prefix bytes,
UTF-16 bytes and low-BYTE length behavior.

Before-port C references and the frozen nearest-team original corpus remain
preserved. The separate title seal executes 64 original calls over two controls:
six roots/406 bytes plus the actual C-locale helper/223 bytes, 60 returns and
four nil-read faults. All 160 root instruction starts complete in both controls.
Nil-query zero and the current `416690` no-op remain separately declared native
policies. No full helper-branch or all33 original CPU equivalence is inferred.

Normal and strict/highres Go 1.27.3 each pass **19,555 IDs** with no failures or
skips: all preceding 18,938 plus 617 new IDs. Strict compiles `cgocheck2` and
`checkptr=2`. All 8,912 source files remain unchanged across the combined parent,
and all six mains have full bytes preserved before execution. Independent
qualification SHA-256 is
`7a85968a24532c05ce5c6f27a5386a231ee314a1be6634e1559ea65cdfb7e5a4`.
It verifies source-bound assertions and transport, not all per-call raw numeric
arrays, private native state or old corpus recertification. The compile/layout
test-only failures remain preserved alongside the passing focused attempt.

Clean code `b1ade25b23ec` supplies the normal/HD macOS ARM64 products with
Go 1.27.3, embedded `vcs.modified=false` and natural zero build/help exits.
Independent two-product qualification SHA-256 is
`dd37b75d48c7f2e78b48583cde37d5ca4194d33239518f410abef4945101c84a`.
The later documentation update does not replace this code/product source binding.
All ten actual shop/hearing games have same-PID product PREEXEC preservation
and natural zero exits. Independent product/game qualification SHA-256 is
`183eb09c00714c083cec271235bc12fbf70a92a0b90ec85a1f5e94b4786c3bb5`.
It verifies eight sale sound-307 events, twenty prepared hearing conditions,
eight native turns and twelve negative controls. All 56 capture paths have
actual-view or exact encoded-byte alias closure over 32 distinct PNG streams:
thirteen current actual-view paths, forty prior-view aliases and three current
aliases. Known MISSING shopkeeper heading and clipping limits remain open.

Validation stays local to macOS; no GitHub CI is queried, awaited or downloaded.
Complete native MODE/available-WORK evidence, executed mains, current products
and canonical failure/success archives remain preserved. Only byte-verified,
owned duplicates were removed: nine focused paths, ten full-native paths,
ten game runtimes and 56 temporary PNGs. Both large metadata JSON files retain
their whole bytes in a verified ZIP; the redundant raw files alone were removed.
Original 4,439 assets, 126 Save files, products and all canonical archives remain
exact. These diagnostics add no all33 original CPU equivalence, all-entry raw
array recertification, entry-specific gameplay, default NPC hearing, physical
speaker or nine-cell campaign credit. Other pointer graphs and campaign final
audits remain open; overall 100% is not established. The older 50/33 counts
below describe the preceding checkpoint.

## GAME1 startup, pause and comparator entries, 2026-11-09

`sub_4134D0`, `sub_4137E0`, `sub_413890`, `sub_4138E0`,
`sub_413900`, `sub_413920`, `sub_4139B0` and `sub_413F60`
now use Go algorithms with unchanged public C signatures. The retained old eight
C bodies already used native pointers; this migration supplies no fabricated
pointer-truncation negative control. Their original bodies and public header
remain exact, with only eight literal `#if 0` wrappers added to GAME1.c.

The C DWORD globals 251708/712/716/720/744 use actual symbol addresses, not
same-offset memmap shadows. Scratch 251740 and buffer 251636 retain their real
memmap owners. Scalar accesses stay BYTE1 or DWORD4. Focus callbacks remain
typed `void (*)(int)` addresses in the existing native C registration arrays;
no original PE callback DWORD is manufactured into a Go pointer. The const C
qsort adapter passes both full pointers to the private Go export. Its 84-byte
record still has a DWORD at byte 80: pointer width does not expand this scalar
layout. The two ordered reads return raw SUB32 wrap, without sign normalization.

Before R installation, 535 frozen inputs produce 1,070 independently checked
original rows over both controls, covering all 302 bytes and 87 starts. Both pure
helpers execute actual guest instructions. Music's external tail host return is
separate from a real original RET. Forty-six native old-C positive controls pass.
Current four-global startup and disabled two-BYTE/nil CD policies are explicitly
separate from the original fifth clear/helper and drive suffix behavior.

Normal and strict/highres Go 1.27.3 each pass **18,938 IDs**, preserving the old
17,854 regression IDs and all 8,883 source files. All six actual mains are saved
before execution. Independent new-family checks verify each profile's 1,070
calls: 162 compatibility-policy calls plus 908 core calls against both controls,
1,816 original comparisons, 690 core returns and 218 read faults. Four full data
pages, independently initialized BEFORE bytes, complete native operand metadata,
ordered reads/writes/service effects, packet consumption and full callback/string
role words are checked. Defaults, pinned Go owners, GC, observer reentry and
larger-page canaries have passing source-bound assertions; separate raw default
or GC-event logs are not retained. Per-call PID/main-SHA/child-Wait receipts,
physical fault-PC/nil-EA/private register/stack equivalence and window-focus
dispatch are not claimed. The old memdump/qsort parent pointer graph remains open.

Qualification: `actual-next-eight18938-independently-qualified-v1.json`, SHA-256
`830928063b3615df579c8e46f8c254b2a570b357e27fb5c54a1f752118a66632`.
The retained source inventory is **50 Go connections and 33 C algorithms out of
83 definitions**; all old body hashes remain exact. Older regression PASS is
separate from recertifying all older raw numeric corpora. Test-only import and
JSON decoder failures and the explicitly interrupted failed strict attempt are
preserved.

Clean code `62036a0a9` also qualifies both Go 1.27.3 macOS ARM64 products and
all ten normal/HD shop/hearing games, with actual product PREEXEC bytes and
natural zero parent exits. Independent product/game qualification SHA-256 is
`6956a141743cb841979712095fd464d6b31c7bfbc6a57a110e2e34f51f73bd6d`.
Twenty prepared hearing conditions include eight native turns and twelve
negative controls; eight sale sound-307 events use the OpenAL null sink.
All 56 captures have actual-view or exact encoded-byte alias closure. These
diagnostics do not prove default NPC settings, physical audio, execution of all
eight raw entries during gameplay or campaign completion. Known shopkeeper
heading/clipping limitations remain. Original assets/Save, references, products,
executed mains and failure/success records stay preserved in canonical storage;
only verified owned duplicates are removed. This final evidence update changes
the two documents only; 33 C algorithms and overall 100% remain open.

## GAME1 tile-list entries, 2026-11-09

`sub_411350`, `sub_4113A0` and `sub_411490` now use native Go algorithms
through their unchanged typed C signatures. Four node scalars remain int32;
Next stays a full native pointer at offset 16 (ARM64 node size 24). XY is
exactly two DWORDs. Variation is read before Edge, the actual 490 and 3A0
bodies execute in order, Value is lazy, and Next is read last. Aligned Go
pointer slots use typed loads; unaligned foreign slots use native-width
`memcpy`. No DWORD is reconstructed as a Go pointer. EDGE uses its actual
60-byte BYTE-table owner, distinct from the native TileDef array.

Before implementation, 1,049 frozen inputs yield 2,098 independently qualified
original rows across two FPU controls, covering all 412 bytes and 180 instruction
starts in both controls. All three original bodies, their internal CALLs and
the 48-byte switch table are sealed; no host branch/return model substitutes
for these helper bodies. Nine actual old-C positive controls pass separately.
The former C already had typed native pointers, so no new truncation crash is
claimed for this algorithm migration.

Go 1.27.3 normal and strict/highres each pass **17,854 selected IDs** with no
skips or failures, retaining all preceding 15,728 IDs. The complete 8,872-file
source lease remains exact throughout both profiles, and all six actual mains
are retained before execution. Independent raw comparison checks each
profile's 2,098 Go/C calls against both original controls, including independent
input/original BEFORE reconstruction, full native pointer identities, every
bank canary, ordered reads/real helper returns, input consumption and AFTER.
There are 1,602 primary returns and 496 primary read faults per profile.

Ten native-width controls include eight actual unfinished PTR8 reads across
20–23-byte node guards; they earn no complete-original prefix equivalence.
Eight default, four pinned-Go and four reentry calls bring each profile to
2,124 calls and 4,232 original-control comparisons. Sixteen individually pinned
owner operations and 124 forced GCs are source-bound controls, not separate
raw GC-event logs. The initial independent reader's empty-trace rejection is
retained; its format-check correction changes no engine, input or reference.

Qualification: `actual-tile-list-three17854-independently-qualified-v2.json`,
SHA-256 `2000cdc42a7a38c3a9decc1ad8a75f3e19d9e1d22f929af4df9374d98523cec7`.
All 83 former C definition/body hashes and the public header remain exact.
The static inventory has **42 Go connections / 41 remaining C algorithms**.
Next entries begin `4134D0`, `4137E0`, `413890`; removed installation-drive
fallback/state policy and native focus dispatch require separate review.
The exact clean implementation `954abbf97` subsequently builds normal/HD
Go 1.27.3 products and passes all ten existing shop/hearing product E2Es.
Independent product, raw/canonical log and 56-frame checks qualify; current
products and every executed main remain preserved. Eight native test/WORK
trees, ten E2E trees and 56 temporary PNGs are removed only after complete
canonical-byte and fresh-identity checks. Original 4,439 assets and 126 Save
files remain exact. Details and hashes are in the completion document.
Physical fault-PC/nil-EA/private stack/GPR/FPU equivalence, partial pointer
writes, full caller/gameplay, physical speakers, the prepared shopkeeper
heading/clipping and nine chapter 1–3 final audits remain separate gates.
Overall porting remains incomplete.

## GAME1 SecretWall list entries, 2026-11-08

`sub_410550`, `sub_410730` and `sub_4107A0` now execute native Go
algorithms through their existing C signatures. Next, Wall, the actual C head
and returned references retain full pointer width; wall ID remains an unsigned
WORD and the other record scalars remain DWORDs. Lookup retains the second
Wall read on success. Clear caches Next before Free and clears the head at
the end. Remove caches the live-head reload before Next and unlinks before
Free. The former C bodies already used native pointers: this is an algorithm
migration, without a new pointer-truncation defect claim.

Before implementation, 273 independently frozen inputs produced 546 actual
original records in two FPU controls: 408 returns and 138 protected read
faults. All 189 function bytes and 75 instruction starts are covered in both
controls. Eight CALL sites remain declared First/Next/Free boundaries; the
three helper bodies receive no original execution credit. Nine actual
standalone old-C positive controls pass separately.

Go 1.27.3 normal and strict/highres each pass all **15,728 selected IDs**
across three packages, with no skips. Whole-source leases contain 8,861 files
and remain exact before and after both profiles. A separate reader compares
1,092 actual Go/C corpus calls per profile with the matching original control:
816 returns and 276 read faults. It independently reconstructs native BEFORE
fields and canaries, checks ordered operands and full service pointers,
actual base/offset/effective addresses, consumed and unused inputs, and all
six native AFTER banks. Eighteen extra calls cover actual production-head
defaults, pinned Go owners and observer reentry; ten individually pinned
owners and eighteen forced GCs are exercised per profile.

For 148 returned remove calls per profile, original post-Free EAX remains
raw diagnostic data. The current typed-C cached removed-pointer return is
checked separately against the observed Free argument. A Free DWORD is never
converted into a native pointer; all ordered effects and final fields remain
compared. Individual injected service/stimulus writes are source-bound and
corroborated by consumption, following reads and AFTER, rather than separate
write events.

Independent qualification is
`actual-secret-wall-three15728-independently-qualified-v2.json`, SHA-256
`a8f494da86413d59632581cd8178947e07b84c43b032fb47eff21b8bdc082a3d`.
The earlier reader's missing historical negative-record reference remains a
retained rejection; its explicit historical-reference correction changes no
game code, native comparison or original fixture.

The exact 300-, 346- and 554-byte former C definitions remain under three
`#if 0` guards; the public header and all 83 retained definition hashes are
unchanged. GAME1 now has **39 Go connections and 44 remaining C algorithms
out of 83 definitions**. Next source-order entries are `411350`, `4113A0`
and `411490`. Native fault-PC/physical nil-EA/private stack/FPU equivalence,
partial pointer writes and shifted/cross-kind aliases remain unqualified.
The parent WORD discrepancy, `4E2220`, the separate `418AE0` high-bit key,
shopkeeper heading and nine chapter 1-3 final audits remain open. Overall
porting is incomplete.

## GAME1 wall attachment and NPC damage, 2026-11-08

`sub_410390` now executes in native Go through its existing C entry. It
keeps the drawable and wall references at full pointer width, while the
original coordinate pair remains exactly two signed DWORDs in eight bytes.
The drawable pointer store, cached flag byte, live coordinate queries,
polygon byte and final wall byte retain their original order. The complete
567-byte former C definition stays under `#if 0`; the public header and all
existing original fixtures remain unchanged.

Before implementation, 375 independently frozen inputs produced 750 original
records across two FPU controls, covering all 145 bytes and 57 instruction
starts: 626 returns and 124 protected read faults. Each normal/strict profile
compares 1,500 actual Go/C calls with the matching original control: 1,252
returns and 248 read faults, including ordered scalar/full-pointer accesses,
current eight-byte query coordinates and complete mapped AFTER/canary banks.
Fourteen extra default-service, observer and pinned-Go calls exercise 36
forced GCs per profile. External helper bodies, physical fault addresses/PCs,
private stack/FPU state and shifted/partial-pointer alias domains receive no
new equivalence credit.

The map-transition `IndexObject` nil dereference at the former line 308 is
reproduced with an outside-grid collider before the fix. Collider bounds are
now intersected with the finite AI grid without turning an empty intersection
into a border obstacle. Invalid point markers and waypoints are skipped;
valid overlaps and iteration behavior remain. All 1,045 selected AI-path IDs
pass in both modes, including the unchanged existing original corpus.

The reported NPC1343/Pixie672 IMPACT shape has a pure missile as both source
and weapon, with no unit update. It now enters the existing NPC case-11 tail
instead of reporting an unsupported shape. That shared NPC case-3/11 tail
also follows the original toward-zero amount/carry/Quest conversions; the
global rounding helper is unchanged. The old nearest-rounding carry assertions
now look up unique rows from the unchanged SHA-pinned 6,622-row original
numeric fixture. Twenty-two new Pixie IDs and 87 related existing IDs pass
in both modes, including actual collision-to-HP/attribution integration.
Original whole nonnil caller execution is not inferred from numeric fixtures.

Go 1.27.3 normal and strict/highres each pass all 14,625 selected IDs across
three packages with no skips. The source lease is byte-exact before and
after both modes; a separate data reader verifies actual retained executions
and raw comparisons. Earlier compile and functional failures remain retained
as failed epochs. GAME1 has **36 Go connections and 47 remaining C algorithms
out of 83 definitions**, with every retained definition hash unchanged.
Next source-order entries are `410550`, `410730` and `4107A0`.

The parent WORD fault discrepancy, `4E2220`, the separate `418AE0` high-bit
key issue, shopkeeper heading and all nine chapter 1-3 final audits remain
open. These scoped checks establish neither full gameplay completion nor
100% porting.

## GAME1 eligibility entry, 2026-11-08

`sub_40AA70` now executes in native Go through the existing
`int(nox_playerInfo*)` C ABI. Cached info/player identities retain full pointer
width. Info BYTE52 is distinct from the other settings getter; player status,
team key and iterator state retain DWORD reads at named native offsets 4976,
2072 and 2144. The lazy global read, cached key, two independent Check96 calls,
two unsigned low-AL results and signed count/limit comparisons preserve the
original order. The iterator now uses `int32(raw) > 0`, correcting the former
C unsigned comparison for high-bit state words. The former C entry already
used native pointers; this is an algorithm migration and a signed-comparison
fix, without a new entry-pointer truncation claim.

Before implementation, 662 independent inputs produced 1,324 qualified
original records in two FPU controls: 1,286 returns and 38 read faults. All
346 bytes and 127 instruction starts are covered in both controls. Eighteen
call sites remain external boundaries; their eleven helper bodies receive
no original execution credit. Expected original rows never enter native
workers.

Go 1.27.3 normal and strict/highres each pass 11,955 selected test IDs and
three packages. An independent reader verifies 2,648 actual Go/C corpus calls
per profile against the corresponding original control-word record: 2,572
returns and 76 read faults. Full int32 return bits, ordered BYTE/DWORD loads,
service words/arguments/full pointers, input consumption and every mapped
native AFTER/canary byte agree. Ten additional linked-default/pinned-Go calls
pass, including three individually pinned Players and eighteen forced GCs
per profile. Default controls use the actual production global, C AL bridge
and separate C limit symbol; these extra results are source-bound assertions,
not invented raw default-service traces. Whole-source snapshots before and
after both profiles match exactly.

The exact 1,351-byte former C definition remains under `#if 0`, with the public
header unchanged. Current GAME1 accounting is **35 Go connections and 48
remaining C algorithms out of 83 definitions**; all 83 retained definition
hashes remain exact. Next source-order entries are `410390`, `410550`,
`410730` and `4107A0`. The clean `34916ce3a` normal/HD products build and pass
help checks; they precede this entry and provide no new-entry gameplay credit.
Native physical fault-PC/nil-EA/register/stack/FPU equivalence, original
external helper bodies, the separate high-bit key issue in `418AE0`, `4E2220`,
the parent WORD discrepancy, the missing shopkeeper heading and all nine
chapter 1-3 final audits remain open. Overall porting is incomplete.

## GAME1 sudden-death entry, 2026-11-08

`sub_40A970` now executes in native Go through its existing `void(void)` C
ABI. Cached player identities retain full pointer width; the status read is
a DWORD with mask `0x100`, using native offset 4976. Four production C state
cells and the Go frame counter retain the original store and service order,
including the final game-flag clear. Default FPU control is `0x0E3F`, with
`0x023F` also checked. The exact former C definition remains under `#if 0`;
the public header is unchanged. Its former typed pointers were already
native-width, so this migration does not establish a new truncation bug.

Before implementation, 403 independent inputs produced 806 qualified original
records in both FPU controls: 784 returns and 22 read faults. The seal includes
the 140-byte caller and actual 17-byte `419A70` conversion body, 47 instruction
starts in total. Each normal/strict profile executes 1,612 Go/C corpus calls
against the corresponding original control-word records. Ordered reads,
service words, converted integers, complete mapped AFTER banks and canaries
agree. Eight additional default-path/pinned-Go calls pass, with 36 forced GCs
and 32 witnesses of actual production C cell pointers per profile.

Go 1.27.3 normal and strict/highres each pass 9,294 test IDs and three packages.
Independent readers bind actual executable bytes, bounded child Wait records
and raw observations. Expected original rows are never sent to native workers.
Of 3,224 intermediate float-store pairs per profile, 3,172 match bit for bit;
the other 52 differ only in private NaN sign/payload. Every non-NaN store and
every public converted integer matches. Private NaN payload, physical x87
register/stack/status, native fault-PC and nil effective-address equivalence
remain unqualified. No reference bytes or assertions were changed.

The first build failed on an unused import in the new test. Its full failure
record remains preserved, with zero legacy/native entry execution credit.
The successful retry removes only that unused test import.

At that checkpoint, GAME1 accounting was **34 Go connections and 49 remaining C algorithms
out of 83 definitions**; all 83 retained C definition hashes remain exact.
Next is `40AA70`: the original signed `field_2140 > 0` predicate differs from
the current unsigned C comparison and is awaiting original boundary-input
verification. The clean `552cf07ce` products already pass ten local shop/hearing
scenarios; those products precede this new entry and give it no gameplay
credit. Original external helper bodies, `4E2220`, the parent WORD discrepancy,
the missing shopkeeper heading and all nine chapter 1-3 final audits remain
open. Overall porting is incomplete.

## GAME1 info and active-player entries, 2026-11-08

`sub_40A740` and `sub_40A770` now execute in native Go through their existing
C ABI. Info, player, unit, update and opaque team identities retain full
pointer width. The original cached-info-before-flag order, BYTE status/index
loads, lazy engine-flag DWORD read and fresh global unit scan for each team
are preserved. No team-membership filter is added. The previous C definitions
already used native pointers; these are two algorithm migrations, rather than
two newly discovered pointer truncations. Their exact bodies remain under
`#if 0` and public headers are unchanged.

Before implementation, 394 independent inputs produced 788 qualified original
records in two FPU modes: 652 returns and 136 read faults. The complete caller
bodies cover 230 bytes and 85 instruction starts. Each normal/strict profile
makes 788 actual Go/C corpus calls and 1,576 comparisons with those records.
Ordered reads, raw service words, full relocated pointers, aliases, guarded
prefixes and whole mapped native AFTER/canary banks agree. Thirty-four separate
default-path and pinned-Go calls also pass, including actual player/unit/team
iterators and 64 forced GCs per profile. The total is 822 new entry calls per
profile. Go 1.27.3 normal and strict/highres each pass 7,671 test IDs and three
packages. An independent reader verifies actual main bytes and bounded child
execution records; original expected rows are never sent to native workers.

At that checkpoint, GAME1 accounting was **33 Go connections and 50 remaining C algorithms
out of 83 definitions**. All 83 retained C definition hashes remain exact.
Next source-order work is `40A970`, with its distinct DWORD status bit and
actual x87 store/conversion boundary, followed by `40AA70`. Physical fault-PC,
nil effective-address and native register/stack/FPU equivalence remain
unqualified. External service bodies, the retained `4E2220` traversal, the
parent WORD discrepancy and the nine chapter 1-3 final audits remain open.
The clean `2c0e65c2a` products pass ten shop/hearing scenarios; those products
precede these two new ports and supply no new-entry gameplay credit.

## GAME1 timer and setting entry follow-up, 2026-11-08

Ten additional algorithms now execute in Go: timer entries `40A1F0`,
`40A230`, `40A250`, `40A310`, and scalar entries `40AA00`, `40AA20`,
`40AA30`, `40AA40`, `40AA60`, `40E090`. Native scalar cell pointers exceed
4 GiB. The timer clock retains all 64 bits; deadline stores remain separate
low then high DWORD writes, and the live selected limit remains a BYTE read
after the clock. Public C widths are checked at compile time. These scalar
migrations do not imply ten pre-existing truncated pointer arguments.

The 172-byte/60-instruction original timer callers have 2,246 qualified
recordings from 1,123 inputs and two FPU controls. Per normal/strict profile,
2,246 Go/C corpus calls match both original records 4,492 times, including
ordered scalar/service observations and all 140 state bytes. Eight separate
default/reentry calls also pass. The six settings callers have 75-byte/
22-instruction static seals and 224 actual Go/C comparisons with the exact
preserved compiled C6, plus fourteen order/alias calls. All 104 state bytes
and the actual independent C-global neighbour canary agree. E090 clears its
separate C global, and its void return is excluded from return comparisons.

Both Go 1.27.3 profiles pass 6,853 test IDs and three packages; independent
raw-result/main-byte verification also passes. The initial unused-import
compile failure remains preserved, with no native legacy execution claim.
Only that new-test import changed for the successful retry. Exact prior C
bodies, source fixtures and actual executables remain retained.

At that checkpoint, GAME1 accounting was **31 Go connections / 52 remaining C algorithms
among 83 definitions**. Original helper bodies, protected/invalid-index timer
faults, native physical register/stack/FPU equivalence, the retained `4E2220`
PE32 traversal, the parent WORD discrepancy and campaign final audits remain
open. Source-order work continues with `40A740/770/970`.

## GAME1 Material7 and armor coefficient pointers, 2026-11-08

`sub_4133D0` no longer narrows its typed object argument, InitData or cached
descriptor to a DWORD. Object fields and material slot one use named native
offsets, and the descriptor cache is a server-owned typed pointer with reload
and free invalidation. The original cache-read/eager-InitData/resolver/class/
slot order and raw identity comparison are retained. Warm opaque descriptors
are compared as pointers without adding struct-alignment admission.

`sub_415BD0` now takes `nox_object_t*` and loads the definition's float at its
native offset 76. Its exact old C body remains under `#if 0`; a preserved-C
high-address crash and a separate argument-widened positive control are retained.
The old raw `4E2220` loop remains a PE32 carrier path with an explicit call-site
compatibility cast; it is not included in the native-pointer claim.

Normal/strict Go 1.27.3 each pass 4,477 tests and three packages. Checks cover
28 actual Material7 Go/C calls, pinned graphs and eight in-call GCs, opaque
identities, server cache isolation and invalidation, and 5,134 armor coefficient
Go/C calls with 2,560 typed C conversion comparisons. Original nonfinite/x87
physical return behavior and nil/partial-fault address equivalence remain open.
Two initial failures from new-test C-heap initialization were corrected in a
separate retry without changing production code, assertions or IEEE expectations;
all failed evidence and actual mains are preserved.

Six additional scalar entries also migrate to Go: `40A180`, `40A220`,
`40A3C0`, `40A6A0`, `40A6B0` and `40A710`. Each profile compares 1,180,680
actual calls with preserved old C, native live cells and state canaries.
Their prior pointers already had native width. At that earlier checkpoint,
GAME1 accounting was 21 Go connections and 62 remaining C algorithms among 83 definitions.
The parent WORD discrepancy and campaign/product completion gates remain open.

## GAME1 setting-flag entries, 2026-11-08

Five public C entries, `409E40`, `409E70`, `409EC0`, `409EF0` and `409F40`,
now call native Go bodies. Each retains its actual `int(int)` ABI and its
exact old C definition under `#if 0`. Subflags and the dirty flag remain
separate actual C globals with compile-time/runtime DWORD width checks;
no PE-address alias, adjacent struct or Go shadow replaces either cell.
Full signed return words, the second E70 read, EC0's cached result and
dirty-before-flags stores, and EF0's conditional live reload are preserved.

The complete original 256-byte/77-instruction caller set has 2,522 qualified
two-FPU recordings. Normal/strict Go 1.27.3 each pass 3,915 tests and two
packages, including 2,522 actual Go/C calls and 5,044 original comparisons.
Actual global addresses exceed 4 GiB; VM-bounded neighbouring bytes, raw
returns, ordered operands and independent service mutations agree. Declared
external helpers receive no body-execution credit. The prior five scalar
bodies did not truncate pointers; these are algorithm migrations.

Current GAME1 accounting is 13 Go connections and 70 remaining C algorithms
out of 83 definitions. Clean pre-flags `532dc25ae` products also pass the ten
unchanged shop/hearing scenarios in normal/HD with retained original assets
and Save data. The native parent WORD discrepancy and campaign gates remain
open; this accounting is not an overall completion percentage.

## Raw Skeleton pointer entry, 2026-11-08

`sub_4E23C0` now transports three native object pointers and two signed
32-bit values through its existing public C entry. Its exact 1,027-byte C
transcription, including its final newline, is preserved under `#if 0`.
The Go caller uses named native field offsets and preserves the original
DWORD X-before-Y copies, signed direction WORD, low-bit Direction test,
UpdateData pointer cached before Action, byte480-before-byte481 reads and
full signed parent result. The ordinary registered bool route is unchanged.

The complete 237-byte/91-instruction original has 502 independently checked
recordings from 251 chosen inputs and two FPU controls. Each local Go 1.27.3
normal/strict profile passes 1,391 test IDs and two packages. Its 502 actual
Go/C child calls each compare both original controls: 446 return naturally
and 56 retain actual captured native read faults. All selected ordered reads,
service arguments, scalar/pointer values, input mutations, returns and full
foreign-bank canaries agree with the original field projection. Children get
only independently selected inputs; original observations stay in the parent.

Compiled C/Go object offsets are 60/76/80/128/872. Compiled Monster animation
offsets are 552/553: Waypoints alignment contributes four extra bytes beyond
pointer widening. Go/C tests also pass with explicitly pinned Go heap graphs
and GC during Action after the update pointer has been cached. Actual nil-target
Buff/DefaultDamage fallbacks pass without diagnostic observers.

Native fault-PC classification and physical nil effective-address equivalence
remain outside this evidence. Go's actual nil trap and the numeric field address
are recorded separately. Declared helper boundaries supply no helper-body or
gameplay execution credit; the earlier parent WORD difference remains open.
Fresh product/E2E and the nine campaign audits remain separate completion gates.

## GAME1 source-order port, 2026-11-07

`sub_409A70` now implements its six-DWORD live settings lookup in Go behind
the unchanged `int(short)` C entry. The exact original C definition remains
under `#if 0`. The prior C implementation already kept its internal table
pointer native-width; this is an algorithm migration, not a newly fixed
pointer truncation. Existing `sub_409B50` and `sub_409B80` ports are retained.

Go 1.27.3 local macOS ARM64 normal and highres/strict profiles each pass
11 test IDs and two packages. The new original-table comparisons exercise
three variants and two input signs through Go and the real C entry (3,072
calls per profile); the existing stock consumer covers all 65,536 input
words. Separate preserved-C comparisons retain four address placements and
canaries. The final profiles preserve their actual compiled C executable,
source, input, pre-execution hashes and all three child results. Earlier
successful profiles lacked that ephemeral C executable and remain recorded
with this limitation. No new original execution is claimed for this leaf.

The source-order inventory contains 83 `GAME1.c` `sub_*` definitions. This
change moves one of the 76 remaining C algorithms to Go: eight entries now
have Go connections and 75 remain to migrate. These are inventory states,
not an overall completion percentage. Raw Skeleton and campaign gates remain
open as described below.

## Raw damage checkpoint, 2026-11-07 (WIP)

Raw DefaultDamage `4E0B30` and the four scalar callers `4E24B0`, `4E24E0`,
`4E2520` and `4E2560` now retain three native object pointers at their existing
public C entries. Signed 32-bit damage/type/result transport, full DWORD type
comparisons, wrapping MechGolem doubling and conditional parent calls are
preserved. Their old C bodies remain exact under `#if 0`; ordinary bool
registrations keep their existing behavior.

The scalar callers have complete original 196-byte/75-instruction evidence and
430 independently verified original boundary executions. Normal/strict Go 1.27.3
each pass 871 tests and two packages, including actual high-address C consumers.
Full parent normal/strict profiles each execute 2,076 Go and 2,076 C rows:
2,074 pass and two cross-page WORD rows fail per backend. All 8,348 test IDs run;
8,342 pass and six fail, counting the four failed leaves and two parent tests.
The original `02` versus actual native `AA` post-fault byte remains unqualified.
Independent ordered observations retain full pointer identities and caller
scratch; this difference is not evidence of a truncated pointer.

Fresh normal/HD products pass ten unchanged shop/hearing scenarios from the
same dirty source epoch. Canonical products, executed mains, failures, original
assets/Save and 56 frame paths remain preserved after byte-verified duplicate
cleanup. Physical audio and direct raw-entry gameplay coverage retain their
separate limits. Earlier inventory counts below describe their recorded source
epochs; this checkpoint establishes no new global count or campaign completion.

## Earlier checkpoints

`sub_409B50` now uses full native pointers and a Go implementation behind its
unchanged public const-source C signature. Its unsigned count remains 32-bit;
it is a value, not a pointer. The complete 41-byte original scan and ordered
DWORD/BYTE copies qualify fifty original calls across two FPU controls, with
forty returns and ten original-only fault observations. Local macOS ARM64 Go
1.27.3 normal/strict integration checks each retain all thirty-one
setter/getter/static IDs within forty-one passing tests and two package
passes. Forty Go/C comparisons and eight exact Go allocation-end NUL checks
preserve overlap order, high pointer bits and the shared getter alias. The
original C transcription and all executed mains remain preserved. The prior
setter already accepted a native pointer, so no integer candidate is removed;
native fault, DF=1, four-GiB scan and whole GUI/campaign results are not claimed.

`sub_409B80` in `GAME1.c` now exports its original `char*` entry from Go and
returns the same C-owned shared map-name buffer at full native width. Its
old C return already preserved pointers; no integer candidate is removed.
The complete six-byte original body is sealed separately and sixteen actual
original calls qualify across two FPU controls. Local Go 1.27.3 normal and
strict macOS ARM64 checks each pass nine tests and two packages, including
six actual C-consumer/string cases above 4 GiB. Setter visibility, stable
aliases, first-NUL/non-UTF-8 copies and earlier Go-copy independence pass.
The old C statement, actual test mains and raw evidence remain preserved;
buffer capacity, whole GUI routes and campaign completion are not inferred.

Raw DefaultDamage now has an independently verified original reference before
its pointer ABI changes. The 1,034-case primary set and four separately frozen
Shield inputs execute 2,076 parent entries under two FPU modes. Their union
completes all 592 parent instructions; 1,902 returns and 174 faults are retained
as observations. Full memory, original bytes, argument transport, caller scratch,
register/FPU records, fault extents and protection conditions pass the separate
streaming verifier. Actual predicate112/34 and converter17 bodies execute when
reached; the other 25 direct targets/three indirect sites remain declared
external services. The main 3,373/713 manifest is unchanged; the additive 14/3
manifest passes the actual local Go 1.27.3 byte verifier.

Failed recorder/wrapper/metadata checks remain alongside corrected tools.
The original cross-page WORD-store observation contains two partial-fragment
records. An isolated native ARM64 STRH control instead leaves the mapped byte
unchanged on SIGBUS. Both observations remain exact; physical fault-state
equivalence, Go-engine behavior and actual Windows/x86 atomicity are unqualified.
The raw parent still takes integer object arguments, its C call edges remain,
and ordinary Go aliases are unchanged. No integer candidate is removed: counts
stay 266 integer/41 typed candidates with ten overlaps. Whole consumers, the
parent port and nine final campaign audits remain open.

Pre-ABI controls additionally pass normal/strict local Go 1.27.3: four tests
per profile, no skips/failures/warnings and four preserved PREEXEC mains. Healthy
high C memory, sixteen C/Go layout facts, real Buff queries and nil-target return
precede the actual old-C fault. Its earliest observed access is the typed Buffs
callee using the truncated target plus344, before the parent class read. Source
and weapon losses have transport-echo evidence only. No native parent engine
is credited by those separate pre-ABI controls.

The subsequent complete 1,038-case initial-memory fixture now qualifies in
local normal and strict highres/cgocheck2/checkptr2 Go 1.27.3 runs. Each has
1,044 raw test passes, zero skips/failures/warnings and two package passes
counted separately. Per mode, 158,814 Go and 158,814 standalone C reads include
1,342 protected accesses; 157,472 readable C loads and 2,076 actual five-word
echoes also pass. Independent reconstruction checks exact test identities,
four PREEXEC mains across both modes, actual ARM64 load opcodes/fault PCs,
full-width pointer relocation, complete Go AFTER pages and PE region guards.
C whole-page memcmp certificates are retained, with no independent healthy
C AFTER snapshots. Twenty-three offsets have actual C named-field or explicit
fixture-view evidence; four Monster offsets remain Go-derived adapters.

The prepared bounded runner additionally passes a fresh focused normal run:
1,044 tests, two byte-identical qualified Go mains and a distinct automatically
built worker with actual header bytes preserved. Independent reconstruction
of that worker's actual records also passes. Make and strict-shell preparation
are now connected and independently qualify focused local runs of 1,045 and
1,089 tests, respectively: three PREEXEC mains each, zero skips/failures/warnings,
and exact initial-field/fault reconstruction. The strict route retains both
distinct `noxmovie.test` mains by package identity; no package scope is pruned.
The normal route exercises the original Make environment defaults. Neither
focused result claims a complete default Make or `./...` run. The fixture Go
sources select only Darwin ARM64. The failed
relative-seal-path invocation and corrected absolute-path run remain, with
all original inputs/expected values unchanged. No parent algorithm executes
in this fixture, and no integer candidate leaves the inventory. Whole
consumers, the raw parent port, physical post-fault equivalence and the nine
final campaign audits remain open. No GitHub CI is awaited or downloaded.

Damage predicates now qualify the complete recorded local five-profile scope
at clean `4303b7607`: normal/highres 809,810 passes each, server variants
803,591 each and strict highres 673,611, with no failures/warnings and 26
actual PREEXEC mains. All 93 inputs and 11,086 original-family leaves remain.
All five actual bfa baseline primary comparisons retain every shared raw ID/
action/multiplicity, adding exactly 4,416 IDs; address rows remain explicit
and unnormalized. Normal/HD products also pass ten fresh original shop/hearing
scenarios. Targeted image evidence covers 56 paths with recorded merchant-name
and actor-visibility limits. Full failure records, canonical products/test
mains, 4,439 assets and 126 Save files remain after byte-verified duplicate
cleanup. The global-marker wrapper and tuple/JSON aggregate errors are
preserved with their separate corrected v2 evidence; gameplay code and
original YAML/assertions did not change for these corrections.

Static counts stay 266 integer/41 typed candidates with ten overlaps; the
broader checks remove no additional entry. Raw `4E0B30` still narrows object
arguments. Its next reference plan distinguishes the existing 1,792-byte
seal from the 1,783-byte body/nine trailing NOPs, and 50 direct call sites
to 28 unique targets from three indirect sites. This is static preparation:
no new parent input freeze, original execution or native implementation is
credited. Nine campaign final audits and overall completion remain open.

Raw damage predicates also pass actual pinned Go-heap integration after
`884557bf4`: local normal/strict HD Go 1.27.3 each pass 4,417 events, no
skips/failures/warnings and four PREEXEC mains. The previous 4,394 events
and all 92 committed inputs remain exact. Four allocations are individually
pinned per graph; five default public-C cases keep observers nil, and sixteen
Go/C cases force 44 GCs with sixteen declared stimuli. Whole allocation
bytes, cached values, full arguments and thirteen C return traces agree.
Original 1,948 engine leaves and C-owned fault domains remain. Static counts
stay 266 integer/41 typed candidates; products, broader profiles, the raw
integer parent and nine final campaign audits remain open.

Raw `4E1400`/`4E1470` now use typed object-pointer C exports after consumer
freeze `c03b34f56`. Local macOS ARM64 Go 1.27.3 normal/strict HD each pass
4,394 events with zero skips/failures/warnings and four PREEXEC mains.
The 1,948 actual Go/public-C original comparisons preserve lazy BYTE/DWORD
reads, full UseData pointers, cached values and 180 fault outcomes. Old C
bodies511/231 remain exact under `#if 0`; ordinary guarded aliases remain.
The raw DefaultDamage parent still takes integer pointers: three call-site
casts retain existing DWORD bits and do not restore lost native bits.

The new phase protects 89 immutable inputs and admits only reviewed shared
C/header changes plus the native Go file, totaling 92 inputs. Readiness449/
nativeGo1216/typed declaration/four ABI32 fallback anchors stay exact. The
host scan now reports 191 C files, 477 headers, 3,457 definitions, 266 integer
candidates, 41 typed candidates, ten overlaps and no warnings. Only these
two entries leave the previous integer set. Actual Go-heap/GC, current
products, broader profiles, the raw parent and nine final audits remain open.

Raw damage predicates now also pass whole consumers after `2ed97ebe7`:
local normal/strict HD Go 1.27.3 each pass 2,449 events, zero skips/failures/
warnings and four actual PREEXEC mains. All 91 phase inputs remain exact.
The added 1,954 events cover 487 actual C argument transports, 487 typed
runtime cases and 974 complete reference-guided transports, retaining the
preceding 495 IDs and six actual old-C fault children. Both raw integer
ABIs/bodies remain and native engine leaves are zero. No additional integer
candidate, product or campaign audit is credited.

Raw damage predicates `4E1400`/`4E1470` now pass independent initial-memory
checks before implementation: local macOS ARM64 Go 1.27.3 normal/strict HD
each pass 495 events with zero skips/failures/warnings and four PREEXEC
mains. All 487 heaps, 2,922 typed probes, 1,222 unreadable domains, 15 layout
facts and 89 frozen phase inputs qualify. Three actual old-C fault children
per mode follow healthy controls, covering source, weapon UseData and friendly
weapon pointer loss. A metadata-only GOENV query failure is retained; it
started no build or test main. The raw integer ABI/body still remain and no
native engine or complete consumer is present. No static candidate, product
or campaign audit is credited. See the latest completion checkpoint.

Readiness now qualifies the complete recorded local five-profile scope at
clean `bfa2fea41`: normal/highres 805,394 passes each, server variants 799,175
each and strict highres 669,195, all with zero failures/warnings and 26 actual
PREEXEC mains. Each retains 9,138 original Go/C leaves and all 76 phase inputs.
All five actual e92/bfa primary comparisons preserve every shared old test
ID/action/multiplicity and add exactly 1,038 raw Readiness IDs; address-only
differences remain explicit and unnormalized. Normal/HD products also pass
ten original shop/hearing scenarios. Targeted images and byte-preserving
cleanup are complete, with generic merchant-name and hearing actor-visibility
limitations recorded. Static counts remain 268 integer/41 typed candidates;
no additional entry is removed by these broader checks. Nine final chapter
audits and overall completion remain open. See the latest completion evidence.

Raw Readiness also passes actual pinned Go-heap integration after `7b8e947f3`.
Normal/strict highres each pass 1,040 events, preserving the preceding 1,029
and adding eleven. Six Go objects per graph are individually pinned; three
default public-C cases use nil observers, and six Go/C cases force 36 GCs
during reads with eight declared stimuli per mode. Full object bytes and
actual C argument/signed-return traces agree. Both modes enable cgo checks;
all four PREEXEC mains exit unforced 0, without skips, failures or warnings.
All 75 preceding inputs remain exact. Products and broader local profiles
remain pending; this test addition removes no further pointer candidate.

Raw Readiness is now a native object-pointer Go export after consumer freeze
`93f175870`. Local Go 1.27.3 normal/strict highres each pass 1,029 events,
including 452 actual Go/public-C original comparisons. All 70 frozen inputs
and the original 449-byte C transcription remain exact. The implementation
retains the entry table read, full callback identity and matching-slot reload;
the ordinary guarded alias remains unchanged. Four ABI32 fallback calls now
pass their original 32-bit address bits explicitly to the pointer signature.
The canonical host scan reports 191 C files, 476 headers, 3,459 definitions,
268 integer candidates, 41 typed candidates, ten overlaps and zero warnings.
Only Readiness leaves the preceding integer-candidate set. Actual Go-heap GC,
current products and broader local profiles remain separate pending gates.

Raw Readiness whole consumers now pass local macOS Go 1.27.3 normal and
strict highres: 577 events each, no skips, failures or build warnings, and
four preserved actual mains. All 71 sealed inputs remain exact. The checks
include 113 C argument transports, 113 typed-runtime cases, 226 complete
reference-guided transports and the existing native callback collision and
old-C fault controls. The raw integer ABI/body and absent engine remain;
no integer candidate is removed. See the Readiness projection document in
`src/legacy/testdata` and the latest completion checkpoint.

Independent Readiness initial memory now passes local normal/strict HD
Go1.27.3:119events each,113whole heaps,1808typed probes,52unreadable domains,
15C/Go facts and actual old-C high-pointer fault controls after healthy data
checks. Four PREEXEC mains/parents exit0. Three original NONMATCH callback
rows additionally prove equal low32/different full native pointers, including
a deliberate bad uint32 comparison. Original input/result/initial gzip and
prior56/reference8 remain exact; v1 actual success is retained separately.
The raw integer ABI and body still remain; whole consumers/engine/products
are pending. No new integer candidate or chapter is credited.

All five current local macOS Go1.27.3 profiles at clean `e92c637c0` now pass:
normal/highres804356PASS46skip each, server/server-highres798137PASS7skip
each, strict-highres668157PASS10skip. Failures/build warnings0;26actual
PREEXEC mains/full logs archived and verified before owned duplicate cleanup.
Each profile retains8686fixed original-family leaf IDs(3838Go/4848C), including
792raw Regeneration leaves, plus ten default C calls/twelve in-call GCs.
All56frozen/fullsource inputs remain exact. The earlier unexecuted plan's
1010phantom Go leaves are corrected against the two C-only test roots;
test command scope and all earlier C assertions stay unchanged.

Readiness reference `e92c637c0` separately freezes68original bytes/113inputs/
twoFPU proofs. Actual CLI validates the existing3373code/713data manifest
and new dedicated3code/0data manifest without changing the old56. Its raw C
integer ABI, independent native initial/consumer/engine/product gates remain
open. No additional pointer candidate is removed. Current normal/HD raw test
names differ only in three actual-address names per run; shared actions/counts
agree, with the six raw names retained. No historical deleted-primary replay
or new CI credit is claimed. Nine chapter final audits and overall100 remain
open. See the latest `porting-completion.md` checkpoints for exact scopes.

Clean `07e7913a3` normal/highres-tag Go1.27.3 ARM64 products now pass24actual
raw Regeneration public C calls,6positive/18negative, unforced native exits0.
Actual stock Sword/LeatherArmor use the human Warrior and WizardRobe an owned
NPC with ordinary class admission; descriptors/clocks/default services stay
unchanged and observers nil. Both complete health guards and restoration,
native allocations above4GiB and real Sword mouse equipment are checked.
Shared-boundary robe positives do not add differential-period proof beyond
the focused actual-default tests. Both actual frames are viewed and archived;
only verified owned duplicates are removed. Existing56frozen inputs,4439asset
hashes and126Save files remain exact. Other entries, full current local
profiles and nine chapter final audits remain open; overall100 remains false.

At clean `3cb7514e9`, the next product setup reaches the original armor
class gate: Warrior cannot pick up WizardRobe. Its complete failure is kept.
The diagnostic now uses the human Warrior for Sword/LeatherArmor and an
actual stock owned NPC for WizardRobe, with ordinary inventory admission,
the real per-holder HP maximum/interval and both complete health guards.
No class/flag/descriptor/clock or production engine changes are made. Focused
normal/strict HD each pass1641events with eight PREEXEC mains, zero warnings
and all56frozen inputs exact. A human Wizard robe, rebuilt products and full
current local profiles remain open.

Both clean `8a7d74829` normal/HD Go1.27.3 products build and start. The first
normal product E2E equips a stock Regeneration1 Sword, then stops before its
first raw call because the new diagnostic budget1200 is below the unchanged
stock1440-frame interval. Its failure/product/runtime are retained. Only the
diagnostic wait budget increases to3600; production behavior and frozen56
inputs stay exact. Actual two-product E2E and full current local profiles
remain pending; no pointer candidate or chapter is newly credited.

Actual default Regeneration services now pass eight prepared public-C
conditions and two pinned Go graph calls with twelve in-HP-read GCs per local
profile. Normal/strict HD Go1.27.3 each pass 1,641 events, including all792
original leaves and the existing registered schedule regression. Eight actual
PREEXEC mains are retained; all56 frozen inputs remain exact. The compiled
product diagnostic remains unexecuted. Full local profiles/products and the
nine chapter final audits remain open; no additional candidate is removed.

Raw Regeneration is now a native two-pointer void Go export, after consumer
freeze `88625f307`. Exact old C bytes remain under `#if 0` and all 56 frozen
inputs stay exact. Local normal/strict HD Go 1.27.3 each pass 1,631 events,
including 792 actual Go/C complete original leaves and six PREEXEC mains
across both modes. Actual default services, GC and product gates remain.

The remeasured canonical host-normal inventory is **191 C/475 headers,
3,460 active definitions, 269 integer candidates, 41 typed candidates and ten
overlaps, zero warnings**. Prior source reproduces 3,461/270/41/10 over
191 C/474 headers. The sole removed integer identity is raw Regeneration;
all other integer/typed identities and all 474 old headers stay exact. One
dedicated Regeneration header is added. Cgo preambles and other target macros
remain outside this view. Counts do not supply a completion percentage.
See the latest local checkpoint in `porting-completion.md` for evidence.

Complete Regeneration consumers subsequently pass local Go1.27.3 normal/strict
HD,839 events each:396 whole original transports,198 frozen-runtime cases,
actual existing HP helpers and complete read/call/memory/DIV fault prefixes.
Six PREEXEC mains are retained. The raw engine remains absent and its original
consumer has zero engine leaves; no candidate is removed. The latest user
instruction continues local macOS tests without waiting for GitHub CI.
See `src/legacy/testdata/regeneration_raw_native_projection_4e01d0.md`.

Current preparation, 2026-11-06: raw Regeneration `4E01D0` retains its active
two-integer C ABI. Original reference `ccceee5ba` covers 257 complete bytes
and 198 independent inputs. Native initial controls pass Go 1.27.3 normal and
strict HD, 205 events each: 198 whole heaps, 36 read-fault domains, 11 C/Go layout
facts and actual unchanged-C high-pointer faults after positive controls.
Four PREEXEC mains and all preparation/cache failures are preserved.
Complete consumers and raw implementation remain pending; no integer candidate
is removed. See `testdata/regeneration_raw_initial_memory_4e01d0.md` under
`src/legacy`. The active branch is `port/go1.27-multiarch`; recovery gaps and
current evidence are recorded in the latest `porting-completion.md` checkpoint.

This inventory prioritizes remaining public C ABIs after the native point,
occupancy, retreat and ROAM lifecycle ports. It does not supply a porting
denominator or count confirmed runtime crashes.

Latest direct-product checkpoint, 2026-11-06: clean source `d009345015`
passes20conditions and80actual public-C calls in normal/HD Go 1.27.3 ARM64
products. Actual stock item/inventory construction, mouse equipment messages
and default owner-chain services use live addresses above4GiB. Original
`MagicReflection1`/`Grip1` descriptors are disabled placeholders: two negative
conditions per product preserve that state; eight others configure only
private complete copies with real C callbacks and diagnostic strengths/slots.
No natural stock active-effect or earned campaign credit is claimed. Health,
original descriptors and all38frozen inputs remain unchanged. Focused normal/
strict HD each pass5,381events, including1,010complete actual-C original
leaves and all prior identities/actions/multiplicities. Eight PREEXEC mains,
both earlier diagnostic failures and all products are retained; verified raw
duplicates are removed. The previous full5/CI6/hearing runs retain their
`5eee32f9b` execution scope;8,151of8,152old source inputs are exact, with only
the existing E2E parser gaining the diagnostic action. See
`ART/four-modifier-native-direct-product-go1273-v3/current-code-four-raw-entry-direct-product-final-qualified-v1.json`.
Other raw pointer entries and all nine chapter final audits remain open.

Latest full validation checkpoint, 2026-11-06: clean source `5eee32f9b`
passes all five recorded macOS Go 1.27.3 profiles and all six current native
Windows/Linux CI profiles. Each profile executes1,010actual-public-C whole
original comparisons and20in-callback forced GCs. All prior primary test
identities/actions/multiplicities remain; each unrestricted profile adds
3,469passing four-entry events. All26local PREEXEC mains, six CI products
and161CI manifest mains are retained and compiler/hash/architecture checked
(Windows6PREEXEC, Linux155existing POSTEXEC). Whole CI ZIPs match API digests.
Actual clean normal/HD products also pass20NPC/Troll hearing cases, including
eight physical sound-directed turns and12negative controls. This verifies
prepared hearing conditions, with no hearing production defect found there;
it does not exercise the four raw modifier entries directly in a live product.
The later checkpoint above adds direct four-entry product/default-gameobject
diagnostics. Other raw pointer entries and nine chapter final audits remain
open. No overall100 credit.
See `porting-completion.md` and
`ART/four-modifier-native-ci-go1273/six-native-four-modifier-Go1273-current-code-final-qualified-v1.json`.

Four-entry engine checkpoint, 2026-11-06: after consumer freeze
`8b6368bdb`, Inversion `4E03D0`, Grip `4E0480`, Grip traversal `4E03F0`
and inventory checker `4FA4F0` now expose native-pointer C signatures and
actual Go exports. The exact four old C bodies remain under `#if 0`.
All 38 sealed reference/initial/getter/consumer/helper inputs stay exact.
Go 1.27.3 normal/strict HD each pass **3,492 events**, zero failures/skips/
warnings: **1,010 actual-public-C complete original comparisons** per mode
(596 predicate/traversal and414checker), including62original fault leaves,
whole native/PE32 heaps, ordered accesses, real parent/callback arguments,
returns and scratch. Five pinned Go-object graphs additionally exercise all
four C entries with20in-callback forced GCs per mode and guarded scratch.
All six actual mains are retained before execution; parents/drivers exit
unforced zero. The predicates and the existing ordinary late-defend path
now use the original signed strength comparison; two earlier incorrect
unsigned-high expectations are corrected against sealed original setge/setl.
Full current-source local profiles, products/E2Es and six native CI profiles
remain pending. This checkpoint supplies no overall completion percentage.
Evidence: `ART/modifier-inversion-inventory-4fa4f0-native-entry-port/four-entry-native-normal-and-strict-qualified-v1.json`.

The first full normal gate then has802,752PASS/46skip and one failure: the
older real `4E1320` C-entry integration test still expects unsigned behavior
for `MinInt32`. Its two expected values are corrected to original Inversion0/
Grip1; production engines and all38frozen inputs stay exact. All six actual
mains and complete parent/child failure output are retained before failed
WORK cleanup. The five full profiles must rerun on this corrected test source.

Preparation notes below describe the earlier source states at their freezes.

An additive pre-engine `4FA4F0` reference now executes the complete original
checker159, parent38 and Inversion22 on 207 inputs per FPU profile, with no
helper substitution. All219 bytes, 2,269 external accesses,267parent/263effect
calls and thirteen faults are observed per profile. Complete heap/stack/
caller/register/write/return/fault outputs reproduce byte-exactly. The raw
two-integer C ABI remains active. Independent native initial/default-parent/
fault controls and complete consumers must precede this coupled four-entry
engine. Earlier freezes and all3373code/713data seals remain unchanged; no
integer candidate is removed. See `inversion_inventory_original_4fa4f0.md`.

Independent native checker memory then passes 427 normal/strict HD Go 1.27.3
events each: 207 whole initial heaps,242selected nil/protected read domains,
eighteen C/Go layout comparisons and267actual default C parent calls with
four original-matching helper faults. Each profile also proves the actual
old raw checker fails above4GiB after valid native/parent/whole-heap controls.
Six PREEXEC parent mains and all expected child-fault logs are retained.
Old C bodies/prototypes and the helper/earlier freezes remain exact; complete
checker consumers and all four raw engines remain pending.

The complete checker consumer preparation subsequently passes **2,464 events**
in each Go 1.27.3 normal/strict HD profile, with zero skips/failures/warnings.
All 414 full native transports and 2,269 original accesses per FPU profile
agree, including complete native/PE32 final heaps and fault prefixes. The
actual parent helper's field reads are now optionally observed, with its
traversal algorithm unchanged; all 267 calls/four faults and complete read
sequences pass again. Whole stack/write/caller/return certificates and the
actual C call transport consumer are compiled before any of the four raw
engines. Six PREEXEC mains per preparation version are retained. All original
and earlier freezes remain exact; C bodies/prototypes remain old. See
`inversion_inventory_native_projection_4fa4f0.md`. This is preparation,
without native-engine or product/CI credit.

The actual raw Marker `54E460`, Boulder `54E4B0` and MonsterGenerator
`54E630` entries now have native object arguments, following native initial
memory `4a66bcefa` and whole-consumer `31e0dfea7` freezes. Exact old bodies
remain under `#if 0`; original references/getters/consumers stay unchanged.
Go 1.27.3 normal/strict HD selections each pass 8,071 events, including
800 complete original Go/C leaves, nine actual default-C conditions and two
forced-GC callback cases; eight actual mains are retained before execution.
The initial integration failure is a malformed unowned-Marker expectation:
real owner lookup returns terminal self and its nil update faults originally.
Only the new test is corrected; engine/headers/frozen consumers stay exact.
At exact clean source `239638703`, normal/HD product E2Es, all five recorded
local profiles and all six native CI profiles now pass. The 26 local mains
are retained before execution; every prior test identity/action/multiplicity
is retained in each profile, with 1,426 new passing events. Full archives and
compiler metadata qualify all six CI products and 161 manifest mains
(Windows six PREEXEC, Linux existing 155 POSTEXEC). Exact scope, event counts,
prepared diagnostic limits and artifact references are in the latest
`porting-completion.md` checkpoint. Remaining raw entries and nine chapter
final audits remain open.

The next reference-only freeze `26086de87` covers Inversion `4E03D0`, Grip
traversal `4E03F0` and Grip `4E0480`: 183 original bytes, 298 inputs per FPU
profile and 18 intentional original memory faults each. All original
instruction bytes execute, including actual nested Grip calls, with complete
byte-identical output reproduction. The predicates use signed comparisons;
their decompiled C still uses unsigned fields and integer effect pointers.
The original C bodies remain active pending independent native initial
memory, actual unchanged-C fault controls and a full consumer freeze. No
integer candidate is removed by this preparation. The seal inventory is
3,373 code / 713 data ranges; the previous code/data entries stay exact.
Earlier preparation notes below describe the source states at their freezes.

Independent initial memory then passes normal/strict HD Go 1.27.3, 306 events
each: all 298 full initial heaps, thirteen C/Go layout facts, 136 selected
nil/protected read-fault domains and three actual old-C high-pointer faults
per profile. The old C bodies and integer prototypes remain active; only the
previously missing four-integer traversal prototype is declared. Four parent
mains are retained before execution, with complete expected child-fault logs.
Original goldens/producers/inputs stay exact. Full consumers must freeze next;
no native implementation or candidate-removal credit is added.

The full consumer preparation subsequently passes 1,205 events in each normal/
strict HD Go 1.27.3 profile. Each replays 596 complete native-memory transports,
including all 1,411 original access observations per FPU profile, ordered
read/write/fault prefixes and full final heaps. Frozen typed getters/stores,
actual six-address C callback transport and compiled consumer machine code
precede the raw engines; native-engine leaves remain zero. The old integer
C bodies and all original/initial references stay exact. Read-only dependency
analysis adds the still-active raw two-integer `4FA4F0` checker to this coupled
family; its ordinary Go collision path does not close the public C entry.
Its separate original/initial/full-consumer freeze is the next prerequisite.

The repeated 191-source/474-header host-normal scan finds **3,465 active
definitions / 274 integer candidates / 41 typed narrowing candidates / ten
overlaps / zero warnings**. Only the three entries leave that inventory.
Eight other integer arguments receive semantic notes without closing whole
functions/returns or lowering the candidate count. Cgo preambles and other
target macros remain outside this triage view. See `porting-completion.md`
and `ART/marker-boulder-generator-native-entry-port/three-entry-native-Go1273-normal-and-strict-final-qualified-v1.json`.

Complete three-entry original-output consumers are subsequently frozen with
all three C ABIs still integers and no raw native method in the repository.
Normal/strict HD Go 1.27.3 selections each pass 2,246 events and 400 initial
transport leaves, with six pre-execution mains retained. Independent original
execution adds 3,743 memory observations per FPU profile while reproducing
both old complete output bytes exactly. Whole heap/write-prefix/world/fault
checks and the explicitly compositional image projection are documented in
the new native projection reference. Absent-engine branches receive no native
leaf credit; the actual three ports remain the next gate.

The three remaining raw Marker/Boulder/MonsterGenerator entries now have
an independent 200-input native initial-memory fixture, frozen before
implementation. Go 1.27.3 normal/strict HD selections each pass 1,835 events
and 184 actual protected/null read-fault domains, with six actual mains
retained before execution. Whole 64-KiB native guards and PE32 projection
stay exact; observed C/Go fields agree. Unobserved update-data bytes are
guarded, without an arbitrary cross-structure alias claim. Their C ABIs
remain integers and complete consumers/engines remain pending. Evidence and
scope are recorded in `porting-completion.md`; no native-port credit is added.

Current clean macOS normal/HD Go 1.27.3 products at `c41351758` also pass
20 actual NPC/Troll hearing cases, including eight native turns toward
sound and twelve negative controls. Both processes exit zero unforced;
inputs remain exact and all products/runtime/frames are compactly retained.
These prepared diagnostics change no production hearing/C signatures and
supply no direct raw-death-entry or earned campaign credit. The previous
declared properties and mock-client-output limits remain; see
`porting-completion.md` for the source-bound current product evidence.

Clean `0894002e4` passes five local Go 1.27.3 configurations in the recorded
six-client/five-server-package scope, with 26 actual mains retained before
execution. Windows `42058922871` and Linux `42058956843` then pass all six
native configurations. Every profile preserves all 1,144 credit Go/C leaves,
five actual credit C conditions and the earlier Skull/Polyp/Skeleton/Golem
contracts. Whole ZIPs/products/API digests/raw jobs are retained. All six
products and 161 retained manifest executables identify Go 1.27.3; Windows
retention is pre-execution, Linux's existing 155-manifest collector is
post-execution. Counts, exact scopes and qualifiers are in
`porting-completion.md`. These are source-bound passing gates, not an
overall porting percentage or completion of the remaining raw entries.

Reference-only `4cd52e87d` freezes the other three entries' 200-case complete
initial PE32 heap digests and read/fault domains. A separate actual original
emulator setup check verifies every case without executing a raw body.
Native memory relocation/guards and complete frozen consumers remain
required before implementing those entries. Existing Go/C/header code is
unchanged by this data/producer/documentation addition.

Go 1.27.3 checkpoint `74813f67c` passes both Windows architectures, with
complete products and six pre-execution mains retained and compiler-verified.
All four Linux variants fail only the six particle compressed-PNG hashes
and their parent. Test-only `53f9fef33`, following the independent original
hash/pixel freeze at `4f1481e9e`, preserves every original MD5 label and
compares all original pixels instead of version-dependent compression bytes.
Full stock-asset renderer normal/strict selections each pass 438 events,
including 276 goldens and a native pixel-change rejection control. Production
renderer and C signatures are unchanged; these related passes do not supply
fresh full-suite/native-CI or overall completion credit.

At `dd1632f88`, raw credit `4D61B0` now accepts `nox_object_t*`; its old C
body remains exact under `#if 0`. Go 1.27.3 normal/strict HD selections each
pass 1,633 events, including 1,144 complete original Go/C leaves and five
actual C-constructor conditions, with complete native memory guards. It
preserves the byte flag gate, cached update and player reload after counter
increment. The first failed alias comparison and an independent no-engine
strict reproduction identify a missing fifth initial flags byte in the
fixture. Fixture-only `3d4546be8` corrects that transport; original outputs,
the frozen consumer and engine stay unchanged. The no-engine green gate
passes 299 events before the native comparisons pass. See the full scope
and retained failure/correction evidence in `porting-completion.md`.

The fresh host-normal CPP scan at that engine covers 191 C sources and 474
headers: **3,468 active definitions / 277 integer candidates / 41 typed-pointer
narrowing candidates / ten overlaps / zero warnings**. Cgo preambles and
other target macros remain outside this triage view. Marker, Boulder and
MonsterGenerator raw entries still need their native fixtures/consumers and
ports; their reference gates below do not close those ABIs.

The active toolchain is **Go 1.27.3** at `9b7a7fa2d`. Original-only
`18bb167e4` freezes all 679 body bytes for Marker `54E460`, Boulder `54E4B0`,
MonsterGenerator `54E630` and credit `4D61B0`, before native implementation.
`90d9e324b` verifies independent complete layouts, eight actual C constructors,
whole-record guards and real typed owner/audio controls, then records all
four unchanged old C entries faulting above 4 GiB. Credit's independent
286-case initial memory and complete consumers are frozen at `9bdcf473c`
and `b2e8af7fd`; Go 1.27.3 normal/strict gates pass 304/299 events, with
zero native implementation/comparison credit at that gate. The first three entries
still require native memory fixtures and complete consumers before engine
changes. Those preparation gates alone close none of the four raw ABIs.

Go 1.27.3 normal/HD products at `e1113a42b` verify 20 actual NPC/Troll hearing
cases, including eight native turns toward real server sound events before
visual acquisition and twelve negative controls. Related normal/strict
selections each pass 16,845 events. Production hearing code and C signatures
are unchanged. Declared diagnostic properties, mock client output and
retained failures/products/mains are detailed in `porting-completion.md`.
The latest complete six native CI results at `772da88ee` still use Go 1.26.9;
fresh full Go 1.27.3 qualification remains open.

At `6f12e2168`, the native shop sale binding restores original sound307
instead of791. Normal/strict related selections each pass192 events, and
normal/HD grouped/single and equipped-sale UI replays verify real OpenAL
source/buffer submission. This Go binding correction changes no C pointer
signatures. `1439d7502` independently fixes mutable access-list isolation in
the E2E harness. Exact products at `836d9c7c2` also pass both five-entry Skull
stock-map diagnostics. See the current evidence and retention correction in
`porting-completion.md`; historical ROOT retention claims predate user cleanup.

The host-normal CPP view after the five Skull entries verifies 191 C sources
and 474 headers: **3,469 active definitions / 278 integer candidates / 41
typed-pointer narrowing candidates / ten overlaps / zero warnings**. It is
bound to `4ecdd9ed8` C/header bytes, unchanged by the shop fix. The fresh
full-five local gates and all six retained Windows/Linux artifact ZIPs pass.
Scope, actual-main preservation differences and remaining historical evidence
gaps are stated in `porting-completion.md`. These counts remain triage views,
not a percentage or a confirmed-defect denominator. Evidence is under
`/Users/ghost/codex-sandbox/artifacts/opennox-port-20261105/skull-update-native-entry-port`.

Shop sale rejections at `766134ea8` resolve the original `Trade.c` strings
and reuse the typed native `004D9EB0` line-message service. English/Korean
tests decode real queued packets, and both clean-product shop UI replays
verify the displayed quest rejection and unchanged item/gold. This Go adapter
correction leaves original C bodies and the C pointer inventory unchanged.

Fireball Staff shop repair is restored at `487729563` in the existing native
`005108D0`/`00510AE0` transaction. Wand charge fields use `WandUseData`, recharge
uses the already ported `0053C520` service, and the typed charge report binding
updates the actual client inventory. The original C bodies and the host C
inventory are unchanged. Normal/strict selections and both clean-product
shop UI replays pass; see the Fireball Staff finding in `porting-completion.md`.

MechGolem `549E90` and Golem `549FA0` now have typed native object C
arguments at `5aaefaaa5`, with both original C bodies exact under `#if 0`.
Original seals `ba9f9ddff` and `41b5f6083` precede the engine: all **691**
instruction bytes, **164** inputs and both FPU profiles execute, including
22 reentrant calls and ten memory/two divide errors per profile. Both old
actual C entries SIGSEGV above 4 GiB after seven real C constructors,
independent complete layouts and whole-record controls pass. Normal/strict
selections each pass **11,409** events, including **656** complete original
Go/C leaves and six actual default-C conditions. Forced stack/GC regressions
pass **450** events under GOGC10/cgocheck2/checkptr2/count50.

The raw core preserves the live packed name tables, unsigned mass bytes,
shared extracted C cache, mode/FPS reloads, signed modulo, allocation/fault
order, FPU spills and reentry. It reuses a GC-safe native source reference;
ordinary dispatcher admission and index normalization remain separate.
Fixture-only `1b5a6afb9` corrects mapped/unmapped address transport after an
independent no-engine strict control proves that a relocated invalid mass
or slot address can land in another native mapping. All prior assertions,
expected data and producers stay unchanged; the other nine original files
remain exact. Controlled out-of-table projections supply no general native
named-field alias certificate. Full local/native CI/product evidence and
all failed/preparation files are retained in `porting-completion.md` and
`ROOT/golem-death-native-entry-port`.

Read-only follow-up identifies raw ImpEgg `54CAE0`, Polyp `54CB10` and
Potion `54CBB0` bodies totaling **214 original bytes**. Polyp already has
one complete 150-byte precision seal; it is not counted as a new execution
or new raw/native qualification. Their ordinary native callbacks do not
close these raw C ABIs. Byte/call provenance is retained in
`ROOT/next-object-death-original-PE-byte-and-call-audit-20261104-v1.json`.

Reference commit `fd15e447c` independently executes all 214 body bytes and
the existing 17-byte FISTP helper for 149 inputs in each PC53 to-zero/nearest
profile. Each profile includes 31 original reentrant calls, 110 actual FISTP
calls and 15 memory faults, with full heap/image/caller guards. Seventy-three
complete outputs differ by profile. Only ImpEgg's 33 and Potion's 31 bytes
lacked prior seals; the Polyp/converter bytes receive no new body credit.
Inputs, producers, outputs and proofs reproduce before native engine changes.
All three unchanged actual C entries fault above 4 GiB after independent
complete C/Go layouts, a real ToxicCloud C constructor/native update record,
whole-record checks and all three real typed C audio controls pass. Two
preparation failures remain preserved: a missing declaration header, then
missing audio initialization in the minimal fixture. No old-C fault credit
is assigned to those failed preparations.
Evidence:
`ROOT/simple-polyp-death-native-entry-port/complete-original-reference-reproduced-before-native-v1.json`.
The separate before-engine gate is retained in
`ROOT/simple-polyp-death-native-entry-port/actual-old-C-causality-before-native-qualified-v1.json`.

At `7b71b8889`, the three public entries accept native `nox_object_t*` arguments, with
the exact old C bodies retained under `#if 0`. Polyp keeps the live numeric
type cache, Y-before-X fault order, entry-cached update pointer, signed FPS,
PC53 multiplication/binary32 spill/FISTP, and audio-before-delete order.
Ordinary Go dispatcher admission remains unchanged. Normal and strict HD
`cgocheck2`/`checkptr=2` selections pass **3,334 / 3,335** test events,
including **596** complete original Go/C leaves and seven actual default-C
server conditions. All executed mains and preparation failures are retained.

The first strict run fails in the reference fixture's partial-flags object
alignment. Replaying its exact retained main without the death engine
reproduces that failure. Fixture-only `7329b1c27` aligns the native object and
projects the mapped flags to the unchanged PE32 heap location. An independent
strict memory control then passes eight events. Engine files, original
inputs, producers, goldens and assertions stay unchanged; both complete
selections pass after that correction. This declared fixture correspondence
does not certify arbitrary native named-field aliasing.

Clean normal/HD products at observer-only `db9cdda44` pass the stock-map
direct-C diagnostic for all three entries. Prepared stock RedPotion records
exercise the class-unrestricted raw contract, preserving living host HP150
and poison0. The unowned ToxicCloud has native source/update pointers above
4 GiB and initial original-stock lifetime450; it replicates to the client
and the final frames show its green particles. Real queued audio supplies
three entry events plus one independent server-service control. The first
product attempt's incorrect immediate-audio assertion stays failed and
preserved. The observer correction changes none of the seven engine/test
files or six frozen original files. Both products exit normally, release
ports and retain private runtime data. No earned kill/campaign or physical
audible-fidelity credit is assigned.

Windows runs `41969128461` at engine `7b71b8889` and `41970766870` at product
source `db9cdda44` both pass: amd64 **786,234 / 11 skips**, 386 **766,504 /
94**. Each job preserves all 596 original Go/C leaves and seven actual
default-C conditions. Every full archive, API digest, raw job log, six
pre-execution PE mains and two clean server products per run are retained
and hash/header/source verified. Both exact primary 386 PEs retain the
existing Skeleton/Golem GOGC10 replay without rebuilding. This does not
recover the three missing mains or prove the old cancelled run's cause.

The current host-normal CPP view verifies **191 C sources / 474 headers**:
**3,474 active definitions / 283 integer-cast candidates / 41 typed-pointer
narrowing candidates / ten overlaps / zero warnings**. Three definitions and
only two explicit integer-cast candidates disappear: Potion's former pointer
conversion was implicit. The WIP capture remains labelled with its original
source; all 665 current C/header inputs and commit blobs are hash-bound to
clean `7b71b8889`, along with all 191 retained CPP outputs/logs. Cgo preambles
and other target macros remain outside
the scan; these counts are triage, not remaining-defect or completion totals.

Read-only next-target analysis identifies five remaining raw Skull/ArrowTrap
entries `54F9A0` through `54FC50`, totaling **751** original body bytes.
Reachable CFG boundaries exclude separate RET stubs at `54FCB0`/`54FCC0`;
the first 773-byte boundary overcount and its correction are retained.
The independent 753-input producer subsequently executes all 751 body bytes
and the existing 17-byte converter in both FPU profiles, with 53 original
reentries, 402 original enumeration callback calls, 276 actual FISTP calls
and 26 memory faults per profile. All heap/image/caller/nonvolatile/FPU
guards pass, and 303 complete outputs differ between profiles. Independent
reproduction matches encoded/decoded outputs and proof bytes. The original
reference adds five complete function seals and ten file-backed data records
(104 bytes), preserving every prior record and its order. The unchanged CLI
verifies all **3,367 code / 707 data ranges**. It rejects an unmarked new
executable-data record and a falsely file-backed virtual-zero tail. The
existing 4.0 constant and 2048-byte direction table receive no new data
credit. Virtual-zero and malformed-index fixture scope is explicit in
`src/server/testdata/skull_raw_native_original_54f9a0.md`.
The original reference alone assigns no actual unchanged-C fault credit.
A separate before-engine gate subsequently proves all five actual unchanged C entries SIGSEGV above
4 GiB after 19 independent native layout checks, three real C constructors,
52-byte update/eight-byte collide records, typed C direction/audio and whole
allocation/data guards pass. The two-package selection passes 39 events;
both actual pre-execution mains are retained. The first preparation wrongly
compared the complete C prefix912 to Go's allocation928. Fixture-only size
accounting now asserts the shared prefix and the two native-word Go suffix
independently, retaining every whole-allocation guard. That failed source,
logs and both executed mains remain preserved, with no old-C crash credit.
Before-engine evidence:
`ROOT/skull-update-native-entry-port/actual-old-C-causality-before-native-qualified-v2.json`.
Before native implementation, the independent whole-heap correspondence
passes all 753 inputs with highres/cgocheck2/checkptr=2 (757 test events).
The full immutable-output consumer compiles with the same independent
controls (758 test events), explicitly assigning no native comparison credit
while the five methods are absent. Original-only regeneration reproduces
both transport fixtures exactly. Both executed mains are retained before
startup. Proofs: `ROOT/skull-update-native-entry-port/independent-strict-memory-before-native-qualified-v1.json`
and `ROOT/skull-update-native-entry-port/frozen-complete-consumer-before-native-qualified-v1.json`.
The five public C entries now accept native `nox_object_t*` arguments.
All five exact old C bodies and intermediate helpers remain under `#if 0`.
The raw core retains all six live numeric dwords, cached update/collide
pointers, signed direction indexing, PC53 arithmetic/spills/FISTP, exact
fault and callback order, and shared-marker reentry. Ordinary Go dispatcher
admission and direction normalization are unchanged. Normal and strict HD
`cgocheck2`/`checkptr=2` selections each pass **7,147** test events in four
packages, including **3,012** complete immutable original Go/C leaves and
**12** actual default-C server conditions. All earlier selected terminal
identities/actions/multiplicities remain; normal and strict terminals match.
All eight actual selected mains were retained before execution, and all
13 frozen original/getter/consumer/transport files stay exact. The first
normal wrapper fails only in its post-test collector (expected three mains,
retained four); its successful tests/four mains and failed metadata remain
preserved. A fresh four-main normal gate passes completely. Subsequent full
local, product, CPP and native-CI results are recorded in the current retained
verification section of `porting-completion.md`. Selected proofs:
`ROOT/skull-update-native-entry-port/normal-related-original-and-default-C-qualified-v2.json`
and `ROOT/skull-update-native-entry-port/strict-HD-related-original-and-default-C-qualified-v1.json`.
Full original reproduction evidence:
`ROOT/skull-update-native-entry-port/complete-original-reference-reproduced-before-native-v1.json`.
Evidence: `ROOT/skull-update-native-entry-port/original-PE-byte-and-call-audit-v2.json`.

Before this three-entry port, the Golem host-normal CPP view verifies
**191 C sources / 474 headers**:
**3,477 active definitions / 285 integer-cast candidates / 41 typed-pointer
narrowing candidates / ten overlaps / zero warnings**. Only these two
integer candidates disappear relative to the Skeleton view. The original
working-tree capture at `41b5f6083` is not relabelled; its exact C/header
inputs are rebound to clean `5aaefaaa5`. Cgo preambles and other target
macros remain outside this scan. These are triage views, not a completion
percentage or remaining-defect denominator.

The Skeleton family now ports raw public entries `54A310`, `54A390`,
`54A4C0` and `54A750` to native object pointers. Engine `12af92f96`, completed
by `7c00b5765` and `aa26dfa24`, preserves all four old C bodies under
`#if 0`; ordinary Go dispatch/admission remains separate. Frozen original
execution covers all 1,189 bytes, 169 inputs and both FPU profiles, including
memory/divide faults.
Actual old C entries each fault above 4 GiB after complete native-layout and
real constructor/service positive controls. Selected normal/strict gates
each pass **10,720** events, including **676** complete original Go/C leaves
and all six default C conditions. The exact 386 PE GC10 replay exposes a live nil+offset
Go pointer during callback allocation/stack growth at `a24399707`. `aa26dfa24`
retains a native source reference across callbacks and projects malformed
nil-source addresses only in C. Valid origins stay typed pointers. Three
internal service argument types and fixture transport are corrected without
changing ported public C signatures, method behavior or golden/expected data;
the other eight reference files remain exact. Three forced-stack regressions pass **150** events
under GOGC10/cgocheck2/checkptr2/count50. Full local/product/native CI evidence
and all preserved failures are recorded in `porting-completion.md`.

The preserved host scan, captured at the original WIP and rebound by hashes
to clean `8882981b0`, verifies all **191 C sources and 474 headers**. Those
C/header inputs also remain exact at `aa26dfa24`; its new cgo preamble
helpers are outside this scan. It finds
**3,479 active definitions / 287 integer-cast candidates / 41 typed-pointer
narrowing candidates / ten overlaps / zero preprocessor warnings**. Four old
C definitions disappear, but only two cast candidates disappear: the other
two were forwarding entries. This covers host macOS ARM64 normal C
preprocessing, excluding cgo preambles and other target configurations;
these views are triage, not a remaining-defect denominator. Full original
preprocessed outputs and provenance remain under
`ROOT/skeleton-death-drop-native-entry-port`.

At `616b61d`, host macOS ARM64 normal preprocessing of all **191** top-level
`src/legacy/*.c` files finds **3,502** active function definitions. Real clang
preprocessing excludes the preserved `#if 0` original bodies; all 191
preprocessor logs have zero warnings. Full preprocessed outputs, input
hashes, function/caller inventories and scripts are retained under
`ROOT/monster-retreat-predicates-native-entry-port`.

The heuristic finds **303** definitions casting an integer parameter to a
pointer. **289** have a fixed 32-bit `int`/`int32_t`/`uint32_t` operand, while
**14** use native-width or other integer types. Of the 289, **26** have direct
non-test Go `C.symbol` references and **171** have direct active C callers.
References may store or compare callback addresses rather than invoke them.
Forwarding entries without their own cast, cgo preambles, aliases, pointer
returns, globals and integer fields need separate analysis. Other target
preprocessor configurations also remain outside this host inventory.

Semantic inspection is required before changing a candidate:

| Example | Classification |
| --- | --- |
| `548630` collision target and `548830` door data | Already use `uintptr_t`; these are not fixed 32-bit argument truncations. |
| `4597E0` server-options hide argument | Integer control flag, also returned through a legacy pointer-shaped result; a cast alone does not establish a pointer input. |
| `56F9E0` protected-mana argument | Integer protection handle resolved by `56F590`, with sentinel results; changing it to an object pointer would change its meaning. |
| `549D80` and related death symbols | Some Go references classify callback addresses and select existing native implementations. Raw C ABI and actual dispatcher behavior must be checked separately. |
| `545520`, `545580`, `5455B0` | Actual legacy-int public entries each SIGSEGV above 4 GiB after independent native-layout, typed sound/action/buff and whole-allocation positive controls pass. Frozen original comparison precedes the native port. |

The three retreat predicates now use typed `nox_object_t*` public arguments
and shared native Go cores. Their result remains the original canonical
integer. Raw entries retain original faults and short circuits; ordinary
callers keep their earlier admission checks. Original C bodies remain exact
under `#if 0`, and the frozen original getters/consumers are unchanged.

At ported source `43ae3d9`, the repeated host scan finds **3,499 active C
definitions / 301 integer-cast candidates / 287 fixed-32-bit candidates**.
Only the resume and cast predicates disappear from the cast inventory; the
third, forwarding retreat predicate had no cast of its own. All three typed
public ABIs are independently typechecked and compared with the original.
The fourteen native-width/other candidates remain separated. This scan
introduces no new fixed-32-bit candidate or preprocessor warning.

At that preceding `43ae3d9` source, raw `5456D0` still has an integer object
argument. No direct production call is found outside its definition/header;
the ordinary dispatcher already uses a native Go update. That classification
does not close indirect reachability. The subsequent actual C reproduction
SIGSEGV above 4 GiB after independent layouts, real native owner/Push
composition and typed sound/action/buff/predicate controls pass.

Runtime `8212e8c` ports raw `5456D0` to a typed `nox_object_t*` entry and shares
its update core with the ordinary dispatcher. The raw entry retains original
faults, pointer-load order, cached MUD/live owner behavior, unspilled distance
arithmetic and ordered action writes. Ordinary admission, including guarded
missing health, remains intact. Five original references are frozen before
engine changes, including a corrected C-prefix-versus-Go-EXT layout check;
the initial preparation failure and actual old C crash remain preserved.

The repeated `8212e8c` host scan finds **3,498 active C definitions / 300
integer-cast candidates / 286 fixed-32-bit candidates**. Only `5456D0` leaves
the fixed-width inventory. Fourteen native-width/other candidates remain
separate; no new candidate or preprocessor warning is introduced. Integer
object entries `5455E0`, `545790` and `5457C0` remain follow-up targets at that
source.
Their ordinary native dispatcher implementations do not port those raw C ABIs.

Runtime `7d404efc` subsequently ports ROAM start `545790` and cancel `5457C0`
to native object arguments and pointer returns, sharing their stores with the
ordinary dispatcher while retaining its admission rules. Clean `f6a41c78e`
has identical engine inputs and a corrected opaque-bit fixture preparation.
Both old actual C entries SIGSEGV above 4 GiB after native positive controls.
The independent original executes 75 instruction bytes and 1,321 contexts
per FPU profile, including six faults. The first frozen fixture's unaligned
typed-pointer construction fails strict preparation; a raw-byte replacement
is proved identical across complete 64 KiB allocations at untouched pre-engine
source, retaining original goldens/expected values/guards and all failed files.
The original four other references remain exact. Corrected normal/strict
selections each pass 21,126 events, with matching identities and no new skips.

The repeated host scan finds **3,496 active C definitions / 298 integer-cast
candidates / 284 fixed-32-bit candidates**. Only the two ROAM entries leave
the fixed-width inventory; no new candidate or preprocessor warning is added.
Raw cancel keeps signed unchecked native array arithmetic. Diagnostic
out-of-array projections do not establish named-field alias equivalence after
pointer expansion or certify malformed states; valid slots and ordinary
empty/invalid-head behavior have separate checks.

An additional scan reuses all **191 hash-verified preprocessed C inputs** at
`f6a41c78e`. It finds **45 definitions** explicitly casting a standalone pointer
parameter value to a fixed 32-bit integer; eleven overlap the earlier
all-integer candidate inventory. Scalar field/index/call operands are excluded;
the preliminary broader 100-match inventory remains preserved. Twelve of the
45 have direct non-test Go C.symbol references, which may store callback
addresses rather than invoke them. These views must not be added into a
remaining-defect total. They are semantic triage, with local aliases,
pointer-shaped returns,
callbacks, globals/fields and other narrowing expressions still outside the
expression scan. The full inventory and provenance are retained in
`ROOT/monster-roam-lifecycle-native-entry-port/active-C-typed-pointer-parameter-narrowing-inventory-v2.json`.

For example, `544A00` accepts `nox_object_t*` but explicitly casts it to `int`
before computing the position address. It also returns an integer-shaped
object pointer and stores the selected pointer in uint32 scratch. Its callback
`544A40` still takes two integer object arguments; outer edible-check `5455E0`
is another legacy-int public entry. This family requires its own original
execution, actual C reproductions, native scratch/return and ordinary-caller
analysis before porting. No crash or completion credit is inferred from these
new static findings alone.

Runtime `c351c49c0` subsequently ports all three edible-family entries.
`5455E0` returns a native action-slot pointer, `544A00` uses native search
input/result/shared scratch, and `544A40` accepts native candidate/owner
pointers. Raw live globals, nested search and fault order remain distinct
from ordinary admission and local scratch. All three actual old C entries
fault above 4 GiB after native positive controls. Eight original references
are frozen before the engine change; each FPU profile executes 3,406
contexts, including thirteen faults, across 439 instruction bytes. Normal
and strict selections each pass 14,761 events, including fifty fresh actual-C
production children, without changing those frozen references.

The new hash-bound host scan finds **3,493 active definitions / 297 integer-
cast candidates**. Only `544A40` leaves that cast view: forwarding `5455E0`
was missed because it had no cast of its own, while `544A00` belonged to the
typed-parameter view. The latter now contains **44 definitions**, eleven
overlapping the integer-candidate view. Preprocessor warnings remain zero.
Source/header hashes are independently rebound to clean `c351c49c0`;
earlier working-tree snapshots are not relabelled as clean-source captures.
Both inventories and their provenance remain under
`ROOT/monster-retreat-edibles-native-entry-port`.

The remaining direct Go references need semantic classification: several
parser casts reduce successful non-null results to a boolean, some references
identify callbacks that already dispatch to native Go implementations, and
others call raw C entries. In particular secret-wall packet entries `4DF120`
and `4DF180` accept pointers but narrow them internally, and truncate player-
unit traversal pointers. Their existing ordinary Go implementation does not
close those public C paths. Original execution and actual C reproduction
must precede a port; a static match supplies no crash or completion credit.

Runtime `f6672bd1c` ports both secret-wall packet entries to the existing
native wall record type and full native unit/update/player pointers. Raw
live traversal, cached wall WORD, unchanged packet identity, zero return and
original faults remain separate from the ordinary guarded player snapshot.
Seven original references are frozen before engine changes; both original
FPU profiles execute 172 instruction bytes and 1,188 contexts, including
26 faults, with byte-identical outputs. Both actual old C entries SIGSEGV
above 4 GiB after independent native controls. Normal/strict comparisons and
fourteen fresh default-C pipeline children pass in each profile.

The repeated host scan now finds **3,491 active definitions / 297 integer-
parameter cast candidates**. The separate typed-pointer view contains **42
definitions**, retaining eleven overlaps with the integer view. Only
`4DF120` and `4DF180` leave the typed view; all prior integer-candidate
identities remain. All 191 C inputs and headers match clean `f6672bd1c`, with
zero preprocessor warnings. The original working-tree captures retain their
actual provenance under `ROOT/secret-wall-state-native-entry-port`.

At the preceding `f6672bd1c` source, raw wall-destroyed entry `4DF0A0` remains
an integer-pointer candidate.
Its existing ordinary Go counterpart does not port its public C ABI. The
complete 118-byte original body has distinct direct-Send return and recipient-
32 broadcast paths. Both original FPU profiles execute 1,974 contexts,
including 43 faults, with byte-identical complete outputs. This is original
execution evidence only: its native getter/consumer, actual C reproduction
and port are subsequent gates under `ROOT/wall-destroyed-native-entry-port`.

Runtime `c335ee23d` subsequently ports that public entry to a native wall
argument, retaining the signed C-int recipient/result and direct-Send versus
recipient-32 broadcast behavior. The raw player traversal core is shared
with both preceding secret-wall entries. Six references are frozen before
engine changes; the actual old C entry faults above 4 GiB in two routing
modes after native controls. Normal/strict selections each pass 6,363 events,
including thirteen new and fourteen preceding actual default-C children,
without changing the 4,312 prior inputs or six frozen references.

The new hash-bound host scan finds **3,490 active definitions / 296 integer-
parameter cast candidates**. Only `4DF0A0` leaves that view. The typed-pointer
view retains **42 definitions / eleven integer-view overlaps**, with zero
preprocessor warnings. All 191 C source/header hashes match clean
`c335ee23d`; no working-tree capture is retrospectively relabelled.

Engine `b65798bc5` then ports raw Chakram driver `4EB250` and callback
`4EB340`, including native object/point/result pointers and shared pointer
scratch. Original integer class bits replace the old numeric float cast.
Ten references are frozen first; the actual old driver crashes above 4 GiB
after native controls, while the callback reproduces a valid-target class
failure. Complete original execution covers 376 bytes and 272 contexts per
FPU profile, including nine faults. Ordinary native startup rounding also
fails 75 of 99 original comparisons before correction; those failures remain
preserved. Normal/strict selections subsequently each pass 7,618 events,
including fourteen actual default-C production children per profile.

The hash-bound host scan now finds **3,488 definitions / 294 integer-cast
candidates**. Only these two entries leave that view. The typed-pointer view
has **41 candidates / ten overlaps**, with only `4EB340` removed. All other
candidate identities remain and all 191 preprocessing logs have zero
warnings. The manifest's existing 4,043 records remain unchanged. Working-
tree snapshots keep their original source labels and are rebound by exact
C source/header hashes to the committed engine. A clean ARM64 product and
six host E2Es subsequently qualify. Linux's four native CI gates and a
fresh full Windows amd64/386 run pass. The earlier Windows 386 original
Trigger nil-fault process termination and its exact binaries remain
preserved; corrected exact-PE replays complete but do not identify its
cause. Both macOS 26/15 CI runs fail to acquire runners. Five native local
ARM64 test profiles pass instead. Across those five local and six remote
profiles, every earlier terminal identity/action/multiplicity remains,
with 5,176 additional passes per ordinary profile and 5,181 in strict,
zero new skips and zero emitted warnings. This supplies no passing macOS
remote-CI credit.

The next review covers raw death entries `549D80`, `549E00`, `549E70` and
`54A250`, whose public object arguments are still fixed-width integers.
Existing Go references select ordinary native callbacks and do not close
those raw ABIs. Pre-engine original execution covers all 282 instruction
bytes and 305 inputs per FPU profile, with byte-identical complete outputs.
Eight faults occur in explicitly declared world position reads; none occur
in these original instructions. Reentrant calls execute original bodies,
and complete image/64 KiB heap/caller/register/FPU checks pass. A push-world
mutation demonstrates that later damage/spark calls read the live point.
Six original input/producer/golden/provenance files are subsequently frozen
before the engine change. Both published producers replay byte for byte.
Four nonoverlapping manifest seals add these 282 bytes, retaining every
previous record, its order and metadata: **3,354 code / 693 data / 4,047
total records**; all 21 verifier unit checks pass. The full Go CLI still
rejects the earlier `5471F4` jump-table data in an executable section; the
unchanged preceding manifest reproduces the identical rejection. All 4,047
range hashes and raw PE bounds pass an independent byte check, with both
CLI failures retained and no passing full-CLI credit. The initial publication
preflight rejects a global sort before any source mutation because it would
change the prior record order; the failed preparation is retained and the
new records are appended instead. Getter/consumer freezing, actual
above-4-GiB C reproductions and the port remain pending under
`ROOT/monster-death-explosion-native-entry-port`; no crash or new native-port
credit is inferred from this preparation.

After the ten-reference freeze at `995a01e0f`, the four entries are ported to
native object arguments and the ordinary callback shares their live-point
core. All four actual old C calls previously faulted above 4 GiB after
native layout/full-record and real typed C packet/push controls. Eleven of
289 ordinary original comparisons exposed copied-point mismatches before
the correction. Normal/strict selections now each pass 9,455 events, retain
all 7,618 preceding wall/Chakram events and nineteen earlier death controls,
and add 1,818 family events. Eight fresh default-C root-server children per
profile qualify real map/ray/C-target/packet/audio/deletion composition;
the target damage endpoint is explicitly captured, not an HP-gameplay claim.
All frozen references stay unchanged. The repeated hash-bound host scan
finds **3,484 definitions / 290 integer-cast candidates**: only these four
leave that inventory. Typed-pointer candidates remain **41**, with ten
overlaps and zero preprocessing warnings.

Engine `46f8d9027` subsequently qualifies all seven controlled native-product
E2E lanes and all eleven local/remote test profiles. Each profile preserves
the complete preceding terminal test identities/actions/multiplicities and
adds only 1,818 passes, with no new skips or build warnings. The exact-source
products, available full test executables, ZIPs and complete logs are retained
under `ROOT/monster-death-explosion-native-entry-port` and its product E2E
directory. Full profile counts and scope are recorded in the completion
document; static triage counts still earn no completion percentage.

Next, raw `nox_xxx_monsterDeadTroll_54A270(int)` narrows the source object,
narrows the constructor result again for `createAt`, and reads the created
object's update pointer through PE32 `uint32_t[187]`. Independent native C
controls establish prefix size 912, position offsets 60/64, update offset
872 and an eight-byte update pointer. Its already sealed original body is
154 bytes, plus the 17-byte original converter. Two FPU profiles each execute
all 171 bytes across 136 inputs, retain eleven original faults and differ in
49 complete outputs; both producer and input replays are byte-exact. These
preparations live in `ROOT/troll-death-native-entry-port`. No old-C runtime
crash or native-port credit is inferred before the pending frozen consumers,
positive-control fault reproduction and engine change. The user's remaining
five historical product copies are nonessential; further copy search and
recovery gates are stopped, retaining their hash/missing-state records.

Build, strict-pointer, product and CI results belong in
[porting-completion.md](porting-completion.md), with their exact source
revision. Static inventory changes alone earn no campaign, audio, runtime
stability or overall 100% completion credit.

The Troll references are subsequently frozen in `f77fdc7f6` and `d1dd3eb02`
before changing the public entry. The actual old C call faults above 4 GiB
after the full native layout and real C constructor/update/whole-record
controls pass. The native entry uses a typed source, typed constructor result
and entry-cached native update pointer; its exact old C body remains under
`#if 0`. All 136 inputs in both FPU profiles compare through Go and the real
C symbol without changing any frozen reference. Four default-C root children
also exercise real create/ownership/AUD/balance/FPS, including a live update
replacement and allocation failure before a nil source is read.

The repeated host scan has **3,483 active definitions / 289 integer-cast
candidates**; only this one old C definition and its public declaration change
in the tracked C/header sources. The independent typed-pointer pass retains
**41 candidates / ten overlaps**. The unchanged ordinary callback keeps its
existing admission guards. Product and full native CI qualification remain
separate gates.
