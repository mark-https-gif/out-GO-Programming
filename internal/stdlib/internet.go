package stdlib

import (
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

// internetModule exposes internet-facing helpers: public IP, DNS, network
// interfaces, port checks/scans and UPnP port mapping (opening ports on the
// router automatically).
func internetModule() *module.Module {
	m := module.New("internet")

	// public_ip() returns the public (WAN) IP as seen from the internet.
	m.Set("public_ip", func(args ...object.Object) object.Object {
		ip, err := fetchPublicIP()
		if err != "" {
			return errObj("internet::public_ip: " + err)
		}
		return &object.String{Value: ip}
	})

	// resolve(host) returns an array of IPv4/IPv6 addresses for a hostname.
	m.Set("resolve", func(args ...object.Object) object.Object {
		host, ok := requireString("resolve", argsValue(args, 0))
		if !ok {
			return errObj("internet::resolve expects STRING host")
		}
		ips, err := net.LookupHost(host)
		if err != nil {
			return errObj("internet::resolve: " + err.Error())
		}
		return stringsToArray(ips)
	})

	// reverse(ip) returns an array of hostnames for an IP address.
	m.Set("reverse", func(args ...object.Object) object.Object {
		ip, ok := requireString("reverse", argsValue(args, 0))
		if !ok {
			return errObj("internet::reverse expects STRING ip")
		}
		names, err := net.LookupAddr(ip)
		if err != nil {
			return errObj("internet::reverse: " + err.Error())
		}
		return stringsToArray(names)
	})

	// interfaces() lists local network interfaces with their details.
	m.Set("interfaces", func(args ...object.Object) object.Object {
		return listInterfaces()
	})

	// port_check(ip, port [, timeout_ms]) reports whether a TCP port is open.
	m.Set("port_check", func(args ...object.Object) object.Object {
		ip, ok1 := requireString("port_check", argsValue(args, 0))
		port, ok2 := argsInt(args, 1)
		if !ok1 || !ok2 {
			return errObj("internet::port_check expects STRING ip, INTEGER port [, INTEGER timeout_ms]")
		}
		timeout := 1000 * time.Millisecond
		if ms, ok := argsInt(args, 2); ok {
			timeout = time.Duration(ms) * time.Millisecond
		}
		return boolObj(tcpOpen(ip, int(port), timeout))
	})

	// port_scan(ip, start, end [, timeout_ms]) returns an array of open TCP ports.
	m.Set("port_scan", func(args ...object.Object) object.Object {
		ip, ok1 := requireString("port_scan", argsValue(args, 0))
		start, ok2 := argsInt(args, 1)
		end, ok3 := argsInt(args, 2)
		if !ok1 || !ok2 || !ok3 {
			return errObj("internet::port_scan expects STRING ip, INTEGER start, INTEGER end [, INTEGER timeout_ms]")
		}
		timeout := 500 * time.Millisecond
		if ms, ok := argsInt(args, 3); ok {
			timeout = time.Duration(ms) * time.Millisecond
		}
		if start < 1 {
			start = 1
		}
		if end > 65535 {
			end = 65535
		}
		if end < start {
			return errObj("internet::port_scan: end must be >= start")
		}
		return portScan(ip, int(start), int(end), timeout)
	})

	// port_free(port) reports whether a local TCP port can be bound.
	m.Set("port_free", func(args ...object.Object) object.Object {
		port, ok := argsInt(args, 0)
		if !ok {
			return errObj("internet::port_free expects INTEGER port")
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return boolObj(false)
		}
		ln.Close()
		return boolObj(true)
	})

	// open_port(port, protocol [, description]) opens a port on the router via
	// UPnP (AddPortMapping). protocol is "tcp" or "udp". Returns true on success.
	m.Set("open_port", func(args ...object.Object) object.Object {
		return upnpMap(args, true)
	})

	// close_port(port, protocol) removes a UPnP port mapping from the router.
	m.Set("close_port", func(args ...object.Object) object.Object {
		return upnpMap(args, false)
	})

	// router_external_ip() returns the router's external IP via UPnP.
	m.Set("router_external_ip", func(args ...object.Object) object.Object {
		ip, err := upnpExternalIP()
		if err != "" {
			return errObj("internet::router_external_ip: " + err)
		}
		return &object.String{Value: ip}
	})

	// --- network server (create a network / multi-client endpoint) ---

	// server_create(port [, protocol]) starts a TCP (default) or UDP server.
	m.Set("server_create", func(args ...object.Object) object.Object {
		return serverCreate(args)
	})

	// server_port(handle) returns the actual bound port.
	m.Set("server_port", func(args ...object.Object) object.Object {
		return serverPort(args)
	})

	// server_clients(handle) returns an array of connected client ids.
	m.Set("server_clients", func(args ...object.Object) object.Object {
		return serverClients(args)
	})

	// server_poll(handle) returns the next {client, data} message, or null.
	m.Set("server_poll", func(args ...object.Object) object.Object {
		return serverPoll(args)
	})

	// server_send(handle, client, data) sends to one client. Returns bool.
	m.Set("server_send", func(args ...object.Object) object.Object {
		return serverSend(args)
	})

	// server_broadcast(handle, data) sends to all clients. Returns count.
	m.Set("server_broadcast", func(args ...object.Object) object.Object {
		return serverBroadcast(args)
	})

	// server_stop(handle) closes the server and all its clients.
	m.Set("server_stop", func(args ...object.Object) object.Object {
		return serverStop(args)
	})

	m.Desc = "Internet helpers: public IP, DNS, interfaces, port check/scan, UPnP port mapping and TCP/UDP servers"
	return m
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func boolObj(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

func stringsToArray(items []string) object.Object {
	elems := make([]object.Object, len(items))
	for i, s := range items {
		elems[i] = &object.String{Value: s}
	}
	return &object.Array{Elements: elems}
}

func tcpOpen(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, itoa(int64(port))), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func portScan(ip string, start, end int, timeout time.Duration) object.Object {
	type result struct{ port int }
	out := make(chan result)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 256)

	for p := start; p <= end; p++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if tcpOpen(ip, p, timeout) {
				out <- result{port: p}
			}
		}(p)
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	ports := make([]int, 0)
	for r := range out {
		ports = append(ports, r.port)
	}
	sort.Ints(ports)

	elems := make([]object.Object, len(ports))
	for i, p := range ports {
		elems[i] = &object.Integer{Value: int64(p)}
	}
	return &object.Array{Elements: elems}
}

func listInterfaces() object.Object {
	ifaces, err := net.Interfaces()
	if err != nil {
		return errObj("internet::interfaces: " + err.Error())
	}
	elems := make([]object.Object, 0, len(ifaces))
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		ipList := make([]string, 0, len(addrs))
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				ipList = append(ipList, ipnet.IP.String())
			}
		}
		h := &object.Hash{Pairs: make(map[uint64]object.HashPair)}
		hashSetStr(h, "name", iface.Name)
		hashSetStr(h, "mac", iface.HardwareAddr.String())
		hashSetStr(h, "ips", strings.Join(ipList, ","))
		hashSetBool(h, "up", iface.Flags&net.FlagUp != 0)
		hashSetBool(h, "loopback", iface.Flags&net.FlagLoopback != 0)
		hashSetBool(h, "virtual", isVirtualInterface(iface.Name))
		elems = append(elems, h)
	}
	return &object.Array{Elements: elems}
}

