package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"forge.rights.ninja/jeff/goedlink/n8"
	"forge.rights.ninja/jeff/goedlink/nesrom"
)

var N8 n8.N8

// copyFile copies a file on the N8.
//
// Prefix the source or destination string with `sd:` to specify a location on the N8 SD card.
func copyFile(source string, destination string) {
	N8.CopyFile(source, destination)
	fmt.Printf("[Copy File] \"%s\" copied to \"%s\"\n", source, destination)
}

// info prints the current configuration from the N8.
func info() {
	N8.ExitServiceMode()
	config := N8.GetConfig()
	fmt.Println("[Info]")
	config.PrintFull()
}

// appmode switches the N8 out of service mode
func appmode() {
	fmt.Println("[App Mode]")
	N8.ExitServiceMode()
	fmt.Println("[App Mode] ok")

}

// reboot sends a reboot command to the N8.
func reboot() {
	N8.Reboot()
	fmt.Println("[Reboot] N8 is rebooting")
}

// servicemode switches the N8 to service mode.
func servicemode() {
	fmt.Println("[Service Mode]")
	N8.EnterServiceMode()
	fmt.Println("[Service Mode] ok")

}

// mkdir creates a directory on the N8.
func mkdir(path string) {
	N8.MakeDir(path)
	fmt.Printf("[mkdir] \"%s\" created \n", path)
}

// recovery runs N8 recovery operation.
func recovery() {
	fmt.Println("[recovery] EDIO core recovery...")
	N8.Recovery()
	fmt.Println("[recovery] ok")
}

// loadRom loads ROM or OS.
//
// Auto detects normal ROM or `ROM_TYPE_OS`, optionally loads
// provided mappe data, and starts the ROM. Prints MapConfig.
func loadRom(romPath string, mapPath string) {
	rom, err := nesrom.NewNesRom(romPath)
	if err != nil {
		log.Fatalf("[loadRom] rom error: %v", err) // TODO: add a better error here
	}
	rom.Print()

	if rom.GetType() == nesrom.ROM_TYPE_OS {
		N8.LoadOS(rom, mapPath)
	} else {
		N8.LoadNewGame(romPath, mapPath)
	}

	N8.GetConfig().Print()
}

// fpgaInit initializes N8 FPGA.
//
// Reads FPGA init data from file and writes to FPGA.
func fpgaInit(path string, length uint32) {
	var buf []uint8

	if path != "" {
		file, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("[fpgaInit] error reading file %s: %v", path, err)
		}
		buf = make([]uint8, len(file))
		copy(buf, file)
	} else {
		if length == 0 {
			log.Fatalln("[fpgaInit] length must be specified when reading from standard input")
		}
		buf = make([]uint8, length)
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
}

// writeFlash writes data to N8 flash.
//
// Reads data from file and writes to flash.
func writeFlash(path string, address uint32) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[writeMemory] error reading file %s: %v", path, err)
	}

	N8.WriteFlash(address, file, (uint32)(len(file)))
}

// writeMemory writes data to N8 memory.
//
// Reads data from file if path specified, otherwise reads from standard input.
func writeMemory(path string, address uint32, length uint32) {
	var buf []uint8 = make([]uint8, length)

	if path == "" {
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
		file, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("[writeMemory] error reading file %s: %v", path, err)
		}
		copy(buf, file)
	}
	N8.WriteMemory(address, buf, (uint32)(len(buf)))
	// TODO: some sort of output with details of the write operation.
}

// readMemory reads data from memory address.
//
// Writes data to file if path specified, otherwise prints to standard output.
func readMemory(path string, address uint32, length uint32) {
	var buf []uint8 = make([]uint8, length)
	N8.ReadMemory(address, buf, (uint32)(len(buf)))

	if path == "" {
		fmt.Printf("[Read Memory]\n address $%04x-$%04x:\n", address, address+length)

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
		err := os.WriteFile(path, buf, 0644)
		if err != nil {
			log.Fatalf("[readMemory] error writing to file %s: %v", path, err)
		}
	}

	os.Exit(0)
}

