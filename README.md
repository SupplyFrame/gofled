Go based FLED interface
===============================

Responsibilities
	* Sending LED data to strip
		- this should run in a channel so we can push a slice of data into it and have it send in the background

	* Basic UI for managing which incoming data streams get displayed
		- multiple sockets can be receiving data at once
		- display usually only selects the first connected stream and displays this
		- UI allows you to see what is in all the streams via a Javascript UI with a websocket for data
		- UI allows you to select from the streams to display on the LEDs

	* A stream can configure various properties when it is setup or later via REST interface
		- transitions can be signalled by the stream data including a special code, when detected FLED will hold the last animation frame and blend it into subsequent frames based on a selected transition from the available transition effects (wipe, slide, zoom, blend, decimate, etc)
		- alternatively no transition signals will be sent by client and the client will handle its own transitions
		- buffer length can be set by the stream or REST protocol, this defines how many frames to buffer before showing animations otherwise a default of 10 frames will be buffered


Implementation
==============

Sources connect on websocket at /source
After handshaking to websocket protocol they are sent a UUID identifying that source
This UUID is used in other actions to modify/listen to the source

Broker manages communications

Everytime a 'frame' comes in to a source socket we store it in the sources buffer of frames
Source buffer of frames is a buffered channel of size X, specified at source creation time

Each frame event from socket, pushes onto buffered channel frames, blocking if buffer is full, if blocking then we need to keep reading from the ws and discarding data

Renderer - this pulls a frame from the active source, sending it over the spi port

SourceHandler - receives a websocket frame, pushes it onto frame buffered channel

ClientHandler - this is a websocket that comes from a client wishing to preview a source, this peeks at the current frame from the buffer

handlers:

/source 	- upgrades to websocket
			- stores reference to ws and notifies Broker that a new source is available
			- 



Sources need to push frames onto a queue
	Every time we push a frame onto the queue
		if the queue is full, we pop a frame from the queue and that becomes our current active frame, then we enqueue the new frame
		if the queue is not full we push the frame onto the queue
		a frame of data could be indicated as a 'transition start'
			if the transition start is set, then the Source grabs a reference to the frame
			subsequent frames are then processed by a transition function before being stored in the queue.
			on every subsequent frame after transition start we must move transition start to the next frame in the queue
			thereby allowing cumulative transitions to take effect
			this allows transitions to be implemented in a consistent way even if a source is rendering multiple different animations.
	Access to current active frame should be mutexed or through a channel somehow

A Source object manages the queue, the current active frame, and a websocket that is the source of data

The renderer keeps track of current active source and last source.
	When transition between sources is in effect, renderer passes the latest frame from both last source and active source to the transition method.
		Transition method blends between the two sources, returning a buffer of frame data as a result which is what is rendered
	When no transition is in effect then renderer pulls the latest frame from the active Source and sends it over SPI

	Renderer runs as fast as possible and records framerate for querying later

Clients are websockets that have been connected.
	Each client represents a web browser that is interested in one or more Sources.

/sources handler renders a page that contains javascript that connects to the /clients websocket
	javascript queries for a list of available sources
		websocket responds with UUID list of each source, along with information about where it came from etc
	javascript then subscribes to the sources its interested in with websocket command
		websocket now sends data for each source registered 



First thing, setup a source.

	Source receives frames in websocket messages
	It pushes the frames onto a buffered channel to maintain a blocking queue
		When pushing use a select statement to determine if it will block



If a client connects by tcp or websocket we create a FrameSource for them and start the reader process after allocating an appropriate reader func
	
	reader func simply handles communication with the socket or tcp conn creating Command objects
	Command objects are passed to ProcessCommand function on FrameSource which switches appropriately:

	Commands contain the read data packet and a type indicating what the data packet is
	Command.FRAME - a packet of frame data, this should be added to the source
	Command.TRANSITION - source is requesting that a transition be rendered into the source
	Command.ATTENTION - source is requesting that it become the focus of attention (useful for alerts, use wisely)
	Command.CLOSING - notifying the renderer that we're about to close this connection it triggers a transition and forces the renderer to choose a new source

	ProcessCommand pushes data onto the FrameSource, or sends out signals on the 'out' chan which indicate state changes to the renderer manager


Blender - picks a source and renders it to the activeSource
		- can have multiple sources selected at once
		- various algorithms should allow different sources to display
		- for instance a source could issue an 'overlay' command, making it lay over the top of the current activeSource
		- we could transition between sources, blending back and forth between them