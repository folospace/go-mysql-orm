package orm

type InfoLogger interface {
    Info(args ...any)
}

type ErrorLogger interface {
    Error(args ...any)
}

var infoLogger InfoLogger
var errorLogger ErrorLogger

func SetInfoLogger(l InfoLogger) {
    infoLogger = l
    infoLogger.Info("set info logger")
}

func SetErrorLogger(l ErrorLogger) {
    errorLogger = l
    errorLogger.Error("set error logger")
}
