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
