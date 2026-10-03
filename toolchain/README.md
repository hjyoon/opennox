# Go 도구체인 정책

이 포팅 브랜치의 유일한 지원 도구체인은 `go1.26.5`이다. 버전의 단일 텍스트 기준은 `go-version.txt`이며, `src/go.mod`에는 언어 기준 `go 1.26.0`과 권장 도구체인 `toolchain go1.26.5`를 함께 선언한다.

빌드와 테스트는 다음 래퍼로 실행한다.

```sh
./scripts/go.sh version
./scripts/go.sh -C src test ./internal/noxbuild
```

Linux 64비트의 native-width CGo 회귀는 `make test-linux-pie`로 root·server·legacy 전체를 시험한다. 기본 비-PIE 실행 파일에서는 C heap이 4GiB 아래에 놓일 수 있어 고주소 포인터 시험의 전제가 성립하지 않을 수 있다. PIE 시험은 그 전제를 유지하는 별도 검증 게이트이며, 일반 제품의 빌드·실행 검증을 대체하지 않는다.

2026-10-04 player self-report의 장비 인챈트 바이트 보고 누락을 복원했다. 원본 `004D99A7..004D99E1`은 진입 시의 update 포인터에서 Player를 읽어 마지막 보고 바이트와 object `+440`의 하위 바이트를 비교하고, 변경 시 `004D8840`의 reliable `5B <item-enchantment-byte>` 패킷을 보낸다. 전송 실패도 결과를 무시하며, 전송 뒤 같은 update 포인터에서 Player와 해당 바이트를 다시 읽어 캐시를 갱신한다. `+440`은 native `Field110`이며, 중독 강도 `Poison540` (`+540`)과는 무관하다. 최초 연결의 필드 오해를 headless 실행에서 발견해 바로잡았다. 새 high-C-owned root 회귀는 65,536개 바이트 비교 조합, 256개 unsigned 수신자, 골드 콜백 이후 재로딩, 전송 중 update/Player/마스크가 바뀌는 경우 및 실패 후 캐시 갱신을 검사한다. 상위 24비트만 바뀌거나 중독이 적용되었을 때 잘못 보고하지 않고, 원본 마스크와 중독 상태·타이머도 보존하는 독립 회귀를 추가했다. 원본 코드 manifest에는 보고 slice·sender·3바이트 padding의 해시 범위만 추가했다. Quest key 보고는 아직 별도 누락 slice다.

Windows PowerShell에서는 다음과 같이 실행한다.

```powershell
.\scripts\go.ps1 version
.\scripts\go.ps1 -C src test ./internal/noxbuild
```

실행할 제품이 현재 clean source revision과 정확히 일치하는지는 `noxbuild -verify`로 확인한다. 이 검사는 ELF, PE, Mach-O를 호스트 종류와 무관하게 읽어 Go 1.26.5, 지원 대상 tuple, 전체 `vcs.revision`, `vcs.modified=false`를 강제한다. 특히 무시된 `build/` 아래에 남은 구 실행 파일을 다시 실행하는 일을 차단한다.

```sh
./scripts/go.sh -C src run ./internal/noxbuild \
  -go=../scripts/go.sh -verify \
  ../build/linux-amd64/opennox ../build/linux-amd64/opennox-server
```

```powershell
.\scripts\go.ps1 -C src run ./internal/noxbuild `
  -go=go -verify `
  ..\build\windows-amd64\opennox.exe ..\build\windows-amd64\opennox-server.exe
```

래퍼는 외부 `GOROOT`를 제거하고 `GOTOOLCHAIN=go1.26.5`와 빈 `GOEXPERIMENT`를 강제한 뒤, 실제 `GOVERSION`이 정확히 일치하는지 검사한다. 이로써 goenv 같은 버전 관리자가 다른 표준 라이브러리 경로를 주입하는 경우도 차단한다. `GO` 환경 변수에는 공백 없는 Go 실행 파일 경로 하나만 지정할 수 있다. `internal/noxbuild`도 자신을 컴파일한 Go와 자식 빌드에 쓰는 Go를 각각 검사한다.

