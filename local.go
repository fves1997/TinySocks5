package Socks5

import (
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
	defer localConn.Close()
	log.Println("接收到程序请求", localConn.RemoteAddr(), "==>", localConn.LocalAddr())
	/* 连接代理服务器 */
	proxyConn, err := net.DialTCP("tcp", nil, localSocks.ProxyAddress)
	if err != nil {
		log.Println("连接代理服务器", err)
		log.Println("代理结束")
		return
	}
	defer proxyConn.Close()
	log.Println("开始转发数据")
	isOver := make(chan bool)

	go func() {
		/* 将代理服务器的结果返回给应用 */
		log.Println("程序<<==========代理服务器")
		for {
			if err = localSocks.transprotData(proxyConn, localConn); err != nil {
				proxyConn.Close()
				localConn.Close()
				break
			}
		}
		log.Println("程序<<==========代理服务器，传输完成")
		isOver <- true
	}()

	/* 将程序请求转发给代理服务器 */
	log.Println("程序==========>>代理服务器")
	for {
		if err := localSocks.transprotData(localConn, proxyConn); err != nil {
			proxyConn.Close()
			localConn.Close()
			break
		}
	}
	log.Println("程序==========>>代理服务器，传输完成")
	<-isOver
	log.Println("代理结束")
}

func (localSocks *LocalSocks) transprotData(from *net.TCPConn, to *net.TCPConn) error {
	for {
		n, err := from.Read(byteBuffer)
		if n > 0 {
			_, err := to.Write(byteBuffer[:n])
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
}
