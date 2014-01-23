package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	//"sync"
	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
	"io"
	"strings"
	"time"
)

var Cfg *goconfig.ConfigFile
var ListenPort string

func main() {
	// Load config
	LoadConfig("conf/conf.ini")

	// open a logger
	runtime.GOMAXPROCS(runtime.NumCPU())
	LOGFILE, _ := os.OpenFile("portcheck.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	defer LOGFILE.Close()
	log.SetOutput(LOGFILE)

	// create a TCP server
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ListenPort)
	if err != nil {
		log.Println("[TCPERR]", "net.ResolveTCPAddr: ", err.Error())
	}
	tcps, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		log.Println("[TCPERR]", "net.ListenTCP: ", err.Error())
		return
	}
	defer tcps.Close()

	fmt.Printf("Start PortCheck Server Succeed, Listen on %s port.\r\n", ListenPort)

	for {
		conn, err := tcps.Accept()
		if err != nil {
			log.Println("[TCPERR]", "tcps.Accept: ", err.Error())
			return
		}

		fmt.Printf("[%s] TCP Client: %s Connected.\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			conn.RemoteAddr())

		go TCPClientHandle(conn)
	}
}

// TCP Client data handle
func TCPClientHandle(conn net.Conn) {
	defer func() {
		log.Println("[TCP]", "TCP Client Close: <-", conn.RemoteAddr())
		fmt.Printf("[%s] TCP Client %s Closed.\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			conn.RemoteAddr())
		conn.Close()
	}()

	var buf [2048]byte
	var rip string
	var rport string

	//
	// start ...
	//
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(30)))
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Println("[TCPERR]", "conn.Read: ", err.Error())
			if err == io.EOF {
				return
			}
			return
		}

		s := strings.Fields(string(buf[:n]))
		if len(s) != 2 {
			log.Println("[TCP]", "Recv error format strings")
			return
		}

		// copy
		//rip = s[0]
		ripAddr, _ := net.ResolveTCPAddr("tcp4", conn.RemoteAddr().String())
		rip = ripAddr.IP.String()
		rport = s[1]
		fmt.Printf("[%s] %s Recv: IP->%s, PORT->%s.\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			conn.RemoteAddr(),
			rip,
			rport)

		// Determine whether the IP is the domain name
		// if IP is the domain name, TODO DNS
		// Do nothing...

		// TCP & UDP link test
		go TCPLink(rip, rport)
		go UDPLink(rip, rport)

		log.Println("[TCP]", conn.RemoteAddr(), "Recv:", n, "Bytes")
		log.Println("[TCP]", conn.RemoteAddr(), buf[:n])

		//
		// end ...
		//
	}
}

// For TCP link test
func TCPLink(rip, rport string) {
	servAddr := rip + ":" + rport
	data := servAddr + " OK"

	log.Println("[TCPLink] Send to", servAddr)
	fmt.Printf("[%s] [TCPLink] Send to %s ...\r\n",
		time.Now().Format("2006-01-02 15:04:05"),
		servAddr)

	// connect to TCP server
	tcpAddr, err := net.ResolveTCPAddr("tcp4", servAddr)
	if err != nil {
		log.Println("[TCPLink] ResolveTCPAddr failed:", err.Error())
	}
	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		log.Println("Dial failed:", err.Error())
		fmt.Printf("[%s] [TCPLink] Connect to %s failed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		return
	}
	defer func() {
		fmt.Printf("[%s] [TCPLink] %s connect closed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		time.Sleep(time.Second)
		conn.Close()
	}()

	// send IP:PORT to server for TCPLink test
	_, err = conn.Write([]byte(data))
	if err != nil {
		log.Println("[TCPLink] Write to server failed:", err.Error())
		fmt.Printf("[%s] [TCPLink] Send to %s failed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		return
	}

	fmt.Printf("[%s] [TCPLink] Send to %s succeed\r\n",
		time.Now().Format("2006-01-02 15:04:05"),
		servAddr)
}

// For UDP link test
func UDPLink(rip, rport string) {
	servAddr := rip + ":" + rport
	data := servAddr + " OK"

	log.Println("[UDPLink] Send to", servAddr)
	fmt.Printf("[%s] [UDPLink] Send to %s ...\r\n",
		time.Now().Format("2006-01-02 15:04:05"),
		servAddr)

	// connect to UDP server
	udpAddr, err := net.ResolveUDPAddr("udp4", servAddr)
	if err != nil {
		log.Println("[UDPLink] ResolveUDPAddr failed:", err.Error())
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		log.Println("Dial failed:", err.Error())
		fmt.Printf("[%s] [UDPLink] Connect to %s failed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		return
	}
	defer func() {
		fmt.Printf("[%s] [UDPLink] %s connect closed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		time.Sleep(time.Second)
		conn.Close()
	}()

	// send IP:PORT to server for UDPLink test
	_, err = conn.Write([]byte(data))
	if err != nil {
		log.Println("[UDPLink] Write to server failed:", err.Error())
		fmt.Printf("[%s] [UDPLink] Send to %s failed\r\n",
			time.Now().Format("2006-01-02 15:04:05"),
			servAddr)
		return
	}

	fmt.Printf("[%s] [UDPLink] Send to %s succeed\r\n",
		time.Now().Format("2006-01-02 15:04:05"),
		servAddr)
}

// LoadConfig loads configuration file.
func LoadConfig(cfgPath string) {
	if !com.IsExist(cfgPath) {
		os.Create(cfgPath)
	}

	var err error
	Cfg, err = goconfig.LoadConfigFile(cfgPath)
	if err != nil {
		panic("Fail to load configuration file: " + err.Error())
	}

	ListenPort, err = Cfg.GetValue("server", "listen")
	if err != nil {
		panic("Fail to load configuration file: cannot find key docs_js_path")
	}
}
