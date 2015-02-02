package main

// A Message contains the process ID, message type and process output.
type Message struct {
	ID string
	Type string
	Body []byte `json:"-"` // don't send body in json encodings
	Dest chan *Message `json:"-"` // don't send in json
	Data map[string]interface{}
}

// A Broker broadcasts messages to multiple clients.
type Broker struct {
	clients map[chan *Message]bool
	joining chan chan *Message
	leaving chan chan *Message
	messages chan *Message
}

// Start managing client connections and message broadcasts.
func (b *Broker) Start() {
	go func() {
		for {
			select {
			case s := <-b.joining:
				b.clients[s] = true
			case s := <-b.leaving:
				delete(b.clients, s)
			case message := <-b.messages:
				for s := range b.clients {
					if message.Dest == nil {
						s <- message
					} else if message.Dest == s {
						s <- message
					}
				}
			}
		}
	}()
}

// NewBroker creates a Broker instance.
func NewBroker() *Broker {
	return &Broker{
		make(map[chan *Message]bool),
		make(chan (chan *Message)),
		make(chan (chan *Message)),
		make(chan *Message),
	}
}
