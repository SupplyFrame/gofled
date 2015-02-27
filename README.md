GoFLED
======

A Fantastic LED project in Go.

Installation
============

```bash
./build-pru.sh
go build
```

Now you need to configure both the PRU-1 and the SPI-1. We're using the default Debian image on our BBB, this comes with the Adafruit SPI, and a default PRU-1 overlay.

```bash
echo ADAFRUIT-SPI1 > /sys/devices/bone_capemgr.*/slots
echo BB-BONE-PRU-01 > /sys/devices/bone_capemgr.*/slots
modprobe uio-pruss
```

These can be setup to activate automatically by modifying your uEnv.txt in the root of your SD card. The UIO-Pruss module can be enabled in /etc/modules.

Running
=======
Change to the directory where you cloned this repository then:

```bash
sudo LD_LIBRARY_PATH=. ./gofled
```

Theory
======
GoFLED is a Go based server application that can receive multiple simultaenous streams of RGB color information and render it out to APA102 LED strips over SPI.

It is primarily setup for LED matrix operation, but could be converted for other uses easily.

Each source of data sends RGB values as byte arrays over a TCP socket, protocol is a simple one consisting of a 32bit unsigned int indicating length of message, then a command byte, then the remainder of the message.

Source Controls
===============
Each source can control how it will be used by GoFLED.

Sources have 2 modes, Active and Inactive.

Active sources are up for use by the blender in any way it chooses. They can be selected in the UI, and can be automatically merged between them.

Inactive sources are ignored by the blender. They are shown in the live view of the UI, but cannot be merged in without specifically being enabled at source.
Inactive sources can be temporarily activated by sending an Overlay command, this has a max duration during which time the source will be added to the active set at the top of the stack and blended in using its blend mode.

All sources have a blend mode, defaulting to Add.

The simplest new source will just start sending frame data. It creates an active source by default and will be available for display by the blender.

To setup an Inactive Source, first send the active command, then start sending data.

To bring a source to the top and display it temporarily over all others call the overlay command.



Blending effects processing
===========================

Each source can be blended over other sources
Need a suite of blend functions for different types of blending
	These should be of the form func(dst byte[], src byte[], amount float64)
	In an additive blend amount would be applied as a scalar to src before blending allowing us to control how much of the src is applied in the function
Each frame source has a blend mode and an amount value (defaults to 1.0)
If we wish to blend in a source over time, we would move amount from 0.0 to 1.0

For overlays, we would blend them in over a short period and then blend them out over a short period.
