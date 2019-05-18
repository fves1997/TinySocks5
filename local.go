package Socks5

import (
	"io"
	"log"
	"net"
)

type LocalSocks struct {
	LocalAddress *net.TCPAddr
	ProxyAddress *net.TCPAddr
}

func init() {
	log.SetPrefix("[Local]")
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func NewLocalSocks(localHost string, proxyHost string) *LocalSocks {
	localTcpAddress, err := net.ResolveTCPAddr("tcp", localHost)
	if err != nil {
		log.Fatal("本地Host配置错误")
	}
	proxyTcpAddress, err := net.ResolveTCPAddr("tcp", proxyHost)
	if err != nil {
		log.Fatal("ServerHost配置错误")
	}
	return &LocalSocks{
		LocalAddress: localTcpAddress,
		ProxyAddress: proxyTcpAddress,
	}
}

func (localSocks *LocalSocks) Listener() {
	listener, err := net.ListenTCP("tcp", localSocks.LocalAddress)
	if err != nil {
		log.Fatal("本地监听创建失败...", err)
	}
	log.Println("开始接收连接...")
	for {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("接收Tcp连接失败,", err)
			continue
		}
		go localSocks.handleTcpConn(tcpConn)
	}
}

func (localSocks *LocalSocks) handleTcpConn(localConn *net.TCPConn) {
	/* 连接代理服务器 */
	proxyConn, err := net.DialTCP("tcp", nil, localSocks.ProxyAddress)
	if err != nil {
		localConn.Close()
		log.Println(err)
		log.Println("代理结束")
		return
	}
	go func() {
		defer localConn.Close()
		defer proxyConn.Close()
		log.Printf("%s===>%s===>%s", localConn.RemoteAddr(), localConn.LocalAddr(), proxyConn.RemoteAddr())
		_, _ = io.Copy(proxyConn, localConn)
	}()
	go func() {
		defer proxyConn.Close()
		defer localConn.Close()
		log.Printf("%s<===%s<===%s", localConn.RemoteAddr(), localConn.LocalAddr(), proxyConn.RemoteAddr())
		_, _ = io.Copy(localConn, proxyConn)
	}()
}
