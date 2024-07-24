package json

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type token struct {
	pos  int
	b    string
	kind Kind
}

func parseJSON(s string) []token {
	res := []token{}
	Parse([]byte(s))(func(pos int, b []byte, kind Kind) {
		res = append(res, token{pos, string(b), kind})
	})
	return res
}

func TestParseJSON(t *testing.T) {
	for name, tt := range map[string]struct {
		input string
		want  []token
	}{
		"valid JSON object with string values": {
			//                11111111112222222222333
			//      012345678901234567890123456789012
			input: `{"name":"John","city":"New York"}`,
			want: []token{
				{0, `{`, KindObjectBegin},
				{1, `"name"`, KindObjectKey},
				{8, `"John"`, KindString},
				{15, `"city"`, KindObjectKey},
				{22, `"New York"`, KindString},
				{32, `}`, KindObjectEnd},
			},
		},
		"valid JSON object with integer values": {
			//                111111111122
			//      0123456789012345678901
			input: `{"age":30,"year":2022}`,
			want: []token{
				{0, `{`, KindObjectBegin},
				{1, `"age"`, KindObjectKey},
				{7, `30`, KindNumber},
				{10, `"year"`, KindObjectKey},
				{17, `2022`, KindNumber},
				{21, `}`, KindObjectEnd},
			},
		},
		"valid JSON object with mixed values": {
			//                11111111112222222222333333333344
			//      012345678901234567890123456789012345678901
			input: `{"name":"John","age":30,"city":"New York"}`,
			want: []token{
				{0, `{`, KindObjectBegin},
				{1, `"name"`, KindObjectKey},
				{8, `"John"`, KindString},
				{15, `"age"`, KindObjectKey},
				{21, `30`, KindNumber},
				{24, `"city"`, KindObjectKey},
				{31, `"New York"`, KindString},
				{41, `}`, KindObjectEnd},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, parseJSON(tt.input))
		})
	}
}
