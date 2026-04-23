package middleware

import (
	"testing"
	"time"
)

func TestNewIPRateLimiter(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 3,
		Burst:             5,
	}

	rl := NewIPRateLimiter(config)
	if rl == nil {
		t.Fatal("expected rate limiter, got nil")
	}

	if rl.config.RequestsPerMinute != 3 {
		t.Errorf("expected 3 requests per minute, got %d", rl.config.RequestsPerMinute)
	}

	if rl.config.Burst != 5 {
		t.Errorf("expected burst of 5, got %d", rl.config.Burst)
	}
}

func TestIPRateLimiterMap_Allows(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 10,
		Burst:             5,
	}

	rlm := NewIPRateLimiterMap(config, 1*time.Minute)

	ip := "192.168.1.1"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !rlm.Allow(ip) {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th request should be rejected
	if rlm.Allow(ip) {
		t.Error("6th request should be rejected")
	}
}

func TestIPRateLimiterMap_DifferentIPs(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerMinute: 10,
		Burst:             2,
	}

	rlm := NewIPRateLimiterMap(config, 1*time.Minute)

	// Each IP should have independent limit
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// IP1 hits limit
	for i := 0; i < 2; i++ {
		if !rlm.Allow(ip1) {
			t.Errorf("IP1 request %d should be allowed", i+1)
		}
	}

	if rlm.Allow(ip1) {
		t.Error("IP1 should be rate limited")
	}

	// IP2 should still be allowed
	if !rlm.Allow(ip2) {
		t.Error("IP2 should not be rate limited yet")
	}
}

func TestRateLimitConfig_Defaults(t *testing.T) {
	if LoginRateLimiterConfig.RequestsPerMinute != 3 {
		t.Errorf("LoginRateLimiterConfig should be 3, got %d", LoginRateLimiterConfig.RequestsPerMinute)
	}

	if AttendanceRateLimiterConfig.RequestsPerMinute != 10 {
		t.Errorf("AttendanceRateLimiterConfig should be 10, got %d", AttendanceRateLimiterConfig.RequestsPerMinute)
	}

	if LoginRateLimiterConfig.Burst != 5 {
		t.Errorf("LoginRateLimiterConfig.Burst should be 5, got %d", LoginRateLimiterConfig.Burst)
	}

	if AttendanceRateLimiterConfig.Burst != 15 {
		t.Errorf("AttendanceRateLimiterConfig.Burst should be 15, got %d", AttendanceRateLimiterConfig.Burst)
	}
}
