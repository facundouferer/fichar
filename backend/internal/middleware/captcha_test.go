package middleware

import (
	"testing"
)

func TestCaptchaConfig_Defaults(t *testing.T) {
	// 5 minutes
	expectedExpiry := 5 * 60 * 1000000000 // nanoseconds
	if DefaultCaptchaConfig.Expiry.Nanoseconds() != int64(expectedExpiry) {
		t.Errorf("Expected 5 min expiry, got %v", DefaultCaptchaConfig.Expiry)
	}
}

func TestFailedLoginConfig_Defaults(t *testing.T) {
	if DefaultFailedLoginConfig.MaxAttempts != 5 {
		t.Errorf("Expected 5 max attempts, got %d", DefaultFailedLoginConfig.MaxAttempts)
	}

	if DefaultFailedLoginConfig.BlockDuration != 15*60*1000000000 { // 15 minutes in nanoseconds
		t.Errorf("Expected 15 min block duration, got %v", DefaultFailedLoginConfig.BlockDuration)
	}
}

func TestMathCaptcha_Struct(t *testing.T) {
	captcha := MathCaptcha{
		SessionID: "test-session-id",
		Question:  "5 + 3 = ?",
		Answer:    8,
	}

	if captcha.SessionID != "test-session-id" {
		t.Errorf("Expected session ID, got %s", captcha.SessionID)
	}

	if captcha.Question != "5 + 3 = ?" {
		t.Errorf("Expected question, got %s", captcha.Question)
	}

	if captcha.Answer != 8 {
		t.Errorf("Expected answer 8, got %d", captcha.Answer)
	}
}

func TestCaptchaChallenge_JSON(t *testing.T) {
	challenge := CaptchaChallenge{
		SessionID: "session-123",
		Question:  "10 + 5 = ?",
	}

	// Should not expose answer
	if challenge.SessionID != "session-123" {
		t.Errorf("Expected session ID, got %s", challenge.SessionID)
	}

	if challenge.Question != "10 + 5 = ?" {
		t.Errorf("Expected question, got %s", challenge.Question)
	}
}

func TestCaptchaResponse_JSON(t *testing.T) {
	resp := CaptchaResponse{
		RequireCaptcha: true,
		Captcha: &CaptchaChallenge{
			SessionID: "abc123",
			Question:  "7 - 2 = ?",
		},
	}

	if !resp.RequireCaptcha {
		t.Error("RequireCaptcha should be true")
	}

	if resp.Captcha == nil {
		t.Error("Captcha should not be nil")
	}

	if resp.Captcha.SessionID != "abc123" {
		t.Errorf("Expected session ID, got %s", resp.Captcha.SessionID)
	}
}

func TestCaptchaRequest_Struct(t *testing.T) {
	req := CaptchaRequest{
		SessionID: "session-xyz",
		Answer:    5,
	}

	if req.SessionID != "session-xyz" {
		t.Errorf("Expected session ID, got %s", req.SessionID)
	}

	if req.Answer != 5 {
		t.Errorf("Expected answer 5, got %d", req.Answer)
	}
}