따라서 `GOEXPERIMENT=cgocheck2 ./scripts/go.sh ...`는 strict CGo 검증이 아니다. POSIX 환경의 테스트 전용 실행 경로는 고정 도구체인을 먼저 찾은 뒤 `cgocheck2`와 `CGO_ENABLED=1`의 실제 적용을 확인한다. 릴리스 빌드 래퍼의 정책은 바꾸지 않으며, 테스트 인수를 생략하면 전체 패키지를 실행한다.

```sh
sh ./scripts/test-cgocheck2.sh
sh ./scripts/test-cgocheck2.sh -run TestNative -count=3 . ./legacy ./server
```

macOS/ARM64의 GUI 회귀는 `NOX_E2E_SEAT=headless`로 실행한다. 다음 통합 시나리오는 Con01a 곰 전투, Con02a 네크로맨서 소환·공격·퇴장, 늑대 Charm 60회와 Henrick 추종, 이후 ESC 메뉴 입력을 같은 세션에서 검증한다. 맵 trigger로 플레이어 위치만 준비하며 대화 버튼은 실제 mouse 입력으로 누른다. 일반·HD 제품 모두 통과했으며, 네크로맨서 소환 시 frozen 상태와 이벤트 종료 후 조작/cinematic 복귀를 구분한다. 이 검증은 SDL 화면 출력이나 실제 OpenAL 재생 검증을 대신하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-conjurer-headless.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/solo-conjurer-headless.yaml
```

`host-game-flame-monster-food.yaml`도 일반·HD headless 실행에서 player HP `150→148`, Spider의 RedApple `40→45`, Meat `40→50`, Mushroom 독 `4→0`과 소비된 필드 오브젝트 제거를 확인했다. 휴면 tutorial NPC/Wolf의 원본 `Cur=Max=0`은 사망 상태가 아니며, 일반 몬스터의 food/retreat AI와 구분한다. 통합 이벤트의 종료가 통과하더라도 미이식 피해 분기는 별도로 검사해야 한다.

`host-game-fist-unit-damage.yaml`은 실제 Wizard host 메뉴 시작 뒤 정상 NoxScript object-to-position API로 Fist 1..5레벨을 플레이어→stock NPC 및 NPC→플레이어 방향으로 시전한다. 두 유닛의 원래 `PlayerDamage` callback, 자연 충돌, 서버 HP/피격 marker/귀속, fractional carry, 실제 client 피해 표시와 발사체 제거를 검사하며 일반·HD headless 각각 10회, 합계 20회가 exit 0이다. 원본 피해 `50/100/200/300/400`, Wizard의 시작 armor `0.23000002` 및 연속 시전의 소수 누적을 유지하고 NPC→플레이어 실효 피해 `44/89/177/265/354`를 확인했다. 모든 시전에서 중복 거부, 18틱의 명중, 36틱의 world/owned/client drawable 제거를 관찰하며 native object/update 포인터는 4GiB를 넘는다. 위치·검증용 HP 2,000·NPC 대기 AI만 fixture이고 피해 결과·패킷·물리 상태를 주입하지 않는다. 자율 NPC 주문 선택, incantation/mana, 유닛 사망·campaign trigger·다른 모든 피해 종류나 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다.

기존 배럴 Fist 제어 시나리오의 첫 historical PNG는 앞선 색상 복원 이후 불일치한다. 이번 피해 수정 전 `e882c931b`의 별도 clean worktree에서도 같은 불일치를 확인했고 현재 `_got` PNG와 byte-identical이었다. 기존 공개 기준 PNG를 바꾸거나 override하지 않았다. 피해 수정 전 worktree에서 얻은 private 비교 프레임 10개를 기준으로 일반·HD의 배럴 5레벨 피해/제거 및 RGBA 비교를 모두 통과했으며 이를 historical PNG 통과로 세지 않는다. 원본 자산·개인 Save/config와 private 참조 이미지는 저장소에 추가하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-fist-unit-damage.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-fist-unit-damage.yaml
```