// printUsage prints the command usage.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n%s -d <device> [command]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  appmode\n")
	fmt.Fprintf(os.Stderr, "        leave service mode\n")
	fmt.Fprintf(os.Stderr, "  cp --source <source> --destination <destination>\n")
	fmt.Fprintf(os.Stderr, "        copy files, prefix with `sd:` to copy to/from SD card\n")
	fmt.Fprintf(os.Stderr, "  initfpga --path <path>\n")
	fmt.Fprintf(os.Stderr, "        initialize FPGA\n")
	fmt.Fprintf(os.Stderr, "  getrtc\n")
	fmt.Fprintf(os.Stderr, "        get real-time clock\n")
	fmt.Fprintf(os.Stderr, "  loadrom --rom <path_to_rom> [--map <path_to_rom>]\n")
	fmt.Fprintf(os.Stderr, "        load ROM and (optionally) mapper\n")
	fmt.Fprintf(os.Stderr, "  mkdir --path <path> \n")
	fmt.Fprintf(os.Stderr, "        make directory\n")
	fmt.Fprintf(os.Stderr, "  readmemory --address <address> --length <length> [--path <path_to_file>]\n")
	fmt.Fprintf(os.Stderr, "        read memory standard output or to file\n")
	fmt.Fprintf(os.Stderr, "  reboot\n")
	fmt.Fprintf(os.Stderr, "        send reboot command\n")
	fmt.Fprintf(os.Stderr, "  recovery\n")
	fmt.Fprintf(os.Stderr, "        run EDIO recovery\n")
	fmt.Fprintf(os.Stderr, "  servicemode\n")
	fmt.Fprintf(os.Stderr, "        enter service mode\n")
	fmt.Fprintf(os.Stderr, "  setrtc [--time <time>] \n")
	fmt.Fprintf(os.Stderr, "        set real-time clock, specify time in format `YYYY-DD-MM HH:mm:SS`, defaults to `now`\n")
	fmt.Fprintf(os.Stderr, "  writeflash --address <address> --length <length> [--path <path_to_file>]\n")
	fmt.Fprintf(os.Stderr, "        write standard input or file to flash\n")
	fmt.Fprintf(os.Stderr, "  writememory --address <address> --length <length> [--path <path_to_file>]\n")
	fmt.Fprintf(os.Stderr, "        write standard input or file to  memory\n")
}

// getDevice set the N8 address specified by the `--device` flag.
//
// SetReturns only the remaining args, striping out `--device` and
// its argument.
func getDevice(args []string) []string {
	var result []string
	skipNext := false

	for _, arg := range args {
		if skipNext {
			N8.Address = arg
			skipNext = false
			continue
		}
		if arg == "--device" {
			skipNext = true
			continue
		}
		result = append(result, arg)
	}

	return result
}

