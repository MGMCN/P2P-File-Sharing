package handler

type Manager struct {
	handlers map[string]BaseStreamHandler
}

func NewHandlerManager() *Manager {
	return &Manager{}
}

func (m *Manager) InitHandlerManager() {
	m.handlers = make(map[string]BaseStreamHandler)

	echoHandler := NewEchoHandler()
	echoProtocolID := "/echo/0.0.1"
	echoHandler.initHandler(echoProtocolID)
	m.handlers[echoProtocolID] = echoHandler

	searchHandler := NewSearchHandler()
	searchProtocolID := "/search/0.0.1"
	searchHandler.initHandler(searchProtocolID)
	m.handlers[searchProtocolID] = searchHandler

	downloadHandler := NewDownloadHandler()
	downloadProtocolID := "/download/0.0.1"
	downloadHandler.initHandler(downloadProtocolID)
	m.handlers[downloadProtocolID] = downloadHandler

	leaveHandler := NewLeaveHandler()
	leaveProtocolID := "/leave/0.0.1"
	leaveHandler.initHandler(leaveProtocolID)
	m.handlers[leaveProtocolID] = leaveHandler
}

func (m *Manager) GetHandlers() map[string]BaseStreamHandler {
	return m.handlers
}

// GetSenderHandler Not graceful.
// If you don't need a handler, just remove the case corresponding to that handler here.
func (m *Manager) GetSenderHandler(command string) BaseStreamHandler {
	switch command {
	case "echo":
		//log.Println("Get echo sender")
		return m.handlers["/echo/0.0.1"]
	case "search":
		//log.Println("Get search sender")
		return m.handlers["/search/0.0.1"]
	case "download":
		//log.Println("Get download sender")
		return m.handlers["/download/0.0.1"]
	case "leave":
		//log.Println("Get download sender")
		return m.handlers["/leave/0.0.1"]
	default:
		//log.Println("Get default sender and do nothing")
	}
	return nil
}
