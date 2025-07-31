package lg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/goccy/go-json"

	"github.com/go-logfmt/logfmt"

	goutils "github.com/hoveychen/go-utils"
)

type key int

const logContextKey key = iota

// LogContext indicates all the contextual logs in a request.
type LogContext struct {
	msg    []string  // Descriptions
	keys   []string  // Keys for contextual KV
	values []string  // Values for contextual KV
	uuid   int64     // uuid for identifying the same request
	from   string    // direct service caller name
	since  time.Time // request start time
}

// EncodingType represent the type to use for log encoding.
type EncodingType string

const (
	// TypeJSON encodes log into one-line json.
	TypeJSON EncodingType = "json"

	// TypeLogfmt encodes log into logfmt(structlog) format.
	TypeLogfmt EncodingType = "logfmt"
)

var encoding = TypeLogfmt

// SetEncoding changes the log encoding type.
func SetEncoding(t EncodingType) {
	encoding = t
}

// JSON returns the log in JSON format.
func (lc *LogContext) JSON() string {
	buf := &bytes.Buffer{}

	msg := strings.Join(lc.msg, " ")
	if len(msg) > 0 {
		buf.WriteString(msg)
		buf.WriteRune(' ')
	}

	out := map[string]string{}
	if len(lc.keys) != len(lc.values) {
		Error("Invalid numbers of keys vs. numbers of values")
		return msg
	}
	for i := 0; i < len(lc.keys); i++ {
		out[lc.keys[i]] = lc.values[i]
	}
	if lc.uuid > 0 {
		out["uuid"] = strconv.FormatInt(lc.uuid, 36)
	}
	if lc.from != "" {
		out["from"] = lc.from
	}
	if !lc.since.IsZero() {
		out["escaped"] = time.Since(lc.since).String()
	}
	encoder := json.NewEncoder(buf)
	encoder.Encode(out)
	return buf.String()
}

// Logfmt returns the log in LOGFMT format.
func (lc *LogContext) Logfmt() string {
	msg := strings.Join(lc.msg, " ")
	if len(lc.keys) != len(lc.values) {
		Error("Invalid numbers of keys vs. numbers of values")
		return msg
	}

	var buf bytes.Buffer

	encoder := logfmt.NewEncoder(&buf)
	if lc.uuid > 0 {
		if err := encoder.EncodeKeyval("uuid", strconv.FormatInt(lc.uuid, 36)); err != nil {
			Error("Encoding logfmt", err)
		}
	}
	if lc.from != "" {
		if err := encoder.EncodeKeyval("from", lc.from); err != nil {
			Error("Encoding logfmt", err)
		}
	}
	if !lc.since.IsZero() {
		if err := encoder.EncodeKeyval("escaped", time.Since(lc.since).String()); err != nil {
			Error("Encoding logfmt", err)
		}
	}
	for i := 0; i < len(lc.keys); i++ {
		if err := encoder.EncodeKeyval(lc.keys[i], lc.values[i]); err != nil {
			Error("Encoding logfmt", err)
		}
	}
	str := buf.String()
	if str == "" {
		return msg
	}

	return msg + " " + color.MagentaString(str)
}

// String returns a string representing the log.
func (lc *LogContext) String() string {
	switch encoding {
	case TypeJSON:
		return lc.JSON()
	case TypeLogfmt:
		return lc.Logfmt()
	}
	return ""
}

func (lc *LogContext) Map() map[string]string {
	lgContextMap := map[string]string{}
	for i := 0; i < len(lc.keys); i++ {
		lgContextMap[lc.keys[i]] = lc.values[i]
	}
	return lgContextMap
}

func (lc *LogContext) Message() string {
	return strings.Join(lc.msg, ",")
}

// Empty returns true if the log context contains nothing.
func (lc *LogContext) Empty() bool {
	return len(lc.msg)+len(lc.keys)+len(lc.values) == 0 && lc.uuid == 0
}

