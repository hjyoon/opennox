# 비공개 원본 오라클 매니페스트

이 디렉터리에는 사용자가 보유한 `nox/` 기준본의 **경로, 바이트 수, SHA-256**만 보관한다. `GAME.EXE`, 맵, 음성, 영상 등 원본 자산 자체를 소스 저장소나 공개 CI에 복사하지 않는다.

`nox-2023-1003-01.json`은 다음 규칙으로 봉인한다.

- 모든 일반 파일을 상대 경로 순서로 정렬한다.
- 각 파일의 크기와 SHA-256을 기록한다.
- `(경로 NUL 크기 NUL 파일 SHA-256 LF)` 레코드 전체의 SHA-256을 `tree_sha256`으로 기록한다.
- 심볼릭 링크, 특수 파일, 대소문자만 다른 충돌 경로와 이식 불가능한 경로는 거부한다.
- `seal`은 기존 매니페스트를 덮어쓰지 않는다. 기준본을 바꾸려면 기존 파일을 보존하고 새 ID와 새 파일명으로 별도 검토한다.

검증은 저장소 루트에서 실행한다.

```sh
make oracle-verify
make oracle-code-verify
make oracle-test
```

`game-exe-functions.json`은 포팅한 원본 함수의 VA, 정확한 함수 길이와 범위 SHA-256을 기록한다. 검증기는 전체 트리 매니페스트와 오라클 ID 및 `GAME.EXE` 해시를 교차 확인하고, PE32/I386 형식, 이미지 베이스 `0x00400000`, 함수와 jump table이 실행 섹션에 있는지, 상수 테이블이 비실행 섹션에 있는지와 각 범위 해시를 검사한다. 현재는 `0x004D6A20`, `0x004D8270`, `0x004D9EB0`, SpellProjectile 의존 함수 `0x004E03D0/0x004E0A70`, 주소 순서 복원 `0x004E44F0..0x004EB7FF`, Bomb/Boom/Die/Glyph/SparkExplosion/WallReflect/Pixie/WallReflectSpark/OwnCollide/Spark/Webbing/Flag/Barrel/AudioEvent/Pentagram/Sign/TrapDoor/Teleport/AwardSpell/Fist/TeleportWake/Chakram/Arrow/Harpoon이 사용하는 `0x004EC520..0x004EC5A5`, `0x004EEBF0..0x004EEC7C`, Chakram bolt damage `0x004EF1E0..0x004EF26F`, SpellProjectile이 도달하는 `0x004FA2B0..0x004FA58E`, Boom/Pixie/Chakram 방향 함수 `0x00509ED0..0x00509F1F`, `0x00536D80..0x00536E73`, `0x00537760..0x0053776E`, Chakram update `0x0053DCC0..0x0053DDEF`, `0x0054E170..0x0054E45D`, DeathBall Door 반사 `0x0057B770..0x0057B801`과 shared 벽 반사 `0x0057B810..0x0057B841`을 합계 229개 코드 범위로 봉인했다.

기존 방향 벡터·임계값·반환표, collision 상수와 표·문자열·등록 레코드에 더해 SpellProjectile의 벽 반사 `0.0f`, projectile/DeathBall 반사 `0.1f`, InversionEffect부터 TeleportWake까지 복원한 collision 등록 레코드와 NUL 이름, Boom의 방향 상수와 다섯 balance 문자열, SparkExplosion의 정확한 반경 상수와 parser 형식 문자열, Chest의 SilverKey·잠금 메시지 블록, Spark와 Webbing이 각각 참조하는 별도 WebbingSlow 메시지, WallReflect/WallReflectSpark/Pixie shared parser 형식, AudioEvent의 `%s` parser 형식, DeathBall의 32쌍 Door 방향표와 두 balance 문자열을 봉인했다. 여기에 Chakram의 binary64 `0.5`, binary32 공격 반경 `30`, retarget epsilon·거리 상수, 256쌍 binary32 방향표, random 호출의 원본 debug 경로, `ArcherBolt`·`BoltSoloDamageMin` 문자열, collide/update 등록 레코드와 두 NUL 이름을 추가했다. Arrow가 공유하는 binary64 `0.5`, 별도 위치의 `ArcherBolt` 문자열, collide 등록 레코드와 NUL 이름에 Harpoon의 `HarpoonDamage` balance 키·collide 등록 레코드·NUL 이름을 더해 합계 108개 비실행 데이터 범위를 검사한다. `004E8E50`이 반환하는 `0x007531F8`, Chakram retarget scratch `0x007531F0..0x00753250`, Chest feedback timestamp `0x00753258`, Harpoon damage cache `0x00753298`, Chakram과 DeathBall이 공유하는 trace point `0x00833EB8`과 full ready dword `0x00833EC0`, OwnCollide·BarrelCollide·AudioEventCollide·ChakramUpdate가 읽는 frame `0x0084EA04`, Webbing·ChakramUpdate가 읽는 FPS dword `0x0085B3FC`는 PE의 초기화 데이터 범위 밖 BSS이므로 별도 데이터 해시 범위를 만들지 않는다.