`host-game-push-spell.yaml`은 실제 Wizard host 시작 뒤 플레이어의 정상 script cast 1..5레벨과 stock NPC의 `AIActionCastOnObj`→자연 cast-animation frame→`CastSpellByUser4FDD20`→Push selector 경로를 검사한다. 일반·HD headless 각각 6회가 exit 0이며, NPC는 원본 mode-selected power 3으로 11틱 후 animation frame 4에서 시전하고 12틱 후 행동 종료와 서버·client 이동을 확인했다. 원본 `PushPowerCoeff=15`, 반경 600/10, 거리 감쇠와 대상 mass를 유지하며 정확한 서버 force, 실제 packet을 받은 drawable 이동, 1회 cast audio event와 HP 불변을 확인한다. native 시전자·대상·NPC update 포인터는 4GiB를 넘는다. 기존 6-int C 진입점의 시전자 절단 및 PE32 `+56` 위치 참조를 실제 public wrapper의 실패 회귀로 재현했고, native `PosVec`(wide host `+60`)를 쓰는 Go cast로 복원했다. 원본 executable/자산은 검사 전후 동일하다. 위치·대기 AI·명시적인 주문/대상 선택만 fixture이며 force·cast frame·피해·이동 packet·client 위치를 주입하지 않는다. 자율 NPC 주문 선택, 플레이어 incantation/mana, 특정 campaign trigger 또는 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-push-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-push-spell.yaml
```

`host-game-pull-spell.yaml`은 같은 실제 Wizard host 시작과 정상 script cast 1..5레벨, stock NPC의 자연 `AIActionCastOnObj` 시전 경로에서 Pull을 검사한다. 일반·HD headless 각각 6회, 합계 12회가 exit 0이다. 원본 `PullPowerCoeff=15`, 반경 600/10, 거리 감쇠와 대상 mass를 유지하며, 서버의 정확한 음수 force와 실제 이동 packet을 받은 client drawable의 **안쪽 이동**을 확인했다. NPC는 원본 mode-selected power 3으로 11틱 후 animation frame 4에서 시전하고 12틱 후 행동 종료와 양쪽 이동을 확인했다. 모든 시전에서 cast audio event는 1회이며 NPC HP 150과 Wizard HP 75는 변하지 않는다. 4GiB 초과 native 포인터를 실제 공개 Pull wrapper에 넘기면 기존 6-int C callee의 시전자 절단 및 PE32 `+56` 접근으로 크래시가 발생하는 실패 회귀를 먼저 재현했고, wide host의 실제 `PosVec`(offset `+60`)를 사용하는 Go cast로 복원했다. 원본의 signed DWORD level 곱셈→FCHS 부호 반전→binary32 spill 순서, 음수 0과 세 번째 오브젝트의 cast audio 귀속은 별도 함수 회귀에서 검사한다. 기존 Push 시나리오도 일반·HD 각각 6회 재실행해 바깥쪽 이동이 유지됨을 확인했으며 원본 executable/자산은 검사 전후 동일하다. 위치·대기 AI·명시적인 주문/대상 선택만 fixture이고 force 결과·cast frame·HP·이동 packet·client 위치는 주입하지 않는다. 자율 NPC 주문 선택, 플레이어 incantation/mana, campaign trigger 또는 Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-pull-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-pull-spell.yaml
```

