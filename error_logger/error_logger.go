package error_logger

type Options map[string]interface{}

type ErrorLogger interface {
	Capture(error, *Options) error
}
