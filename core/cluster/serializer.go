package cluster

import "encoding/json"

type Serializer interface {
	Serialize(v any) ([]byte, error)
	Deserialize(data []byte, v any) error
}

type JSONSerializer struct {
}

func NewJSONSerializer() Serializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
