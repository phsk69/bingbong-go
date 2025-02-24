package timing

import (
	"sync"
	"time"
)

// RenderTiming holds timing information for page and template rendering
type RenderTiming struct {
	PageStart        time.Time
	PageDuration     time.Duration
	TemplateStart    time.Time
	TemplateDuration time.Duration
	mutex            sync.Mutex
	templateMeasured bool
	pageMeasured     bool
}

// NewRenderTiming creates a new timing structure with page start time
func NewRenderTiming() *RenderTiming {
	return &RenderTiming{
		PageStart: time.Now(),
	}
}

// StartTemplate marks the start of template rendering
func (rt *RenderTiming) StartTemplate() {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.templateMeasured = false
	rt.TemplateStart = time.Now()
}

// EndTemplate marks the end of template rendering
func (rt *RenderTiming) EndTemplate() {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if !rt.templateMeasured && !rt.TemplateStart.IsZero() {
		rt.TemplateDuration = time.Since(rt.TemplateStart)
		rt.templateMeasured = true
	}
}

// EndPage marks the end of page processing
func (rt *RenderTiming) EndPage() {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if !rt.pageMeasured {
		rt.PageDuration = time.Since(rt.PageStart)
		rt.pageMeasured = true
	}
}

// GetPageDuration returns the page render duration in milliseconds with high precision
func (rt *RenderTiming) GetPageDuration() float64 {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	var duration time.Duration
	if !rt.pageMeasured {
		duration = time.Since(rt.PageStart)
	} else {
		duration = rt.PageDuration
	}

	// Convert to milliseconds with floating point precision
	return float64(duration.Microseconds()) / 1000.0
}

// GetTemplateDuration returns the template render duration in milliseconds with high precision
func (rt *RenderTiming) GetTemplateDuration() float64 {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	if !rt.templateMeasured {
		if !rt.TemplateStart.IsZero() {
			// If not measured but started, calculate on the fly
			duration := time.Since(rt.TemplateStart)
			return float64(duration.Microseconds()) / 1000.0
		}
		return 0
	}

	// Convert to milliseconds with floating point precision
	return float64(rt.TemplateDuration.Microseconds()) / 1000.0
}
