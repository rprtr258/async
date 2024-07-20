package json

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJSON(t *testing.T) {
	tests := map[string]struct {
		input []byte
		want  map[string]any
	}{
		"valid JSON object with string values": {
			input: []byte(`{"name":"John","city":"New York"}`),
			want: map[string]any{
				"name": "John",
				"city": "New York",
			},
		},
		"valid JSON object with integer values": {
			input: []byte(`{"age":30,"year":2022}`),
			want: map[string]any{
				"age":  30.0,
				"year": 2022.0,
			},
		},
		"valid JSON object with mixed values": {
			input: []byte(`{"name":"John","age":30,"city":"New York"}`),
			want: map[string]any{
				"name": "John",
				"age":  30.0,
				"city": "New York",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseJSON(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