`host-game-fumble-spell.yaml`은 실제 Warrior host 메뉴 시작 뒤 플레이어→stock NPC의 정상 script object cast 요청 1..5, 플레이어→일반 Troll의 drop-all 분기, stock NPC→플레이어의 자연 cast-animation 경로를 검사한다. 일반·HD headless 각각 7회, 합계 14회가 exit 0이다. 원본 targeted Magic projectile은 요청 level과 별개로 mode-selected power 3을 유지하며 Fumble 효과는 level을 사용하지 않는다. 장착 무기·wand·방패만 떨어뜨리고 몸통 방어구와 미장착 예비 장비를 유지하는 판정은 함수 회귀로 고정했고, 실제 NPC/플레이어의 무기·방패 드롭 및 client 장비 mask 동기화도 확인했다. 일반 몬스터는 인벤토리 두 개가 모두 떨어지고, 최초 시전자가 아닌 **명중한 Magic의 위치**를 기준으로 원본 force 50과 mass에 따른 정확한 힘 및 서버·client의 바깥쪽 이동을 확인한다. NPC HP 150, Troll HP 80, Warrior HP 150은 변하지 않는다. NPC는 11틱 후 animation frame 4에서 자연 시전하고 20틱 후 명중, 22틱 후 행동 종료·투사체 제거·client replay가 완료됐다. native unit/item/projectile/update 포인터는 4GiB를 넘는다. 모든 드롭의 실제 서버 상태와 시야 안 드롭의 client drawable을 검사하며, 원본의 무작위 reachable 위치가 벽 뒤인 드롭은 표시 결과 검증에서 구분한다. 위치·대기 AI·정상 아이템 지급/장착·명시적인 주문/대상 선택만 fixture이고 드롭·장비 flag·힘·cast frame·HP·투사체·client 결과를 주입하지 않는다. 원본 executable/자산은 검사 전후 동일하다. GameBall/Shopkeeper 분기는 별도 함수 회귀 범위이며, 자율 NPC 주문 선택·플레이어 incantation/mana·campaign trigger·Linux/Windows/SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-fumble-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-fumble-spell.yaml
```

`host-game-player-status-animation.yaml`은 실제 메뉴로 시작한 regular host에서 정상 enchant API로 제자리 플레이어에게 Held(스턴)·Confused·AntiMagic·Charming·Shield·Slowed를 각각 90틱 적용한다. 초기 무적은 강제로 제거하지 않고 자연 만료를 기다린다. 일반·HD headless와 일반 독립 재실행에서 실제 클라이언트 패킷의 buff/HUD 동기화, 두 시점의 원본 효과 sprite 픽셀 및 서로 다른 애니메이션 프레임, 자연 만료 후 제거를 확인했다. Slowed는 실제 화면 안의 YellowBubbleParticle 생성과 만료 후 0개를 검사한다. 효과별 시작·진행·해제의 18개 PNG도 모든 실행에서 일치했다. 원래 로더의 packed DWORD 저장으로 native 애니메이션 cache가 nil이 되는 실패를 먼저 재현했으며, cache 초기화와 세션 정리를 별도 함수 단위로 복원했다. 상태 준비만 fixture이고 client buff·화면 출력·만료를 주입하지 않는다. Stun 주문의 조건별 Held/Slowed 선택, 적의 주문 명중·이동 제어, 다른 상태 종류·캐릭터 자세나 SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-animation.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-animation.yaml
```

`host-game-player-status-spell.yaml`은 실제 Wizard 메뉴 선택 뒤 level 1 Confuse와 Stun을 self-target으로 시전한다. `host-game-player-status-spell-warrior.yaml`은 실제 Warrior 메뉴로 Stun의 Slowed 분기를 확인한다. 정상 `SpellAccept4FD400` selector→6-argument C cast 진입→buff 적용→실제 client packet/HUD→그리기→자연 만료 경로이며, enchant를 직접 적용하는 앞 시나리오와 구분한다. 원본 Confuse 90틱·Stun 60틱과 class 선택을 바꾸지 않는다. Wizard의 Held 머리 위 효과는 두 서로 다른 sprite frame에서 픽셀이 일치했고, Warrior는 원본대로 Held가 아닌 Slowed의 노란 입자가 두 시점에서 7개·6개 표시된 뒤 사라졌다. 4GiB 초과 native player/argument 포인터를 유지하며, C-heap argument를 명시적으로 초기화하는 회귀도 검사한다.

일반·HD headless 네 실행에서 시작·진행·해제의 새 PNG 9개가 decoded RGBA 기준으로 정확히 일치했다. 공유 시각 검사 helper 변경 후 기존 6종 상태의 PNG 18개도 일반 회귀 실행에서 그대로 일치했다. self-target server cast 준비만 fixture이며 적의 주문 명중·플레이어 incantation 입력·mana 소비, 모든 캐릭터 자세나 SDL/OpenAL 출력을 대신하지 않는다. Conjurer/기타 class byte와 Monster Mass 분기는 별도 native 함수 회귀 범위다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-spell.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-spell.yaml
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-player-status-spell-warrior.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-player-status-spell-warrior.yaml
```

`host-game-warrior-abilities.yaml`은 regular host의 정상 초기화로 습득한 Warrior 스킬 5개를 실제 A/S/D/F/G 키로 각각 두 번 사용한다. 돌진 피해·범위 내 함성의 anti-magic/stun·작살의 피해와 끌어오기·가볍게 걷기의 은신 유지 및 이동 중 공격으로 해제·늑대의 눈의 투명 적 감지를 확인한다. 각 사용의 서버 상태·실제 패킷의 HUD·종료·쿨다운 중 재입력 거부·ready 보고·재사용을 검사하며 일반·HD headless 실행 모두 통과했다. Troll HP 2,000, 정지 AI 대상·위치·투명 Spider와 quickbar 배치는 명시적인 fixture다. 스킬 실행·종료·쿨다운을 직접 주입하지 않는다. 원본 HarpoonDuration=0과 TreadLightlyDuration=99,999는 변경하지 않으며, 벽 돌진·PvP·campaign의 레벨별 검증이나 SDL/OpenAL 검증을 대신하지 않는다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-warrior-abilities.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-warrior-abilities.yaml
```

