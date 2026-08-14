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

`game-exe-functions.json`은 포팅한 원본 함수의 VA, 정확한 함수 길이와 범위 SHA-256을 기록한다. 검증기는 전체 트리 매니페스트와 오라클 ID 및 `GAME.EXE` 해시를 교차 확인하고, PE32/I386 형식, 이미지 베이스 `0x00400000`, 함수와 jump table이 실행 섹션에 있는지, 상수 테이블이 비실행 섹션에 있는지와 각 범위 해시를 검사한다. 현재는 `0x004D6A20`, `0x004D8270`, `0x004D9EB0`, SpellProjectile 의존 함수 `0x004E03D0/0x004E0A70`, `0x004E44F0..0x004EA37D`, `0x004EADE0`, Bomb/Boom/Die/Glyph/SparkExplosion/WallReflect/Pixie/WallReflectSpark/OwnCollide/Spark가 사용하는 `0x004EC520..0x004EC5A5`, `0x004EEBF0..0x004EEC7C`, SpellProjectile이 도달하는 `0x004FA2B0..0x004FA58E`, Boom/Pixie 방향 함수 `0x00509ED0..0x00509F1F`, `0x00536D80..0x00536E73`, `0x00537760..0x0053776E`, `0x0054E170..0x0054E45D`, DeathBall Door 반사 `0x0057B770..0x0057B801`과 shared 벽 반사 `0x0057B810..0x0057B841`의 176개 함수와 다섯 dispatch table 범위를 합계 181개 코드 범위로 봉인했다. 기존 방향 벡터·임계값·반환표, collision 상수와 표·문자열·등록 레코드에 더해 SpellProjectile의 벽 반사 `0.0f`, projectile/DeathBall 반사 `0.1f`, InversionEffect·SpellProjectileCollide·BombCollide·BoomCollide·DieCollide·GlyphCollide·SparkExplosionCollide·ChestCollide·WallReflectCollide·WallReflectSparkCollide·PixieCollide·OwnCollide·SparkCollide·YellowStarShotCollide·DeathBallCollide·DeathBallFragmentCollide 등록 레코드와 NUL 이름, Boom의 방향 상수와 다섯 balance 문자열, SparkExplosion의 정확한 반경 상수와 parser 형식 문자열, Chest의 SilverKey·잠금 메시지 블록, Spark의 WebbingSlow 메시지, WallReflect/WallReflectSpark/Pixie shared parser 형식, DeathBall의 32쌍 Door 방향표와 두 balance 문자열까지 합계 59개 비실행 데이터 범위를 별도 봉인했다. `004E8E50`이 반환하는 `0x007531F8`, Chest feedback timestamp `0x00753258`, DeathBall trace point `0x00833EB8`과 full ready dword `0x00833EC0`, OwnCollide frame `0x0084EA04`는 PE의 초기화 데이터 범위 밖 BSS이므로 별도 데이터 해시 범위를 만들지 않는다. `004FA2B0`은 이번 충돌 경로가 도달하는 state 13과 18..20 및 고정 mapping만 의미 복원 완료로 세며, 다른 호출자의 동적 ability/weapon 분기는 아직 전역 완료로 주장하지 않는다. Chest가 호출하는 `004EDF00`과 `004EDA40`의 내부 ABI32 본체와 trace helper `00537760`을 쓰는 DeathBall 외 다섯 raw C caller의 ABI32 반환 경계도 별도 복원 범위다. 함수 하나를 포팅할 때마다 어셈블리에서 경계를 확인한 뒤 이 목록과 의미 설명을 확장하며, 다음 순차 함수는 Webbing collision `004EA380`이다.

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
