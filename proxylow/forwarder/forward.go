package prowxylow

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/songgao/water"
)

//netsh interface ip set address name="Ehternet 2" source=static addr=10.1.0.10 mask=255.255.255.0 gateway=none

//network : 192.168.1.10/24
//windows设置虚拟网卡
func tuntapRun(rch chan []byte, wch chan []byte, remoteAddr string, network string, DNS string) {
	//启动虚拟网卡之前设置已连接的网卡
	err:=setDNS(DNS)
	if err!=nil {
		log.Fatal(err)
	}
	
	//init tun/tap device
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID: "tap0901",
			Network:     network,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
	defer ifce.Close()
	//prepare work
	index := strings.LastIndex(network, ".")
	idxLast := strings.LastIndex(network, "/")
	gw := network[:index+1] + "0"
	addr := network[:idxLast]

	cmd := fmt.Sprintf("interface ip set address name=\"%s\" source=static addr=%s mask=255.255.255.0 gateway=%s", ifce.Name(), addr, gw)
	err = exec.Command("netsh", strings.Split(cmd, " ")...).Run()
	if err != nil {
		log.Fatal(err)
	}
	//需要等待一段时间给系统识别网卡
	time.Sleep(time.Millisecond * 5000)

	log.Println(ifce.Name())
	log.Println(addr)
	log.Println(cmd)
	log.Println(gw)

	//获取remoteAddr前ip地址
	index = strings.LastIndex(remoteAddr, ":")
	tmpIP := remoteAddr[:index]
	err = setRouteConfig(tmpIP, gw, DNS)
	if err != nil {
		log.Fatal(err)
	}
	

	//forward work

	go func() {
		rbuf := make([]byte, 4096)
		for {
			n, err := ifce.Read(rbuf)
			if err != nil {
				log.Fatal(err)
			}
			log.Print(rbuf[12:16])
			log.Print(" -> ")
			log.Println(rbuf[16:20])
			rch <- rbuf[:n]
		}
	}()
	//不阻塞改函数会导致ifce read 报错 ：The handle is invalid.
	for {
		select {
		case wbuf := <-wch:
			log.Print(wbuf[12:16])
			log.Print(" -> ")
			log.Println(wbuf[16:20])
			ifce.Write(wbuf)
		}
	}
}

/*
	method ：tcp or udp
	network : 192.168.1.10/24
*/
func ForwarderRun(method string, remoteAddr string, network string, DNS string) error {

	rch := make(chan []byte, 4096)
	wch := make(chan []byte, 4096)

	go tuntapRun(rch, wch, remoteAddr, network, DNS)

	//forward work
	if method == "tcp" {
		err := tcp4Forward(remoteAddr, rch, wch)
		if err != nil {
			return err
		}
	} else {
		err := udp4Forward(remoteAddr, rch, wch)
		if err != nil {
			return err
		}
	}
	return nil
}

func tcp4Forward(remoteAddr string, rch chan []byte, wch chan []byte) error {
	serAddr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp4", nil, serAddr)

	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case sdBuf := <-rch:

				_, err := conn.Write(sdBuf)
				if err != nil {
					conn.Close()
					log.Fatal(err)
				}
			}
		}
	}()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				log.Fatal(err)
			}
			wch <- buf[:n]
		}
	}()
	return nil
}

func udp4Forward(remoteAddr string, rch chan []byte, wch chan []byte) error {
	serAddr, err := net.ResolveUDPAddr("udp4", remoteAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp4", nil, serAddr)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case sedbuf := <-rch:
				_, err := conn.Write(sedbuf)
				if err != nil {
					conn.Close()
					log.Fatal(err)
				}
			}
		}
	}()

	go func() {
		rbuf := make([]byte, 4096)
		for {
			n, err := conn.Read(rbuf)
			if err != nil {
				conn.Close()
				log.Fatal(err)
			}
			wch <- rbuf[:n]
		}
	}()

	return nil
}

//windows 命令行配置路由表
//tunNet :192.168.1.0
func setRouteConfig(remoteAddr string, tunNet string, DNS string) error {
	localNetwork, err := findLocalNetwork()
	if err != nil {
		return err
	}
	//删除默认route,会断网
	cmd := "delete 0.0.0.0 mask 0.0.0.0"
	err = exec.Command("route", strings.Split(cmd, " ")...).Run()
	if err != nil {
		return err
	}
	//将远程服务器走正常route
	cmd = fmt.Sprintf("add %s %s", remoteAddr, localNetwork)
	err = exec.Command("route", strings.Split(cmd, " ")...).Run()
	if err != nil {
		return err
	}
	//设置dns服务器走正常route
	cmd = fmt.Sprintf("add %s %s", DNS, localNetwork)
	err = exec.Command("route", strings.Split(cmd, " ")...).Run()
	if err != nil {
		return err
	}
	//将所欲流量走tun/tap
	cmd = fmt.Sprintf("add 0.0.0.0 mask 0.0.0.0 %s metric 60", tunNet)
	err = exec.Command("route", strings.Split(cmd, " ")...).Run()
	if err != nil {
		return err
	}

	return nil
}

func findLocalNetwork() (string, error) {
	ret, err := exec.Command("route", "print").CombinedOutput()
	if err != nil {
		return "", err
	}
	outText := string(ret)
	index := strings.Index(outText, "0.0.0.0")
	outText = outText[index+len("0.0.0.0"):]
	index = strings.Index(outText, "0.0.0.0") + len("0.0.0.0")
	outText = outText[index:]
	i := 0
	for ; i < len(outText); i++ {
		if 48 <= outText[i] && outText[i] <= 57 {
			index = i
			for {
				i++
				if outText[i] == ' ' {
					goto findLocalNetwork
				}
			}
		}
	}
findLocalNetwork:
	outText = outText[index:i]

	return outText, nil
}

//启用utf8编码
func findConnectedIfce() string {
		exec.Command("chcp", "65001").Run()
	out, _ := exec.Command("netsh", strings.Split("interface show interface", " ")...).CombinedOutput()
	text := string(out)

	index := strings.Index(text, "Connected")
	text = text[index+len("Connected"):]
	for i := 0; i < len(text); i++ {
		if text[i] != ' ' {
			text = text[i:]
			break
		}
	}
	for i := 0; i < len(text); i++ {
		if text[i] == ' ' {
			text = text[i:]
			break
		}
	}
	for i := 0; i < len(text); i++ {
		if text[i] != ' ' {
			text = text[i:]
			break
		}
	}
	index=strings.Index(text,"\r\n")
	text=text[:index]
	return text
}

//windows命令行配置dns
func setDNS(DNS string) error{
	ifceName:=findConnectedIfce()
	cmd:=fmt.Sprintf("interface ip set dns \"%s\" static %s primary",ifceName,DNS)
	return exec.Command("netsh",strings.Split(cmd," ")...).Run()
}
