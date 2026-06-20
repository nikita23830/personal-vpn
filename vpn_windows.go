//go:build windows

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wintun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

const modeName = "клиент"

const (
	adapterName = "MyVPN"
	clientMask  = "255.255.255.0"
	vpnGateway  = "10.8.0.1"
	defaultDNS  = "8.8.8.8"
)

var (
	iphlpapi                        = windows.NewLazySystemDLL("iphlpapi.dll")
	procConvertInterfaceLuidToIndex = iphlpapi.NewProc("ConvertInterfaceLuidToIndex")
	procConvertInterfaceLuidToAlias = iphlpapi.NewProc("ConvertInterfaceLuidToAlias")
)

func luidToIndex(luid uint64) uint32 {
	var idx uint32
	procConvertInterfaceLuidToIndex.Call(
		uintptr(unsafe.Pointer(&luid)),
		uintptr(unsafe.Pointer(&idx)),
	)
	return idx
}

func luidToAlias(luid uint64) string {
	buf := make([]uint16, 257) // NDIS_IF_MAX_STRING_SIZE + 1
	ret, _, _ := procConvertInterfaceLuidToAlias.Call(
		uintptr(unsafe.Pointer(&luid)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret != 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
}

func adapterHasIP(_ string, ip string) bool {
	out, _ := exec.Command("ipconfig").CombinedOutput()
	cleaned := strings.ReplaceAll(string(out), "\x00", "")
	return strings.Contains(cleaned, ip)
}

func isElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()
	return token.IsElevated()
}

func run() {
	serverFlag := flag.String("server", "", "server address host:port, e.g., 203.0.113.10:9990")
	gwFlag := flag.String("gw", "", "Physical gateway IP (autodetect by default)")
	dnsFlag := flag.String("dns", defaultDNS, "DNS server for the tunnel")
	key := flag.String("key", "change-this-secret", "shared secret key (the same on both the server and the client)")
	flag.Parse()

	if *serverFlag == "" {
		log.Fatal("Specify -server host:port")
	}

	if !isElevated() {
		log.Fatal("ADMINISTRATOR PERMISSIONS ARE REQUIRED. Close the program, then open cmd or PowerShell. " +
			"Run it using “Run as administrator” and launch .\\vpn.exe from there. " +
			"Without admin privileges, netsh won't be able to assign an IP address and routes—the tunnel won't work.")
	}

	initCrypto(*key)
	if *key == "change-this-secret" {
		log.Println("NOTE: The default key is used. Specify your own using -key and use the same key on the server.")
	}

	raddr, err := net.ResolveUDPAddr("udp", *serverFlag)
	if err != nil {
		log.Fatalf("incorrect server address: %v", err)
	}
	serverIP := raddr.IP.String()

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("Unable to open UDP: %v", err)
	}
	defer conn.Close()

	adapter, err := wintun.CreateAdapter(adapterName, "Wintun", nil)
	if err != nil {
		log.Fatalf("Wintun: %v (Run as administrator? Is wintun.dll nearby?)", err)
	}
	session, err := adapter.StartSession(0x400000) // 4 МиБ
	if err != nil {
		adapter.Close()
		log.Fatalf("StartSession: %v", err)
	}

	assignedIP := register(conn)
	log.Printf("A virtual IP address has been obtained: %s", assignedIP)

	gw := *gwFlag
	if gw == "" {
		gw = findGateway()
	}
	if gw == "" {
		cleanup(session, adapter, serverIP, 0)
		log.Fatal("Unable to determine the gateway; please specify -gw <IP>")
	}
	log.Printf("physical gateway: %s", gw)

	luid := adapter.LUID()
	ifIndex := luidToIndex(luid)
	alias := luidToAlias(luid)
	if alias == "" {
		alias = adapterName
	}
	log.Printf("interface: alias=%q, index=%d", alias, ifIndex)
	configure(luid, ifIndex, alias, assignedIP, serverIP, gw, *dnsFlag)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Completion, route rollback...")
		removeRoutes(serverIP, ifIndex)
		os.Exit(0)
	}()

	go func() {
		for {
			time.Sleep(25 * time.Second)
			conn.Write(seal([]byte{typeRegister}))
		}
	}()

	log.Println("The VPN is active. All traffic is routed through the server. Press Ctrl+C to exit.")
	go tunToUDP(session, conn)
	udpToTun(session, conn)
	runtime.KeepAlive(adapter)
}

