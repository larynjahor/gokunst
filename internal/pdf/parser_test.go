package pdf

import (
	"bytes"
	"testing"

	"github.com/larynjahor/gokunst/internal/ds"
	"github.com/larynjahor/gokunst/internal/ps"
	"github.com/larynjahor/gokunst/internal/types"
	"github.com/stretchr/testify/require"
)

func TestParseObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Object
	}{{
		name:  "indirect direct",
		input: "0 1 obj <</Name <</Name /Val>> /Name1 [1 2 3]>> endobj",
		want: types.Dict{
			Name("/Name"): types.Dict{
				Name("/Name"): Name("/Val"),
			},
			Name("/Name1"): types.Array{Integer(1), Integer(2), Integer(3)},
		},
	},
		{
			name:  "basic",
			input: "1",
			want:  Integer(1),
		},
		{
			name:  "empty array",
			input: "[]",
			want:  types.Array{},
		},
		{
			name:  "empty dict",
			input: "<<>>",
			want:  types.Dict{},
		},
		{
			name:  "nested dict",
			input: "<</Name <</Name /Val>> /Name1 [1 2 3]>>",
			want: types.Dict{
				Name("/Name"): types.Dict{
					Name("/Name"): Name("/Val"),
				},
				Name("/Name1"): types.Array{Integer(1), Integer(2), Integer(3)},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			p := &objectParser{l: ps.NewLexer(r), sourceReader: nil, st: ds.NewStack[ps.Lexem]()}

			obj, err := p.ParseObject()
			require.NoError(t, err)

			require.Equal(t, obj, tc.want)
		})
	}
}
