# Hypnos

Native Windows tray app that sends `F24` every 59 seconds.

Hypnos exists for rough workdays when you need a little room to breathe.
It keeps your Windows session from looking idle while you step away, wind down, stretch, get water, or take a short reset without fighting aggressive AFK checks.

The app runs in the Windows tray. Right-click the icon for `Turn on permanent`, `Turn off`, `Timer`, `Settings`, `About it`, or `Exit`.

## Build

```powershell
.\build.ps1
```

The build output is:

```text
dist\Hypnos.exe
```

If Go is not installed, `build.ps1` downloads a portable Go toolchain into `.tools`. It is only needed for building.

## Start

```powershell
.\dist\Hypnos.exe
```

## Features

- single Windows exe
- no installation for the user
- no admin rights
- no external runtime like Python or .NET
- sends `F24` every 59 seconds
- tray icon only
- quick timers from 30 minutes to 3 hours
- settings with exact time
- optional shutdown when the selected end time is reached