func register(conn *net.UDPConn) net.IP {
	buf := make([]byte, 1500)
	for attempt := 0; attempt < 10; attempt++ {
		conn.Write(seal([]byte{typeRegister}))
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			continue
		}
		if msg, ok := open(buf[:n]); ok && len(msg) >= 5 && msg[0] == typeRegReply {
			conn.SetReadDeadline(time.Time{})
			return net.IPv4(msg[1], msg[2], msg[3], msg[4]).To4()
		}
	}
	log.Fatal("The server did not respond to the registration request")
	return nil
}

func tunToUDP(session wintun.Session, conn *net.UDPConn) {
	event := session.ReadWaitEvent()
	out := make([]byte, 65536)
	for {
		packet, err := session.ReceivePacket()
		if err == nil {
			out[0] = typeData
			n := copy(out[1:], packet)
			conn.Write(seal(out[:n+1]))
			session.ReleaseReceivePacket(packet)
			continue
		}
		if errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
			windows.WaitForSingleObject(event, windows.INFINITE)
			continue
		}
		log.Printf("ReceivePacket: %v", err)
		return
	}
}

func udpToTun(session wintun.Session, conn *net.UDPConn) {
	buf := make([]byte, 65535)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("UDP read: %v", err)
			return
		}
		msg, ok := open(buf[:n])
		if !ok || len(msg) < 1 || msg[0] != typeData {
			continue
		}
		payload := msg[1:]
		packet, err := session.AllocateSendPacket(len(payload))
		if err != nil {
			continue
		}
		copy(packet, payload)
		session.SendPacket(packet)
	}
}

func sh(name string, args ...string) {
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		log.Printf("%s %v: %v: %s", name, args, err, string(out))
	}
}

func configure(luid uint64, ifIndex uint32, alias string, clientIP net.IP, serverIP, gw, dns string) {
	wluid := winipcfg.LUID(luid)
	addr, _ := netip.AddrFromSlice(clientIP.To4())
	if err := wluid.SetIPAddresses([]netip.Prefix{netip.PrefixFrom(addr, 24)}); err != nil {
		log.Printf("ATTENTION: Unable to assign IP %s: %v — full tunnel may not work", clientIP, err)
	} else {
		log.Printf("The IP address %s has been assigned to the adapter (via the API)", clientIP)
	}

	sh("netsh", "interface", "ipv4", "set", "interface", alias, "metric=1")
	sh("netsh", "interface", "ipv4", "set", "subinterface", alias, "mtu=1400", "store=persistent")
	sh("netsh", "interface", "ip", "set", "dns", "name="+alias, "static", dns)

	sh("route", "add", serverIP, "mask", "255.255.255.255", gw, "metric", "1")

	ifs := fmt.Sprint(ifIndex)
	sh("route", "add", "0.0.0.0", "mask", "128.0.0.0", vpnGateway, "metric", "1", "IF", ifs)
	sh("route", "add", "128.0.0.0", "mask", "128.0.0.0", vpnGateway, "metric", "1", "IF", ifs)

	sh("netsh", "interface", "ipv6", "add", "route", "::/1", "interface="+ifs, "metric=1")
	sh("netsh", "interface", "ipv6", "add", "route", "8000::/1", "interface="+ifs, "metric=1")

	log.Printf("Routes have been configured (full tunnel, IF=%s, IPv6 blocked)", ifs)
}

func removeRoutes(serverIP string, ifIndex uint32) {
	sh("route", "delete", "0.0.0.0", "mask", "128.0.0.0", vpnGateway)
	sh("route", "delete", "128.0.0.0", "mask", "128.0.0.0", vpnGateway)
	sh("route", "delete", serverIP)
	if ifIndex != 0 {
		ifs := fmt.Sprint(ifIndex)
		sh("netsh", "interface", "ipv6", "delete", "route", "::/1", "interface="+ifs)
		sh("netsh", "interface", "ipv6", "delete", "route", "8000::/1", "interface="+ifs)
	}
}

func cleanup(session wintun.Session, adapter *wintun.Adapter, serverIP string, ifIndex uint32) {
	removeRoutes(serverIP, ifIndex)
	session.End()
	adapter.Close()
}

func findGateway() string {
	out, err := exec.Command("route", "print", "-4", "0.0.0.0").Output()
	if err != nil {
		return ""
	}
	for _, ln := range splitLines(string(out)) {
		f := fields(ln)
		if len(f) >= 4 && f[0] == "0.0.0.0" && f[1] == "0.0.0.0" {
			if ip := net.ParseIP(f[2]); ip != nil && f[2] != "0.0.0.0" {
				return f[2]
			}
		}
	}
	return ""
}

func splitLines(s string) []string {
	var res []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	return append(res, s[start:])
}

func fields(s string) []string {
	var res []string
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
			i++
		}
		j := i
		for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\r' {
			j++
		}
		if j > i {
			res = append(res, s[i:j])
		}
		i = j
	}
	return res
}
