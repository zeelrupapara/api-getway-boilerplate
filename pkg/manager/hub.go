// Developer: Saif Hamdan

package manager

import (
	"reflect"
	"sync"
	"syscall"
	"time"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/logger"

	"github.com/gofiber/websocket/v2"
)

type Handler func(*Ctx) error
type ErrorHandler func(err error) *model.Event
type RouterMap map[model.EventType]Handler

const (
	PongWait     = 10000 * time.Millisecond // pongWait should always be more than the ping interval
	PingInterval = 5000 * time.Millisecond
)

type Hub struct {
	fd         int
	Clients    ClientMap
	ClientList ClientList
	RouterMap  RouterMap
	sync.RWMutex
	Log          *logger.Logger
	ErrorHandler ErrorHandler
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		Clients:      make(ClientMap),
		ClientList:   make(ClientList, 0),
		RouterMap:    make(RouterMap),
		Log:          log,
		ErrorHandler: DefaultErrorHandler,
	}
}

// get All active clients
func (h *Hub) GetAll() ClientList {
	return h.ClientList
}

// Get One
func (h *Hub) Get(sessionsId string) (*Client, bool) {
	h.RLock()
	defer h.RUnlock()

	v, ok := h.Clients[sessionsId]
	return v, ok
}

// Get One By Id
// func (h *Hub) GetById(id string) (*Client, bool) {
// 	h.RLock()
// 	defer h.RUnlock()

// 	for _, v := range h.Clients {
// 		if id == v.Id {
// 			return v, true
// 		}
// 	}

// 	return nil, false
// }

// addClient will add clients to our clientList
func (h *Hub) Store(client *Client) error {
	// fd := websocketFD(client.Conn)
	// err := unix.EpollCtl(h.fd, syscall.EPOLL_CTL_ADD, fd, &unix.Epollmodel.Event{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	// if err != nil {
	// 	return err
	// }

	// Lock so we can manipulate
	h.Lock()
	defer h.Unlock()

	h.Clients[client.SessionId] = client
	h.ClientList = append(h.ClientList, client)

	h.clientsCount()
	return nil
}

// remove client
func (h *Hub) Delete(sessionId string) error {
	h.Lock()
	defer h.Unlock()

	// Check if Client exists, then delete it
	if client, ok := h.Clients[sessionId]; ok {
		// close client
		client.close()

		// fd := websocketFD(client.Conn)
		// err := unix.EpollCtl(h.fd, syscall.EPOLL_CTL_DEL, fd, nil)
		// if err != nil {
		// 	return err
		// }

		// remove client
		delete(h.Clients, sessionId)
		for i := range h.ClientList {
			if h.ClientList[i].SessionId == sessionId {
				h.ClientList = append(h.ClientList[:i], h.ClientList[i+1:]...)
				break
			}
		}

	}

	h.clientsCount()

	return nil
}

// func (h *Hub) Wait() ([]*websocket.Conn, error) {
// 	events := make([]unix.EpollEvent, 100)
// 	n, err := unix.EpollWait(h.fd, events, 100)
// 	if err != nil {
// 		return nil, err
// 	}

// 	h.RLock()
// 	defer h.RUnlock()

// 	var connections []*websocket.Conn
// 	// for i := 0; i < n; i++ {
// 	// 	conn := h.Clients[int(events[i].Fd)]
// 	// 	connections = append(connections, conn)
// 	// }
// 	return connections, nil
// }

// addClient will add clients to our clientList
func (h *Hub) Update(newClient *Client) {
	// Lock so we can manipulate
	h.Lock()
	defer h.Unlock()

	h.Clients[newClient.SessionId] = newClient

	h.clientsCount()
}

// addClient will add clients to our clientList
func (h *Hub) UpdateSessionId(oldSessionId string, newSessionId string) {
	// Lock so we can manipulate
	h.Lock()
	defer h.Unlock()

	h.Clients[newSessionId] = h.Clients[oldSessionId]
	delete(h.Clients, oldSessionId)

	h.clientsCount()
}

func (h *Hub) DeleteAll() {
	h.Lock()
	defer h.Unlock()

	h.Clients = make(ClientMap)
	h.ClientList = make(ClientList, 0)

	h.clientsCount()
}

func (h *Hub) clientsCount() {
	h.Log.Logger.Info("WS Clients count: ", len(h.ClientList))
	h.Log.Logger.Info("WS FD count: ", h.fd)
}

func (h *Hub) RegisterRoute(event model.EventType, handler Handler) {
	h.RouterMap[event] = handler
}

func (h *Hub) SetErrorHandler(cb ErrorHandler) {
	h.ErrorHandler = cb
}

func DefaultErrorHandler(err error) *model.Event {
	return &model.Event{Type: model.EventType_InternalError}
}

func SetMaxWebsocketConnections() {
	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
}

func websocketFD(conn *websocket.Conn) int {
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
