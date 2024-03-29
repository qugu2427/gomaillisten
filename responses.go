package smtpin

type response struct {
	keepAlive    bool
	respond      bool
	upgradeToTls bool
	statusCode   uint16
	msg          string
	extendedMsgs []string
}

func (r response) withMsg(msg string) response {
	r.msg = msg
	return r
}

func (r response) withExtMsgs(msgs []string) response {
	r.extendedMsgs = msgs
	return r
}

var (
	resGreeting = response{
		true,
		true,
		false,
		codeReady,
		"ESMTP Service Ready",
		nil,
	}
	resHello = response{
		true,
		true,
		false,
		codeOk,
		"Hello",
		nil,
	}
	resInvalidClrf = response{
		true,
		true,
		false,
		codeSyntaxErr,
		"Syntax error: invalid crlf",
		[]string{"Crlf must only occur once at end of each request"},
	}
	resUnknownCmd = response{
		true,
		true,
		false,
		codeSyntaxErr,
		"Syntax error: unknown command",
		nil,
	}
	resInvalidArgNum = response{
		true,
		true,
		false,
		codeSyntaxErr,
		"Syntax error: invalid number of arguments",
		nil,
	}
	resCmdObsolete = response{
		true,
		true,
		false,
		codeNotImplemented,
		"Command not implemented: command obsolete",
		nil,
	}
	resCmdDisabled = response{
		true,
		true,
		false,
		codeNotImplemented,
		"Command not implemented: command disabled",
		nil,
	}
	resCantParseAddr = response{
		true,
		true,
		false,
		codeSyntaxErr,
		"Syntax error: unable to parse valid email address from message",
		nil,
	}
	resNoop = response{
		true,
		true,
		false,
		codeOk,
		"No operation",
		nil,
	}
	resPktTooBig = response{
		true,
		true,
		false,
		codeSyntaxErr,
		"Syntax error: packet too big",
		nil,
	}
	resBye = response{
		false,
		true,
		false,
		codeBye,
		"Goodbye",
		nil,
	}
	resReset = response{
		true,
		true,
		false,
		codeOk,
		"Session reset",
		nil,
	}
	resInvalidSequence = response{
		true,
		true,
		false,
		codeOk,
		"Invalid command sequence",
		nil,
	}
	resAcceptingMailFrom = response{
		true,
		true,
		false,
		codeOk,
		"Accepting mail",
		nil,
	}
	resRcptAdded = response{
		true,
		true,
		false,
		codeOk,
		"Added recipient",
		nil,
	}
	resMailAccepted = response{
		true,
		true,
		false,
		codeOk,
		"Mail accepted",
		nil,
	}
	resStartMail = response{
		true,
		true,
		false,
		codeStartMail,
		"Start mail",
		nil,
	}
	resBlank = response{
		true,
		false,
		false,
		0,
		"",
		nil,
	}
	resSpfErr = response{
		true,
		true,
		false,
		codeActionAborted,
		"Spf check error",
		nil,
	}
	resSpfFail = response{
		true,
		true,
		false,
		codeActionNotTaken,
		"Spf check failed",
		nil,
	}
	resConnUpgrade = response{
		true,
		true,
		true,
		codeReady,
		"Ready for tls upgrade",
		nil,
	}
	resNoTls = response{
		true,
		true,
		false,
		codeNotImplemented,
		"Command not implemented: tls not available",
		nil,
	}
	resFailedTls = response{
		true,
		true,
		false,
		codeAuthFailure,
		"Tls handshake failed",
		nil,
	}
	resMsgTooBig = response{
		true,
		true,
		false,
		codeMsgTooBig,
		"Message too big",
		nil,
	}
	resNotLocal = response{
		true,
		true,
		false,
		codeNotLocal,
		"User not local",
		nil,
	}
	resDkimFailed = response{
		true,
		true,
		false,
		codeActionNotTaken,
		"Dkim authentication failed",
		nil,
	}
)
