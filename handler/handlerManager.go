package handler

type Manager struct {
	handlers map[string]BaseStreamHandler
}

func NewHandlerManager() *Manager {
	return &Manager{}
}

func (m *Manager) InitHandlerManager() {
	m.handlers = make(map[string]BaseStreamHandler)

	//echoHandler := NewEchoHandler()
	//echoProtocolID := "/echo/0.0.1"
	//echoHandler.initHandler(echoProtocolID)
	//m.handlers[echoProtocolID] = echoHandler

	stateHandler := NewStateHandler()
	stateProtocolID := "/state/0.0.1"
	stateHandler.initHandler(stateProtocolID)
	m.handlers[stateProtocolID] = stateHandler

	searchHandler := NewSearchHandler()
	searchProtocolID := "/search/0.0.1"
	searchHandler.initHandler(searchProtocolID)
	m.handlers[searchProtocolID] = searchHandler
}

func (m *Manager) GetHandlers() map[string]BaseStreamHandler {
	return m.handlers
}

// GetSenderHandler Not graceful.
// If you don't need a handler, just remove the case corresponding to that handler here.
func (m *Manager) GetSenderHandler(command string) BaseStreamHandler {
	switch command {
	//case "echo":
	//	log.Println("Get echo sender")
	//	return m.handlers["/echo/0.0.1"]
	case "search":
		//log.Println("Get search sender")
		return m.handlers["/search/0.0.1"]
	case "state":
		//log.Println("Get state sender")
		return m.handlers["/state/0.0.1"]
	default:
		//log.Println("Get default sender and do nothing")
	}
	return nil
}
