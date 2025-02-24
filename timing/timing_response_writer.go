package timing

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TimingResponseWriter wraps a gin.ResponseWriter to capture and inject timing data
type TimingResponseWriter struct {
	gin.ResponseWriter
	timing      *RenderTiming
	buffer      *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// NewTimingResponseWriter creates a new response writer that injects timing information
func NewTimingResponseWriter(w gin.ResponseWriter, timing *RenderTiming) *TimingResponseWriter {
	return &TimingResponseWriter{
		ResponseWriter: w,
		timing:         timing,
		buffer:         &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code but doesn't write it immediately
func (w *TimingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	// Don't call the underlying WriteHeader yet - we'll do that when we flush
}

// Write buffers the data for HTML or passes it through for other content types
func (w *TimingResponseWriter) Write(data []byte) (int, error) {
	// If not HTML content, pass through directly
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		if !w.wroteHeader {
			w.ResponseWriter.WriteHeader(w.statusCode)
			w.wroteHeader = true
		}
		return w.ResponseWriter.Write(data)
	}

	// For HTML, buffer the response to modify it later
	return w.buffer.Write(data)
}

// WriteString is a convenience method that calls Write
func (w *TimingResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Flush writes the collected data with timing information
func (w *TimingResponseWriter) Flush() {
	if w.buffer.Len() == 0 {
		// Nothing to flush
		return
	}

	// End the page timing now
	w.timing.EndPage()

	// Get the HTML content
	html := w.buffer.String()

	// Check if this is HTML content - first check the header
	contentType := w.Header().Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml")

	// Also check if the content looks like HTML even if the header doesn't indicate it
	if !isHTML && (strings.Contains(html, "<html") || strings.Contains(html, "<body")) {
		isHTML = true
	}

	if !isHTML {
		// Not HTML content, write as-is
		if !w.wroteHeader {
			w.ResponseWriter.WriteHeader(w.statusCode)
			w.wroteHeader = true
		}
		w.ResponseWriter.Write(w.buffer.Bytes())
		return
	}

	// Format the timing info with microsecond precision for values < 1ms
	var pageTimeStr, templateTimeStr string

	pageDuration := w.timing.GetPageDuration()
	if pageDuration < 1.0 {
		pageTimeStr = fmt.Sprintf("%.2f μs", pageDuration*1000) // Convert to microseconds
	} else {
		pageTimeStr = fmt.Sprintf("%.2f ms", pageDuration)
	}

	templateDuration := w.timing.GetTemplateDuration()
	if templateDuration < 1.0 {
		templateTimeStr = fmt.Sprintf("%.2f μs", templateDuration*1000) // Convert to microseconds
	} else {
		templateTimeStr = fmt.Sprintf("%.2f ms", templateDuration)
	}

	// Prepare the timing footer HTML
	timingFooter := fmt.Sprintf(`<footer class="footer footer-center p-4 bg-base-200 text-base-content">
		<div class="flex justify-center items-center">
			<p>Copyright © 2024 - All rights reserved. Powered by bingbong-go Page: <strong>%s</strong> Template: <strong>%s</strong></p>
		</div>
	</footer>`, pageTimeStr, templateTimeStr)

	// Find where to insert the footer
	footerPos := strings.LastIndex(html, "<footer class=\"footer footer-center")
	closingBodyPos := strings.LastIndex(html, "</body>")
	closingDivPos := strings.LastIndex(html, "</div>")

	var resultHTML string
	if footerPos >= 0 && closingBodyPos > footerPos {
		// Replace existing footer with timing info
		resultHTML = html[:footerPos] + timingFooter + html[closingBodyPos:]
	} else if closingBodyPos >= 0 {
		// No footer found, but body closing tag exists - insert timing footer
		resultHTML = html[:closingBodyPos] + timingFooter + html[closingBodyPos:]
	} else if closingDivPos >= 0 && strings.Contains(html, "site-wrapper") {
		// Look for site-wrapper div closing
		wrapperDivEnd := strings.LastIndex(html[:closingDivPos], "site-wrapper")
		if wrapperDivEnd >= 0 {
			// Find the actual closing div for site-wrapper
			for i := wrapperDivEnd; i < len(html); i++ {
				if i+6 < len(html) && html[i:i+6] == "</div>" {
					resultHTML = html[:i] + timingFooter + html[i:]
					break
				}
			}
		}

		// If we couldn't find a proper place, just use the last closing div
		if resultHTML == "" {
			resultHTML = html[:closingDivPos] + timingFooter + html[closingDivPos:]
		}
	} else {
		// Fallback: just append the footer
		resultHTML = html + timingFooter
	}

	// Write the status code and headers if not already written
	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.wroteHeader = true
	}

	// Write the modified HTML
	w.ResponseWriter.Write([]byte(resultHTML))

	// Clear the buffer
	w.buffer.Reset()
}

// WriteHeaderNow implements the gin.ResponseWriter interface
func (w *TimingResponseWriter) WriteHeaderNow() {
	if !w.wroteHeader {
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.wroteHeader = true
	}
}

// Status implements the gin.ResponseWriter interface
func (w *TimingResponseWriter) Status() int {
	return w.statusCode
}

// Size implements the gin.ResponseWriter interface
func (w *TimingResponseWriter) Size() int {
	return w.buffer.Len()
}

// Written implements the gin.ResponseWriter interface
func (w *TimingResponseWriter) Written() bool {
	return w.wroteHeader
}

// Hijack implements the http.Hijacker interface
func (w *TimingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Push implements the http.Pusher interface
func (w *TimingResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
