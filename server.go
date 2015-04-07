package main

import (
	"fmt"
	"time"
	"runtime"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
	"net/http"
	"net"
	"encoding/binary"
	"os"
	"html/template"
	"encoding/json"
	"strconv"
	"os/signal"
)

var (
	b *Broker
	blender *Blender
	ledWidth = 52
	ledHeight = 34
	numLEDs = ledWidth * ledHeight
	addr = ":9000"
	tcpAddr = ":9001"
	quit = make(chan int)
	quitting = false
	upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

const (
	TCP_READ_LEN	= 0 // waiting for header
	TCP_READ_DATA	= 1 // reading packet data
)

func XY(pos int) (x int, y int) {
	ledPos := pos / 3
	x = ledPos % ledWidth
	y = ledPos / ledWidth
	return x,y
}
func POS(x int, y int) (pos int) {
	pos = (y*ledWidth + x) * 3
	return pos
}

func tcpHandler(conn net.Conn) {
	state := TCP_READ_LEN
	lenBuff := make([]byte, 0, 4)
	var dataBuff []byte

	src := NewFrameSource(numLEDs, 10, conn.RemoteAddr())

	// wait for 10 seconds before activating the source to allow TCP buffer to build up
	time.Sleep(10*time.Second)

	blender.joining <- src

	fmt.Println("Active source initialized")

	defer func() {
		// cleanup the socket and remove the source from our data set
		blender.leaving <- src
	}()

	start := time.Now()

	for {
		// create a buffer and read some bytes from the connection
		data := make([]byte, 8192, 8192)
		bytesReceived, err := conn.Read(data)
		// bail on any error from the connection
		if err != nil {
			fmt.Println("ERROR : ", err.Error())
			conn.Close()
			break;
		}
		// loop over all the bytes we received
		for ; bytesReceived > 0; {
			// are we waiting 
			if state == TCP_READ_LEN {
				// how many more bytes do we need in our lenbuff?
				lenRemaining := cap(lenBuff) - len(lenBuff)
				if lenRemaining > bytesReceived {
					lenRemaining = bytesReceived
				}
				// slice off just the number of bytes we need from data and append to the lenBuff
				lenBuff = append(lenBuff, data[:lenRemaining]...)
				// reset data to point to the remaining bytes
				data = data[lenRemaining:]
				bytesReceived = bytesReceived - lenRemaining
				// is the len buffer now full?
				if len(lenBuff) == cap(lenBuff) {
					// received the whole string, conver the byte array to a uint32
					packetLen := binary.LittleEndian.Uint32(lenBuff)
					// reset len buffer for the next time through
					lenBuff = make([]byte, 0, 4)
					// setup data buffer with appropriate capacity
					dataBuff = make([]byte, 0, packetLen)
					// change state
					state = TCP_READ_DATA
				}
			}
			// we're reading the data part of the packet
			if state == TCP_READ_DATA {
				// how many more bytes do we need in our data array?
				dataRemaining := cap(dataBuff) - len(dataBuff)
				if dataRemaining > bytesReceived {
					dataRemaining = bytesReceived
				}
				// slice off just what we need
				dataBuff = append(dataBuff, data[:dataRemaining]...)
				// reset data buffer to point to whats left
				data = data[dataRemaining:]
				bytesReceived = bytesReceived - dataRemaining

				// is the buffer full?
				if len(dataBuff) == cap(dataBuff) {
					// all data received
					//fmt.Println("Received all data :", dataBuff)
					state = TCP_READ_LEN

					// parse out the command bytes
					cmdBytes := dataBuff[:4]
					cmd := binary.LittleEndian.Uint32(cmdBytes)
					
					// send the command byte and the data slice into the parser
					src.ParseCommand(Command(cmd), dataBuff[4:])

					// frame received...how many milliseconds have elapsed?
					elapsed := time.Since(start)
					// calculate duration of sleep based on framesource fps
					sleepMs := (1000.0 / float64(src.fps)) - elapsed.Seconds()*1000.0
					// sleep so that we rate limit the connection to max 150fps
					time.Sleep(time.Duration(sleepMs)*time.Millisecond);
					start = time.Now()
				}
			}
		}
	}
	fmt.Println("Connection finished.")
}


// implement a tcp socket that receives our framedata protocol
func StartTCP() {
	l, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Listening for TCP on ", tcpAddr)

	defer l.Close()

	for {
		// accept a connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		fmt.Printf("Received connection %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
		// spawn a process to handle this connection
		go tcpHandler(conn)
	}
}


// implements a websocket source, so that browser apps can send data to the system as any other source
func sourceHandler(w http.ResponseWriter, r *http.Request) {
	c, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Close notification unsupported!\n", http.StatusInternalServerError)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			return
		}
		fmt.Println("Websocket upgrade failed: %s\n", err)
	}

	// create a source for this socket
	src := NewFrameSource(numLEDs, 10, ws.RemoteAddr())

	defer func() {
		ws.Close()
		blender.leaving <- src
	}()
	blender.joining <- src

	closer := c.CloseNotify()

	for {
		// read messages from websocket
		messageType, msg, err := ws.ReadMessage()
		if (err != nil) {
			fmt.Println("Error : ", err)
			return
		}
		// check type of message
		// if message type is binary assume it is frame data
		if messageType == websocket.BinaryMessage {
			// its a binary message, it must be frame data, parse it into a struct of frame data
			fmt.Println("Binary message received.")
			// read first 4 bytes into uint32 as command bit
			cmdBytes := msg[:4]
			cmd := binary.LittleEndian.Uint32(cmdBytes)
			
			// send the command byte and the data slice into the parser
			src.ParseCommand(Command(cmd), msg[4:])

			// pause to achieve 30fps
			time.Sleep(33*time.Millisecond);
		} else {
			// its a text message, echo it out for now
			fmt.Println("Text message : ", msg)
		}

		if <-closer {
			fmt.Println("Closing connection\n")
			return
		}
	}
}

