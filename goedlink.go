package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"forge.rights.ninja/jeff/goedlink/n8"
	"forge.rights.ninja/jeff/goedlink/nesrom"
)

var commands = map[string]func([]string){
	"appmode":     AppMode,
	"cp":          Copy,
	"info":        Info,
	"initfpga":    InitFpga,
	"getrtc":      GetRtc,
	"loadrom":     LoadRom,
	"mkdir":       MakeDirectory,
	"readmemory":  ReadMemory,
	"reboot":      Reboot,
	"recovery":    Recovery,
	"servicemode": ServiceMode,
	"setrtc":      SetRtc,
	"writeflash":  WriteFlash,
	"writememory": WriteMemory,
}

var N8 n8.N8

// AppMode switches the N8 out of service mode
func AppMode(args []string) {
	fs := flag.NewFlagSet("appmode", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		fmt.Println("[App Mode]")
		N8.ExitServiceMode()
		fmt.Println("[App Mode] ok")
		os.Exit(0)
	}

	fs.Usage()
}

// Copy copies a file on the N8.
//
// Prefix the source or destination string with `sd:` to specify a location on the N8 SD card.
func Copy(args []string) {
	fs := flag.NewFlagSet("copy", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	source := fs.String("source", "", "path to copy from, prefix with `sd:` for file on the SD card")
	destination := fs.String("destination", "", "path to copy to, prefix with `sd:` for destination on the SD card")
	fs.Parse(args)

	if *device != "" && *source != "" && *destination != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		N8.CopyFile(*source, *destination)
		fmt.Printf("[Copy] \"%s\" copied to \"%s\"\n", *source, *destination)
		os.Exit(0)
	}

	fs.Usage()
}

// Info prints the current configuration from the N8.
func Info(args []string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		N8.ExitServiceMode()
		config := N8.GetConfig()
		fmt.Println("[Info]")
		config.PrintFull()
		os.Exit(0)
	}

	fs.Usage()
}

// InitFpga initializes N8 FPGA.
//
// Reads FPGA init data from file and writes to FPGA.
func InitFpga(args []string) {
	fs := flag.NewFlagSet("initfpga", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	path := fs.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	length := fs.Int64("length", 0, "(required) number of bytes to read (eg. '0x40', '64', etc)")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		var buf []uint8
		if *path != "" {
			file, err := os.ReadFile(*path)
			if err != nil {
				log.Fatalf("[fpgaInit] error reading file %s: %v", *path, err)
			}
			buf = make([]uint8, len(file))
			copy(buf, file)
		} else {
			if *length == 0 {
				log.Fatalln("[fpgaInit] length must be specified when reading from standard input")
			}
			buf = make([]uint8, *length)
			reader := bufio.NewReader(os.Stdin)
			for {
				_, err := reader.Read(buf)

				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Println("[writeMemory] error reading input:", err)
					return
				}
			}
		}

		N8.FpgaInit(buf, nil)
		os.Exit(0)
	}

	fs.Usage()
}

// GetRtc returns the time currently set on the N8 RTC.
func GetRtc(args []string) {
	fs := flag.NewFlagSet("getrtc", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		rtc := N8.GetRtc()
		rtc.Print()
		os.Exit(0)
	}

	fs.Usage()
}

// LoadRom loads ROM or OS.
//
// Auto detects normal ROM or `ROM_TYPE_OS`, optionally loads
// provided mappe data, and starts the ROM. Prints MapConfig.
func LoadRom(args []string) {
	fs := flag.NewFlagSet("loadrom", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	romPath := fs.String("rom", "", "path to rom")
	mapPath := fs.String("map", "", "path to copy from, prefix with `sd:` for file on the SD card")
	fs.Parse(args)

	if *device != "" && *romPath != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		rom, err := nesrom.NewNesRom(*romPath)
		if err != nil {
			log.Fatalf("[loadRom] rom error: %v", err) // TODO: add a better error here
		}
		rom.Print()

		if rom.GetType() == nesrom.ROM_TYPE_OS {
			N8.LoadOS(rom, *mapPath)
		} else {
			N8.LoadGame(*romPath, *mapPath)
		}

		N8.GetConfig().Print()
		os.Exit(0)
	}

	fs.Usage()
}

// MakeDirectory creates a directory on the N8.
func MakeDirectory(args []string) {
	fs := flag.NewFlagSet("mkdir", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	path := fs.String("path", "", "directory to create on the SD card")
	fs.Parse(args)

	if *device != "" && *path != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		N8.MakeDir(*path)
		fmt.Printf("[mkdir] \"%s\" created \n", *path)
		os.Exit(0)
	}

	fs.Usage()
}

