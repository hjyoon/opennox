package opennox

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
	"unsafe"

	"github.com/opennox/libs/ifs"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
)

var (
	nox_enable_audio     = 1
	dword_5d4594_805980  *audioStructXxx
	dword_5d4594_1193336 uint32
)

var (
	audioTimer93944 ail.Timer = math.MaxUint32
	audioDev        ail.Driver
	audioMaster     *timer.TimerGroup
)

func sub_43EFD0(a1 unsafe.Pointer) int {
	s := (*legacy.AudioVoiceNative)(a1).Sample
	s.Smp.End()
	legacy.AudioSampleFinishNative(s)
	return 0
}

func sub_43EC30(a1 unsafe.Pointer) int {
	smp := audioDev.AllocateSample()
	s, _ := alloc.New(legacy.AudioSample{})
	(*legacy.AudioVoiceNative)(a1).Sample = s
	s.Dev = audioDev
	s.Field1 = a1
	s.Smp = smp
	b1, _ := alloc.Make([]byte{}, 16384)
	b2, _ := alloc.Make([]byte{}, 16384)
	s.Data1 = &b1[0]
	s.Data2 = &b2[0]
	if smp == 0 {
		return -2147221504 // 0x80040000
	}
	smp.SetUserData(s)
	return 0
}

func sub_43ECB0(a1 unsafe.Pointer) int {
	v := (*legacy.AudioVoiceNative)(a1)
	p := v.Sample
	if p.Smp != 0 {
		p.Smp.Release()
	}
	if p2 := p.Data1; p2 != nil {
		alloc.Free(p2)
	}
	if p2 := p.Data2; p2 != nil {
		alloc.Free(p2)
	}
	alloc.Free(p)
	v.Sample = nil
	return 0
}

func sub_43E940(a1 unsafe.Pointer) int {
	if startAudioServices() != 0 {
		return -2147221504 // 0x80040000
	}
	if a1 != nil {
		(*legacy.AudioDeviceNative)(a1).Limit = 1
	}
	return 0
}

func startAudioServices() int {
	if audioTimer93944 != math.MaxUint32 {
		return 0
	}
	ail.Startup()
	audioTimer93944 = ail.RegisterTimer(func(u uint32) {
		sub_486EF0()
		// Native FX bypass the old event/voice service. Apply its live FX
		// group to those voices too, without touching legacy/movie samples.
		nativeAudioFX.mu.Lock()
		if len(nativeAudioFX.voices) != 0 {
			fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
			fx.Update()
			nativeAudioFX.serviceAmbientLoopsLocked()
			enabled := legacy.Sub_453070() != 0
			for _, voice := range nativeAudioFX.voices {
				eventVolume, ok := voice.UserData().(uint32)
				if !ok || voice.Status() != 4 {
					continue
				}
				if !enabled {
					// Discard queued hardware buffers as well as the AIL state.
					// Re-enabling admits future events, not these old sounds.
					voice.Init()
					continue
				}
				mixedVolume := uint64(eventVolume) * uint64(fx.Timers[0].Current>>16) / VolumeMax
				volume := int(127 * mixedVolume >> 14)
				if volume > 127 {
					volume = 127
				}
				voice.SetVolume(volume)
			}
		}
		nativeAudioFX.mu.Unlock()
		legacy.MusicModule.Sub_43D2D0()
		(*timer.TimerGroup)(legacy.Get_dword_587000_127004()).ClearUpdated()
	})
	if audioTimer93944 == math.MaxUint32 {
		return -2147221504 // 0x80040000
	}
	audioTimer93944.SetFrequency(30)
	audioTimer93944.Start()
	legacy.Sub_42EBB0(1, legacy.Get_sub_43E910(), 0, "Audio")
	legacy.Sub_42EBB0(2, legacy.Get_sub_43E8E0(), 0, "Audio")
	return 0
}

func sub_43E9F0() {
	if audioTimer93944 != math.MaxUint32 {
		audioTimer93944.Stop()
		audioTimer93944.Release()
		audioTimer93944 = math.MaxUint32
	}
	ail.Shutdown()
}

func noxAudioServeT(a1 int) {
	noxAudioServe()
}

func noxAudioServe() { // sub_43F1A0
	if audioDev != 0 {
		ail.Serve()
	}
}

func sub_43EA20(a1 unsafe.Pointer) int {
	//char v2[16]; // [esp+4h] [ebp-10h]

	ctx := (*legacy.AudioContextNative)(a1)
	v2 := sub_43EA90(unsafe.Pointer(&ctx.Format60[0]))
	audioDev = sub_43EAD0(v2)
	if audioDev == 0 {
		return -2147221504 // 0x80040000
	}
	*memmap.PtrInt32(0x587000, 93948) = 0
	ctx.Limit = 24
	return 0
}

type audioStruct1 struct {
	Field0  uint16
	Field2  uint16
	Field4  uint32
	Field8  uint32
	Field12 uint16
	Field14 uint16
}

