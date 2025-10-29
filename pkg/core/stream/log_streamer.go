package stream

import "context"

type LogStreamer struct {
	dir string
	api HandleLogStream
}

func NewLogStreamer(d string, a HandleLogStream) *LogStreamer {
	return &LogStreamer{
		dir: d,
		api: a,
	}
}

type HandleLogStream interface {
	Stream(ctx context.Context, id string, logChan chan<- string) error
}
