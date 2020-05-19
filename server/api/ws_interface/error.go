package ws_interface

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
	return &HandlerError{
		Code:          code,
		InternalError: internal,
	}
}

const (
	BadRequest        WsErrorCode = 4000
	RequestParseError WsErrorCode = 4001
	AuthCheckError    WsErrorCode = 4002
	UnauthorizedError WsErrorCode = 4003

	ContentNotFoundError WsErrorCode = 4004
	CasinoNotFoundError  WsErrorCode = 4005
	GameNotFoundError    WsErrorCode = 4006
	SessionNotFoundError WsErrorCode = 4007

	SessionInvalidStateError WsErrorCode = 4100

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

	case SessionInvalidStateError:
		return "action while session invalid state"
	case InternalError:
		return "internal server error"
	default:
		return "unknown error"
	}
}
