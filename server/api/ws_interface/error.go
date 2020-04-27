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
	BadRequest           WsErrorCode = 4000
	RequestParseError    WsErrorCode = 4001
	AuthCheckError       WsErrorCode = 4002
	UnauthorizedError    WsErrorCode = 4003
	ContentNotFoundError WsErrorCode = 4004

	InternalError WsErrorCode = 5000
)

func GetErrorMsg(code WsErrorCode) string {
	switch code {
	case BadRequest:
		return "bad request"
	case RequestParseError:
		return "request parse error"
	case UnauthorizedError:
		return "unauthorized error"
	case InternalError:
		return "internal server error"
	default:
		return "unknown error"
	}
}
