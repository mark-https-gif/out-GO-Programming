package stdlib

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// net handle registry. Handles are integer ids mapping to live sockets,
// owned by OUT scripts. The interpreter is single-threaded, so no locking
// is strictly needed, but we keep a mutex for safety with background readers.
var (
	netMu        sync.Mutex
	netUDP       = make(map[int64]*net.UDPConn)
	netTCP       = make(map[int64]*net.TCPConn)
	netListeners = make(map[int64]*net.TCPListener)
	netNextID    int64 = 1
)

// netModule exposes low-level UDP/TCP sockets for VoIP and raw networking.
func netModule() *module.Module {
	m := module.New("net")

	// --- UDP ---

	// udp_open(port) opens a UDP socket bound to a local port.
	// port 0 picks an ephemeral port (sending only).
	m.Set("udp_open", func(args ...object.Object) object.Object {
		port := int64(0)
		if len(args) > 0 {
			p, ok := args[0].(*object.Integer)
			if !ok {
				return errObj("net::udp_open expects INTEGER port")
			}
			port = p.Value
		}
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: int(port)})
		if err != nil {
			return errObj("net::udp_open: " + err.Error())
		}
		netMu.Lock()
		id := netNextID
		netNextID++
		netUDP[id] = conn
		netMu.Unlock()
		return &object.Integer{Value: id}
	})

	// udp_send(sock, ip, port, data) sends a datagram to ip:port.
	m.Set("udp_send", func(args ...object.Object) object.Object {
		sock, ok1 := argsInt(args, 0)
		ip, ok2 := requireString("udp_send", argsValue(args, 1))
		port, ok3 := argsInt(args, 2)
		data, ok4 := requireString("udp_send", argsValue(args, 3))
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return errObj("net::udp_send expects INTEGER sock, STRING ip, INTEGER port, STRING data")
		}
		netMu.Lock()
		conn, ok := netUDP[sock]
		netMu.Unlock()
		if !ok {
			return errObj("net::udp_send: invalid socket")
		}
		addr := &net.UDPAddr{IP: net.ParseIP(ip), Port: int(port)}
		if _, err := conn.WriteToUDP([]byte(data), addr); err != nil {
			return errObj("net::udp_send: " + err.Error())
		}
		return NULL
	})

	// udp_recv(sock, maxlen) reads one datagram. Returns NULL on timeout.
	m.Set("udp_recv", func(args ...object.Object) object.Object {
		sock, ok1 := argsInt(args, 0)
		maxlen, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("net::udp_recv expects INTEGER sock, INTEGER maxlen")
		}
		netMu.Lock()
		conn, ok := netUDP[sock]
		netMu.Unlock()
		if !ok {
			return errObj("net::udp_recv: invalid socket")
		}
		buf := make([]byte, maxlen)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok2 := err.(net.Error); ok2 && ne.Timeout() {
				return NULL
			}
			return errObj("net::udp_recv: " + err.Error())
		}
		return &object.String{Value: string(buf[:n])}
	})

	m.Set("udp_close", func(args ...object.Object) object.Object {
		sock, ok := argsInt(args, 0)
		if !ok {
			return errObj("net::udp_close expects INTEGER sock")
		}
		netMu.Lock()
		conn, ok := netUDP[sock]
		if ok {
			delete(netUDP, sock)
		}
		netMu.Unlock()
		if !ok {
			return errObj("net::udp_close: invalid socket")
		}
		conn.Close()
		return NULL
	})

	// --- TCP ---

	// tcp_listen(port) starts a TCP listener on a local port.
	m.Set("tcp_listen", func(args ...object.Object) object.Object {
		port, ok := argsInt(args, 0)
		if !ok {
			return errObj("net::tcp_listen expects INTEGER port")
		}
		ln, err := net.Listen("tcp", ":"+itoa(port))
		if err != nil {
			return errObj("net::tcp_listen: " + err.Error())
		}
		netMu.Lock()
		id := netNextID
		netNextID++
		netListeners[id] = ln.(*net.TCPListener)
		netMu.Unlock()
		return &object.Integer{Value: id}
	})

	// tcp_accept(listener) blocks until a client connects and returns a conn.
	m.Set("tcp_accept", func(args ...object.Object) object.Object {
		lnID, ok := argsInt(args, 0)
		if !ok {
			return errObj("net::tcp_accept expects INTEGER listener")
		}
		netMu.Lock()
		ln, ok := netListeners[lnID]
		netMu.Unlock()
		if !ok {
			return errObj("net::tcp_accept: invalid listener")
		}
		conn, err := ln.Accept()
		if err != nil {
			return errObj("net::tcp_accept: " + err.Error())
		}
		netMu.Lock()
		id := netNextID
		netNextID++
		netTCP[id] = conn.(*net.TCPConn)
		netMu.Unlock()
		return &object.Integer{Value: id}
	})

	// tcp_connect(ip, port) dials a TCP server and returns a conn.
	m.Set("tcp_connect", func(args ...object.Object) object.Object {
		ip, ok1 := requireString("tcp_connect", argsValue(args, 0))
		port, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("net::tcp_connect expects STRING ip, INTEGER port")
		}
		conn, err := net.Dial("tcp", ip+":"+itoa(port))
		if err != nil {
			return errObj("net::tcp_connect: " + err.Error())
		}
		netMu.Lock()
		id := netNextID
		netNextID++
		netTCP[id] = conn.(*net.TCPConn)
		netMu.Unlock()
		return &object.Integer{Value: id}
	})

	// tcp_send(conn, data) writes all bytes to the connection.
	m.Set("tcp_send", func(args ...object.Object) object.Object {
		connID, ok1 := argsInt(args, 0)
		data, ok2 := requireString("tcp_send", argsValue(args, 1))
		if !ok1 || !ok2 {
			return errObj("net::tcp_send expects INTEGER conn, STRING data")
		}
		netMu.Lock()
		conn, ok := netTCP[connID]
		netMu.Unlock()
		if !ok {
			return errObj("net::tcp_send: invalid connection")
		}
		if _, err := conn.Write([]byte(data)); err != nil {
			return errObj("net::tcp_send: " + err.Error())
		}
		return NULL
	})

	// tcp_recv(conn, maxlen) reads up to maxlen bytes. Returns NULL on timeout
	// or when the peer closes the connection.
	m.Set("tcp_recv", func(args ...object.Object) object.Object {
		connID, ok1 := argsInt(args, 0)
		maxlen, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("net::tcp_recv expects INTEGER conn, INTEGER maxlen")
		}
		netMu.Lock()
		conn, ok := netTCP[connID]
		netMu.Unlock()
		if !ok {
			return errObj("net::tcp_recv: invalid connection")
		}
		buf := make([]byte, maxlen)
		n, err := conn.Read(buf)
		if err != nil {
			if n > 0 {
				return &object.String{Value: string(buf[:n])}
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				return NULL
			}
			return errObj("net::tcp_recv: " + err.Error())
		}
		return &object.String{Value: string(buf[:n])}
	})

	m.Set("tcp_close", func(args ...object.Object) object.Object {
		connID, ok := argsInt(args, 0)
		if !ok {
			return errObj("net::tcp_close expects INTEGER conn")
		}
		netMu.Lock()
		conn, ok := netTCP[connID]
		if ok {
			delete(netTCP, connID)
		}
		netMu.Unlock()
		if !ok {
			return errObj("net::tcp_close: invalid connection")
		}
		conn.Close()
		return NULL
	})

	// close(handle) closes any open net handle (udp socket, acceptance, listener).
	m.Set("close", func(args ...object.Object) object.Object {
		id, ok := argsInt(args, 0)
		if !ok {
			return errObj("net::close expects INTEGER handle")
		}
		netMu.Lock()
		if c, ok := netUDP[id]; ok {
			delete(netUDP, id)
			netMu.Unlock()
			c.Close()
			return NULL
		}
		if c, ok := netTCP[id]; ok {
			delete(netTCP, id)
			netMu.Unlock()
			c.Close()
			return NULL
		}
		if l, ok := netListeners[id]; ok {
			delete(netListeners, id)
			netMu.Unlock()
			l.Close()
			return NULL
		}
		netMu.Unlock()
		return errObj("net::close: invalid handle")
	})

	// set_timeout(handle, ms) sets a read deadline (default: no timeout).
	m.Set("set_timeout", func(args ...object.Object) object.Object {
		id, ok1 := argsInt(args, 0)
		ms, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("net::set_timeout expects INTEGER handle, INTEGER ms")
		}
		deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
		netMu.Lock()
		if c, ok := netUDP[id]; ok {
			c.SetReadDeadline(deadline)
			netMu.Unlock()
			return NULL
		}
		if c, ok := netTCP[id]; ok {
			c.SetReadDeadline(deadline)
			netMu.Unlock()
			return NULL
		}
		netMu.Unlock()
		return errObj("net::set_timeout: invalid handle")
	})

	// --- helpers ---

	// local_ip() returns the primary LAN IPv4 address. It uses the OS route
	// lookup (source address for an outbound packet) so VPN/virtual adapters
	// are not chosen by mistake.
	m.Set("local_ip", func(args ...object.Object) object.Object {
		if ip := primaryIPv4(); ip != "" {
			return &object.String{Value: ip}
		}
		return errObj("net::local_ip: no LAN address found")
	})

	// hostname() returns the machine hostname.
	m.Set("hostname", func(args ...object.Object) object.Object {
		h, err := os.Hostname()
		if err != nil {
			return errObj("net::hostname: " + err.Error())
		}
		return &object.String{Value: h}
	})

	m.Desc = "Low-level TCP/UDP sockets (VoIP, raw networking)"
	return m
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}