`host-game-warrior-charge-collisions.yaml`은 실제 A 키 입력으로 서버 플레이어 fixture와 기존 stock 벽에 각각 두 번 돌진한다. 일반·HD headless에서 플레이어 HP `2,000→1,850`, 벽 충돌의 자기 피해·Held·종료·쿨다운 재입력 거부·HUD ready·재사용을 확인했다. 정상 join/관전자 종료로 초기화한 두 번째 서버 플레이어의 HP·배치와 quickbar만 준비하며 능력·collision callback·피해·CollisionWall·buff 종료·쿨다운을 주입하지 않는다. 원격 클라이언트의 네트워크 입력 대결이나 FlagBall GameBall drop을 대신하지 않으며, 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-game-warrior-charge-collisions.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-game-warrior-charge-collisions.yaml
```

`solo-conjurer-chapter2-field-guide-shop.yaml`은 일반·HD headless 클라이언트에서 Con02a의 실제 Mystic 상점 열기, Wasp 책 구매·인벤토리 사용·습득 알림, 상점 재입장과 세션 해제를 검증한다. 구매 자금 10,000과 Urchin 보상은 명시적인 fixture이며, Wasp는 실제 맵 상점 정의에서 생성된다. 실제 마우스 입력과 서버·클라이언트 패킷 처리로 Wasp 가격 100, 잔액 `10,000→9,900`, 재고 `8→7`, 책 소비 `1→0`, 습득 레벨 `0→1`을 확인하고 재입장 후에도 품절을 검사한다. 이 시나리오도 다른 게임 E2E와 동시에 실행하지 않는다.

`solo-conjurer-pet-autosave-load.yaml`은 seed 프로세스에서 실제 주문으로 만든 Wolf 1마리·Urchin 2마리와 Con02a AUTOSAVE를 준비하고 정상 종료한 뒤, 새 프로세스의 실제 메뉴에서 로드한다. 두 일반·HD headless 실행에서 script ID·HP·owner·HUD·minimap·client drawable, 이어지는 Con03a 출구 전환과 player 이동을 확인했다. 기대값 JSON은 격리된 save 루트의 저장 슬롯 밖에 있으며 소환수를 생성·복원하지 않는다. Pixie 저장·Quest 단계·개인 기존 save는 이 시나리오의 범위가 아니다. seed와 load YAML을 같은 디렉터리에 두고 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-conjurer-pet-autosave-load.yaml
```

`host-quest-twenty-stages.yaml`은 실제 Warrior Quest 메뉴부터 stock 출구로 1→20단계를 진행한다. 각 단계의 지도·playable player·native 생성기 최대 수, 전환 후 실제 이동 19회, 5/20단계 미니언의 HP와 inventory 연결을 검사한다. 일반·HD headless 모두 제품당 생성기 2,596개 기록과 20단계 강화 경계의 High `8→16`·Normal `5→10`·Low `2→4`·Singular `1→1`을 확인하고 정상 종료했다. player의 출구 배치만 fixture이며 stage·spawn·HP·생성기 cap·UI 해제·이동은 주입하지 않는다. 생성기의 실제 최대 spawn 개체 수·boss 전투/drop·온라인 Quest 소환수 보존·20단계 이후나 SDL/OpenAL 검증은 별도다. 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-quest-twenty-stages.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-quest-twenty-stages.yaml
```

`host-quest-minion-combat.yaml`은 실제 Quest 메뉴와 네 번의 stock 출구 전환으로 5단계 네크로맨서를 생성한 뒤, 플레이어만 근처에 배치한다. 일반·HD headless에서 원래 AI가 2틱 뒤 적을 감지하고 실제 전기 피해로 player HP `450→397`과 client 피해 표시를 만들었다. 실제 mouse 추적·이동·공격으로 Necromancer HP `100→0`, 사망과 원래 inventory의 CommonSpellBook 필드 방출을 확인했다. 몬스터 생성·AI·aggression·HP·damage·death·보상을 주입하지 않으며 일반 독립 재실행에서도 두 전투 PNG와 기존 단계 결과 PNG가 일치했다. 이는 해당 네크로맨서의 전투 검증이며 Hecubah나 모든 피해 종류, 주변 FlyingGolem의 별도 PIERCE shape guard, 온라인 대전 검증을 포함하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/host-quest-minion-combat.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/host-quest-minion-combat.yaml
```

