package n8

import (
	"encoding/hex"
	"fmt"

	"forge.rights.ninja/jeff/goedlink/nesrom"
)

const CONFIG_BASE = 32

const (
	CFG_MIR_H   uint8 = 0
	CFG_MIR_V   uint8 = 1
	CFG_MIR_4   uint8 = 2
	CFG_MIR_1   uint8 = 3
	CFG_CHR_RAM uint8 = 4
	CFG_SRM_OFF uint8 = 8
)

const (
	CTRL_RST_DELAY uint8 = 0x01
	CTRL_SS_ON     uint8 = 0x02
	CTRL_SS_BTN    uint8 = 0x08
	CTRL_UNLOCK    uint8 = 0x80
)

type MapConfig struct {
	binaryConfig []uint8
	MapIndex     uint8
	PrgSize      uint32
	ChrSize      uint32
	SrmSize      uint32
	MasterVol    uint8
	SSKeyMenu    uint8
	SSKeySave    uint8
	SSKeyLoad    uint8
	MapCfg       uint8
	Ctrl         uint8
}

//
// Map Config
//

func (c *MapConfig) GetBinary() []uint8 {
	return c.binaryConfig
}
func (c *MapConfig) GetMapIndex() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+0] | ((c.GetBinary()[CONFIG_BASE+2] & 0xf0) << 4))
}
func (c *MapConfig) GetSubmap() uint8 {
	return c.MapCfg >> 4
}
func (c *MapConfig) GetPrgSize() uint32 {
	return (uint32)(8192 << (c.GetBinary()[CONFIG_BASE+1] & 0x0f))
}
func (c *MapConfig) GetChrSize() uint32 {
	return (uint32)(8192 << (c.GetBinary()[CONFIG_BASE+2] & 0x0f))
}
func (c *MapConfig) GetSrmSize() uint32 {
	return (uint32)(128 << (c.GetBinary()[CONFIG_BASE+1] >> 4))
}
func (c *MapConfig) GetMasterVol() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+3])
}
func (c *MapConfig) GetSSKeyMenu() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+8])
}
func (c *MapConfig) GetSSKeySave() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+5])
}
func (c *MapConfig) GetSSKeyLoad() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+6])
}
func (c *MapConfig) GetMapCfg() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+4])
}
func (c *MapConfig) GetCtrl() uint8 {
	return (uint8)(c.GetBinary()[CONFIG_BASE+7])
}

// NewMapConfigFromBinary parses MapConfig from serialized config data
func NewMapConfigFromBinary(data []uint8) *MapConfig {
	var c MapConfig

	buf := make([]uint8, CONFIG_BASE+16)
	copy(buf, data)
	c.binaryConfig = buf

	c.Parse()
	return &c
}

// NewMapConfig returns new default MapConfig
func NewMapConfig() *MapConfig {
	binaryConfig := make([]uint8, CONFIG_BASE+16)
	return &MapConfig{
		binaryConfig: binaryConfig,
		MapIndex:     255,
		SSKeyLoad:    0xff,
		SSKeySave:    0xff,
		SSKeyMenu:    0xff,
	}
}

// NewConfigFromNesRom returns a MapConfig for a given NesRom
func NewConfigFromNesRom(rom *nesrom.NesRom) *MapConfig {
	c := NewMapConfig()
	c.MapIndex = rom.GetMapper()

	switch rom.GetMirroring() {
	case nesrom.MIR_HOR:
		c.MapCfg |= CFG_MIR_H
	case nesrom.MIR_VER:
		c.MapCfg |= CFG_MIR_V
	case nesrom.MIR_4SC:
		c.MapCfg |= CFG_MIR_4
	}

	if rom.GetChrSize() == 0 {
		c.MapCfg |= CFG_CHR_RAM
	}

	c.PrgSize = rom.GetPrgSize()
	c.ChrSize = rom.GetChrSize()
	c.SrmSize = rom.GetSrmSize()

	c.MasterVol = 8
	c.SSKeyMenu = 0x14 // start + down
	c.SSKeySave = 0xff // 0x14
	c.SSKeyLoad = 0xff // 0x18

	return c
}

// Parse initializes MapConfig values from its own raw binary data.
func (c *MapConfig) Parse() {
	c.MapIndex = uint8(c.GetMapIndex())
	c.PrgSize = c.GetPrgSize()
	c.ChrSize = c.GetChrSize()
	c.SrmSize = c.GetSrmSize()
	c.MasterVol = c.GetMasterVol()
	c.SSKeyMenu = c.GetSSKeyMenu()
	c.SSKeySave = c.GetSSKeySave()
	c.SSKeyLoad = c.GetSSKeyLoad()
	c.MapCfg = c.GetMapCfg()
	c.Ctrl = c.GetCtrl()
}

// PrintFull prints all details about the MapConfig in human-readable format.
func (config *MapConfig) PrintFull() {
	fmt.Printf(" mapper.....%d sub.%d\n", config.GetMapIndex(), config.GetSubmap())
	fmt.Printf(" prg size....%dK\n", config.GetChrSize()/1024)
	chrType := ""
	if config.GetMapCfg()&CFG_CHR_RAM != 0 {
		chrType = "ram"
	}
	fmt.Printf(" chr size....%dK %s\n", config.ChrSize/1024, chrType)
	srmState := ""
	if config.MapCfg&CFG_SRM_OFF != 0 {
		srmState = "srm off"
	} else if config.SrmSize < 1024 {
		srmState = fmt.Sprintf("%dB ", config.SrmSize)
	} else {
		srmState = fmt.Sprintf("%dK ", config.SrmSize/1024)
	}
	fmt.Printf(" srm size....%s\n", srmState)
	fmt.Printf(" master vol..%d\n", config.MasterVol)

	mir := "?"
	switch config.MapCfg & 3 {
	case CFG_MIR_H:
		mir = "h"
	case CFG_MIR_V:
		mir = "v"
	case CFG_MIR_4:
		mir = "4"
	case CFG_MIR_1:
		mir = "1"
	}
	fmt.Printf(" mirroring...%s\n", mir)
	fmt.Printf(" cfg bits....%08b\n", config.MapCfg)
	fmt.Printf(" menu key....0x%02X\n", config.SSKeyMenu)
	fmt.Printf(" save key....0x%02X\n", config.SSKeySave)
	fmt.Printf(" load key....0x%02X\n", config.SSKeyLoad)
	fmt.Printf(" rst delay...%s\n", boolToYesNo(config.Ctrl&CTRL_RST_DELAY != 0))
	fmt.Printf(" save state..%s\n", boolToYesNo(config.Ctrl&CTRL_SS_ON != 0))
	fmt.Printf(" ss button...%s\n", boolToYesNo(config.Ctrl&CTRL_SS_BTN != 0))
	fmt.Printf(" unlock......%s\n", boolToYesNo(config.Ctrl&CTRL_UNLOCK != 0))
	fmt.Printf(" ctrl bits...%08b\n", config.Ctrl)
	config.Print()
}

// Print prints the hex-formated data in the config.
func (config *MapConfig) Print() {
	fmt.Printf(" CFG0: %s\n", hex.EncodeToString(config.binaryConfig[CONFIG_BASE:CONFIG_BASE+8]))
	fmt.Printf(" CFG1: %s\n", hex.EncodeToString(config.binaryConfig[CONFIG_BASE+8:CONFIG_BASE+16]))
}

//
// Misc
//

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
