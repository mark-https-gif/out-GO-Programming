package stdlib

import (
	"net"
	"sync"

	"github.com/out-lang/out/internal/object"
)

// netServer is a small event-driven TCP/UDP server. It runs accept/read loops
// in the background and queues incoming messages so a single-threaded OUT
// script can poll them with server_poll().
type netServer struct {
	handle int64
	proto  string

	ln    *net.TCPListener
	uconn *net.UDPConn

	mu         sync.Mutex
	clients    map[int64]*serverClient
	nextClient int64
	closed     bool

	events chan serverEvent
	done   chan struct{}
}

type serverClient struct {
	id     int64
	conn   *net.TCPConn
	udpAdr *net.UDPAddr
	sendMu sync.Mutex
}

type serverEvent struct {
	client int64
	data   string
}

var (
	srvMu      sync.Mutex
	servers    = make(map[int64]*netServer)
	srvNextID  int64 = 1
)

func (s *netServer) acceptLoop() {
	for {
		conn, err := s.ln.AcceptTCP()
		if err != nil {
			return
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			conn.Close()
			return
		}
		id := s.nextClient
		s.nextClient++
		s.clients[id] = &serverClient{id: id, conn: conn}
		s.mu.Unlock()
		go s.readTCPClient(id, conn)
	}
}

func (s *netServer) readTCPClient(id int64, conn *net.TCPConn) {
	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			s.push(id, string(buf[:n]))
		}
		if err != nil {
			s.mu.Lock()
			delete(s.clients, id)
			s.mu.Unlock()
			conn.Close()
			return
		}
	}
}

func (s *netServer) readUDPLoop() {
	buf := make([]byte, 65536)
	for {
		n, addr, err := s.uconn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		s.mu.Lock()
		var cid int64 = -1
		for id, c := range s.clients {
			if c.udpAdr != nil && c.udpAdr.String() == addr.String() {
				cid = id
				break
			}
		}
		if cid == -1 {
			cid = s.nextClient
			s.nextClient++
			s.clients[cid] = &serverClient{id: cid, udpAdr: addr}
		}
		s.mu.Unlock()
		s.push(cid, string(buf[:n]))
	}
}

func (s *netServer) push(client int64, data string) {
	select {
	case s.events <- serverEvent{client: client, data: data}:
	default:
		// queue full — drop to avoid blocking network loops
	}
}

// ---------------------------------------------------------------------------
// OUT-facing API
// ---------------------------------------------------------------------------

func serverCreate(args []object.Object) object.Object {
	port, ok := argsInt(args, 0)
	if !ok {
		return errObj("internet::server_create expects INTEGER port [, STRING protocol]")
	}
	proto := "tcp"
	if len(args) > 1 {
		if p, ok := requireString("server_create", args[1]); ok {
			proto = p
		}
	}

	s := &netServer{
		proto:   proto,
		clients: make(map[int64]*serverClient),
		events:  make(chan serverEvent, 1024),
		done:    make(chan struct{}),
	}

	switch proto {
	case "tcp":
		ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: int(port)})
		if err != nil {
			return errObj("internet::server_create: " + err.Error())
		}
		s.ln = ln
		go s.acceptLoop()
	case "udp":
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: int(port)})
		if err != nil {
			return errObj("internet::server_create: " + err.Error())
		}
		s.uconn = conn
		go s.readUDPLoop()
	default:
		return errObj("internet::server_create: protocol must be 'tcp' or 'udp'")
	}

	srvMu.Lock()
	id := srvNextID
	srvNextID++
	s.handle = id
	servers[id] = s
	srvMu.Unlock()
	return &object.Integer{Value: id}
}

func serverLookup(args []object.Object, fn string) (*netServer, object.Object) {
	id, ok := argsInt(args, 0)
	if !ok {
		return nil, errObj("internet::" + fn + " expects INTEGER server handle")
	}
	srvMu.Lock()
	s, ok := servers[id]
	srvMu.Unlock()
	if !ok {
		return nil, errObj("internet::" + fn + ": invalid server handle")
	}
	return s, nil
}

// serverPoll returns the next incoming message as {client, data}, or null.
func serverPoll(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_poll")
	if e != nil {
		return e
	}
	select {
	case ev := <-s.events:
		h := &object.Hash{Pairs: make(map[uint64]object.HashPair)}
		hashSetInt(h, "client", ev.client)
		hashSetStr(h, "data", ev.data)
		return h
	default:
		return NULL
	}
}

func serverSend(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_send")
	if e != nil {
		return e
	}
	cid, ok1 := argsInt(args, 1)
	data, ok2 := requireString("server_send", argsValue(args, 2))
	if !ok1 || !ok2 {
		return errObj("internet::server_send expects INTEGER handle, INTEGER client, STRING data")
	}
	s.mu.Lock()
	c, ok := s.clients[cid]
	s.mu.Unlock()
	if !ok {
		return boolObj(false)
	}
	return boolObj(sendToClient(s, c, data))
}

func serverBroadcast(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_broadcast")
	if e != nil {
		return e
	}
	data, ok := requireString("server_broadcast", argsValue(args, 1))
	if !ok {
		return errObj("internet::server_broadcast expects INTEGER handle, STRING data")
	}
	s.mu.Lock()
	clients := make([]*serverClient, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	count := int64(0)
	for _, c := range clients {
		if sendToClient(s, c, data) {
			count++
		}
	}
	return &object.Integer{Value: count}
}

func sendToClient(s *netServer, c *serverClient, data string) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.udpAdr != nil {
		if s.uconn == nil {
			return false
		}
		_, err := s.uconn.WriteToUDP([]byte(data), c.udpAdr)
		return err == nil
	}
	if c.conn == nil {
		return false
	}
	_, err := c.conn.Write([]byte(data))
	return err == nil
}

func serverClients(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_clients")
	if e != nil {
		return e
	}
	s.mu.Lock()
	ids := make([]int64, 0, len(s.clients))
	for id := range s.clients {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	elems := make([]object.Object, len(ids))
	for i, id := range ids {
		elems[i] = &object.Integer{Value: id}
	}
	return &object.Array{Elements: elems}
}

func serverPort(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_port")
	if e != nil {
		return e
	}
	if s.ln != nil {
		return &object.Integer{Value: int64(s.ln.Addr().(*net.TCPAddr).Port)}
	}
	if s.uconn != nil {
		return &object.Integer{Value: int64(s.uconn.LocalAddr().(*net.UDPAddr).Port)}
	}
	return NULL
}

func serverStop(args []object.Object) object.Object {
	s, e := serverLookup(args, "server_stop")
	if e != nil {
		return e
	}
	srvMu.Lock()
	delete(servers, s.handle)
	srvMu.Unlock()

	s.mu.Lock()
	s.closed = true
	clients := make([]*serverClient, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.clients = make(map[int64]*serverClient)
	s.mu.Unlock()

	if s.ln != nil {
		s.ln.Close()
	}
	if s.uconn != nil {
		s.uconn.Close()
	}
	for _, c := range clients {
		if c.conn != nil {
			c.conn.Close()
		}
	}
	close(s.done)
	return NULL
}

func hashSetInt(h *object.Hash, key string, val int64) {
	ko := &object.String{Value: key}
	h.Pairs[object.HashKey(ko)] = object.HashPair{Key: ko, Value: &object.Integer{Value: val}}
}

