package logger

// Debug 输出调试信息 在release下不会输出
func Debug(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

// Info 输出常规日志信息
func Info(format string, v ...interface{}) {
	logger.Info(format, v...)
}

// Warn 输出警告
// 含堆栈信息
func Warn(format string, v ...interface{}) {
	logger.Warn(format, v...)
}

// Error 输出错误信息
// 含堆栈信息
func Error(format string, v ...interface{}) {
	logger.Error(format, v...)
}

// Panic 输出致命信息 并抛出中断
func Panic(format string, v ...interface{}) {
	logger.Panic(format, v...)
}

// Logger 输出日志的接口
// 上层应用可实现该接口，通过BindLogger函数来接管工具库日志输出
type Logger interface {
	// Debug 输出调试信息 在release下不会输出
	Debug(format string, v ...interface{})
	// Info 输出常规日志信息
	Info(format string, v ...interface{})
	// Warn 输出警告
	Warn(format string, v ...interface{})
	// Error 输出错误信息
	Error(format string, v ...interface{})
	// Panic 输出致命信息 并抛出中断
	Panic(format string, v ...interface{})
}

//////////////////////////////////////////////////////////////////

var (
	logger Logger
)

