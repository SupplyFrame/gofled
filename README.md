GoFLED
======

A Fantastic LED project in Go.

Installation
============

./build-pru.sh
go build

Running
=======
sudo LD_LIBRARY_PATH=. ./gofled

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

To setup an Inactive Source, first send the SetActive(False) command, then start sending data.

To bring a source to the top and display it temporarily over all others call SetOverlay(duration)

