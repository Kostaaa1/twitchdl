package twitch

import (
	"io"
	"os"

	"github.com/Kostaaa1/twitchdl/types"
)

type progressWriter struct {
	writer     io.Writer
	slug       string
	progressCh chan<- types.ProgresbarChanData
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err == nil {
		select {
		case pw.progressCh <- types.ProgresbarChanData{
			Text:  pw.slug,
			Bytes: n,
		}:
		default:
		}
	}
	return n, err
}

func (pw *progressWriter) SetWriter(writer io.Writer) {
	pw.writer = writer
}

func (c *Client) NewProgressWriter(slug, dstPath string) (*progressWriter, error) {
	f, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}

	pw := &progressWriter{
		writer:     f,
		slug:       slug,
		progressCh: c.progressCh,
	}
	return pw, nil
}
