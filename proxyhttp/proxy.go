package proxyhttp

import (
	_ "fmt"
	"io"
	"net"
	"strings"
)

//HTTPProxy 代理人
type HTTPProxy struct {
	addr string
}

//RunServer 启动服务器
func (h *HTTPProxy) RunServer(addr string) {
	h.addr = addr
	tcpAddr, _ := net.ResolveTCPAddr("tcp", h.addr)
	lister, _ := net.ListenTCP("tcp", tcpAddr)
	for {
		conn, _ := lister.Accept()
		go h.forward(conn)
	}
}

//forward 转发数据包
func (h *HTTPProxy) forward(conn net.Conn) {
	buf := make([]byte, 4096)
	recvLen, _ := conn.Read(buf)
	req := string(buf[:recvLen])

	//解析出域名，端口和 http类型
	domain, port, httpsFlag := h.parseHeader(&req)

	//查询IP地址
	ip, _ := net.ResolveIPAddr("ip", domain)

	// fmt.Println(domain)
	// fmt.Println(ip.String())
	// fmt.Println(port)

	//与服务器建立tcp连接
	serAddr, _ := net.ResolveTCPAddr("tcp", ip.String()+":"+port)
	serconn, err := net.DialTCP("tcp", nil, serAddr)
	if err != nil {
		return
	}

	if httpsFlag {
		//https
		conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		//http
		serconn.Write(buf[:recvLen])
	}

	go func() {

		for {
			io.Copy(serconn, conn)
		}
	}()
	go func() {

		for {
			io.Copy(conn, serconn)
		}
	}()
}

//parseHeader 解析出域名，端口和 http类型
func (h *HTTPProxy) parseHeader(req *string) (domain string, port string, httpsFlag bool) {
	url := ""
	count := 0
	index := 0

	for i := 0; i < len((*req)); i++ {
		if (*req)[i:i+1] == " " {
			count++
			if count == 2 {
				url = (*req)[index+1 : i]
				break
			} else {
				index = i
			}
		}
	}
	port = "80"
	domain = ""
	httpsFlag = false

	//解析出域名和端口
	if strings.Contains(url, "http://") {
		tmpURL := url[len("http://"):]
		tmpIndex := strings.Index(tmpURL, "/")
		if tmpIndex == -1 {
			//没找到http://www.angelbeats.top:5660/
			tmpIndex = strings.Index(tmpURL, ":")
			if tmpIndex == -1 {
				//没有自定义端口
				domain = tmpURL
			} else {
				port = tmpURL[tmpIndex+1:]
			}
		} else {
			pos := strings.Index(tmpURL, ":")
			if pos == -1 {
				//没有自定义端口
				domain = tmpURL[:tmpIndex]
			} else {
				domain = tmpURL[:pos]
				port = tmpURL[pos+1 : tmpIndex]
			}

		}
	} else {
		httpsFlag = true
		tmpIndex := strings.LastIndex(url, ":")
		if tmpIndex == -1 {
			domain = url
		} else {
			domain = url[:tmpIndex]
			port = url[tmpIndex+1:]
		}
	}
	return domain, port, httpsFlag
}
