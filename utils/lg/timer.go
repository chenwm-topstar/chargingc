package lg

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Timer is a helper to time multiple tick.
// Simple usage like:
// ```
// t := lg.NewTimer("Search similar", itemID)
// defer lg.Info(t)
// meta := getMeta(itemID)
// t.Tick("Get meta")
// results := querySimilar(itemID)
// t.Tick("Query")
// return formatResult(results)
// ```
type Timer struct {
	start       time.Time
	last        time.Time
	timerMsg    string
	durations   map[string]time.Duration
	spansOrder  []string
	parentTimer *Timer
	lock        sync.Mutex
}

// NewTimer returns a Timer object.
func NewTimer(msg ...interface{}) *Timer {
	return &Timer{
		start:     time.Now(),
		last:      time.Now(),
		timerMsg:  fmt.Sprint(msg...),
		durations: make(map[string]time.Duration),
	}
}

// Tick records the time ecaplsed since start of timer.
func (t *Timer) Tick(msg ...interface{}) {
	strMsg := t.timerMsg + " | " + fmt.Sprint(msg...)
	t.addSpan(strMsg, time.Since(t.last))
	t.lock.Lock()
	t.last = time.Now()
	t.lock.Unlock()
}

func (t *Timer) addSpan(msg string, d time.Duration) {
	if t.parentTimer != nil {
		t.parentTimer.addSpan(msg, d)
	}
	t.lock.Lock()
	old, exists := t.durations[msg]
	if !exists {
		t.durations[msg] = d
		t.spansOrder = append(t.spansOrder, msg)
	} else {
		t.durations[msg] = old + d
	}
	t.lock.Unlock()
}

func (t *Timer) String() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	lastTick := time.Since(t.last)
	total := time.Since(t.start)
	var ret []string
	for _, msg := range t.spansOrder {
		ret = append(ret, fmt.Sprintf("%s: %s", msg, t.durations[msg]))
	}

	if lastTick > time.Millisecond {
		ret = append(ret, fmt.Sprintf("Rest: %s", lastTick))
	}
	ret = append(ret, fmt.Sprintf("Total: %s", total))
	return strings.Join(ret, "\n")
}

func (t *Timer) Duration() time.Duration {
	return time.Since(t.start)
}

// SubTimer starts a new timer whose data returns to its
// parent timer.
func (t *Timer) SubTimer(msg ...interface{}) *Timer {
	return &Timer{
		start:       time.Now(),
		last:        time.Now(),
		timerMsg:    t.timerMsg + " > " + fmt.Sprint(msg...),
		parentTimer: t,
		durations:   make(map[string]time.Duration),
	}
}

type Span struct {
	msg      string
	duration time.Duration
}
