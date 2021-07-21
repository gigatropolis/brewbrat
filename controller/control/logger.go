package control

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"time"
)

const (
	LOG_QUEUE_SIZE = 100
	LOG_SYNC_DELAY = 2
)

const LOG_LEVEL_NOT_SET = 0
const (
	LogLevelDebug = (1 << iota)
	LogLevelMessage
	LogLevelWarning
	LogLevelPass
	LogLevelFail
	LogLevelResults
	LogLevelError
	LogLevelAll
)

type logArg struct {
	level   uint64
	pattern string
	args    []interface{}
}

type logStream struct {
	ChnLogInput chan logArg
	level       uint64
	logger      *log.Logger
}

func (log *logStream) Start() {

	log.ChnLogInput = make(chan logArg, LOG_QUEUE_SIZE)

	go func() {
		for message := range log.ChnLogInput {
			log.logger.Printf(message.pattern, message.args...)
		}
	}()

}

func (log *logStream) sync() {
	for len(log.ChnLogInput) > 0 {
		time.Sleep(time.Millisecond * LOG_SYNC_DELAY)
	}
}

type Logger struct {
	chnInput    chan logArg
	loggers     map[string]logStream
	debugMode   bool
	initialized bool
	faulted     bool
	END         string
}

// Init method will automatically be called before logger is used but user can call if desired.
func (clog *Logger) Init() {
	if clog.initialized {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			clog.faulted = true
			panic(r)
		}
	}()
	clog.loggers = make(map[string]logStream)
	clog.chnInput = make(chan logArg, LOG_QUEUE_SIZE)
	clog.debugMode = false
	if runtime.GOOS == "linux" {
		clog.END = "\n"
	} else {
		clog.END = "\r\n"
	}
	clog.initialized = true

	go func() {
		for message := range clog.chnInput {
			for _, logger := range clog.loggers {

				if ((message.level & LogLevelDebug) != 0) &&
					((logger.level&(LogLevelMessage|LogLevelAll) != 0) && clog.debugMode) {
					logger.ChnLogInput <- message
					continue
				}

				if (logger.level & LogLevelAll) != 0 {
					logger.ChnLogInput <- message
					continue
				}

				if ((logger.level & LogLevelResults) != 0) &&
					(message.level&(LogLevelPass|LogLevelFail) != 0) {
					logger.ChnLogInput <- message
					continue
				}

				logLevel := uint64(message.level & logger.level)
				if logLevel != 0 {
					logger.ChnLogInput <- message
					continue
				}
			}
		}
	}()
}

func (clog *Logger) ready() bool {
	if clog.initialized {
		return true
	}
	if clog.faulted {
		return false
	}
	clog.Init()
	return clog.initialized
}

func (clog *Logger) Add(name string, level uint64, stream io.Writer) {
	if !clog.ready() {
		return
	}
	if _, ok := clog.loggers[name]; !ok {
		stream := logStream{level: level, logger: log.New(stream, "", log.Ldate|log.Ltime|log.Lmicroseconds)}
		stream.Start()
		clog.loggers[name] = stream
	}
}

func (clog *Logger) Printf(level uint64, value string, args ...interface{}) {
	arg := logArg{level, value, args}
	clog.chnInput <- arg
}

// Sync will block until all messages have been sent to all log streams
// and all log streams have cleared there channels.
func (clog *Logger) Sync() {
	if !clog.ready() {
		return
	}
	for len(clog.chnInput) > 0 {
		time.Sleep(time.Millisecond * LOG_SYNC_DELAY)
	}
	for _, logger := range clog.loggers {
		logger.sync()
	}
}
func (clog *Logger) SetDebug(mode bool) {
	if clog.ready() {
		clog.debugMode = mode
	}
}

func (clog *Logger) IsDebugSet() bool {

	return clog.debugMode
}

func (clog *Logger) LogError(errMsg string, args ...interface{}) {
	if clog.ready() {
		clog.Printf(LogLevelError, fmt.Sprintf("ERROR::%s%s", errMsg, clog.END), args...)
	}
}

func (clog *Logger) LogDebug(DebugMsg string, args ...interface{}) {
	if clog.ready() {
		if clog.debugMode {
			clog.Printf(LogLevelDebug, fmt.Sprintf("DEBUG::%s%s", DebugMsg, clog.END), args...)
		}
	}
}

func (clog *Logger) LogWarning(warnMsg string, args ...interface{}) {
	if clog.ready() {
		clog.Printf(LogLevelWarning, fmt.Sprintf("ERROR::%s%s", warnMsg, clog.END), args...)
	}
}

func (clog *Logger) LogPass(passMsg string, args ...interface{}) {
	if clog.ready() {
		clog.Printf(LogLevelPass, fmt.Sprintf("PASS::%s%s", passMsg, clog.END), args...)
	}
}

func (clog *Logger) LogFail(failMsg string, args ...interface{}) {
	if clog.ready() {
		clog.Printf(LogLevelFail, fmt.Sprintf("FAIL::%s%s", failMsg, clog.END), args...)
	}
}

func (clog *Logger) LogMessage(msg string, args ...interface{}) {
	if clog.ready() {
		clog.Printf(LogLevelMessage, fmt.Sprintf("MSG::%s%s", msg, clog.END), args...)
	}
}