func sub_43EA90(a2 unsafe.Pointer) *audioStruct1 {
	return &audioStruct1{
		Field0:  1,
		Field2:  *(*uint16)(unsafe.Add(a2, 12)),
		Field4:  *(*uint32)(unsafe.Add(a2, 8)),
		Field8:  *(*uint32)(unsafe.Add(a2, 20)),
		Field12: *(*uint16)(unsafe.Add(a2, 16)) * *(*uint16)(unsafe.Add(a2, 12)),
		Field14: 8 * *(*uint16)(unsafe.Add(a2, 16)),
	}
}

func sub_43EAD0(a1 *audioStruct1) ail.Driver {
	if dr := ail.WaveOutOpen(); dr != 0 {
		return dr
	}
	_ = ail.LastError()
	return 0
}

func sub_43EC10() int {
	if audioDev != 0 {
		audioDev.Close()
		audioDev = 0
	}
	return 0
}

func sub_43F130() ail.Driver {
	return audioDev
}

func sub_43ED00(a1p unsafe.Pointer) int {
	v := (*legacy.AudioVoiceNative)(a1p)
	s := v.Sample
	smp := s.Smp
	smp.Init()
	s.Field4 = unsafe.Pointer(v.Buffer.Format)
	if s.Field4 == nil {
		s.Field4 = unsafe.Pointer(&v.Context.Format60[0])
	}
	v3 := int32(legacy.Sub_43F0E0(s.Field4))
	smp.SetType(v3, 0)
	if *(*uint32)(unsafe.Add(s.Field4, 4)) == 2 {
		smp.SetADPCMBlockSize(*(*uint32)(unsafe.Add(s.Field4, 24)))
	}
	sub_43F060(a1p)
	s.Flag7 = 0
	s.Field3 = 0
	smp.RegisterEOBCallback(func() {
		v := smp.UserData().(*legacy.AudioSample)
		legacy.Sub_43EE00(v)
	})
	smp.RegisterEOSCallback(func() {
		legacy.Sub_43EDB0(smp)
	})
	legacy.Sub_43EE00(s)
	return 0
}

func sub_43F060(a1p unsafe.Pointer) int {
	v := (*legacy.AudioVoiceNative)(a1p)
	s := v.Sample
	smp := s.Smp
	smp.SetVolume(int((127 * (v.Mixed.Timers[0].Current >> 16)) >> 14))
	smp.SetPan(int((127 * (v.Mixed.Timers[2].Current >> 16)) >> 14))
	p2 := s.Field4
	v2 := legacy.Sub_486640(unsafe.Pointer(&v.Mixed), int(*(*uint32)(unsafe.Add(p2, 8))))
	smp.SetPlaybackRate(v2)
	return 0
}

func nox_xxx_parseSoundSetBin_424170(path string) error {
	if legacy.Nox_xxx_parseSoundSetBin_424170(path) == 0 {
		return fmt.Errorf("failed to load sounds")
	}
	return nil
}

func nox_audio_initall(a3 int) int {
	if nox_enable_audio == 0 {
		return 1
	}
	if a3 != 0 {
		if sub_4311F0() == 0 {
			return 0
		}
		legacy.Set_dword_587000_81128(unsafe.Pointer(audioMaster))
		if err := nativeAudioFX.init(audioDev, "audio"); err != nil {
			audioEffectsLog.Printf("cannot initialize effects: %v", err)
			shutdownNativeAudio()
			return 0
		}
	}
	(*timer.TimerGroup)(memmap.PtrOff(0x5D4594, 805884)).Init()
	(*timer.TimerGroup)(legacy.Get_dword_587000_93164()).Init()
	(*timer.TimerGroup)(legacy.Get_dword_587000_122852()).Init()
	(*timer.TimerGroup)(legacy.Get_dword_587000_127004()).Init()
	legacy.Dialogs.Nox_xxx_WorkerHurt_44D810()
	legacy.MusicModule.Init()
	v1 := configGetVolume(VolumeMusic)
	if v1 == 0 {
		legacy.Sub_43DC00()
	}
	(*timer.Timer)(legacy.Get_dword_587000_93164()).SetRaw(uint32(v1))
	v2 := configGetVolume(VolumeDialog)
	if v2 == 0 {
		legacy.Sub_44D960()
	}
	(*timer.Timer)(legacy.Get_dword_587000_122852()).SetRaw(uint32(v2))
	v3 := configGetVolume(VolumeFX)
	if v3 == 0 {
		legacy.Sub_453050()
	}
	(*timer.Timer)(legacy.Get_dword_587000_127004()).SetRaw(uint32(v3))
	return 1
}

func sub_4311F0() int {
	if startAudioServices() != 0 {
		return 0
	}
	audioDev = ail.WaveOutOpen()
	if audioDev == 0 {
		sub_43E9F0()
		return 0
	}
	audioMaster, _ = alloc.New(timer.TimerGroup{})
	audioMaster.Init()
	legacy.Set_dword_5d4594_805984(nil)
	return 1
}

func shutdownNativeAudio() {
	legacy.Game2AudioShutdownNative()
	nativeAudioFX.close()
	legacy.Set_dword_587000_81128(nil)
	if audioMaster != nil {
		alloc.Free(audioMaster)
		audioMaster = nil
	}
	sub_43EC10()
	sub_43E9F0()
}

