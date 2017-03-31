package core

import (
	"flag"
	log "github.com/kermitbu/grapes/log"
	"io"
	"net"
)

func (c *CoreServer) Handle(id uint16, f handleFunc) {

	if c.allHandlerFunc == nil {
		c.allHandlerName = make(map[uint16]string)
		c.allHandlerFunc = make(map[uint16]handleFunc)
	}
	if _, ok := c.allHandlerFunc[id]; ok {
		log.Warn("Register called twice for handles ", id)
	}
	c.allHandlerFunc[id] = f

}

func (c *CoreServer) deliverMessage(conn net.Conn, msghead *MessageHead, body []byte) {
	if handler, ok := c.allHandlerFunc[msghead.Cmd]; ok {

		req := &GRequest{connect: &conn, head: msghead, DataLen: msghead.BodyLen, DataBuffer: body}
		rsp := &GResponse{connect: &conn}
		handler(req, rsp)
	} else {
		log.Warn("Never register processing method [%v]", msghead.Cmd)
	}
}

var port = flag.String("port", "10000", "指定服务器监听的端口号")
var conf = flag.String("conf", "", "指定服务器的配置文件")

func (c *CoreServer) InitComplete() {
	// 作为客户端，连接服务器，并准备接收数据

	// for i := 0; i < 4; i++ {
	// 	addr, err := net.ResolveTCPAddr("tcp", ":4040")
	// 	checkErr(err)
	// 	conn, err := net.DialTCP("tcp", nil, addr)
	// 	checkErr(err)

	// 	allClientConnects[addr.String()] = conn

	// 	defer conn.Close()
	// 	go handlClientConn(conn)
	// }
	// 作为服务器端监听端口
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:"+*port)
	if err != nil {
		log.Fatal(err.Error())
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("服务器正常启动,开始监听%v端口", *port)

	complete := make(chan int, 1)

	go func(listen *net.TCPListener) {
		for {
			conn, err := listen.Accept()
			if err != nil {
				log.Fatal(err.Error())
			}
			go c.handleServerConn(conn)
		}
	}(listen)

	<-complete
}

const BufLength = 1024

func (c *CoreServer) handleServerConn(conn net.Conn) {
	log.Info("===>>> New Connection ===>>>")

	head := new(MessageHead)

	hasError := false
	unhandledData := make([]byte, 0)

	for false == hasError {
		buf := make([]byte, BufLength)
		for {
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				hasError = true
			}

			unhandledData = append(unhandledData, buf[:n]...)

			if n != BufLength {
				break
			}
		}

		log.Debug("接收到数据：%v", unhandledData)

		for nil == head.Unpack(unhandledData) {
			msgLen := head.BodyLen + uint16(head.HeadLen)
			msgData := unhandledData[:msgLen]
			unhandledData = unhandledData[msgLen:]

			c.deliverMessage(conn, head, msgData[head.HeadLen:])
		}
	}
	log.Info("===>>> Connection closed ===>>>")
}

type handleFunc func(request *GRequest, response *GResponse)

type CoreServer struct {
	allHandlerName    map[uint16]string
	allHandlerFunc    map[uint16]handleFunc
	allClientConnects map[string]*net.TCPConn
}

type ServiceCollection map[string][]ServiceNode
type ServiceNode struct {
	name    string
	connect *net.Conn
}

var services = make(ServiceCollection)

func (c *CoreServer) GetServiceNodes() ServiceCollection {
	return services
}

func (c *CoreServer) GetServiceNodeByName(name string) []ServiceNode {
	nodes := make([]ServiceNode, 0)

	return nodes
}
