package codec

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/riferrei/srclient"
)

// JSON clean codec, without schema registry support. Using without builder.
type JSON[T any] struct{}

func (JSON[T]) Encode(value any) ([]byte, error) {
	val, ok := value.(*T)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", new(T), value)
	}

	data, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("marshal %T: %w", val, err)
	}

	return data, nil
}

func (JSON[T]) Decode(data []byte) (any, error) {
	var val T

	err := json.Unmarshal(data, &val)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %T: %w", val, err)
	}

	return &val, nil
}

// JSONSchema codec, with schema registry support.
type JSONSchema[T any] struct {
	schemaID int
	client   *srclient.SchemaRegistryClient
}

func NewJSONSchema[T any](client *srclient.SchemaRegistryClient, subject, registrySchema string) (*JSONSchema[T], error) {
	schema, err := client.CreateSchema(subject, registrySchema, srclient.Json)
	if err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &JSONSchema[T]{
		schemaID: schema.ID(),
		client:   client,
	}, nil
}

func (s *JSONSchema[T]) Encode(value any) ([]byte, error) {
	val, ok := value.(*T)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", new(T), value)
	}

	data, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("marshal %T: %w", val, err)
	}

	buf := []byte{0}
	sid := make([]byte, 4)
	binary.BigEndian.PutUint32(sid, uint32(s.schemaID))
	buf = append(buf, sid...)
	buf = append(buf, data...)

	return buf, nil
}

func (s *JSONSchema[T]) Decode(data []byte) (any, error) {
	if len(data) < 5 || data[0] != 0 {
		return nil, fmt.Errorf("invalid message format")
	}

	schemaID := binary.BigEndian.Uint32(data[1:5])
	if s.schemaID != int(schemaID) {
		return nil, fmt.Errorf("invalid schema id: %d", schemaID)
	}

	var val T

	err := json.Unmarshal(data[5:], &val)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %T: %w", val, err)
	}

	return &val, nil
}
