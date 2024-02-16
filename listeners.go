package smtpin

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type ListenConfig struct {
	TlsConfig   *tls.Config // if emtpy RequireTls must be false
	RequireTls  bool        // whether to start tcp conn over tls (recomended: true)
	ListenAddr  string      // where to listen (ex: 0.0.0.0)
	MaxPktSize  int         // max allowed size of each packet
	MaxMsgSize  int         // max allowed size of email message
	MailHandler MailHandler
	LogHandler  LogHandler
	Domains     []string // rcpt domains to accept for delivery (accept all domains if emtpy)
	GreetDomain string   // domain witch will be introduced when smtp starts
}

func BasicListenConfig(tlsConfig *tls.Config, port int, mailHandler MailHandler) ListenConfig {
	return ListenConfig{
		tlsConfig,
		tlsConfig == nil,
		fmt.Sprintf("0.0.0.0:%d", port),
		24576,
		24576 * 1000,
		mailHandler,
		func(l LogLevel, s string) { fmt.Println(string(l) + ": " + s) },
		nil,
		"localhost",
	}
}

func Listen(cfg ListenConfig) (err error) {
	var listener net.Listener

	if cfg.TlsConfig == nil {
		cfg.LogHandler(Warn, "tls config is nill")
		if cfg.RequireTls {
			return fmt.Errorf("tls is required, but tls config is nill")
		} else {
			listener, err = net.Listen("tcp", cfg.ListenAddr)
			if err != nil {
				return fmt.Errorf("failed to listen for smtp (no tls) (%s)", err)
			}
		}
	} else {
		if cfg.RequireTls {
			listener, err = tls.Listen("tcp", cfg.ListenAddr, cfg.TlsConfig)
			if err != nil {
				return fmt.Errorf("failed to listen for smtp (tls) (%s)", err)
			}
		} else {
			listener, err = net.Listen("tcp", cfg.ListenAddr)
			if err != nil {
				return fmt.Errorf("failed to listen for smtp (no tls) (%s)", err)
			}
		}
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			cfg.LogHandler(Error, "failed to accept connection: "+err.Error())
			continue
		}
		conn.SetDeadline(time.Now().Add(100 * time.Second))
		go handleConn(conn, &cfg)
	}
}

func handleConn(conn net.Conn, cfg *ListenConfig) {
	defer conn.Close()

	// Initialize smtp session
	s := session{}
	s.listenCfg = cfg
	s.senderAddr = conn.RemoteAddr()
	s.senderIp = conn.RemoteAddr().(*net.TCPAddr).IP
	s.fillDefaultVals()
	senderIpStr := s.senderIp.String()

	// Greet the client
	err := sendRes(conn, resGreeting.withMsg(cfg.GreetDomain+" "+resGreeting.msg), cfg, senderIpStr)
	if err != nil {
		cfg.LogHandler(Error, "failed to greet client: "+err.Error())
		conn.Close()
		return
	}

	// For each tcp packet (i.e each request)
	for {

		// Get the res from the req
		var res response
		pktBuffer := make([]byte, cfg.MaxPktSize)
		pktSize, err := conn.Read(pktBuffer)
		if err != nil {
			cfg.LogHandler(Error, "failed to read packet: "+err.Error())
			return
		} else if pktSize >= cfg.MaxPktSize {
			cfg.LogHandler(Debug, "recieved oversized packet")
			res = resPktTooBig
		} else {
			req := string(pktBuffer[:pktSize])
			cfg.LogHandler(Debug, fmt.Sprintf("%s->%#v", senderIpStr, req))
			res = s.handleReq(req)
		}

		// Tls upgrade (from STARTTLS)
		// TODO test this
		if res.upgradeToTls {
			if cfg.TlsConfig == nil {
				res = resNoTls
			} else {
				sendRes(conn, resConnUpgrade, cfg, senderIpStr)
				tlsConn := tls.Server(conn, cfg.TlsConfig)
				err := tlsConn.Handshake()
				if err != nil {
					res = resFailedTls
					cfg.LogHandler(Error, "failed to start tls: "+err.Error())
				} else {
					conn = net.Conn(tlsConn)
					s.fillDefaultVals()
					res = resGreeting
				}
			}
		}

		// Respond
		if res.respond {
			err = sendRes(conn, res, cfg, senderIpStr)
			if err != nil {
				cfg.LogHandler(Debug, "failed to send response: "+err.Error())
				conn.Close()
				return
			} else if !res.keepAlive {
				cfg.LogHandler(Debug, "ending connection")
				conn.Close()
				return
			}
		}
	}

}

func sendRes(conn net.Conn, res response, cfg *ListenConfig, client string) (err error) {
	if len(res.extendedMsgs) != 0 {
		msgs := append([]string{res.msg}, res.extendedMsgs...)
		msgsLen := len(msgs)
		for i, msg := range msgs {
			var resMsg string
			if i+1 == msgsLen {
				resMsg = fmt.Sprintf("%d %s%s", res.statusCode, msg, crlf)
			} else {
				resMsg = fmt.Sprintf("%d-%s%s", res.statusCode, msg, crlf)
			}
			cfg.LogHandler(Debug, fmt.Sprintf("%s<-%#v", client, resMsg))
			_, err = conn.Write([]byte(resMsg))
			if err != nil {
				return
			}
		}
		return
	}
	resMsg := fmt.Sprintf("%d %s%s", res.statusCode, res.msg, crlf)
	cfg.LogHandler(Debug, fmt.Sprintf("%s<-%#v", client, resMsg))
	_, err = conn.Write([]byte(resMsg))
	return
}
