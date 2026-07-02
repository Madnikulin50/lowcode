package xml_to_map

import (
	"encoding/xml"
	"io"
	"strings"
)

func ParseXMLToMap(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	decoder := xml.NewDecoder(r)

	var currentKey string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			// Capture the tag name to use as our map key
			currentKey = element.Name.Local
		case xml.CharData:
			// Clean up whitespace and assign content to the current key
			value := strings.TrimSpace(string(element))
			if value != "" && currentKey != "" {
				result[currentKey] = value
			}
		case xml.EndElement:
			// Reset current key upon closing tag
			currentKey = ""
		}
	}

	return result, nil
}
