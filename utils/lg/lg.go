package lg

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/fatih/color"
)

var (
	infoLog, debugEnabledLog, debugDisabledLog, warnLog, errLog, fatalLog *log.Logger

	debug   = false
	logTerm = NewTerminal(8 * 1024 * 1024)
)

func init() {
	// Always set nocolor=false, but strip color to stdout when we should not.
	stdoutNoColor := color.NoColor
	color.NoColor = false

	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr
	if stdoutNoColor {
		// TODO(yuheng): Verify performance hit, since every log will be run against regexp.
		stdout = NewStripColorWriter(stdout)
		stderr = NewStripColorWriter(stderr)
	}

	stdout = io.MultiWriter(stdout, logTerm)
	stderr = io.MultiWriter(stderr, logTerm)

	infoLog = log.New(stdout, color.GreenString("[I]"), log.LstdFlags|log.LUTC|log.Lmicroseconds)
	debugEnabledLog = log.New(stdout, color.CyanString("[D]"), log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmicroseconds)
	debugDisabledLog = log.New(logTerm, color.CyanString("[D]"), log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmicroseconds)
	errLog = log.New(stderr, color.RedString("[E]"), log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmicroseconds)
	warnLog = log.New(stdout, color.YellowString("[W]"), log.LstdFlags|log.LUTC|log.Lmicroseconds)
	fatalLog = log.New(stderr, color.RedString("[F]"), log.LstdFlags|log.Llongfile|log.LUTC|log.Lmicroseconds)
	debug = os.Getenv("debug") != ""
}

// EnableDebug ...
func EnableDebug() {
	debug = true
}

func ForkLog(tailBytes int) io.ReadCloser {
	return logTerm.ForkTTY(tailBytes)
}

func doLog(logger *log.Logger, msg string) {
	for _, line := range strings.Split(msg, "\n") {
		logger.Output(3, line)
	}
}

// PanicError provide a quick way to check unexpected errors that should never happen.
// It's basically an assertion that once err != nil, fatal panic is thrown.
func PanicError(err error, msg ...interface{}) {
	if err != nil {
		var s string
		if len(msg) > 0 {
			s = fmt.Sprintf("%+v\n", err) + ":" + fmt.Sprint(msg...)
		} else {
			s = fmt.Sprintf("%+v", err)
		}
		doLog(errLog, s)
		panic(err)
	}
}

// DPanicError provide a quick way to check unexpected errors that should never happen.
// It's almost the same as Check(), except only in debug mode will throw panic.
func DPanicError(err error) {
	if err != nil {
		doLog(errLog, err.Error())
		if debug {
			panic(err)
		}
	}
}

// Error prints error to error output with [ERROR] prefix.
func Error(v ...interface{}) {
	if v[0] != nil {
		doLog(errLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

// Warn prints warn to warn output with [WARN] prefix.
func Warn(v ...interface{}) {
	if v[0] != nil {
		doLog(warnLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

// Info prints info to standard output with [INFO] prefix.
func Info(v ...interface{}) {
	if v[0] != nil {
		doLog(infoLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

// Debug prints info to standard output with [DEBUG] prefix in debug mode.
func Debug(v ...interface{}) {
	if v[0] != nil {
		if IsDebugging() {
			doLog(debugEnabledLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
		} else {
			doLog(debugDisabledLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
		}
	}
}

// TimeFunc prints the info with timing consumed by function.
// It has specified usage like:
//     defer TimeFunc("Hello world")()
func TimeFunc(v ...interface{}) func() {
	start := time.Now()
	return func() {
		Info(append(v, "|", time.Since(start))...)
	}
}

func TimeFuncDebug(v ...interface{}) func() {
	start := time.Now()
	return func() {
		Debug(append(v, "|", time.Since(start))...)
	}
}

// Fatal prints error to error output with [FATAL] prefix, and terminate the
// application.
func Fatal(v ...interface{}) {
	var msgs []string
	for _, i := range v {
		msgs = append(msgs, fmt.Sprintf("%+v", i))
	}
	doLog(fatalLog, strings.Join(msgs, " "))
	os.Exit(1)
}

// Infof except accepting formating info.
func Infof(msg string, v ...interface{}) {
	Infoc(context.Background(), msg, v...)
}

// Warnf except accepting formating info.
func Warnf(msg string, v ...interface{}) {
	Warnc(context.Background(), msg, v...)
}

// Errorf except accepting formating info.
func Errorf(msg string, v ...interface{}) {
	Errorc(context.Background(), msg, v...)
}

// Debugf except accepting formating info.
func Debugf(msg string, v ...interface{}) {
	Debugc(context.Background(), msg, v...)
}

// Fatalf except accepting formating info.
func Fatalf(msg string, v ...interface{}) {
	Fatalc(context.Background(), msg, v...)
}

// PrintJSON outputs any varible in JSON format to console. Useful for debuging.
func PrintJSON(v interface{}) {
	fmt.Println(Jsonify(v))
}

// Jsonify provides shortcut to return an json format string of any varible.
func Jsonify(v interface{}) string {
	d, err := json.MarshalIndent(v, "", "  ")
	DPanicError(err)
	return string(d)
}

func JsonifyNested(v interface{}) string {
	d, err := json.Marshal(v)
	DPanicError(err)
	return string(d)
}

// GetFuncName provides shortcut to print the name of any function.
func GetFuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// TimeFuncDuration returns the duration consumed by function.
// It has specified usage like:
//     f := TimeFuncDuration()
//	   DoSomething()
//	   duration := f()
func TimeFuncDuration() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}

// IsDebugging returns whether it's in debug mode.
func IsDebugging() bool {
	return debug
}
