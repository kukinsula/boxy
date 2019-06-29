package entity

import (
	"encoding/json"
)

type JSONCodec struct{}

func (codec *JSONCodec) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (codec *JSONCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
