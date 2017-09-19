package exporter

type Exporter interface {
	Export(<-chan bool)
}
