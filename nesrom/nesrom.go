package nesrom

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
)

const (
	ADDR_PRG uint32 = 0x00000000
	ADDR_CHR uint32 = 0x00800000
	ADDR_SRM uint32 = 0x01000000
)

const (
	ADDR_OS_PRG uint32 = (ADDR_PRG + 0x7E0000)
	ADDR_OS_CHR uint32 = (ADDR_CHR + 0x7E0000)
)

const (
	ROM_TYPE_NES uint32 = 0
	ROM_TYPE_FDS uint32 = 1
	ROM_TYPE_OS  uint32 = 2

	MIR_HOR uint8 = 0x48 // 'H'
	MIR_VER uint8 = 0x56 // 'V'
	MIR_4SC uint8 = 0x34 // '4'
	MIR_1SC uint8 = 0x31 // '1'

	FDS_DISK_SIZE uint32 = 65500

	MAX_ID_CALC_LEN uint32 = 0x100000
)

type NesRom struct {
	romPath   string
	prg       []uint8
	chr       []uint8
	ines      []uint8
	crc       uint32
	srmSize   uint32
	mapper    uint8
	mirroring uint8
	batRam    bool
	datBase   uint32
	size      uint32
	romType   uint32
	prgAddr   uint32
	chrAddr   uint32
}

// type NesRom struct {
// 	Mapper    int
// 	Mirroring rune
// 	PrgSize   int
// 	ChrSize   int
// 	SrmSize   int
// }

func NewNesRom(path string) (*NesRom, error) {
	if path == "" {
		return nil, fmt.Errorf("ROM is not specified")
	}

	rom, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	n := &NesRom{
		romPath: path,
		size:    (uint32)(len(rom)),
		ines:    make([]uint8, 32),
	}
	copy(n.ines, rom[:32])

	nes := n.ines[0] == 'N' && n.ines[1] == 'E' && n.ines[2] == 'S'
	fds00 := n.ines[11] == 'H' && n.ines[12] == 'V' && n.ines[13] == 'C'
	fds16 := n.ines[11+16] == 'H' && n.ines[12+16] == 'V' && n.ines[13+16] == 'C'

	switch {
	case nes:
		n.romType = ROM_TYPE_NES
		n.datBase = 16
		n.prgAddr = ADDR_PRG
		prgSize := uint32(rom[4]) * 1024 * 16
		chrSize := uint32(rom[5]) * 1024 * 8
		n.srmSize = 8192
		if prgSize == 0 {
			prgSize = 0x400000
		}
		n.mapper = (rom[6] >> 4) | (rom[7] & 0xf0)
		if rom[6]&1 == 0 {
			n.mirroring = MIR_HOR
		} else {
			n.mirroring = MIR_VER
		}
		n.batRam = rom[6]&2 != 0
		if rom[6]&8 != 0 {
			n.mirroring = MIR_4SC
		}
		if n.mapper == 255 {
			n.romType = ROM_TYPE_OS
			n.prgAddr = ADDR_OS_PRG
			n.chrAddr = ADDR_OS_CHR
		}
		n.prg = make([]uint8, prgSize)
		n.chr = make([]uint8, chrSize)
		copy(n.prg, rom[n.datBase:n.datBase+prgSize])
		copy(n.chr, rom[n.datBase+prgSize:n.datBase+prgSize+chrSize])
	case fds00 || fds16:
		n.romType = ROM_TYPE_FDS
		if fds00 {
			n.datBase = 0
		} else {
			n.datBase = 16
		}
		n.prgAddr = ADDR_SRM
		prgSize := ((uint32)(len(rom)) - n.datBase) / FDS_DISK_SIZE * 0x10000
		if prgSize < (uint32)(len(rom)) {
			prgSize += 0x10000
		}
		n.srmSize = 32768
		n.mapper = 254
		n.prg = make([]uint8, prgSize)
		n.chr = make([]uint8, 0)
		var i uint32
		for i = 0; i < prgSize/0x10000; i++ {
			block := FDS_DISK_SIZE
			src := n.datBase + i*FDS_DISK_SIZE
			dst := i * 0x10000
			if src+block > (uint32)(len(rom)) {
				block = (uint32)(len(rom)) - src
			}
			copy(n.prg[dst:], rom[src:src+block])
		}
	default:
		return nil, fmt.Errorf("unknown ROM format")
	}

	crcLen := (uint32)(len(rom)) - n.datBase
	if crcLen > MAX_ID_CALC_LEN {
		crcLen = MAX_ID_CALC_LEN
	}
	n.crc = crc32.ChecksumIEEE(rom[n.datBase : n.datBase+crcLen])

	return n, nil
}

func (n *NesRom) Print() {
	fmt.Printf("Mapper   : %d\n", n.mapper)
	fmt.Printf("PRG SIZE : %dK (%d x 16K)\n", len(n.prg)/1024, len(n.prg)/1024/16)
	fmt.Printf("CHR SIZE : %dK (%d x 8K)\n", len(n.chr)/1024, len(n.chr)/1024/8)
	fmt.Printf("SRM SIZE : %dK\n", n.srmSize/1024)
	fmt.Printf("Mirroring: %c\n", n.mirroring)
	fmt.Printf("BAT RAM  : %s\n", boolToString(n.batRam))
	fmt.Printf("ROM ID   : 0x%08X\n", n.crc)
}

func boolToString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func (n *NesRom) GetPrgSize() uint32 {
	return (uint32)(len(n.prg))
}

func (n *NesRom) GetChrSize() uint32 {
	return (uint32)(len(n.chr))
}

func (n *NesRom) GetSrmSize() uint32 {
	return n.srmSize
}

func (n *NesRom) GetPrgAddr() uint32 {
	return n.prgAddr
}

func (n *NesRom) GetChrAddr() uint32 {
	return n.chrAddr
}

func (n *NesRom) GetMapper() uint8 {
	return n.mapper
}

func (n *NesRom) GetMirroring() uint8 {
	return n.mirroring
}

func (n *NesRom) GetType() uint32 {
	return n.romType
}

func (n *NesRom) GetName() string {
	return filepath.Base(n.romPath)
}

func (n *NesRom) GetPrgData() []uint8 {
	return n.prg
}

func (n *NesRom) GetChrData() []uint8 {
	return n.chr
}

func (n *NesRom) GetRomID() []uint8 {
	bin := make([]uint8, len(n.ines)+4*3)
	ptr := 0

	copy(bin[ptr:], n.ines)
	ptr += len(n.ines)

	binary.LittleEndian.PutUint32(bin[ptr:], uint32(n.size))
	ptr += 4
	binary.LittleEndian.PutUint32(bin[ptr:], uint32(n.crc))
	ptr += 4
	binary.LittleEndian.PutUint32(bin[ptr:], uint32(n.datBase))
	ptr += 4

	return bin
}