func hashSetStr(h *object.Hash, key, val string) {
	ko := &object.String{Value: key}
	h.Pairs[object.HashKey(ko)] = object.HashPair{Key: ko, Value: &object.String{Value: val}}
}

func hashSetBool(h *object.Hash, key string, val bool) {
	ko := &object.String{Value: key}
	h.Pairs[object.HashKey(ko)] = object.HashPair{Key: ko, Value: &object.Boolean{Value: val}}
}

func fetchPublicIP() (string, string) {
	client := &http.Client{Timeout: 10 * time.Second}
	for _, url := range []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	} {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip, ""
		}
	}
	return "", "could not determine public IP"
}

// ---------------------------------------------------------------------------
// UPnP (IGD) — automatic port opening on the router
// ---------------------------------------------------------------------------

type upnpDevice struct {
	controlURL  string
	serviceType string
	baseHost    string
}

type upnpService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

type upnpDeviceXML struct {
	Services []upnpService `xml:"device>serviceList>service"`
	Devices  []upnpDeviceXML `xml:"device>deviceList>device"`
}

// discoverUPnP finds the internet gateway via SSDP and returns its control info.
func discoverUPnP() (*upnpDevice, string) {
	search := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 2\r\n" +
		"ST: urn:schemas-upnp-org:device:InternetGatewayDevice:1\r\n\r\n"

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, "SSDP socket: " + err.Error()
	}
	defer conn.Close()

	dst := &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900}
	if _, err := conn.WriteToUDP([]byte(search), dst); err != nil {
		return nil, "SSDP send: " + err.Error()
	}

	buf := make([]byte, 4096)
	location := ""
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(buf[:n]), "\r\n") {
			if strings.HasPrefix(strings.ToUpper(line), "LOCATION:") {
				location = strings.TrimSpace(line[len("LOCATION:"):])
				break
			}
		}
		if location != "" {
			break
		}
	}
	if location == "" {
		return nil, "no UPnP gateway found (SSDP timed out)"
	}

	dev, dErr := parseDeviceDescription(location)
	if dErr != "" {
		return nil, dErr
	}
	return dev, ""
}

