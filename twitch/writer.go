package twitch

import (
	"os"

	"github.com/Kostaaa1/twitchdl/types"
)

type progressWriter struct {
	writer     *os.File
	progressCh chan<- types.ProgresbarChanData
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err == nil {
		select {
		case pw.progressCh <- types.ProgresbarChanData{
			Text:  pw.writer.Name(),
			Bytes: n,
		}:
		default:
		}
	}
	return n, err
}

func (pw *progressWriter) SetWriter(writer *os.File) {
	pw.writer = writer
}

func (pw *progressWriter) Close() error {
	return pw.writer.Close()
}

func NewProgressWriter(dstPath string, progressCh chan<- types.ProgresbarChanData) (*progressWriter, error) {
	// create file outside, use SetWriter before writing
	f, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	pw := &progressWriter{
		writer:     f,
		progressCh: progressCh,
	}
	return pw, nil
}
