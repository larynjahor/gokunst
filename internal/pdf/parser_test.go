package pdf

import (
	"bytes"
	"io"
	"testing"

	"github.com/larynjahor/gokunst/internal/ps"
	"github.com/larynjahor/gokunst/internal/types"
	"github.com/stretchr/testify/require"
)

func TestParseObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Object
	}{
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

			obj, err := parseObject(ps.NewLexer(r))
			require.NoError(t, err)

			require.Equal(t, obj, tc.want)
		})
	}
}

func TestParser_parseIndirectObject(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Object
	}{
		{
			name: "direct",
			input: `
			0 0 obj <</Type /Direct>>
			endobj
			`,
			want: types.Dict{
				Name("/Type"): Name("/Direct"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trace bytes.Buffer
			r := io.TeeReader(bytes.NewReader([]byte(tt.input)), &trace)

			got, gotErr := parseIndirectObject(ps.NewLexer(r))
			require.NoError(t, gotErr)

			require.Equal(t, tt.want, got)
		})
	}
}