// virtualAdapterMarkers are substrings of interface names that should not be
// treated as the primary LAN address (VPNs, hypervisors, tunnels).
var virtualAdapterMarkers = []string{
	"radmin", "vpn", "virtualbox", "vmware", "hyper-v", "teredo",
	"bluetooth", "loopback", "tap-", "wintun", "wireguard", "openvpn",
	"hamachi", "zerotier", "tailscale",
}

func isVirtualInterface(name string) bool {
	lower := strings.ToLower(name)
	for _, m := range virtualAdapterMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// primaryIPv4 finds the local IPv4 the OS would use for outbound traffic.
func primaryIPv4() string {
	// Route lookup: UDP "connect" never sends a packet, it just asks the OS
	// which source address it would pick for that destination.
	if conn, err := net.Dial("udp", "8.8.8.8:80"); err == nil {
		if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			conn.Close()
			if ip4 := ua.IP.To4(); ip4 != nil && !ip4.IsLoopback() {
				return ip4.String()
			}
		} else {
			conn.Close()
		}
	}

	// Fallback: first non-loopback, non-virtual interface address.
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if isVirtualInterface(iface.Name) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			if ip4 := ipnet.IP.To4(); ip4 != nil && !ip4.IsLoopback() {
				return ip4.String()
			}
		}
	}
	return ""
}