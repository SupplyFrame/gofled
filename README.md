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