# goedlink—cross-platform EverDrive N8 Pro serial control software

![](https://img.shields.io/gitea/v/release/jeff/goedlink?gitea_url=https%3A%2F%2Fforge.rights.ninja)

*Note: this tool is in a very early dev state! Core functionality is working, but this has not been widely tested yet.*

The EverDrive N8 Pro comes with a USB port for software development and system updates over serial, but the client provided by krikzz doesn't play nice with Linux. This tool aims to re-implement the functionality in golang so as to be easily ported to a wide variety of operating systems.

- [Getting Started](#getting-started)
- [Current Functionality](#current-functionality)
- [Build](#build)
- [License](#license)
- [Change Log](#change-log)

## Getting Started

See [below](#build) for instructions on building or download the latest binary for your OS [here](https://forge.rights.ninja/jeff/goedlink/releases/latest).

Plug in your N8 Pro and identify the serial device (see your OS documentation if you're unsure of how to confirm this).

| OS | Likely device names |
| --- | --- |
| Linux | `/dev/ttyACM*`, `/dev/ttyUSB*` |
| FreeBSD | `/dev/cuaU*`, `/dev/ttyU*` |
| macOS | `/dev/tty.*`, `/dev/cu.*` |
| Windows | `COM*` |

You can confirm that everything is working by running the info command:

```sh
goedlink info -d <device_name>
```

For example, on linux:

```sh
$ goedlink-0.1-linux-amd64 info -d /dev/ttyACM0
[Info]
 mapper.....255 sub.0
 prg size....8K
 chr size....8K 
 srm size....128B 
 master vol..0
 mirroring...h
 cfg bits....00000000
 menu key....0x00
 save key....0x00
 load key....0x00
 rst delay...no
 save state..no
 ss button...no
 unlock......yes
 ctrl bits...10000000
 CFG0: ff00000000000080
 CFG1: 0000000000000000
```

## Current Functionality

*Important: this tool is still largely untested!*

```
Usage: goedlink [command] [options]
Available commands:
  goedlink appmode
  goedlink cp
  goedlink getrtc
  goedlink info
  goedlink initfpga
  goedlink loadrom
  goedlink mkdir
  goedlink readmemory
  goedlink reboot
  goedlink recovery
  goedlink servicemode
  goedlink setrtc
  goedlink writeflash
  goedlink writememory

Show subcommand help:
  goedlink [command] -h

---

Usage of appmode:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show appmode command help
Usage of copy:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -destination sd:
        path to copy to, prefix with sd: for destination on the SD card
  -h    show copy command help
  -source sd:
        path to copy from, prefix with sd: for file on the SD card
Usage of info:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show info command help
Usage of initfpga:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show initfpga command help
  -length int
        (required) number of bytes to read (eg. '0x40', '64', etc)
  -path string
        (optional) read data from a file (otherwise data is read from standard input)
Usage of getrtc:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show getrtc command help
Usage of loadrom:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show loadrom command help
  -map sd:
        path to copy from, prefix with sd: for file on the SD card
  -rom string
        path to rom
Usage of mkdir:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show mkdir command help
  -path string
        directory to create on the SD card
Usage of readmemory:
  -address uint
        (required) hex address to read from (eg. '0xa000', '40960', etc)
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show readmemory command help
  -length int
        (required) number of bytes to read (eg. '0x40', '64', etc)
  -path string
        (optional) save data to a file (otherwise data is just printed to standard output)
Usage of reboot:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show reboot command help
Usage of recovery:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    showrecoverycommand help
Usage of servicemode:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show servicemode command help
Usage of setrtc:
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show setrtc command help
  -time YYYY-MM-DD HH:mm:SS
        (optional) time (format YYYY-MM-DD HH:mm:SS) (default "2024-07-08 17:46:01")
Usage of writeflash:
  -address uint
        (required) hex address to write to (eg. '0xa000', '40960', etc)
  -d string
        serial device path (eg, '/dev/ttyACMO0')
  -h    show writeflash command help
  -path string
        (optional) read data from a file (otherwise data is read from standard input)
Usage of writememory:
  -address uint
        (required) hex address to write to (eg. '0xa000', '40960', etc)
  -d string
        serial device path (eg, /dev/ttyACMO0)
  -h    show writememory command help
  -length int
        (required) number of bytes to read (eg. '0x40', '64', etc)
  -path string
        (optional) read data from a file (otherwise data is read from standard input)
```

## Build

To manually build, ensure you're running a compatible version of golang and run:

```sh
go mod tidy
go build -o goedlink-linux-amd64
```

On macOS, this likely requires `CGO_ENABLED=1`:

```sh
go mod tidy
CGO_ENABLED=1 go build -o goedlink-linux-amd64
```

## License

[GNU General Public License, version 2](LICENSE.md)

## Change Log

- `0.1` initial proof-of-concept
