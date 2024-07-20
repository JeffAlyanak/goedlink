# goedlink—cross-platform EverDrive N8 Pro serial control software

*Note: this tool is in a very early dev state! Most functionality is working, with the most notable issue being that the "select game" command not working correctly.*

The EverDrive N8 Pro comes with a USB port for software development and system updates over serial, but the client provided by krikzz doesn't play nice with Linux. This tool aims to re-implement the functionality in golang so as to be easily ported to a wide variety of operating systems.

- [Current Functionality](#current-functionality)
- [Change Log](#change-log)

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

## License

[GNU General Public License, version 2](LICENSE.md)

## Change Log

- `0.1` initial proof-of-concept