func (lc *LogContext) UUID() int64 {
	if lc == nil {
		return 0
	}
	return lc.uuid
}

func (lc *LogContext) From() string {
	if lc == nil {
		return ""
	}
	return lc.from
}

func (lc *LogContext) Since() time.Time {
	if lc == nil {
		return time.Time{}
	}
	return lc.since
}

func parseFmtStr(format string) (msg string, isKV []bool, keys, descs []string) {
	// Format like "% d" will not be supported.
	var msgs []string
	for _, s := range strings.Split(format, " ") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		idx := strings.Index(s, "=%")
		if idx == -1 {
			re, _ := goutils.CompileRegexp("%[^%]+")
			matches := re.FindAllStringIndex(s, -1)
			for i := 0; i < len(matches); i++ {
				isKV = append(isKV, false)
			}
			msgs = append(msgs, s)
			continue
		}
		keys = append(keys, s[:idx])
		descs = append(descs, s[idx+1:])
		isKV = append(isKV, true)
	}
	msg = strings.Join(msgs, " ")
	return
}

// FromContext returns a the potential LogContext object from Context if any.
func FromContext(ctx context.Context) *LogContext {
	if ctx == nil {
		return nil
	}
	lcVal := ctx.Value(logContextKey)
	lc, ok := lcVal.(*LogContext)
	if !ok {
		return nil
	}
	return lc
}

type logable interface {
	Output(calldepth int, s string) error
}

func logc(ctx context.Context, l logable) {
	lc := FromContext(ctx)
	if lc != nil {
		_, ok := l.(*log.Logger)
		// Only output LogFmt for standard logger.
		// Otherwise, output JSON format log.
		if ok {
			msg := lc.Logfmt()
			for _, line := range strings.Split(msg, "\n") {
				l.Output(3, line)
			}
		} else {
			l.Output(3, lc.JSON())
		}
	}
}

// Infoc logs message in logfmt.
func Infoc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, infoLog)
}

// Warnc logs warn message in logfmt.
func Warnc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, warnLog)
}

// Errorc logs error message in logfmt.
func Errorc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, errLog)
}

// Debugc logs debug message in logfmt.
func Debugc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	if IsDebugging() {
		logc(ctx, debugEnabledLog)
	} else {
		logc(ctx, debugDisabledLog)
	}
}

// Fatalc logs debug message in logfmt. see `Fatal()` for more detail.
func Fatalc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, fatalLog)
	os.Exit(1)
}

// Derive extracts log context from src context, to merge it into dest context.
func Derive(src, dest context.Context) context.Context {
	lc := FromContext(src)
	if lc == nil {
		return dest
	}

	return context.WithValue(dest, logContextKey, lc)
}

// With wraps a new context, ands inject log context messge into it.
func With(ctx context.Context, msg string, v ...interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(msg) == 0 && len(v) == 0 {
		return ctx
	}
	lc := FromContext(ctx)
	if lc == nil {
		lc = &LogContext{}
	}
	newLc := *lc
	msgTmpl, isKV, keys, desc := parseFmtStr(msg)
	var msgV []interface{}
	var objV []interface{}
	for i, kv := range isKV {
		var val interface{}
		if i >= len(v) {
			val = "<Missing>"
		} else {
			val = v[i]
		}
		if kv {
			objV = append(objV, val)
		} else {
			msgV = append(msgV, val)
		}
	}
	msg = fmt.Sprintf(msgTmpl, msgV...)
	if msg != "" {
		newLc.msg = append(lc.msg, msg)
	}
	if len(objV) != len(desc) {
		Error("Number of values not equals to number of descriptors")
		return ctx
	}

	newLc.keys = append(newLc.keys, keys...)
	for i := range desc {
		newLc.values = append(newLc.values, fmt.Sprintf(desc[i], objV[i]))
	}

	return context.WithValue(ctx, logContextKey, &newLc)
}