`solo-warrior-purchased-wolf-transition.yaml`은 War01a 원본 세 NPC 이벤트를 끝내고 stock War03b를 로드한 다음, Henrick의 실제 Yes/Done 입력 두 번으로 원래 Wolf1/Wolf2를 구매한다. 일반·HD headless에서 stock 가격 200, 골드 `400→200→0`, 서로 독립적인 Migrate/Monitor bit를 확인하고 War03b→새 War03c→저장된 War03b를 stock 출구 collision으로 왕복했다. 전환 직후와 추가 240틱 뒤의 네 assertion에서 같은 두 객체·script ID·wire ID·HP `40/40`·host owner·client drawable과 player 근처 복원을 확인하며 중복 생성도 검사한다. 네 PNG는 구매 대화와 두 목적지 화면을 기록한다. 초기 War03b 로드, 구매 자금 400, stock StartDialog 서비스 호출과 player의 출구 배치만 fixture다. 늑대 생성·재소유·HP/AI/migration flag·출구 callback을 직접 주입하지 않는다. 전체 campaign 도보 진행·다른 챕터·온라인 Quest의 소환수 보존이나 SDL/OpenAL 검증은 별도이며 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-warrior-purchased-wolf-transition.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/solo-warrior-purchased-wolf-transition.yaml
```

독립 서버의 실제 기동·맵 전환·정상 종료는 다음 headless 회귀로 확인한다. `curl`·`jq`가 필요하며, 고정 도구체인으로 `server` 제품을 빌드한 뒤 개인 save/config를 제외한 임시 데이터 뷰에서 실행한다. E2E 모드로 NAT 포워딩과 공개 서버 등록을 차단하고, 인증된 로컬 API로 `so_beach→estate→trilevel→estate`를 로드한다. 각 맵의 mode·player 수, 실제 game tick 증가와 map load 횟수를 확인하고 console `quit` 이후 10초 안에 프로세스가 exit 0으로 끝나야 통과한다. 원본 데이터는 변경하지 않고 출력 디렉터리에 로그·임시 런타임 데이터를 남긴다. 기본 게임/API 포트는 18610, metrics 포트는 6062이며 다른 서버 E2E와 동시에 쓰지 않는다. `NOX_E2E_SERVER_PORT`·`NOX_E2E_SERVER_METRICS_PORT`로 바꿀 수 있다. 이미 빌드한 제품을 검사할 때에는 절대 경로의 `NOX_E2E_SERVER_BINARY`를 지정한다. 이 검증은 접속한 원격 플레이어의 게임플레이·공개 서버 발견·Quest 소환수 보존을 대신하지 않는다.

```sh
bash ./scripts/run-headless-server-e2e.sh \
  /absolute/path/to/nox /absolute/path/to/server-e2e-output
```

`host-quest-player-death.yaml`은 실제 Quest 메뉴와 네 번의 stock 출구 전환 뒤 자연 전투로 Warrior를 세 번 사망시킨다. 일반·HD headless와 일반 새 프로세스 재실행에서 서버 목숨 `2→1→0→2`, death counter `0→1→2→0`, HP `450→0→450`, 같은 native player identity, 마지막 사망의 결과창과 실제 Continue 입력 부활을 확인했다. 20 ordinary dead frame 뒤 native host PlayerInd 31과 실제 timer child 10712의 empty StaticText data를 관찰하고 Continue 직전 화면도 캡처한다. player placement만 fixture이며 timer/text event·GUI 값·HP·AI·목숨·통계·사망·페널티·부활은 주입하지 않는다. PNG 8개가 일반·HD에서 바이트 단위로 같고 독립 재실행 baseline도 통과했다.

최종 발사체는 제거되었을 수 있어 독점적 lethal attribution은 단정하지 않는다. 실제 전투 골드와 generator/monster/secret counter가 0인 한계, ankh HUD `X 0`의 별도 미완료 경계, remote non-host 결과창·SDL/OpenAL·부활 후 새 이동·다른 캐릭터/피격·온라인 Quest 소환수 보존 범위를 구분한다. host countdown은 `0049B6E0` 복원 뒤 empty text와 실제 화면으로 후속 검증했다. 원본 자산과 개인 Save/config는 변경하지 않는다. PNG가 저장소에 생성되지 않도록 YAML을 임시 시나리오 디렉터리에 복사해 실행하고 다른 GUI E2E와 동시에 실행하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox /absolute/path/to/temp-e2e/host-quest-player-death.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  /absolute/path/to/temp-e2e/host-quest-player-death.yaml
```

