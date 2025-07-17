package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
)

// Logger is a middleware that logs API requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start the timer
		start := time.Now()
		// Process the request
		c.Next() // Call the next handler in the chain

		// Log the request details
		latency := time.Since(start)
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		//log with different colors based on status code
		switch {
		case status >= 500:
			//server error - red
			gin.DefaultWriter.Write([]byte("[ERROR]"))
		case status >= 400:
			//client error - yellow
			gin.DefaultWriter.Write([]byte("[WARN]"))
		default:
			//success - green
			gin.DefaultWriter.Write([]byte("[INFO]"))
		}
		gin.DefaultWriter.Write([]byte(time.Now().Format("2006-01-02 15:04:05")))
		gin.DefaultWriter.Write([]byte(" | "))
		gin.DefaultWriter.Write([]byte(method))
		gin.DefaultWriter.Write([]byte(" | "))
		gin.DefaultWriter.Write([]byte(path))
		gin.DefaultWriter.Write([]byte(" | "))
		gin.DefaultWriter.Write([]byte(latency.String()))
		gin.DefaultWriter.Write([]byte(" | "))
		gin.DefaultWriter.Write([]byte(c.ClientIP()))
		gin.DefaultWriter.Write([]byte("\n"))
	}
}
