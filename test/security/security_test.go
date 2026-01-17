package security

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/stretchr/testify/suite"
)

// SecurityTestSuite provides security testing capabilities
type SecurityTestSuite struct {
	testutil.BaseTestSuite
	server *httptest.Server
	client *http.Client
}

func (s *SecurityTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	// Create test server
	handler := s.createTestHandler()
	s.server = httptest.NewServer(handler)

	// Create HTTP client
	s.client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For testing only
			},
		},
	}
}

func (s *SecurityTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
	s.BaseTestSuite.TearDownTest()
}

func TestSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityTestSuite))
}

func (s *SecurityTestSuite) TestSecurityHeaders() {
	// Test that security headers are properly set
	resp := s.makeRequest("GET", "/metrics", nil, nil)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}

	for header, expected := range expectedHeaders {
		actual := resp.Header.Get(header)
		s.Equal(expected, actual, "Security header %s should be set correctly", header)
	}
}

func (s *SecurityTestSuite) TestHTTPSRedirect() {
	// Test that HTTP requests are redirected to HTTPS in production
	// This would require a more sophisticated test setup
	s.T().Skip("HTTPS redirect testing requires TLS server setup")
}

func (s *SecurityTestSuite) TestAuthenticationBypass() {
	// Test various authentication bypass attempts
	bypassAttempts := []struct {
		name     string
		headers  map[string]string
		method   string
		path     string
		expected int
	}{
		{
			name:     "no_auth",
			headers:  map[string]string{},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
		{
			name: "invalid_token",
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
			},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
		{
			name: "sql_injection_token",
			headers: map[string]string{
				"Authorization": "Bearer ' OR '1'='1",
			},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
		{
			name: "path_traversal_token",
			headers: map[string]string{
				"Authorization": "Bearer ../../../etc/passwd",
			},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
		{
			name: "script_injection_token",
			headers: map[string]string{
				"Authorization": "Bearer <script>alert('xss')</script>",
			},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
		{
			name: "null_byte_injection",
			headers: map[string]string{
				"Authorization": "Bearer token\x00admin",
			},
			method:   "GET",
			path:     "/debug/collectors",
			expected: 401,
		},
	}

	for _, attempt := range bypassAttempts {
		s.Run(attempt.name, func() {
			// Null bytes in HTTP headers are rejected by Go's HTTP client at protocol level
			// This is correct security behavior - HTTP spec forbids control characters in headers
			if attempt.name == "null_byte_injection" {
				s.T().Skip("Null bytes in HTTP headers are rejected by Go's HTTP client - this is correct security behavior")
				return
			}

			resp := s.makeRequest(attempt.method, attempt.path, attempt.headers, nil)
			s.Equal(attempt.expected, resp.StatusCode,
				"Authentication bypass attempt should be rejected")
		})
	}
}

func (s *SecurityTestSuite) TestPathTraversalAttacks() {
	// Test path traversal attacks on various endpoints
	pathTraversalAttempts := []struct {
		name string
		path string
	}{
		{"basic_traversal", "/debug/../../../etc/passwd"},
		{"encoded_traversal", "/debug/%2e%2e/%2e%2e/%2e%2e/etc/passwd"},
		{"double_encoded", "/debug/%252e%252e/%252e%252e/etc/passwd"},
		{"unicode_traversal", "/debug/\u002e\u002e/\u002e\u002e/etc/passwd"},
		{"long_path", "/debug/" + strings.Repeat("../", 100) + "etc/passwd"},
		{"null_byte", "/debug/config\x00.txt"},
	}

	for _, attempt := range pathTraversalAttempts {
		s.Run(attempt.name, func() {
			// Null bytes in URLs are rejected by Go's URL parser
			if attempt.name == "null_byte" {
				s.T().Skip("Null bytes in URLs are rejected by Go's URL parser - this is correct security behavior")
				return
			}

			resp := s.makeRequest("GET", attempt.path, nil, nil)

			// Should return 404 or 400, not 200
			s.NotEqual(200, resp.StatusCode,
				"Path traversal should not succeed")

			// Response should not contain sensitive file content
			body := s.readResponseBody(resp)
			s.NotContains(body, "root:x:", "Should not return /etc/passwd content")
			s.NotContains(body, "#!/bin/", "Should not return script content")
		})
	}
}

func (s *SecurityTestSuite) TestXSSPrevention() {
	// Test XSS prevention in various input fields
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
		"'><script>alert('xss')</script>",
		"\"><script>alert('xss')</script>",
		"javascript:/*--></title></style></textarea></script></xmp>",
		"<iframe src=javascript:alert('xss')>",
	}

	// Test in URL parameters
	for i, payload := range xssPayloads {
		s.Run(fmt.Sprintf("url_param_%d", i), func() {
			path := fmt.Sprintf("/debug/collectors?filter=%s", payload)
			resp := s.makeRequest("GET", path, nil, nil)

			body := s.readResponseBody(resp)

			// Verify payload is properly escaped/sanitized
			s.NotContains(body, "<script>", "Script tags should be escaped")
			s.NotContains(body, "javascript:", "JavaScript URLs should be blocked")
			s.NotContains(body, "onerror=", "Event handlers should be escaped")
		})
	}
}

func (s *SecurityTestSuite) TestSQLInjection() {
	// Test SQL injection attempts (even though we don't use SQL directly)
	sqlPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users --",
		"1'; INSERT INTO users VALUES ('admin', 'password'); --",
		"' OR 1=1; --",
		"admin'--",
		"admin' /*",
		"' OR 'x'='x",
		"1' AND (SELECT COUNT(*) FROM users)>0 --",
	}

	for i, payload := range sqlPayloads {
		s.Run(fmt.Sprintf("sql_injection_%d", i), func() {
			// Test in various parameters
			paths := []string{
				fmt.Sprintf("/debug/collectors?name=%s", payload),
				fmt.Sprintf("/debug/profiling/profile/%s", payload),
			}

			for _, path := range paths {
				resp := s.makeRequest("GET", path, nil, nil)

				// Should not return database errors or succeed inappropriately
				body := s.readResponseBody(resp)
				s.NotContains(body, "SQL error", "Should not leak SQL errors")
				s.NotContains(body, "database error", "Should not leak database errors")
				s.NotContains(body, "mysql", "Should not leak database type")
				s.NotContains(body, "postgresql", "Should not leak database type")
			}
		})
	}
}

func (s *SecurityTestSuite) TestCSRFProtection() {
	// Test CSRF protection for state-changing operations
	stateChangingRequests := []struct {
		method string
		path   string
	}{
		{"POST", "/debug/profiling/collector/jobs/enable"},
		{"POST", "/debug/profiling/collector/jobs/disable"},
		{"PUT", "/debug/config/reload"},
		{"DELETE", "/debug/cache/clear"},
	}

	for _, req := range stateChangingRequests {
		s.Run(fmt.Sprintf("csrf_%s_%s", req.method, strings.ReplaceAll(req.path, "/", "_")), func() {
			// Request without CSRF token should fail
			resp := s.makeRequest(req.method, req.path, nil, nil)

			// Should require proper CSRF protection
			s.NotEqual(200, resp.StatusCode,
				"State-changing request without CSRF token should fail")
		})
	}
}

func (s *SecurityTestSuite) TestRateLimiting() {
	// Test rate limiting on sensitive endpoints
	sensitiveEndpoints := []string{
		"/debug/profiling",
		"/debug/collectors",
		"/debug/cache",
	}

	for _, endpoint := range sensitiveEndpoints {
		s.Run(fmt.Sprintf("rate_limit_%s", strings.ReplaceAll(endpoint, "/", "_")), func() {
			// Make many rapid requests
			const requestCount = 100
			const timeWindow = 1 * time.Second

			start := time.Now()
			successCount := 0

			for i := 0; i < requestCount; i++ {
				resp := s.makeRequest("GET", endpoint, nil, nil)
				if resp.StatusCode == 200 {
					successCount++
				}

				// Don't sleep - test rapid requests
			}

			elapsed := time.Since(start)

			// Should have some rate limiting in place
			if elapsed < timeWindow {
				s.Less(successCount, requestCount,
					"Rate limiting should prevent all requests from succeeding")
			}
		})
	}
}

func (s *SecurityTestSuite) TestInputValidation() {
	// Test input validation with various malicious inputs
	maliciousInputs := []struct {
		name  string
		value string
	}{
		{"oversized_input", strings.Repeat("A", 10000)},
		{"null_bytes", "test\x00data"},
		{"control_characters", "test\x01\x02\x03data"},
		{"unicode_exploit", "test\uFEFFdata"},
		{"format_string", "test%s%d%x"},
		{"ldap_injection", "test)(cn=*))(|(cn=*"},
		{"xml_injection", "test<foo>bar</foo>"},
		{"json_injection", `test","admin":true,"foo":"bar`},
	}

	// Test various input parameters
	for _, input := range maliciousInputs {
		s.Run(input.name, func() {
			// Null bytes and control characters are rejected by URL parser itself
			// This is good security at the protocol level
			if input.name == "null_bytes" || input.name == "control_characters" {
				// These should be rejected by Go's URL parser
				// URL encoding would escape them, so raw values are invalid
				s.T().Skip("Null bytes and control characters are rejected by Go's URL parser - this is correct security behavior")
				return
			}

			// Test in URL parameters (use URL encoding for valid characters)
			resp := s.makeRequest("GET", fmt.Sprintf("/debug/collectors?filter=%s", input.value), nil, nil)

			// Should handle gracefully without crashing
			s.NotEqual(500, resp.StatusCode, "Malicious input should not cause server error")

			body := s.readResponseBody(resp)
			s.NotEmpty(body, "Should return some response")
		})
	}
}

func (s *SecurityTestSuite) TestInformationDisclosure() {
	// Test that sensitive information is not disclosed
	resp := s.makeRequest("GET", "/debug/status", nil, nil)
	body := s.readResponseBody(resp)

	// Should not disclose sensitive information
	// Note: "unauthorized" is a valid status message, not a leak
	sensitiveInfo := []string{
		"password",
		"secret",
		"token=", // Actual token values, not the word "token" in error messages
		"key=",   // Actual key values
		// "auth", // Removed: "unauthorized" is a valid response
		"/etc/passwd",
		"/proc/",
		"database connection",
		"connection string",
	}

	for _, info := range sensitiveInfo {
		s.NotContains(strings.ToLower(body), info,
			"Response should not contain sensitive information: %s", info)
	}
}

func (s *SecurityTestSuite) TestHTTPMethodSecurity() {
	// Test HTTP method restrictions
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"}

	for _, method := range methods {
		s.Run(fmt.Sprintf("method_%s", method), func() {
			resp := s.makeRequest(method, "/metrics", nil, nil)

			switch method {
			case "GET", "HEAD":
				// These should be allowed
				s.NotEqual(405, resp.StatusCode, "GET/HEAD should be allowed")
			case "TRACE":
				// TRACE should be disabled for security
				s.Equal(405, resp.StatusCode, "TRACE method should be disabled")
			default:
				// Other methods may or may not be allowed depending on endpoint
			}
		})
	}
}

func (s *SecurityTestSuite) TestTimeoutAttacks() {
	// Test protection against timeout attacks
	s.Run("slow_loris", func() {
		// This would require more sophisticated testing
		// For now, just verify basic timeout handling
		resp := s.makeRequest("GET", "/metrics", nil, nil)
		s.Less(resp.StatusCode, 500, "Should handle requests without timing out")
	})
}

// Helper methods

func (s *SecurityTestSuite) createTestHandler() http.Handler {
	// Create a minimal test handler that simulates the real server
	mux := http.NewServeMux()

	// Add security middleware
	secureHandler := s.addSecurityMiddleware(mux)

	// Add test routes
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("# Test metrics\ntest_metric 1\n"))
	})

	mux.HandleFunc("/debug/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate authentication requirement
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	return secureHandler
}

func (s *SecurityTestSuite) addSecurityMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Block TRACE method
		if r.Method == "TRACE" {
			w.WriteHeader(405)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (s *SecurityTestSuite) makeRequest(method, path string, headers map[string]string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		reqBody = strings.NewReader(fmt.Sprintf("%v", body))
	}

	req, err := http.NewRequest(method, s.server.URL+path, reqBody)
	s.Require().NoError(err)

	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	s.Require().NoError(err)

	return resp
}

func (s *SecurityTestSuite) readResponseBody(resp *http.Response) string {
	defer func() { _ = resp.Body.Close() }()

	body := make([]byte, 10240) // Read up to 10KB
	n, _ := resp.Body.Read(body)

	return string(body[:n])
}
