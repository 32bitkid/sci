package resources

import (
	"bufio"
	"bytes"
	"io"
)

func ReadText(resource *Resource) ([]string, error) {
	var text []string

	reader := bufio.NewReader(bytes.NewReader(resource.bytes))
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
