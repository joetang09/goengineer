package engineer

import (
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	defaultCat = "default"
)

var (
	loggerHolder = map[string]*logrus.Logger{
		defaultCat: logrus.StandardLogger(),
	}

	defaultLogger = GetLogger(defaultCat)

	logPkg = reflect.TypeOf(configIns).PkgPath() + ".LogCpnt"
)

type LogConfig map[string]struct {
	Out    string
	Level  string
	Format string
	Hooks  []string
}

type LogCpnt struct {
}

func (LogCpnt) Init(options ...interface{}) error {

	if len(options) == 0 {
		return nil
	}

	c, ok := options[0].(*LogConfig)
	if !ok {
		return nil
	}

	for cat, config := range *c {

		var logger *logrus.Logger

		if cat == defaultCat {
			logger = loggerHolder[defaultCat]
		} else {
			logger = logrus.New()
		}
		l, _ := logrus.ParseLevel(config.Level)
		logger.SetLevel(l)

		if config.Format == "json" {
			logger.Formatter = &logrus.JSONFormatter{}
		} else {
			logger.Formatter = &logrus.TextFormatter{}
		}

		outDiv := strings.Split(config.Out, ":")
		if len(outDiv) == 0 {
			outDiv = append(outDiv, "std")
		}
		switch outDiv[0] {

		case "file":
			path := "./logs/default.log"
			if len(outDiv) > 1 && outDiv[1] != "" {
				path = outDiv[1]

				if strings.Index(path, "/") != 0 && strings.Index(path, "logs") != 0 && strings.Index(path, "./logs") != 0 {
					path = "./logs/" + path
				}
			}

			arf, err := newARFile(path, RotateTypeDay)

			// f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			// logger.Out = f
			logger.Out = arf
		case "std":
			fallthrough
		default:
			logger.Out = os.Stdout
		}

		loggerHolder[cat] = logger
	}
	return nil

}

func (LogCpnt) CfgKey() string {
	return "log"
}

func (LogCpnt) CfgType() interface{} {
	return LogConfig{}
}

func (LogCpnt) CfgUpdate(interface{}) {

}

func getLogger(cat string) *logrus.Logger {
	l, ok := loggerHolder[cat]
	if !ok {
		l = loggerHolder[defaultCat]
	}
	return l
}

type LogWrapper struct {
	cat string
}

func GetLogger(cat string) *LogWrapper {
	return &LogWrapper{
		cat: cat,
	}
}

func (l *LogWrapper) GetOut() io.Writer {
	return getLogger(l.cat).Out
}

func (l *LogWrapper) Debug(args ...interface{}) {
	getLogger(l.cat).Debug(args...)
}

func (l *LogWrapper) Print(args ...interface{}) {
	getLogger(l.cat).Print(args...)
}

func (l *LogWrapper) Info(args ...interface{}) {
	getLogger(l.cat).Info(args...)
}

func (l *LogWrapper) Warn(args ...interface{}) {
	getLogger(l.cat).Warn(args...)
}

func (l *LogWrapper) Warning(args ...interface{}) {
	getLogger(l.cat).Warning(args...)
}

func (l *LogWrapper) Error(args ...interface{}) {
	getLogger(l.cat).Error(args...)
}

func (l *LogWrapper) Panic(args ...interface{}) {
	getLogger(l.cat).Panic(args...)
}

func (l *LogWrapper) Fatal(args ...interface{}) {
	getLogger(l.cat).Fatal(args...)
}

func (l *LogWrapper) Debugf(format string, args ...interface{}) {
	getLogger(l.cat).Debugf(format, args...)
}

func (l *LogWrapper) Printf(format string, args ...interface{}) {
	getLogger(l.cat).Printf(format, args...)
}

func (l *LogWrapper) Infof(format string, args ...interface{}) {
	getLogger(l.cat).Infof(format, args...)
}

func (l *LogWrapper) Warnf(format string, args ...interface{}) {
	getLogger(l.cat).Warnf(format, args...)
}

func (l *LogWrapper) Warningf(format string, args ...interface{}) {
	getLogger(l.cat).Warningf(format, args...)
}

func (l *LogWrapper) Errorf(format string, args ...interface{}) {
	getLogger(l.cat).Errorf(format, args...)
}

func (l *LogWrapper) Panicf(format string, args ...interface{}) {
	getLogger(l.cat).Panicf(format, args...)
}

func (l *LogWrapper) Fatalf(format string, args ...interface{}) {
	getLogger(l.cat).Fatalf(format, args...)
}

func (l *LogWrapper) Debugln(args ...interface{}) {
	getLogger(l.cat).Debugln(args...)
}

func (l *LogWrapper) Println(args ...interface{}) {
	getLogger(l.cat).Println(args...)
}

func (l *LogWrapper) Infoln(args ...interface{}) {
	getLogger(l.cat).Infoln(args...)
}

func (l *LogWrapper) Warnln(args ...interface{}) {
	getLogger(l.cat).Warnln(args...)
}

func (l *LogWrapper) Warningln(args ...interface{}) {
	getLogger(l.cat).Warningln(args...)
}

func (l *LogWrapper) Errorln(args ...interface{}) {
	getLogger(l.cat).Errorln(args...)
}

func (l *LogWrapper) Panicln(args ...interface{}) {
	getLogger(l.cat).Panicln(args...)
}

func (l *LogWrapper) Fatalln(args ...interface{}) {
	getLogger(l.cat).Fatalln(args...)
}

func (l *LogWrapper) STDLogger() *log.Logger {

	return log.New(getLogger(l.cat).Out, "["+l.cat+"] ", log.Llongfile|log.Ldate|log.Ltime)

}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Print(args ...interface{}) {
	defaultLogger.Print(args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	defaultLogger.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

func Println(args ...interface{}) {
	defaultLogger.Println(args...)
}

func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func Warningln(args ...interface{}) {
	defaultLogger.Warningln(args...)
}

func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

func Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

func STDLogger() *log.Logger {
	return defaultLogger.STDLogger()

}