`004EAD50`은 전체 실행 코드와 절대 포인터 스캔에서 참조가 없었던 구형 1-mana transfer 함수이므로 등록·호출 경계를 새로 만들지 않았고, 직접 의존하는 `00418AB0/00419130/00419180/004EEB80` 본체의 전역 복원까지 주장하지 않는다. 등록 테이블 소비자 `00536EC0`은 name 기준 16바이트 record의 callback/data-size/parser를 `+4/+8/+12`에서 읽으며 Chakram collision record `005CA418`은 callback `004EAF00`, data size 0을, Arrow collision record `005CA428`은 callback `004EB490`, Harpoon collision record `005CA4D8`은 callback `004EB6A0`과 원본 collide-data size 8을 결속한다. Arrow와 Harpoon collide-data의 두 번째 필드가 객체 포인터이므로 두 native record는 64비트에서 16바이트다. 별도 update table의 `005C9308` record는 이름 `005C9718`, callback `0053DCC0`, 원본 update-data size 28을 결속한다. 이 record 안의 두 객체 포인터 때문에 native update record는 64비트에서 40바이트이고, Chakram과 Arrow 충돌 중 임시 attack record도 원본 32바이트에서 native 48바이트로 넓힌다. Harpoon이 접근하는 `PlayerUpdateData`의 target/bolt/field35/X/Y/frame은 32비트 `132/136/140/144/148/152`, 64비트 `152/160/168/172/176/180`이다. Fist callback identity consumer와 producer/update, TeleportWake producer와 공유 gate의 다른 상위 호출자, Chest 하위 효과, Pentagram 전체 update record, trace helper의 다른 raw C caller는 계속 별도 widening 범위다. Chakram callback·update·retarget·random-reflect·bolt-damage와 Arrow·Harpoon callback의 의미 및 typed 경계는 복원했지만 inventory detach/put/equip, item/player attack effects, map damage, movement update와 일부 객체 생성·삭제 서비스의 내부 ABI32까지 전역 완료로 주장하지 않는다. 특히 Arrow producer `00539D80`의 주변 매개변수와 지역 변수, attack-effect raw C 서비스는 아직 ABI32 가정을 포함한다. 함수 하나를 포팅할 때마다 어셈블리에서 경계를 확인한 뒤 이 목록과 의미 설명을 확장하며, 다음 순차 함수는 MonsterArrow collision `004EB800`이다.

`oracle-test`는 의미 비교 전과 후에 O0과 코드 범위 검증을 실행한다. 테스트 입력은 읽기 전용으로 취급해야 하며, 컨테이너/VM에서는 가능하면 `nox/`를 read-only로 마운트한다. 사후 검사는 잘못된 테스트가 원본을 수정한 경우도 실패로 남긴다.

다른 위치의 정당한 보유본을 쓰려면 절대 경로나 저장소 루트 기준 경로를 넘긴다.

```sh
make oracle-verify NOX_ORACLE_ROOT=/absolute/path/to/nox
```

Windows에서 GNU Make를 사용하지 않는 경우에도 같은 검증기를 실행할 수 있다.

```powershell
$NoxRoot = (Resolve-Path 'C:\Games\Nox').Path
$Manifest = (Resolve-Path '.\toolchain\oracle\nox-2023-1003-01.json').Path
.\scripts\go.ps1 -C src run .\internal\noxoracle verify `
  -root $NoxRoot -manifest $Manifest

$CodeManifest = (Resolve-Path '.\toolchain\oracle\game-exe-functions.json').Path
.\scripts\go.ps1 -C src run .\internal\noxoracle code-verify `
  -root $NoxRoot -manifest $CodeManifest -oracle-manifest $Manifest
```

전체 트리 매니페스트는 원본의 **신원과 무결성**을 보장하는 Gate O0이다. 함수 범위 매니페스트는 포팅한 소스가 어느 원본 주소·바이트 범위에서 유래했는지 기계적으로 고정하지만, 해시만으로 의미 동등성을 증명하지는 않는다. 각 함수에는 어셈블리 검토와 원본 또는 검증된 기준 구현을 사용한 차등 시험이 추가로 필요하고, 게임 상태·패킷·프레임 동등성은 상위 게이트에서 별도로 검증한다.
