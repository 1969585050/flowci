package api

import (
	"testing"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := Success(data)

	if resp.Code != 0 {
		t.Errorf("Success() Code = %d, want 0", resp.Code)
	}

	if resp.Message != "success" {
		t.Errorf("Success() Message = %q, want %q", resp.Message, "success")
	}

	if resp.Data == nil {
		t.Error("Success() Data should not be nil")
	}
}

func TestError(t *testing.T) {
	resp := Error(CodeInvalidParam, "test error message")

	if resp.Code != int(CodeInvalidParam) {
		t.Errorf("Error() Code = %d, want %d", resp.Code, CodeInvalidParam)
	}

	if resp.Message != "test error message" {
		t.Errorf("Error() Message = %q, want %q", resp.Message, "test error message")
	}

	if resp.Data != nil {
		t.Error("Error() Data should be nil")
	}
}

func TestErrorWithData(t *testing.T) {
	data := map[string]string{"field": "value"}
	resp := ErrorWithData(CodeNotFound, "resource not found", data)

	if resp.Code != int(CodeNotFound) {
		t.Errorf("ErrorWithData() Code = %d, want %d", resp.Code, CodeNotFound)
	}

	if resp.Message != "resource not found" {
		t.Errorf("ErrorWithData() Message = %q, want %q", resp.Message, "resource not found")
	}

	if resp.Data == nil {
		t.Error("ErrorWithData() Data should not be nil")
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		want    int
		desc    string
	}{
		{CodeSuccess, 0, "success"},
		{CodeInvalidParam, 1001, "invalid parameter"},
		{CodeNotFound, 1002, "resource not found"},
		{CodeDockerConnFailed, 2001, "docker connection failed"},
		{CodeBuildFailed, 2002, "build failed"},
		{CodeDeployFailed, 2003, "deploy failed"},
		{CodeInternalError, 3001, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if int(tt.code) != tt.want {
				t.Errorf("ErrorCode %s = %d, want %d", tt.desc, tt.code, tt.want)
			}
		})
	}
}
