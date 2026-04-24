package secret

import (
	"reflect"
	"testing"
)

func TestMask(t *testing.T) {
	if got := Mask(""); got != "" {
		t.Errorf("Mask(\"\") = %q, want empty", got)
	}
	if got := Mask("password123"); got != "***" {
		t.Errorf("Mask non-empty got %q, want ***", got)
	}
}

type pushReq struct {
	Image    string `json:"image"`
	Username string `json:"username"`
	Password string `json:"password" mask:"true"`
	Token    string `json:"token" mask:"true"`
}

func TestMaskStruct_CopiesAndMasks(t *testing.T) {
	orig := pushReq{
		Image:    "nginx",
		Username: "admin",
		Password: "secret123",
		Token:    "abc",
	}
	got := MaskStruct(orig).(pushReq)

	if got.Password != "***" {
		t.Errorf("Password: got %q, want ***", got.Password)
	}
	if got.Token != "***" {
		t.Errorf("Token: got %q, want ***", got.Token)
	}
	if got.Image != "nginx" || got.Username != "admin" {
		t.Errorf("non-masked fields changed: %+v", got)
	}
	// 原对象不受影响
	if orig.Password != "secret123" {
		t.Errorf("original mutated: %+v", orig)
	}
}

func TestMaskStruct_NilAndNonStruct(t *testing.T) {
	if MaskStruct(nil) != nil {
		t.Error("nil should stay nil")
	}
	if s := MaskStruct("hello"); s != "hello" {
		t.Errorf("non-struct value altered: %v", s)
	}
}

func TestMaskStruct_EmptyFieldStaysEmpty(t *testing.T) {
	r := pushReq{Image: "nginx", Password: ""}
	got := MaskStruct(r).(pushReq)
	if got.Password != "" {
		t.Errorf("empty password masked to %q, want empty", got.Password)
	}
}

func TestMaskStruct_PointerInput(t *testing.T) {
	r := &pushReq{Password: "xxx"}
	got := MaskStruct(r)
	// 指针输入时返回的是新 struct 值（非指针），这是当前实现约定
	if reflect.ValueOf(got).Kind() == reflect.Ptr {
		t.Error("expected returned value to be dereferenced struct")
	}
}
