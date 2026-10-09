<p align="center" style="font-size:32pt;font-style:bold">
    <img src="docs/images/logo.png" width=400>
</p>
<p align="center" style="font-size:12pt;font-style:bold">
    <b>OpenNox</b> is an open-source community collaboration project extending the Nox engine. 
</p>
<p align="center">
    <a href="https://github.com/opennox/opennox/actions"><img alt="OpenNox Build Status (dev)" src="https://github.com/opennox/opennox/actions/workflows/build-and-release.yml/badge.svg"></a>
    <a href="https://www.gnu.org/licenses/gpl-3.0.en.html"><img alt="OpenNox license" src="https://img.shields.io/github/license/opennox/opennox?style=flat"></a>
    <br>
    <a href="https://www.patreon.com/opennox"><img alt="OpenNox on Patreon" src="https://img.shields.io/badge/patreon-Support%20us-blue?logo=patreon&logoColor=white&style=flat"></a>
    <a href="https://discord.gg/HgDUeXhAyW"><img alt="OpenNox on Discord" src="https://img.shields.io/badge/discord-OpenNox-blue?logo=discord&logoColor=white&style=flat"></a>
    <a href="https://matrix.to/#/#opennox:nwca.xyz"><img alt="OpenNox on Matrix" src="https://img.shields.io/badge/matrix-%23opennox-blue?logo=matrix&logoColor=white&style=flat"></a>
</p>

## Features

OpenNox supports all vanilla Nox features. You should be able to complete the campaign and play online with OpenNox.
If something doesn't work, please [open an issue](https://github.com/opennox/opennox/issues/new/choose).

