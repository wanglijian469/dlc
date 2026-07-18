package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

var etagAPIs = map[string]bool{
	"/api/home":          true,
	"/api/site-meta":     true,
	"/api/layout-config": true,
	"/api/menus":         true,
	"/api/categories":    true,
}

type bufferedResponseWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *bufferedResponseWriter) WriteHeader(status int) { w.status = status }
func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}
func (w *bufferedResponseWriter) WriteString(value string) (int, error) {
	return w.body.WriteString(value)
}

func PublicETag() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || !etagAPIs[c.Request.URL.Path] {
			c.Next()
			return
		}
		original := c.Writer
		buffered := &bufferedResponseWriter{ResponseWriter: original, status: http.StatusOK}
		c.Writer = buffered
		c.Next()

		sum := sha256.Sum256(buffered.body.Bytes())
		etag := `"` + hex.EncodeToString(sum[:16]) + `"`
		original.Header().Set("ETag", etag)
		// Homepage and directory data are editable in the CMS. Keep the ETag so
		// unchanged payloads remain inexpensive, but force a normal page refresh
		// to revalidate instead of showing an old banner for up to a minute.
		original.Header().Set("Cache-Control", "public, no-cache")
		if c.GetHeader("If-None-Match") == etag && buffered.status >= 200 && buffered.status < 300 {
			original.WriteHeader(http.StatusNotModified)
			return
		}
		original.WriteHeader(buffered.status)
		if c.Request.Method != http.MethodHead {
			_, _ = original.Write(buffered.body.Bytes())
		}
	}
}