func sub_486F30() int {
	legacy.AudioInitNative()
	dword_5d4594_1193336 = 1
	return 0
}

func sub_486EF0() {
	if dword_5d4594_1193336 != 0 {
		legacy.AudioServeNative()
	}
}

const audioStructXxxNativeSize = 288 + 8*(unsafe.Sizeof(uintptr(0))-4)

var _ = [1]struct{}{}[audioStructXxxNativeSize-unsafe.Sizeof(audioStructXxx{})]

type audioStructXxx struct {
	arr0       *audioStructYyy
	size4      uint
	path2_8    [260]byte
	bagfile268 *legacy.FILE // 67, 268
	field272   *legacy.FILE // 68, 272
	field276   uint32       // 69, 276
	field280   *legacy.FILE // selected native bag/override handle (PE32 +280)
	field284   uint32       // 71, 284
}

func (p *audioStructXxx) C() unsafe.Pointer {
	return unsafe.Pointer(p)
}

func (p *audioStructXxx) Free() {
	if p == nil {
		return
	}
	if p.bagfile268 != nil {
		legacy.Nox_fs_close(p.bagfile268)
		p.bagfile268 = nil
	}
	if p.field272 != nil {
		legacy.Nox_fs_close(p.field272)
		p.field272 = nil
	}
	if p.arr0 != nil {
		alloc.Free(p.arr0)
		p.arr0 = nil
	}
	alloc.Free(p)
}

var _ = [1]struct{}{}[36-unsafe.Sizeof(audioStructYyy{})]

type audioStructYyy struct {
	field0  [32]byte
	field32 uint32
}

func (p *audioStructYyy) C() unsafe.Pointer {
	return unsafe.Pointer(p)
}

func sub_4866F0(path1 string, path2 string) *audioStructXxx {
	idxPath := path1
	if i := strings.LastIndexByte(idxPath, '.'); i >= 0 {
		idxPath = idxPath[:i]
	}
	idxPath += ".idx"
	fi, err := ifs.Open(idxPath)
	if err != nil {
		return nil
	}
	bf1 := binfile.NewFile(fi)
	cf1 := legacy.NewFileHandle(bf1)
	defer legacy.Nox_fs_close(cf1)

	bagPath := path1
	if i := strings.LastIndexByte(bagPath, '.'); i >= 0 {
		bagPath = bagPath[:i]
	}
	bagPath += ".bag"
	fb, err := ifs.Open(bagPath)
	if err != nil {
		return nil
	}

	p, _ := alloc.New(audioStructXxx{})
	p.bagfile268 = legacy.NewFileHandle(binfile.NewFile(fb))
	if p.bagfile268 == nil {
		p.Free()
		return nil
	}
	var hdr [12]byte
	if _, err := bf1.Read(hdr[:12]); err != nil {
		p.Free()
		return nil
	}
	_ = binary.LittleEndian.Uint32(hdr[0:])
	vers := binary.LittleEndian.Uint32(hdr[4:])
	p.size4 = uint(binary.LittleEndian.Uint32(hdr[8:]))
	var arr []audioStructYyy
	if p.size4 > 0 {
		arr, _ = alloc.Make([]audioStructYyy{}, p.size4)
		p.arr0 = &arr[0]
		var buf [36]byte
		if vers != 1 {
			for i := uint(0); i < p.size4; i++ {
				if _, err := bf1.Read(buf[:36]); err != nil {
					p.Free()
					return nil
				}
				copy(arr[i].field0[:], buf[:32])
				arr[i].field32 = binary.LittleEndian.Uint32(buf[32:])
			}
		} else {
			for i := uint(0); i < p.size4; i++ {
				if _, err := bf1.Read(buf[:32]); err != nil {
					p.Free()
					return nil
				}
				copy(arr[i].field0[:], buf[:32])
				arr[i].field32 = 0
			}
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		s1, s2 := alloc.GoStringS(arr[i].field0[:]), alloc.GoStringS(arr[j].field0[:])
		s1, s2 = strings.ToLower(s1), strings.ToLower(s2)
		return s1 < s2
	})
	p.field276 = 0
	if path2 != "" {
		alloc.StrCopyZero(p.path2_8[:], path2)
		if i := alloc.StrLenS(p.path2_8[:]); p.path2_8[i] == '\\' {
			p.path2_8[i] = 0
		}
		var find legacy.WIN32_FIND_DATAA

		if h := legacy.FindFirstFileA(&p.path2_8[0], &find); int(h) != -1 {
			if find.FileAttributes&0x10 != 0 {
				p.field276 = 1
			} else {
				for legacy.FindNextFileA(h, &find) != 0 {
					if find.FileAttributes&0x10 != 0 {
						p.field276 = 1
						break
					}
				}
			}
			legacy.FindClose(h)
		}
		// TODO: strlen()-1 ?
		if i := alloc.StrLenS(p.path2_8[:]); p.path2_8[i] != '\\' {
			p.path2_8[i] = '\\'
			p.path2_8[i+1] = 0
		}
	}
	return p
}
