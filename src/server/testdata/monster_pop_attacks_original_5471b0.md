# Original attack-action pop root

Original GAME.EXE SHA256:
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
No executable or original asset is included.

Both startup FPU controls produce identical 3,270 contexts, including four
raw nil faults. All 65 root instruction bytes execute, using the actual
8-byte jump table and 13-byte switch map. Alignment and switch data are not
instruction coverage. Only Pop50A160 is a declared world service; no complete
Pop/End, campaign or general AI equivalence is claimed.

The root caches its entry MUD across Pop calls and ignores their return.
Signed index -1 reads original MUD+528 (timer prefix); all 24 stack slots and
that explicit alias are covered. Other malformed indices receive no credit.
The root is void; incidental EAX receives no semantic result credit. Inputs
cover action DWORD boundaries/high bits, all accepted actions, repeated
popping, dependencies represented by the declared service and changes of the
unit's current MUD to another record or nil. Whole original heap, code/tables,
nonvolatile registers, stack and FPU control/TOP remain guarded.

Gzip SHA256:
`f745b2204becc35cb9b8657e68c808478b91b5073d095d1cf76ff1cda81ab153`.
Decoded SHA256:
`39192bd5c6fae7ff67914fd4adbe3ed60249f5b90afc7860e7ef951408ac9506`.
Inputs/producer/results and native consumer/getter/golden v2 are frozen
before engine edits. Native comparisons cover ordered services/writes and
the entire C-owned native allocation, including padding/neighbors/high Args.

The public C API originally takes an int address. The frozen getter selects
an exactly typed legacy-int wrapper or a typed-object-pointer wrapper using
C11 _Generic; no incompatible function-pointer cast is used. Before porting,
explicit legacy32 narrowing SIGSEGVs only after independent C/Go field/layout
and whole-record controls and existing typed sound/action getters pass above
4 GiB. The exact binary/raw failure remain private in
`ROOT/monster-pop-attack-actions-native-entry-port`.

The first preparatory layout test incorrectly compared the complete Go Object
against the C prefix. Go adds two EXT uintptr handles; the corrected v2 checks
the exact shared prefix end and both extension offsets separately. That first
failure and v1 consumers remain preserved; original inputs/golden expectations
and engine were unchanged when v2 was frozen.

Separate composition tests use native objects and actual PopAction/Reset
with inactive End handlers, mixed dependent actions, all stack slots and -1.
They supplement the declared-service reference and do not prove complete End
callback or networking behavior. Current strict coverage includes the new
family and the unchanged ordinary MainAI pop regression.
