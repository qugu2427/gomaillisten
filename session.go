package main

import (
	"fmt"
	"net"
	"strings"

	spf "github.com/qugu2427/gospf"
)

/*
since smtp requests are not sent all at once
this is basically a struct to keep track of an
in-propgress smtp request
*/
type session struct {
	senderIp      net.IP
	senderAddr    net.Addr
	saidHello     bool
	extended      bool
	helloFrom     string
	mailFrom      string
	recipients    []string
	bodyStarted   bool
	body          string
	bodyCompleted bool
	listenCfg     *ListenConfig
	spfResult     spf.Result
}

func (s *session) fillDefaultVals() {
	s.saidHello = false
	s.extended = false
	s.helloFrom = ""
	s.mailFrom = ""
	s.recipients = []string{}
	s.bodyStarted = false
	s.body = ""
	s.bodyCompleted = false
	s.spfResult = spf.ResultNone
}

func (s *session) handleReq(req string) response {

	// Insure crlf only appears once at end of string
	if strings.Index(req, crlf) != len(req)-2 {
		return resInvalidClrf
	}

	// Handle body messages
	if s.bodyStarted && !s.bodyCompleted {
		return s.handleBody(req)
	}

	// Translate req into space-seperated args
	req = strings.TrimSuffix(req, crlf)
	args := argSplit(req)
	argsLen := len(args)
	if argsLen == 0 {
		return resNoop
	}

	cmd := strings.ToUpper(args[0])
	if cmd == cmdMail && argsLen > 1 && strings.ToUpper(args[1]) == "FROM" {
		return s.handleMailFrom(req, args, argsLen)
	} else if cmd == cmdRcpt && argsLen > 1 && strings.ToUpper(args[1]) == "TO" {
		return s.handleRcptTo(req, args, argsLen)
	}
	switch cmd {
	case cmdEhlo:
		return s.handleEhlo(args, argsLen)
	case cmdHelo:
		return s.handleHelo(args, argsLen)
	case cmdData:
		return s.handleData()
	case cmdQuit:
		return resBye
	case cmdRset:
		s.fillDefaultVals()
		return resReset
	case cmdVrfy:
		return resCmdDisabled // TODO mabye
	case cmdNoop:
		return resNoop
	case cmdTurn:
		return resCmdObsolete
	case cmdExpn:
		return resCmdDisabled // TODO mabye
	case cmdHelp:
		return resCmdDisabled // TODO mabye
	case cmdSend:
		return resCmdObsolete
	case cmdSaml:
		return resCmdObsolete
	case cmdRelay:
		return resCmdObsolete
	case cmdSoml:
		return resCmdObsolete
	case cmdTls:
		return resCmdObsolete
	case cmdStartTls:
		return resConnUpgrade
	case cmdStartSsl:
		return resCmdObsolete
	case cmdAuth:
		return resCmdDisabled // TODO add optional auth
	}
	return resUnknownCmd
}

func (s *session) handleHelo(args []string, argsLen int) response {
	if s.saidHello {
		return resInvalidSequence
	}
	if argsLen != 2 || args[1] == "" {
		return resInvalidArgNum
	}
	s.helloFrom = strings.TrimSpace(args[1])
	s.saidHello = true
	return resHello.withMsg("Hello " + s.helloFrom)
}

func (s *session) handleEhlo(args []string, argsLen int) response {
	s.extended = true
	if s.saidHello {
		return resInvalidSequence
	}
	if argsLen != 2 || args[1] == "" {
		return resInvalidArgNum
	}
	s.helloFrom = strings.TrimSpace(args[1])
	s.saidHello = true
	return resHello.withMsg("Hello " + s.helloFrom).withExtMsgs([]string{"STARTTLS", fmt.Sprintf("SIZE %d", s.listenCfg.MaxMsgSize)})
}

func (s *session) handleMailFrom(req string, args []string, argsLen int) response {
	if !s.saidHello || s.mailFrom != "" {
		return resInvalidSequence
	}
	if argsLen < 3 {
		return resInvalidArgNum
	}
	emailFound, email := findEmailInLine(req)
	if !emailFound {
		return resCantParseAddr
	}
	senderDomain := email[strings.Index(email, "@")+1:]
	senderIp := net.IP(s.senderIp)
	var err error
	s.spfResult, err = spf.CheckHost(senderIp, senderDomain)
	if err != nil {
		s.listenCfg.LogHandler(Error, "failed to check spf: "+err.Error())
		return resSpfErr
	}
	if s.spfResult == spf.ResultFail {
		return resSpfFail
	}
	s.mailFrom = email
	return resAcceptingMailFrom.withMsg("Accepting mail from " + s.mailFrom)
}

func (s *session) handleRcptTo(req string, args []string, argsLen int) response {
	if !s.saidHello || s.mailFrom == "" {
		return resInvalidSequence
	}
	if argsLen < 3 {
		return resInvalidArgNum
	}
	emailFound, email := findEmailInLine(req)
	if !emailFound {
		return resCantParseAddr
	}
	if s.listenCfg.Domains != nil {
		domain := strings.Split(email, "@")[1]
		allowed := false
		for _, allowedDomain := range s.listenCfg.Domains {
			if domain == allowedDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return resNotLocal
		}
	}
	s.recipients = append(s.recipients, email)
	return resRcptAdded.withMsg("Added recipient " + email)
}

func (s *session) handleData() response {
	if s.bodyStarted || len(s.recipients) == 0 || s.mailFrom == "" || !s.saidHello {
		return resInvalidSequence
	}
	s.bodyStarted = true
	return resStartMail
}

func (s *session) handleBody(req string) response {
	s.body += req
	if strings.HasSuffix(s.body, bodyEnd) {
		s.bodyCompleted = true
		if len(s.body) > s.listenCfg.MaxMsgSize {
			return resMsgTooBig
		}
		if 1 == 2 { // TODO check dkim
			return resDkimFailed
		}
		mail := Mail{
			s.senderAddr,
			s.mailFrom,
			s.body,
			s.spfResult,
		}
		s.listenCfg.MailHandler(&mail)
		return resMailAccepted
	}
	return resBlank
}
