package ws_interface

import "errors"

type WsErrorCode uint64

type WsError struct {
	Code    WsErrorCode `json:"code"`
	Message string      `json:"message"`
}

type HandlerError struct {
	Code          WsErrorCode
	InternalError error
}

func NewHandlerError(code WsErrorCode, internal error) *HandlerError {
	err := HandlerError{
		Code:          code,
		InternalError: internal,
	}
	if internal == nil {
		err.InternalError = errors.New(GetErrorMsg(code))
	}
	return &err
}

const (
	BadRequest        WsErrorCode = 4000
	RequestParseError WsErrorCode = 4001
	AuthCheckError    WsErrorCode = 4002
	UnauthorizedError WsErrorCode = 4003

	ContentNotFoundError  WsErrorCode = 4004
	CasinoNotFoundError   WsErrorCode = 4005
	GameNotFoundError     WsErrorCode = 4006
	SessionNotFoundError  WsErrorCode = 4007
	GameNotListedInCasino WsErrorCode = 4008
	GamePaused            WsErrorCode = 4009
	CasinoPaused          WsErrorCode = 4010

	SessionInvalidStateError WsErrorCode = 4100
	SessionFailedOrFinished  WsErrorCode = 4200

	InternalError WsErrorCode = 5000
)

func GetErrorMsg(code WsErrorCode) string {
	switch code {
	case BadRequest:
		return "bad request"
	case RequestParseError:
		return "request parse error"
	case AuthCheckError:
		return "auth check error"
	case UnauthorizedError:
		return "unauthorized error"

	case ContentNotFoundError:
		return "requesting content not found"
	case CasinoNotFoundError:
		return "casino not found"
	case GameNotFoundError:
		return "game not found"
	case SessionNotFoundError:
		return "session not found"
	case GameNotListedInCasino:
		return "game not listed in casino"
	case GamePaused:
		return "game paused"
	case CasinoPaused:
		return "casino paused"

	case SessionInvalidStateError:
		return "action while session invalid state"
	case SessionFailedOrFinished:
		return "session failed or finished"
	case InternalError:
		return "internal server error"
	default:
		return "unknown error"
	}
}
