//go:build linux

package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"

	"log"

	"github.com/songgao/water"
)

const modeName = "сервер"

const (
	serverTunIP = "10.8.0.1"
	subnet      = "10.8.0.0/24"
)

type server struct {
	ifce *water.Interface
	conn *net.UDPConn

	mu       sync.Mutex
	ip2addr  map[string]*net.UDPAddr
	addr2ip  map[string]net.IP
	nextHost byte
}

func newServer(ifce *water.Interface, conn *net.UDPConn) *server {
	return &server{
		ifce:     ifce,
		conn:     conn,
		ip2addr:  make(map[string]*net.UDPAddr),
		addr2ip:  make(map[string]net.IP),
		nextHost: 2,
	}
}

func (s *server) assign(addr *net.UDPAddr) net.IP {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ip, ok := s.addr2ip[addr.String()]; ok {
		s.ip2addr[ip.String()] = addr
		return ip
	}
	ip := net.IPv4(10, 8, 0, s.nextHost).To4()
	s.nextHost++
	if s.nextHost < 2 {
		s.nextHost = 2
	}
	s.addr2ip[addr.String()] = ip
	s.ip2addr[ip.String()] = addr
	return ip
}

func (s *server) lookup(ip net.IP) *net.UDPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ip2addr[ip.String()]
}

func (s *server) learn(srcIP net.IP, addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.ip2addr[srcIP.String()]; !ok || cur.String() != addr.String() {
		s.ip2addr[srcIP.String()] = addr
		s.addr2ip[addr.String()] = srcIP
	}
}

func (s *server) udpToTun() {
	buf := make([]byte, 65535)
	for {
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("udp read: %v", err)
			continue
		}
		msg, ok := open(buf[:n])
		if !ok || len(msg) < 1 {
			continue
		}
		switch msg[0] {
		case typeRegister:
			ip := s.assign(addr)
			reply := append([]byte{typeRegReply}, ip.To4()...)
			if _, err := s.conn.WriteToUDP(seal(reply), addr); err != nil {
				log.Printf("reg reply: %v", err)
			}
			log.Printf("клиент %s -> %s", addr, ip)
		case typeData:
			if len(msg) < 1+20 {
				continue
			}
			pkt := msg[1:]
			if pkt[0]>>4 != 4 || pkt[12] != 10 || pkt[13] != 8 || pkt[14] != 0 {
				continue
			}
			src := net.IPv4(pkt[12], pkt[13], pkt[14], pkt[15])
			s.learn(src, addr)
			if _, err := s.ifce.Write(pkt); err != nil {
				log.Printf("tun write: %v", err)
			}
		}
	}
}

func (s *server) tunToUDP() {
	buf := make([]byte, 65535)
	out := make([]byte, 65536)
	for {
		n, err := s.ifce.Read(buf)
		if err != nil {
			log.Printf("tun read: %v", err)
			continue
		}
		if n < 20 || buf[0]>>4 != 4 {
			continue
		}
		dst := net.IPv4(buf[16], buf[17], buf[18], buf[19])
		addr := s.lookup(dst)
		if addr == nil {
			continue
		}
		out[0] = typeData
		copy(out[1:], buf[:n])
		if _, err := s.conn.WriteToUDP(seal(out[:n+1]), addr); err != nil {
			log.Printf("udp write: %v", err)
		}
	}
}

func run() {
	listen := flag.String("listen", ":9990", "UDP listen port")
	egress := flag.String("egress", "", "External interface for NAT (auto-detection by default)")
	key := flag.String("key", "change-this-secret", "shared secret key (the same on both the server and the client)")
	flag.Parse()

	initCrypto(*key)
	if *key == "change-this-secret" {
		log.Println("NOTE: The default key is being used. Specify your own using -key and use the same key on the client.")
	}

	ifce, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		log.Fatalf("Failed to create TUN (does it require root privileges?): %v", err)
	}
	log.Printf("TUN: %s", ifce.Name())

	if *egress == "" {
		*egress = defaultEgress()
	}
	setupNetwork(ifce.Name(), *egress)

	udpAddr, err := net.ResolveUDPAddr("udp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("I'm listening UDP %s, egress=%s", *listen, *egress)

	s := newServer(ifce, conn)
	go s.tunToUDP()
	s.udpToTun()
}

func runCmd(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func setupNetwork(tun, egress string) {
	warn := func(err error) {
		if err != nil {
			log.Printf("configuration warning: %v", err)
		}
	}
	warn(runCmd("ip", "addr", "add", serverTunIP+"/24", "dev", tun))
	warn(runCmd("ip", "link", "set", "dev", tun, "mtu", fmt.Sprint(mtu)))
	warn(runCmd("ip", "link", "set", "dev", tun, "up"))
	warn(runCmd("sysctl", "-w", "net.ipv4.ip_forward=1"))

	_ = runCmd("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", subnet, "-o", egress, "-j", "MASQUERADE")
	warn(runCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", subnet, "-o", egress, "-j", "MASQUERADE"))
	warn(runCmd("iptables", "-A", "FORWARD", "-i", tun, "-s", subnet, "-j", "ACCEPT"))
	warn(runCmd("iptables", "-A", "FORWARD", "-o", tun, "-d", subnet, "-j", "ACCEPT"))
	log.Printf("The network is configured (tun=%s, egress=%s)", tun, egress)
}

func defaultEgress() string {
	out, err := exec.Command("sh", "-c", "ip route show default | awk '{print $5; exit}'").Output()
	if err != nil {
		log.Printf("Unable to determine the egress; using eth0: %v", err)
		return "eth0"
	}
	if name := strings.TrimSpace(string(out)); name != "" {
		return name
	}
	return "eth0"
}
