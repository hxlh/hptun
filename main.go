package main

import "proxy/proxyhttp"

func main() {
	proxySer:=&proxyhttp.HTTPProxy{}
	proxySer.RunServer("127.0.0.1:1136")
}