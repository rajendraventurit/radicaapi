package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var localLog *Logger

const defConfPath = "/etc/radica/logger.json"

func init() {
	l := Logger{logLevel: LLAll}
	localLog = &l
}

// Logger is a logger
type Logger struct {
	path         string
	logLevel     LogLevel
	slackWebHook string
}

// LogLevel controls what items are written to the log
type LogLevel int

// LogLevels
const (
	LLError LogLevel = iota
	LLWarning
	LLInfo
	LLDebug
	LLAll
)

type config struct {
	Path         string `json:"path"`
	Level        string `json:"level"`
	SlackWebHook string `json:"slack_webhook"`
}

// Configure will configure the logger using the config defConfPath
// logger will attempt to open and write to the log file on each call
func Configure() error {
	f, err := os.Open(defConfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	conf := config{}
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		return err
	}

	l := Logger{
		path:         os.Getenv("PATH"),
		logLevel:     logLevelFromStr(os.Getenv("LEVEL")),
		slackWebHook: os.Getenv("SLACKWEBHOOK"),
	}
	localLog = &l
	return nil
}

func logLevelFromStr(l string) LogLevel {
	switch l {
	case "ALL":
		return LLAll
	case "WARNING":
		return LLWarning
	case "INFO":
		return LLInfo
	case "DEBUG":
		return LLDebug
	case "ERROR":
		return LLError
	}
	return LLAll
}

// SetLogLevel will set the internal loggers log level
func SetLogLevel(ll LogLevel) {
	localLog.SetLogLevel(ll)
}

// GenMsg will generate a message from a handler
func GenMsg(userid int64, method, url, msg string) string {
	if userid > 0 {
		return fmt.Sprintf("UserID %v %s %s %s", userid, method, url, msg)
	}
	return fmt.Sprintf("%s %s %s", method, url, msg)
}

// GenMsgErr will generate a message from a handler and error
func GenMsgErr(userid int64, method, url string, err error, context ...string) string {
	con := strings.Join(context, " ") + " "
	if userid > 0 {
		return fmt.Sprintf("UserID %v %s %s %s %v", userid, method, url, con, err)
	}
	return fmt.Sprintf("%s %s %s %v", method, url, con, err)
}

// GenHTTPMsg will generate a message from an http handler
func GenHTTPMsg(userid int64, r *http.Request, msg string) string {
	switch {
	case r == nil && userid == 0:
		return msg
	case r == nil:
		return fmt.Sprintf("UserID %v %s", userid, msg)
	default:
		return GenMsg(userid, r.Method, r.URL.String(), msg)
	}
}

// Error will write an error msg
func Error(msg string) {
	localLog.Error(msg)
}

// Errorf will write a formatted error to the default log
func Errorf(format string, ii ...interface{}) {
	Error(fmt.Sprintf(format, ii...))
}

// Warning will write a warning msg
func Warning(msg string) {
	localLog.Warning(msg)
}

// Warningf will write a formatted message
func Warningf(format string, ii ...interface{}) {
	Warning(fmt.Sprintf(format, ii...))
}

// Info will write an info msg
func Info(msg string) {
	localLog.Info(msg)
}

// Infof will write a formatted message
func Infof(format string, ii ...interface{}) {
	Info(fmt.Sprintf(format, ii...))
}

// Debug will write a debug msg
func Debug(msg string) {
	localLog.Debug(msg)
}

// Debugf will write a formatted debug message
func Debugf(format string, ii ...interface{}) {
	Debug(fmt.Sprintf(format, ii...))
}

// Message will write a message to the log regardless of log level
func Message(msg string) {
	localLog.Message(msg)
}

// Messagef will write a formatted message
func Messagef(format string, ii ...interface{}) {
	Message(fmt.Sprintf(format, ii...))
}

// HTTPAccess writes an ACCESS message to the default log
func HTTPAccess(userid int64, r *http.Request) {
	localLog.HTTPAccess(userid, r)
}

// Fatal will write a message to the default log and os.Exit(1)
func Fatal(err error) {
	localLog.Fatal(err)
}

// ErrorHTTP will format a message and write an error to the default log
func ErrorHTTP(userid int64, r *http.Request, msg string) {
	localLog.Error(GenHTTPMsg(userid, r, msg))
}

// ErrorErr will format a message and write an error to the default log
func ErrorErr(userid int64, r *http.Request, err error, context ...string) {
	con := strings.Join(context, " ") + " "
	switch {
	case r == nil && userid == 0:
		localLog.Error(con + err.Error())
	case r == nil:
		localLog.Error(fmt.Sprintf("UserID %v %v %v", userid, con, err))
	default:
		localLog.Error(GenMsgErr(userid, r.Method, r.URL.String(), err, context...))
	}
}

// SetLogLevel sets the log level
func (l *Logger) SetLogLevel(ll LogLevel) {
	l.logLevel = ll
}

// Error will write an error msg
func (l Logger) Error(msg string) {
	l.write("ERROR", msg)
	l.slack(msg)
}

// Errorf writes an error msg with formatting
func (l Logger) Errorf(format string, ii ...interface{}) {
	l.Error(fmt.Sprintf(format, ii...))
}

// Warning will write a warning msg
func (l Logger) Warning(msg string) {
	if l.logLevel < LLWarning {
		return
	}
	l.write("WARNING", msg)
}

// Info will write an info msg
func (l Logger) Info(msg string) {
	if l.logLevel < LLInfo {
		return
	}
	l.write("INFO", msg)
}

// Debug will write a debug msg
func (l Logger) Debug(msg string) {
	if l.logLevel < LLDebug {
		return
	}
	l.write("DEBUG", msg)
}

// Message will write to the log regardless of log level
func (l Logger) Message(msg string) {
	l.write("MESSAGE", msg)
}

// Fatal will write to the log followed by os.Exit(1)
func (l Logger) Fatal(err error) {
	l.write("FATAL", err.Error())
	os.Exit(1)
}

// HTTPAccess will write an ACCESS message to the log
func (l Logger) HTTPAccess(userid int64, r *http.Request) {
	l.write("ACCESS", GenHTTPMsg(userid, r, ""))
}

func (l Logger) write(prefix, msg string) {
	str := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), prefix, msg)

	if l.path != "" {
		f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			os.Stdout.Write([]byte(str))
			return
		}
		defer f.Close()
		_, _ = f.Write([]byte(str))
		return
	}

	os.Stdout.Write([]byte(str))
}

func (l Logger) slack(msg string) error {
	if l.slackWebHook == "" {
		return nil
	}
	p := struct {
		Text string `json:"text"`
	}{
		Text: msg,
	}
	js, err := json.Marshal(&p)
	if err != nil {
		return err
	}
	resp, err := http.Post(l.slackWebHook, "application/json", strings.NewReader(string(js)))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}
	return nil
}
