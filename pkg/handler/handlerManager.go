package handler

import "fmt"

type Manager struct {
	handlers map[string]BaseStreamHandler
}

func NewHandlerManager() *Manager {
	return &Manager{}
}

// InitHandlerManager If you don't need a handler, just remove the corresponding handler here.
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
func (m *Manager) GetSenderHandler(command string) BaseStreamHandler {
	handlerType := fmt.Sprintf("/%s/0.0.1", command)
	if handler, ok := m.handlers[handlerType]; ok {
		return handler
	} else {
		return nil
	}
}