옵션 audit 시나리오 `main-menu-options-audit.yaml`과 `host-game-client-options-audit.yaml`은 일반·HD headless의 실제 입력, native 설정값과 직렬화 결과를 점검한다. 초기 실패와 함수별 보정 이력은 [포팅 인벤토리](PORTING-INVENTORY.md)에 보존한다. 최신 private 회귀의 일반/HD main은 226/228개, 게임 내는 201/203개 assertion 통과·실패 0·정상 exit 0이며 기존 PNG 18개를 override 없이 비교했다. 이는 반복 assertion 수이지 고유 기능 수가 아니다. checkbox·음량/mute·해상도 pending 값·mouse/키 재지정/스크롤·Reset/Defaults·Back/Apply/ESC/Close를 검사하며 E2E의 해상도/디스크 변경 억제는 유지한다.

비-E2E headless framebuffer/API 회귀는 별도로 실제 C apply와 game-entry/menu reset을 호출한다. 일반 8개/HD 10개 해상도·세 signed window mode에서 98/122개 frame 검사와 기존 GUI 네 폰트의 16개 실제 pixel/binding 검사를 각각 3회 통과했다. 버퍼 재생성 뒤 폰트 handle 연결이 끊기던 문제를 고쳤다. 생성한 font/seat fixture를 stock GUI/game-loop 전체의 해상도 적용이나 실제 macOS 창·Retina/물리 오디오 검증으로 확대하지 않는다. 원본 renderer의 기존 encoded PNG MD5 불일치는 변경 전 코드에서도 같으며 별도 미해결 항목이다. 설정 파일의 새 프로세스 복원·원본 default.cfg Reset·실제 OpenAL Soft null 검증도 각각 독립된 근거와 한계를 인벤토리에 기록한다.

패치 버전을 올릴 때에는 다음 항목을 한 변경으로 갱신한다.

1. `toolchain/go-version.txt`
2. `src/go.mod`의 `toolchain` 지시자
3. `src/internal/noxbuild/main.go`의 `requiredGoVersion`
4. Docker 이미지 태그와 CI 매트릭스
5. `BASELINE-linux-386.md` 및 전체 대상 빌드 결과

현재 포팅 단계, 정적 검색 재현법과 확인된 64비트 ABI 단절은 `PORTING-INVENTORY.md`에 기록한다.

첫 구조체 분리의 근거, C/Go 오프셋과 실제 Linux/386 산출물 검증값은 [`ABI-OBJECT.md`](ABI-OBJECT.md)에 기록한다.

사용자가 보유한 원본 데이터에 의존하는 시험은 [`oracle/README.md`](oracle/README.md)의 O0 무결성 게이트를 먼저 통과해야 한다. `make oracle-test`는 봉인된 `nox/` 이외의 데이터로 기대 결과가 바뀌는 것을 차단한다.

임의의 `GOEXPERIMENT`, 호스트 기본 아키텍처 최적화, `latest` 도구체인은 릴리스 산출물에 사용하지 않는다.

Linux Docker 빌더의 Go 베이스는 다음 멀티아키텍처 인덱스로 고정한다.

```text
golang:1.26.5-trixie@sha256:98988b42f3293b627bf07c884ff17181a59501769cd8c06c7ba901e0ce2c9853
```

이 인덱스에서 이 프로젝트가 사용하는 플랫폼은 `linux/386`, `linux/amd64`, `linux/arm/v7`, `linux/arm64/v8`이다.
