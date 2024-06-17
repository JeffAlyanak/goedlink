package config

import (
	"encoding/hex"
	"fmt"
)

const cfgBase = 32

const (
	CfgMirH   = 0
	CfgMirV   = 1
	CfgMir4   = 2
	CfgMir1   = 3
	CfgChrRam = 4
	CfgSrmOff = 8
)

const (
	CtrlRstDelay = 0x01
	CtrlSsOn     = 0x02
	CtrlSsBtn    = 0x08
	CtrlUnlock   = 0x80
)

type MapConfig struct {
	config    []byte
	MapIndex  int
	PrgSize   int
	ChrSize   int
	SrmSize   int
	MasterVol byte
	SSKeyMenu byte
	SSKeySave byte
	SSKeyLoad byte
	MapCfg    byte
	Ctrl      byte
}

func NewMapConfigFromBinary(bin []byte) *MapConfig {

	config := make([]byte, cfgBase+16)
	copy(config, bin)

	return &MapConfig{
		config:    config,
		MapIndex:  getMapIndex(config),
		PrgSize:   getPrgSize(config),
		ChrSize:   getChrSize(config),
		SrmSize:   getSrmSize(config),
		MasterVol: getMasterVol(config),
		SSKeyMenu: getSSKeyMenu(config),
		SSKeySave: getSSKeySave(config),
		SSKeyLoad: getSSKeyLoad(config),
		MapCfg:    getMapCfg(config),
		Ctrl:      getCtrl(config),
	}
}

func NewMapConfig() *MapConfig {
	config := make([]byte, cfgBase+16)
	return &MapConfig{
		config:    config,
		MapIndex:  255,
		SSKeyLoad: 0xff,
		SSKeySave: 0xff,
		SSKeyMenu: 0xff,
	}
}

type NesRom struct {
	Mapper    int
	Mirroring rune
	PrgSize   int
	ChrSize   int
	SrmSize   int
}

func NewMapConfigFromNesRom(rom *NesRom) *MapConfig {
	mc := NewMapConfig()
	mc.MapIndex = rom.Mapper

	switch rom.Mirroring {
	case 'H':
		mc.MapCfg |= CfgMirH
	case 'V':
		mc.MapCfg |= CfgMirV
	case '4':
		mc.MapCfg |= CfgMir4
	}

	if rom.ChrSize == 0 {
		mc.MapCfg |= CfgChrRam
	}

	mc.PrgSize = rom.PrgSize
	mc.ChrSize = rom.ChrSize
	mc.SrmSize = rom.SrmSize

	mc.MasterVol = 8
	mc.SSKeyMenu = 0x14 // start + down
	mc.SSKeySave = 0xff // 0x14
	mc.SSKeyLoad = 0xff // 0x18

	return mc
}

func getMapIndex(config []byte) int {
	return (int)(config[cfgBase+0] | ((config[cfgBase+2] & 0xf0) << 4))
}
func getPrgSize(config []byte) int {
	return (int)(8192 << (config[cfgBase+1] & 0x0f))
}
func getChrSize(config []byte) int {
	return (int)(8192 << (config[cfgBase+2] & 0x0f))
}
func getSrmSize(config []byte) int {
	return (int)(128 << (config[cfgBase+1] >> 4))
}
func getMasterVol(config []byte) byte {
	return (byte)(config[cfgBase+3])
}
func getSSKeyMenu(config []byte) byte {
	return (byte)(config[cfgBase+8])
}
func getSSKeySave(config []byte) byte {
	return (byte)(config[cfgBase+5])
}
func getSSKeyLoad(config []byte) byte {
	return (byte)(config[cfgBase+6])
}
func getMapCfg(config []byte) byte {
	return (byte)(config[cfgBase+4])
}
func getCtrl(config []byte) byte {
	return (byte)(config[cfgBase+7])
}

func (mc *MapConfig) GetBinary() []byte {
	return mc.config
}

func (mc *MapConfig) Submap() byte {
	return mc.MapCfg >> 4
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (mc *MapConfig) PrintFull() {
	fmt.Printf("mapper.....%d sub.%d\n", mc.MapIndex, mc.Submap())
	fmt.Printf("prg size....%dK\n", mc.PrgSize/1024)
	chrType := ""
	if mc.MapCfg&CfgChrRam != 0 {
		chrType = "ram"
	}
	fmt.Printf("chr size....%dK %s\n", mc.ChrSize/1024, chrType)
	srmState := ""
	if mc.MapCfg&CfgSrmOff != 0 {
		srmState = "srm off"
	} else if mc.SrmSize < 1024 {
		srmState = fmt.Sprintf("%dB ", mc.SrmSize)
	} else {
		srmState = fmt.Sprintf("%dK ", mc.SrmSize/1024)
	}
	fmt.Printf("srm size....%s\n", srmState)
	fmt.Printf("master vol..%d\n", mc.MasterVol)

	mir := "?"
	switch mc.MapCfg & 3 {
	case CfgMirH:
		mir = "h"
	case CfgMirV:
		mir = "v"
	case CfgMir4:
		mir = "4"
	case CfgMir1:
		mir = "1"
	}
	fmt.Printf("mirroring...%s\n", mir)
	fmt.Printf("cfg bits....%08b\n", mc.MapCfg)
	fmt.Printf("menu key....0x%02X\n", mc.SSKeyMenu)
	fmt.Printf("save key....0x%02X\n", mc.SSKeySave)
	fmt.Printf("load key....0x%02X\n", mc.SSKeyLoad)
	fmt.Printf("rst delay...%s\n", boolToYesNo(mc.Ctrl&CtrlRstDelay != 0))
	fmt.Printf("save state..%s\n", boolToYesNo(mc.Ctrl&CtrlSsOn != 0))
	fmt.Printf("ss button...%s\n", boolToYesNo(mc.Ctrl&CtrlSsBtn != 0))
	fmt.Printf("unlock......%s\n", boolToYesNo(mc.Ctrl&CtrlUnlock != 0))
	fmt.Printf("ctrl bits...%08b\n", mc.Ctrl)
	mc.Print()
}

func (mc *MapConfig) Print() {
	fmt.Printf("CFG0: %s\n", hex.EncodeToString(mc.config[cfgBase:cfgBase+8]))
	fmt.Printf("CFG1: %s\n", hex.EncodeToString(mc.config[cfgBase+8:cfgBase+16]))
}