// ReadMemory reads data from memory address.
//
// Writes data to file if path specified, otherwise prints to standard output.
func ReadMemory(args []string) {
	fs := flag.NewFlagSet("readmemory", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	path := fs.String("path", "", "(optional) save data to a file (otherwise data is just printed to standard output)")
	address := fs.Uint64("address", 0, "(required) hex address to read from (eg. '0xa000', '40960', etc)")
	length := fs.Int64("length", 0, "(required) number of bytes to read (eg. '0x40', '64', etc)")
	fs.Parse(args)

	if *device != "" && *length != 0 {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		var buf []uint8 = make([]uint8, (uint32)(*length))
		N8.ReadMemory((uint32)(*address), buf, (uint32)(len(buf)))

		if *path == "" {
			fmt.Printf("[Read Memory]\n address $%04x-$%04x:\n", address, (uint32)(*address)+(uint32)(*length))

			for i, b := range buf {
				fmt.Printf(" %02x", b)
				if (i+1)%8 == 0 {
					fmt.Printf("  ")
				}
				if (i+1)%32 == 0 || i+1 == len(buf) {
					fmt.Printf("\n")
				}
			}
		} else {
			err := os.WriteFile(*path, buf, 0644)
			if err != nil {
				log.Fatalf("[readMemory] error writing to file %s: %v", *path, err)
			}
		}
		os.Exit(0)
	}

	fs.Usage()
}

// Reboot sends a reboot command to the N8.
func Reboot(args []string) {
	fs := flag.NewFlagSet("reboot", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		N8.Reboot()
		fmt.Println("[Reboot] N8 is rebooting")
		os.Exit(0)
	}

	fs.Usage()
}

// Recovery runs N8 recovery operation.
func Recovery(args []string) {
	fs := flag.NewFlagSet("recovery", flag.ExitOnError)
	_ = fs.Bool("h", false, "show"+fs.Name()+"command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		fmt.Println("[recovery] EDIO core recovery...")
		N8.Recovery()
		fmt.Println("[recovery] ok")
		os.Exit(0)
	}

	fs.Usage()
}

// ServiceMode switches the N8 to service mode.
func ServiceMode(args []string) {
	fs := flag.NewFlagSet("servicemode", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		fmt.Println("[Service Mode]")
		N8.EnterServiceMode()
		fmt.Println("[Service Mode] ok")
		os.Exit(0)
	}

	fs.Usage()
}

// SetRtc sets the N8 RTC.
//
// Defaults to current time unless user specifies a time.
func SetRtc(args []string) {
	fs := flag.NewFlagSet("setrtc", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	userTime := fs.String("time", time.Now().Format("2006-01-02 15:04:05"), "(optional) time (format `YYYY-MM-DD HH:mm:SS`)")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		t, err := time.Parse("2006-01-02 15:04:05", *userTime)
		if err != nil {
			log.Fatalf("[setrtc] error parsing time string %s: %v", *userTime, err)
		}
		N8.SetRtc(t)
		os.Exit(0)
	}

	fs.Usage()
}

// WriteFlash writes data to N8 flash.
//
// Reads data from file and writes to flash.
func WriteFlash(args []string) {
	fs := flag.NewFlagSet("writeflash", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, '/dev/ttyACMO0')")
	path := fs.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	address := fs.Uint64("address", 0, "(required) hex address to write to (eg. '0xa000', '40960', etc)")
	fs.Parse(args)

	if *device != "" {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		file, err := os.ReadFile(*path)
		if err != nil {
			log.Fatalf("[writeMemory] error reading file %s: %v", *path, err)
		}

		N8.WriteFlash((uint32)(*address), file, (uint32)(len(file)))
		os.Exit(0)
	}

	fs.Usage()
}

// writeMemory writes data to N8 memory.
//
// Reads data from file if path specified, otherwise reads from standard input.
func WriteMemory(args []string) {
	fs := flag.NewFlagSet("writememory", flag.ExitOnError)
	_ = fs.Bool("h", false, "show "+fs.Name()+" command help")
	device := fs.String("d", "", "serial device path (eg, /dev/ttyACMO0)")
	path := fs.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	address := fs.Uint64("address", 0, "(required) hex address to write to (eg. '0xa000', '40960', etc)")
	length := fs.Int64("length", 0, "(required) number of bytes to read (eg. '0x40', '64', etc)")
	fs.Parse(args)

	if *device != "" && *length != 0 {
		N8.InitSerial(*device, time.Second*2)
		defer N8.Port.Close()

		var buf []uint8 = make([]uint8, *length)

		if *path == "" {
			reader := bufio.NewReader(os.Stdin)
			for {
				_, err := reader.Read(buf)

				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Println("[writeMemory] error reading input:", err)
					return
				}
			}
		} else {
			file, err := os.ReadFile(*path)
			if err != nil {
				log.Fatalf("[writeMemory] error reading file %s: %v", *path, err)
			}
			copy(buf, file)
		}
		N8.WriteMemory((uint32)(*address), buf, (uint32)(len(buf)))
		// TODO: some sort of output with details of the write operation.

		os.Exit(0)
	}

	fs.Usage()
}

func main() {
	if len(os.Args) == 1 {
		helptext()
		os.Exit(1)
	}

	cmd, ok := commands[os.Args[1]]
	if !ok {
		helptext()
		os.Exit(1)
	}
	cmd(os.Args[2:])

}

func helptext() {
	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	s := "Usage: goedlink [command] [options]\nAvailable commands:\n"
	for _, k := range keys {
		s += "  goedlink " + k + "\n"
	}
	s += "\nShow subcommand help:\n  goedlink [command] -h\n\n---\n"

	fmt.Println(s)

	AppMode(nil)
	Copy(nil)
	Info(nil)
	InitFpga(nil)
	GetRtc(nil)
	LoadRom(nil)
	MakeDirectory(nil)
	ReadMemory(nil)
	Reboot(nil)
	Recovery(nil)
	ServiceMode(nil)
	SetRtc(nil)
	WriteFlash(nil)
	WriteMemory(nil)
}
