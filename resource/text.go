package resource

import (
	"bufio"
	"bytes"
	"io"
)

type Text []string

func NewText(b []byte) (Text, error) {
	var text Text

	reader := bufio.NewReader(bytes.NewReader(b))
	for {
		if str, err := reader.ReadString(0x00); err == io.EOF {
			break
		} else if err == nil {
			text = append(text, str[:len(str)-1])
		} else {
			return nil, err
		}
	}

	return text, nil
}
