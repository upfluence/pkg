package error_logger

type Options map[string]interface{}

type ErrorLogger interface {
	Capture(error, *Options) error
	Close()
}

func Setup(logger ErrorLogger) {
	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			logger.Capture(err, nil)
			logger.Close()
			panic(err.Error())
		}
	}
}