func parseDeviceDescription(location string) (*upnpDevice, string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(location)
	if err != nil {
		return nil, "fetch description: " + err.Error()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "read description: " + err.Error()
	}

	var doc upnpDeviceXML
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, "parse description: " + err.Error()
	}

	all := append([]upnpService{}, doc.Services...)
	var walk func(d upnpDeviceXML)
	walk = func(d upnpDeviceXML) {
		all = append(all, d.Services...)
		for _, c := range d.Devices {
			walk(c)
		}
	}
	for _, c := range doc.Devices {
		walk(c)
	}

	var chosen *upnpService
	for i := range all {
		st := all[i].ServiceType
		if strings.Contains(st, "WANIPConnection") || strings.Contains(st, "WANPPPConnection") {
			chosen = &all[i]
			break
		}
	}
	if chosen == nil {
		return nil, "router has no WANIPConnection service"
	}

	// Resolve relative controlURL against the description location.
	base := location
	if i := strings.Index(base, "://"); i >= 0 {
		rest := base[i+3:]
		if j := strings.Index(rest, "/"); j >= 0 {
			base = base[:i+3+j]
		}
	}
	control := chosen.ControlURL
	if !strings.HasPrefix(control, "http") {
		if !strings.HasPrefix(control, "/") {
			control = "/" + control
		}
		control = base + control
	}

	host := ""
	if u, err := neturlParseHost(control); err == nil {
		host = u
	}
	return &upnpDevice{controlURL: control, serviceType: chosen.ServiceType, baseHost: host}, ""
}

func neturlParseHost(rawurl string) (string, error) {
	i := strings.Index(rawurl, "://")
	if i < 0 {
		return "", fmt.Errorf("no scheme")
	}
	rest := rawurl[i+3:]
	if j := strings.Index(rest, "/"); j >= 0 {
		rest = rest[:j]
	}
	return rest, nil
}

