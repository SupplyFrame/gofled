package main

import (
	"github.com/gorilla/websocket"
	"fmt"
	"encoding/json"
	"net/http"
)
type ClientWebsocket struct {
	ws *websocket.Conn
	send chan *Message
	broker *Broker
}

func (c *ClientWebsocket) reader() {
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

func (c *ClientWebsocket) writer() {
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
type ClientWebsocketHandler struct {
	broker *Broker
}
func (wsh ClientWebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading websocket : ", err.Error())
		return
	}
	c := &ClientWebsocket{send: make(chan *Message), ws:ws, broker:wsh.broker}
	c.broker.joining <- c.send
	defer func() {
		c.broker.leaving <- c.send
	}()


	go c.writer()

	// update the client with settings and sources
	blender.RefreshSources(c.send)
	blender.RefreshSettings(c.send)
	
	c.reader()
}
