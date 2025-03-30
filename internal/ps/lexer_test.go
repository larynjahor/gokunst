package ps

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Lexem
	}{
		{
			name: "blank",
			input: `%fjsid
			  %% foo

			 %%bar
			`,
			want: []Lexem{},
		},
		{
			name:  "boolean",
			input: `true`,
			want: []Lexem{
				Boolean(true),
			},
		},
		{
			name:  "integer",
			input: `-821`,
			want: []Lexem{
				Integer(-821),
			},
		},
		{
			name:  "float",
			input: `8.21`,
			want: []Lexem{
				Real(8.21),
			},
		},
		{
			name:  "regular string",
			input: `(hello world!)`,
			want: []Lexem{
				String("hello world!"),
			},
		},
		{
			name: "multiline string",
			input: `(hello \
world!)`,
			want: []Lexem{
				String("hello world!"),
			},
		},
		{
			name:  "multiline string",
			input: `(hello\tworld:\)\n)`,
			want: []Lexem{
				String("hello\tworld:)\n"),
			},
		},
		{
			name:  "hex string",
			input: `<68 65 6C 6C 6F>`,
			want: []Lexem{
				String("hello"),
			},
		},
		{
			name:  "name",
			input: `/HelloWorld`,
			want: []Lexem{
				Name("/HelloWorld"),
			},
		},
		{
			name:  "keyword",
			input: `0 1 obj endobj`,
			want: []Lexem{
				Integer(0),
				Integer(1),
				Keyword("obj"),
				Keyword("endobj"),
			},
		},
		{

			name:  "delimeters",
			input: `<<>>[]`,
			want: []Lexem{
				Delimiter("<<"),
				Delimiter(">>"),
				Delimiter("["),
				Delimiter("]"),
			},
		},
		{
			name:  "nested dict",
			input: `<< /Nested<</Dictionary /Value>>>>`,
			want: []Lexem{
				Delimiter("<<"),
				Name("/Nested"),
				Delimiter("<<"),
				Name("/Dictionary"),
				Name("/Value"),
				Delimiter(">>"),
				Delimiter(">>"),
			},
		},
		{
			name:  "dict and array",
			input: `[ (hello) (world) << /Key /Value >> ]`,
			want: []Lexem{
				Delimiter("["),
				String("hello"),
				String("world"),
				Delimiter("<<"),
				Name("/Key"),
				Name("/Value"),
				Delimiter(">>"),
				Delimiter("]"),
			},
		},
		{
			name: "stream",
			input: `
			0 1 obj
			<</Filter/FlateDecode /Length 1488>>
			stream
			agbaagba
			endstream
			endobj
			`,
			want: []Lexem{
				Integer(0),
				Integer(1),
				Keyword("obj"),
				Delimiter("<<"),
				Name("/Filter"),
				Name("/FlateDecode"),
				Name("/Length"),
				Delimiter(">>"),
				Integer(1488),
				Keyword("stream"),
				Keyword("agbaagba"),
				Keyword("endstream"),
				Keyword("endobj"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewBufferString(tc.input)
			l := NewLexer(r)

			var got []Lexem

			for {
				lexem := l.Next()
				if lexem == eof {
					break
				}

				got = append(got, lexem)
			}

			require.ElementsMatch(t, tc.want, got)
		})
	}
}
