package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// Save the original DefaultWriter and restore it after the test
	originalWriter := gin.DefaultWriter
	defer func() {
		gin.DefaultWriter = originalWriter
	}()

	// Set up test mode
	gin.SetMode(gin.TestMode)
	buf := new(bytes.Buffer)
	gin.DefaultWriter = buf

	// Create a test router with the logger middleware
	r := gin.New()
	r.Use(Logger())

	// Add test routes
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	r.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"message": "error"})
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedParts  []string
	}{
		{
			name:           "successful request",
			path:           "/test",
			expectedStatus: 200,
			expectedParts: []string{
				time.Now().Format("2006-01-02"), // date
				"GET",                           // method
				"/test",                         // path
			},
		},
		{
			name:           "error request",
			path:           "/error",
			expectedStatus: 500,
			expectedParts: []string{
				time.Now().Format("2006-01-02"), // date
				"GET",                           // method
				"/error",                        // path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the buffer before each test
			buf.Reset()

			// Create a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			req.RemoteAddr = "192.0.2.1:1234" // Add test IP address
			r.ServeHTTP(w, req)

			// Check the response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Get the log output
			logOutput := buf.String()

			// Check that all expected parts are in the log output
			for _, part := range tt.expectedParts {
				assert.True(t, strings.Contains(logOutput, part),
					"Expected log to contain %q, got: %s", part, logOutput)
			}

			// Check if log entry contains client IP
			assert.True(t, strings.Contains(logOutput, "192.0.2.1"),
				"Log should contain client IP")

			// Verify the basic log format (timestamp | method | path | latency | IP)
			assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \| [A-Z]+ \| /.* \| .* \| .*\n`,
				logOutput, "Log format is incorrect")
		})
	}
}
