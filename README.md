# goedlink—cross-platform EverDrive N8 Pro serial control software

*Note: this tool is in a very early dev state!*

The EverDrive N8 Pro comes with a USB port for software development and system updates over serial, but the client provided by krikzz doesn't play nice with Linux. This tool aims to re-implement the functionality in golang so as to be easily ported to a wide variety of operating systems.

- [Current Functionality](#current-functionality)
- [Change Log](#change-log)

# Current Functionality

*Note: while this tool aims to eventually implement all functionality from edlink, it currently only supports a very small subset*

| Name | Description |
| --- | --- |
| print config | retrieve and print the config data |

Currently, this means you only print the following data (values may differ):

```
mapper.....22 sub.0
prg size....32K
chr size....32K 
srm size....8K 
master vol..128
mirroring...h
cfg bits....00000000
menu key....0xFF
save key....0x00
load key....0x00
rst delay...no
save state..yes
ss button...yes
unlock......yes
ctrl bits...10011110
CFG0: 166202800000009e
CFG1: ffffffffffffffff
```

# Change Log

- `0.1` initial proof-of-concept