For a list of new features see [this page](https://opennox.github.io/docs/opennox/features/index.html).

## Download OpenNox

<a href="https://github.com/opennox/opennox/releases"><img alt="OpenNox releases" src="https://img.shields.io/github/downloads/opennox/opennox/total?style=flat&label=releases"></a>
<a href="https://snapcraft.io/opennox"><img alt="OpenNox Snap package" src="https://img.shields.io/badge/snap-Install-green?logo=snapcraft&logoColor=white&style=flat"></a>

### Release
All release builds are made from the `dev` branch. Recent OpenNox releases can be found [here](<https://github.com/opennox/opennox/releases>).

Linux releases are also available in `stable` channel of our [Snap package](https://snapcraft.io/opennox).

Packages for Arch Linux are available on the AUR: <a href="https://aur.archlinux.org/packages/opennox"><img alt="AUR opennox package" src="https://img.shields.io/aur/version/opennox?style=flat&label=opennox&logo=archlinux"></a> <a href="https://aur.archlinux.org/packages/opennox-bin"><img alt="AUR opennox-bin package" src="https://img.shields.io/aur/version/opennox-bin?style=flat&label=opennox-bin&logo=archlinux"></a>

### Nightly
On each commit, an automated build of the `dev` branch is uploaded.
These builds contain all the latest merged features, but are not yet considered stable for release.
These builds are to help provide an insight to what the next release will contain and should only be used for active playtesting purposes **only**.

Linux nightly builds are also available in `edge` channel of our [Snap package](https://snapcraft.io/opennox).

A package for Arch Linux is available on the AUR: <a href="https://aur.archlinux.org/packages/opennox-git"><img alt="AUR opennox-git package" src="https://img.shields.io/aur/version/opennox-git?style=flat&label=opennox-git&logo=archlinux"></a>

## Build OpenNox
**NOTE: This section is only for people who wish to build the source code locally.**

This porting branch requires exactly Go 1.26.5. Use `./scripts/go.sh` on macOS/Linux or `scripts\\go.ps1` on Windows; see [the toolchain policy](./toolchain/README.md), [the source baseline](./toolchain/BASELINE-linux-386.md), [the private-data oracle policy](./toolchain/oracle/README.md), and [the live portability inventory](./toolchain/PORTING-INVENTORY.md).

For the local reference copy used by this branch, `make oracle-test` first verifies every path, size, and SHA-256 in the private Nox data tree and then runs the implemented semantic comparisons. Original assets remain outside this repository.

### Local GUI regression tests

`scripts/run-headless-gui-e2e.sh` builds the native client and runs a scenario with isolated saves and configuration. Playback defaults to a deterministic headless screen. To exercise the SDL/OpenGL window and presentation on a desktop session, opt in explicitly:

```sh
NOX_E2E_SEAT=sdl scripts/run-headless-gui-e2e.sh \
  /path/to/nox scripts/e2e/host-game-objective-modes.yaml /tmp/opennox-sdl-e2e
```

Set `NOX_E2E_CLIENT_TARGET=client-hd` to test the HD client. Scripted clicks retain the existing 1024×768 reference-screen coordinates and adapt to canvas size, window scaling, letterboxing, and Retina displays; raw recorded events keep their original coordinates. SDL mode requires the platform's native graphics dependencies and access to the desktop window service. The default test driver disables audio; E2E audio-handle tests do not verify hardware sound playback.

The Solo quickbar scenarios verify the real keyboard-to-incantation path, not just the target-mode indicator:

```sh
NOX_E2E_SEAT=headless NOX_E2E_AUDIO=mock scripts/run-headless-gui-e2e.sh \
  /path/to/nox scripts/e2e/solo-wizard-quickbar-buff.yaml /tmp/opennox-quickbar-wizard
```

Use `solo-conjurer-quickbar-buff.yaml` with a separate output directory for Protection from Poison; the Wizard scenario covers Haste and Protection from Fire. Each uses stock book-award insertion, the default self-target bit, a real nugget click to switch targets, and a real spell shortcut. Observations check incantation progress, one stock mana debit, server/client buffs, HUD mana, an unaffected control NPC, preserved quickbar entries, and natural expiry. The explicit starting fixtures are a book award and two ordinary waiting NPCs; these tests do not inject casts, target pointers, mana, or buff timers. They do not establish all spell damage, movement-speed effects, autonomous combat AI, or hardware audio behavior. See [the verification record](toolchain/PORTING-INVENTORY.md#실제-단축키-입력의-자기타인-시전과-자연-만료-회귀).

`host-game-spell-unit-matrix.yaml` checks nine spell families in all six directions between players, ordinary monsters, and scripted NPCs (54 cases). Run it with the same headless/mock wrapper; use a separate output directory and `NOX_E2E_CLIENT_TARGET=client-hd` for HD. It observes real script casts, projectile hits or duration processing, server/client HP and buffs, selected FX, natural status expiry, and unaffected casters/spectators. Position, durable health, waiting AI and ownership are explicit starting fixtures; damage, hit callbacks and network results are not injected. This is not coverage of every spell, keyboard/mana casting, autonomous AI, every campaign encounter, or hardware audio. See [the verification record](toolchain/PORTING-INVENTORY.md#세-unit-종류-사이의-스펠-피해상태자연-만료-검증).

`host-game-trap-projectile-damage.yaml` checks stock ArrowTrap1/2, Skull1..4 and Polyp against players, monsters and NPCs (21 cases). `solo-wizard-chapter8-tower-damage.yaml` checks all three victims against the untouched Wiz08e Force of Nature timer. Both use the same headless/mock wrapper and observe natural launches/collisions, server/client damage and the cloud's first poison tick. Victim placement, durable HP and waiting AI are fixtures; projectile/cast/damage/network outputs are not injected. See [the verification record](toolchain/PORTING-INVENTORY.md#arrowtrapskullpolyp와-원본-챕터8-force-of-nature-피해) for the precise limits.

`host-warrior-switch-visual.yaml` checks six on/off transitions of an original G_Crypts Spike, both secret-wall diagonals through repeated opening/closing and minimap reopening, and ordinary wall creation/deletion/pool reuse. Normal map-script enable calls and wall APIs supply inputs; the live network frame, stock SlaveDraw and actual C minimap raster supply results. The observer's position, minimap zoom and three temporary empty-grid walls are explicit fixtures; client frames/packets, secret-wall states and expected pixels are not injected. Normal and HD headless/mock runs each passed six Spike checks and 50 minimap pixel checks. This does not establish every map, remote-client topology, Windows pixel equivalence or hardware audio. See [the verification record](toolchain/PORTING-INVENTORY.md#스파이크-onoff-외형과-벽-미니맵-수명주기).

`host-wizard-plasma.yaml` checks the stock OblivionOrb's automatic pickup/equip, actual mouse attacks, damage and charge reports, and the live Plasma ray. As in the original game, releasing the attack button retains the acquired target; actual movement cancels the attack, and a second attack naturally exhausts all 250 charges. Both stops must clear the duration, ray, wand flag and native pointer sidecars. Waiting AI, placement and durable target HP are starting fixtures; casts, damage, charge changes and client packets are not injected. Use a byte-identical private copy of the YAML and `NOX_E2E_OVERRIDE=true` to keep its newly captured Screens outside the repository, with separate normal/HD output directories. Both headless/mock clients passed; this does not establish Linux runtime behavior, Windows pixel equivalence or hardware audio. See [the verification record](toolchain/PORTING-INVENTORY.md#플라즈마-지속-스펠의-native64-크래시와-충전광선-수명주기).

`solo-wizard-award-frame-pacing.yaml` checks real XP level-up, stock Haste book acquisition, and all four SetHalberd/Oblivion upgrades. It removes the E2E driver's artificial delay, enables the actual frame limiter, and measures wall time and rendered frames while simulation time is naturally paused; each reward must then finish and resume gameplay naturally. Pause flags, simulation frames, timers and render results are not injected. Use a byte-identical private YAML copy with `NOX_E2E_OVERRIDE=true`, separate normal/HD output directories, and `NOX_E2E_CLIENT_TARGET=client-hd` for HD. Both headless/mock runs passed; measured headless FPS is not a physical display or Linux runtime certification. See [the verification record](toolchain/PORTING-INVENTORY.md#레벨업새-스펠oblivion-무기-업그레이드의-일시정지-프레임-제한).

To verify real OpenAL playback of the original menu effects, chapter music, and speech, enable audio explicitly:

```sh
NOX_E2E_SEAT=sdl NOX_E2E_AUDIO=openal scripts/run-headless-gui-e2e.sh \
  /path/to/nox scripts/e2e/solo-warrior-native-audio.yaml /tmp/opennox-audio-e2e
```

This mode requires access to the default playback device and the complete original game assets. It checks native source allocation and hardware-consumed buffers, excluding discarded buffers and the movie player's separately managed audio. It does not assess audible quality. Movies and playback use real time, so allow several minutes. The default `NOX_E2E_AUDIO=mock` remains deterministic; do not combine `openal` with mock `NOX_E2E_AUDIO_HANDLES=true`.

The opt-in OpenAL integration tests use generated silent PCM/ADPCM on the actual default device and verify pause/resume, short-clip draining, playback accounting, and resource teardown:

```sh
cd src
NOX_TEST_OPENAL=true ../scripts/go.sh test ./legacy/client/audio/ail -run '^TestOpenAL' -count=3
```

### Linux
- [Linux](./docs/build-linux.md)
  
### Windows
- [Windows](./docs/build-windows.md)
- [Windows (on Linux)](./docs/build-windows-on-linux.md)

## Contributing
Read [CONTRIBUTING](CONTRIBUTING.md)!

## Legal

<a href="https://www.gnu.org/licenses/gpl-3.0.en.html"><img alt="OpenNox license" src="https://img.shields.io/github/license/opennox/opennox?style=flat"></a>

This project (OpenNox) is an unofficial community collaboration project for preservation, modding and compatibility purposes.
This project has no direct affiliation with Electronic Arts Inc. and/or the "Nox" brand. "Nox" is an Electronic Arts Inc. brand. All Rights Reserved.

No assets, texts, artwork or other media from the original game(s) is included in this project.
We do not condone piracy in any way, shape or form and encourage users to legally own the original game.

The video game "Nox" is copyright © 2000 Westwood Studios. All Rights Reserved.
Westwood Studios is a trademark or registered trademark of Electronic Arts in the U.S. and/or other countries. All rights reserved.

If not specified otherwise, the source code provided in this repository is licenced under the [GNU General Public License version 3](<https://www.gnu.org/licenses/gpl-3.0.html>). Please see the accompanying LICENSE file.

OpenNox logo created by [@CCHyper](https://github.com/CCHyper) under [CC0 license](https://creativecommons.org/share-your-work/public-domain/cc0/).

OpenNox project additionally follows [C&C Remastered Modding guideline](https://www.ea.com/games/command-and-conquer/command-and-conquer-remastered/modding-faq). All changes to the project MUST follow these rules.