func main() {

	//
	// arg and flag setup
	//
	args := getDevice(os.Args)

	setRtcCmd := flag.NewFlagSet("setrtc", flag.ExitOnError)
	setRtcTime := setRtcCmd.String("time", time.Now().Format("2006-01-02 15:04:05"), "(optional) time (format `YYYY-MM-DD HH:mm:SS`)")

	writeFlashCmd := flag.NewFlagSet("writeflash", flag.ExitOnError)
	writeFlashPath := writeFlashCmd.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	writeFlashAddress := writeFlashCmd.Uint64("address", 0, "(required) hex address to write to (eg. `0xa000`, `40960`, etc)")

	writememoryCmd := flag.NewFlagSet("writememory", flag.ExitOnError)
	writeMemoryPath := writememoryCmd.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	writeMemoryAddress := writememoryCmd.Uint64("address", 0, "(required) hex address to write to (eg. `0xa000`, `40960`, etc)")
	writeMemoryLength := writememoryCmd.Int64("length", 0, "(required) number of bytes to read (eg. `0x40`, `64`, etc)")

	readmemoryCmd := flag.NewFlagSet("readmemory", flag.ExitOnError)
	readMemoryPath := readmemoryCmd.String("path", "", "(optional) save data to a file (otherwise data is just printed to standard output)")
	readMemoryAddress := readmemoryCmd.Uint64("address", 0, "(required) hex address to read from (eg. `0xa000`, `40960`, etc)")
	readMemoryLength := readmemoryCmd.Int64("length", 0, "(required) number of bytes to read (eg. `0x40`, `64`, etc)")

	fpgaInitCmd := flag.NewFlagSet("initfpga	", flag.ExitOnError)
	fpgaInitPath := fpgaInitCmd.String("path", "", "(optional) read data from a file (otherwise data is read from standard input)")
	fpgaInitLength := fpgaInitCmd.Int64("length", 0, "(required) number of bytes to read (eg. `0x40`, `64`, etc)")

	cpCmd := flag.NewFlagSet("cp", flag.ExitOnError)
	cpSource := cpCmd.String("source", "", "path to copy from, prefix with `sd:` for file on the SD card")
	cpDestination := cpCmd.String("destination", "", "path to copy to, prefix with `sd:` for destination on the SD card")

	mkdirCmd := flag.NewFlagSet("mkdir", flag.ExitOnError)
	mkdirPath := mkdirCmd.String("path", "", "directory to create on the SD card")

	loadRomCmd := flag.NewFlagSet("loadrom", flag.ExitOnError)
	romPath := loadRomCmd.String("rom", "", "path to rom")
	mapPath := loadRomCmd.String("map", "", "path to copy from, prefix with `sd:` for file on the SD card")

	flag.Usage = printUsage

	//
	// initial validation
	//
	if N8.Address == "" {
		fmt.Println("device is required")
		flag.Usage()
		os.Exit(1)
	}
	if len(args) < 2 {
		fmt.Println("expected command")
		flag.Usage()
		os.Exit(1)
	}

	//
	// serial init
	//
	N8.InitSerial(time.Second * 2)
	defer N8.Port.Close()

	//
	// command handling
	//
	switch args[1] {
	case "reboot":
		reboot()
		os.Exit(0)
	case "info":
		info()
		os.Exit(0)
	case "appmode":
		appmode()
		os.Exit(0)
	case "servicemode":
		servicemode()
		os.Exit(0)
	case "recovery":
		recovery()
		os.Exit(0)
	case "mkdir":
		mkdirCmd.Parse(args[2:])
		if *mkdirPath == "" {
			log.Fatalln("path required")
		}
		mkdir(*mkdirPath)
		os.Exit(0)
	case "cp":
		cpCmd.Parse(args[2:])

		if *cpSource == "" || *cpDestination == "" {
			log.Fatalln("source and destination required")
		}
		copyFile(*cpSource, *cpDestination)
		os.Exit(0)
	case "loadrom":
		loadRomCmd.Parse(args[2:])

		if *romPath == "" {
			log.Fatalln("rom path required")
		}
		loadRom(*romPath, *mapPath)
		os.Exit(0)
	case "initfpga":
		fpgaInitCmd.Parse(args[2:])
		fpgaInit(*fpgaInitPath, (uint32)(*fpgaInitLength))
		os.Exit(0)
	case "readmemory":
		readmemoryCmd.Parse(args[2:])

		if *readMemoryLength == 0 {
			fmt.Println("missing required field(s)")
			readmemoryCmd.Usage()
			os.Exit(1)
		}

		readMemory(*readMemoryPath, (uint32)(*readMemoryAddress), (uint32)(*readMemoryLength))
		os.Exit(0)
	case "writeflash":
		writememoryCmd.Parse(args[2:])
		writeFlash(*writeFlashPath, (uint32)(*writeFlashAddress))
		os.Exit(0)
	case "writememory":
		writememoryCmd.Parse(args[2:])

		if *writeMemoryLength == 0 {
			fmt.Println("missing required field(s)")
			readmemoryCmd.Usage()
			os.Exit(1)
		}

		writeMemory(*writeMemoryPath, (uint32)(*writeMemoryAddress), (uint32)(*writeMemoryLength))
		os.Exit(0)
	case "setrtc":
		setRtcCmd.Parse(args[2:])

		time, err := time.Parse("2006-01-02 15:04:05", *setRtcTime)
		if err != nil {
			log.Fatalf("[setrtc] error parsing time string %s: %v", *setRtcTime, err)
		}
		N8.SetRtc(time)
		os.Exit(0)
	case "getrtc":
		rtc := N8.GetRtc()
		rtc.Print()
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(1)
	}
}
