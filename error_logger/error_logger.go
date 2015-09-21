package error_logger

import "log"

type Options map[string]interface{}

type ErrorLogger interface {
	Capture(error, *Options) error
}

func Setup(logger ErrorLogger) {
	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			logger.Capture(err, nil)
			log.Fatalf(err.Error())
		}
	}
}
