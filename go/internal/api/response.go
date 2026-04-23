package api

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorCode int

const (
	CodeSuccess ErrorCode = 0

	CodeInvalidParam ErrorCode = 1001
	CodeNotFound     ErrorCode = 1002

	CodeDockerConnFailed ErrorCode = 2001
	CodeBuildFailed      ErrorCode = 2002
	CodeDeployFailed     ErrorCode = 2003

	CodeInternalError ErrorCode = 3001
)

func Success(data interface{}) Response {
	return Response{
		Code:    int(CodeSuccess),
		Message: "success",
		Data:    data,
	}
}

func Error(code ErrorCode, message string) Response {
	return Response{
		Code:    int(code),
		Message: message,
		Data:    nil,
	}
}

func ErrorWithData(code ErrorCode, message string, data interface{}) Response {
	return Response{
		Code:    int(code),
		Message: message,
		Data:    data,
	}
}
