package opbeat

import (
	opbeatClient "github.com/roncohen/opbeat-go"
	"github.com/upfluence/goutils/error_logger"
)

type Logger struct {
	client *opbeatClient.Opbeat
}

func NewErrorLogger() *Logger {
	return &Logger{opbeatClient.NewFromEnvironment()}
}

func (l *Logger) Capture(err error, opts *error_logger.Options) error {
	var options *opbeatClient.Options

	if opts != nil {
		extra := make(opbeatClient.Extra)
		for k, v := range *opts {
			extra[k] = v
		}

		options = &opbeatClient.Options{Extra: &extra}
	}

	return l.client.CaptureError(err, options)
}
