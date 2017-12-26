package log

import (
	"fmt"
	"runtime"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var numBytesToShortenCaller int

func init() {
	// Example filename: /Users/dan/dev/gopath/src/github.com/danielstutzman/sync-log-files-to-db/src/log/encoder.go
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Can't get runtime.Caller(0)")
	}
	numBytesToShortenCaller = len(filename) - len("log/encoder.go")
}

type CustomEncoder struct {
	*zapcore.MapObjectEncoder
	startingTime time.Time
}

func (enc CustomEncoder) Clone() zapcore.Encoder {
	panic("Clone called")
}

func (enc CustomEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	newEncoder := CustomEncoder{
		MapObjectEncoder: zapcore.NewMapObjectEncoder(),
	}
	buf := pool.Get()

	secondsSinceStart := entry.Time.Sub(enc.startingTime).Seconds()
	shortenedCaller := entry.Caller.String()[numBytesToShortenCaller:]
	buf.AppendString(fmt.Sprintf("%5.1f %-5s %s ",
		secondsSinceStart, entry.Level, shortenedCaller))

	buf.AppendString(entry.Message)

	for _, field := range fields {
		// Fix zero times being incorrectly shown as
		// 1754-08-30 22:43:41.128654848 +0000 UTC
		if field.Type == zapcore.TimeType && field.Integer == -6795364578871345152 {
			field.Integer = 0
		}

		field.AddTo(newEncoder)
	}
	for _, field := range fields {
		value := newEncoder.MapObjectEncoder.Fields[field.Key]
		if value == "" {
			buf.AppendString(" ''")
		} else {
			buf.AppendString(fmt.Sprintf(" %v", value))
		}
	}

	buf.AppendByte('\n')

	if entry.Stack != "" {
		buf.AppendString(entry.Stack)
		buf.AppendByte('\n')
	}

	return buf, nil
}

func constructEncoder(config zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return CustomEncoder{
		MapObjectEncoder: zapcore.NewMapObjectEncoder(),
		startingTime:     time.Now(),
	}, nil
}
