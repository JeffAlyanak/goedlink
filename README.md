# goedlink—cross-platform EverDrive N8 Pro serial control software

*Note: this tool is in a very early dev state!*

The EverDrive N8 Pro comes with a USB port for software development and system updates over serial, but the client provided by krikzz doesn't play nice with Linux. This tool aims to re-implement the functionality in golang so as to be easily ported to a wide variety of operating systems.

- [Current Functionality](#current-functionality)
- [Change Log](#change-log)

## Current Functionality

*Note: this tool is still largely untested!*

```
Usage of goedlink:

goedlink -d <device> [command]

Commands:
  appmode
        leave service mode
  cp --source <source> --destination <destination>
        copy files, prefix with `sd:` to copy to/from SD card
  initfpga --path <path>
        initialize FPGA
  getrtc
        get real-time clock
  loadrom --rom <path_to_rom> [--map <path_to_rom>]
        load ROM and (optionally) mapper
  mkdir --path <path> 
        make directory
  readmemory --address <address> --length <length> [--path <path_to_file>]
        read memory standard output or to file
  reboot
        send reboot command
  recovery
        run EDIO recovery
  servicemode
        enter service mode
  setrtc [--time <time>] 
        set real-time clock, specify time in format `YYYY-DD-MM HH:mm:SS`, defaults to `now`
  writeflash --address <address> --length <length> [--path <path_to_file>]
        write standard input or file to flash
  writememory --address <address> --length <length> [--path <path_to_file>]
        write standard input or file to  memory
```

## License

[GNU General Public License, version 2](LICENSE.md)

## Change Log

- `0.1` initial proof-of-concept