type connection struct {
	ws *websocket.Conn
	send chan *Message
	broker *Broker
}

func (c *connection) reader() {
	for {
		messageType, r, err := c.ws.NextReader()
		if err != nil {
			fmt.Println("Error from NextReader :", err.Error())
			break
		}

		if messageType== websocket.CloseMessage {
			fmt.Println("Close Message")
			break
		}
		if messageType != websocket.TextMessage {
			// skip all other message types except TextMessage
			continue
		}
		
		var message map[string] interface{}

		dec := json.NewDecoder(r)
		err = dec.Decode(&message)
		if err != nil {
			fmt.Println("Failed to unmarshal message content : ", err.Error())
			continue
		}
		if message["Type"] == "settings" {
			// cast data to map
			settings := message["Data"].(map[string] interface{})
			// settings received!
			if val, ok := settings["brightness"]; ok {
				brightness := val.(float64)
				blender.brightness = brightness
			}
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for m := range c.send {
		
		// send a message indicating what source the next message goes to
		b, err := json.Marshal(m)
		if err != nil {
			fmt.Println("Failed to encode json : ", err.Error())
			break
		}
		err = c.ws.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			fmt.Println("Error sending message : ", err.Error())
			break
		}	

		if m.Type == "data" {
			err = c.ws.WriteMessage(websocket.BinaryMessage, m.Body)
			if err != nil {
				fmt.Println("Error sending message : ", err.Error())
				break
			}	
		}
	}
	c.ws.Close()
}
type wsHandler struct {
	broker *Broker
}
func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading websocket : ", err.Error())
		return
	}
	c := &connection{send: make(chan *Message), ws:ws, broker:wsh.broker}
	c.broker.joining <- c.send
	defer func() {
		c.broker.leaving <- c.send
	}()

	go c.writer()
	c.reader()
}

func sender() {
	for {
		// queue up a message via the broker
		b.messages <- &Message{ID: "active", Type: "data", Body: blender.GetBuffer()}
		time.Sleep(33*time.Millisecond) // roughly 30fps
	}
}
func sourceSender() {
	for {
		// loop over all sources and send out the raw data from each
		for _,src := range blender.sources {
			b.messages <- &Message{ID: strconv.Itoa(src.ID), Type: "data", Body: src.current}
		}
		time.Sleep(100*time.Millisecond) // 10fps
	}
}
func MaxParallelism() int {
    maxProcs := runtime.GOMAXPROCS(0)
    numCPU := runtime.NumCPU()
    if maxProcs < numCPU {
        return maxProcs
    }
    return numCPU
}
func indexHandler(w http.ResponseWriter, req *http.Request) {
	var indexTemplate = template.Must(template.ParseFiles(
		"templates/_base.html",
		"templates/index.html",
	))
	// render out a list of all sources with UI capable of rendering
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func sourcesHandler(w http.ResponseWriter, req *http.Request) {
	var sourcesTemplate = template.Must(template.ParseFiles(
		"templates/_base.html",
		"templates/sources.html",
	))
	// render out a list of all sources with UI capable of rendering
	if err := sourcesTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func settingsHandler(w http.ResponseWriter, req *http.Request) {
	var settingsTemplate = template.Must(template.ParseFiles(
		"templates/_base.html",
		"templates/settings.html",
	))
	// render out a list of all sources with UI capable of rendering
	if err := settingsTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func main() {
	numCPU := MaxParallelism()
	fmt.Println("MAX CPUS : ", numCPU)
	runtime.GOMAXPROCS(numCPU)

	// setup quit channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			if !quitting {
				fmt.Println("Exiting system...")
				close(quit)
				quitting = true
				// allow everything some time to exit
				time.Sleep(2*time.Second)
				fmt.Println("done.")
    			os.Exit(1)       
			}
		}
	}()

	b = NewBroker()
	b.Start()

	blender = NewBlender(numLEDs,b)
	blender.Start()

	// setup a router and some handlers
	r := mux.NewRouter()
	r.HandleFunc("/sources", sourcesHandler)
	r.HandleFunc("/source", sourceHandler)
	r.HandleFunc("/settings", settingsHandler)
	r.Handle("/client", wsHandler{broker: b})
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	r.HandleFunc("/", indexHandler)

	go Renderer(numLEDs, blender)

	go StartTCP()
	go sender()
	go sourceSender()

	fmt.Println("Starting webserver")
	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(addr)
	
//	time.Sleep(10*time.Second)
	fmt.Println("Exiting")
}