// upnpMap adds (open==true) or removes a port mapping on the router.
func upnpMap(args []object.Object, open bool) object.Object {
	port, ok1 := argsInt(args, 0)
	proto, ok2 := requireString("open_port", argsValue(args, 1))
	if !ok1 || !ok2 {
		return errObj("internet::open_port expects INTEGER port, STRING protocol ('tcp'/'udp')")
	}
	protocol := strings.ToUpper(proto)
	if protocol != "TCP" && protocol != "UDP" {
		return errObj("internet::open_port: protocol must be 'tcp' or 'udp'")
	}
	desc := "OUT"
	if len(args) > 2 {
		if d, ok := requireString("open_port", args[2]); ok {
			desc = d
		}
	}

	dev, derr := discoverUPnP()
	if derr != "" {
		return errObj("internet::open_port: " + derr)
	}

	localIP := primaryIPv4()
	if localIP == "" {
		return errObj("internet::open_port: no local LAN address")
	}

	var body string
	action := "AddPortMapping"
	if open {
		body = soapEnvelope(dev.serviceType, fmt.Sprintf(
			"<u:AddPortMapping xmlns:u=\"%s\">"+
				"<NewRemoteHost></NewRemoteHost>"+
				"<NewExternalPort>%d</NewExternalPort>"+
				"<NewProtocol>%s</NewProtocol>"+
				"<NewInternalPort>%d</NewInternalPort>"+
				"<NewInternalClient>%s</NewInternalClient>"+
				"<NewEnabled>1</NewEnabled>"+
				"<NewPortMappingDescription>%s</NewPortMappingDescription>"+
				"<NewLeaseDuration>0</NewLeaseDuration>"+
				"</u:AddPortMapping>",
			dev.serviceType, port, protocol, port, localIP, xmlEscape(desc)))
	} else {
		action = "DeletePortMapping"
		body = soapEnvelope(dev.serviceType, fmt.Sprintf(
			"<u:DeletePortMapping xmlns:u=\"%s\">"+
				"<NewRemoteHost></NewRemoteHost>"+
				"<NewExternalPort>%d</NewExternalPort>"+
				"<NewProtocol>%s</NewProtocol>"+
				"</u:DeletePortMapping>",
			dev.serviceType, port, protocol))
	}

	if err := soapCall(dev, action, body); err != "" {
		return errObj("internet::open_port: " + err)
	}
	return boolObj(open)
}

func upnpExternalIP() (string, string) {
	dev, derr := discoverUPnP()
	if derr != "" {
		return "", derr
	}
	body := soapEnvelope(dev.serviceType,
		fmt.Sprintf("<u:GetExternalIPAddress xmlns:u=\"%s\"></u:GetExternalIPAddress>", dev.serviceType))

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", dev.controlURL, strings.NewReader(body))
	if err != nil {
		return "", err.Error()
	}
	setSoapHeaders(req, dev.serviceType, "GetExternalIPAddress")
	resp, err := client.Do(req)
	if err != nil {
		return "", err.Error()
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	const tag = "<NewExternalIPAddress>"
	s := string(respBody)
	i := strings.Index(s, tag)
	if i < 0 {
		return "", "router did not return an external IP"
	}
	rest := s[i+len(tag):]
	j := strings.Index(rest, "</NewExternalIPAddress>")
	if j < 0 {
		return "", "malformed external IP response"
	}
	return strings.TrimSpace(rest[:j]), ""
}

func soapEnvelope(serviceType, inner string) string {
	return "<?xml version=\"1.0\"?>" +
		"<s:Envelope xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\" " +
		"s:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\">" +
		"<s:Body>" + inner + "</s:Body></s:Envelope>"
}

func setSoapHeaders(req *http.Request, serviceType, action string) {
	req.Header.Set("Content-Type", "text/xml; charset=\"utf-8\"")
	req.Header.Set("SOAPAction", "\""+serviceType+"#"+action+"\"")
}

func soapCall(dev *upnpDevice, action, body string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", dev.controlURL, strings.NewReader(body))
	if err != nil {
		return err.Error()
	}
	setSoapHeaders(req, dev.serviceType, action)
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Sprintf("router returned HTTP %d", resp.StatusCode)
	}
	return ""
}

func xmlEscape(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
