package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"slices"

	"github.com/larynjahor/gokunst/internal/ds"
	"github.com/larynjahor/gokunst/internal/ps"
	"github.com/larynjahor/gokunst/internal/types"
)

const (
	trailerSeekOffset = 128
)

var (
	ErrXRefNotFound          = fmt.Errorf("xref not found")
	ErrInvalidXRef           = fmt.Errorf("invalid xref")
	ErrInvalidIndirectObject = fmt.Errorf("invalid indirect object")
	ErrInvalidObject         = fmt.Errorf("invalid object")
)

func NewDocument(r io.ReadSeeker) (*Document, error) {
	return &Document{
		r:    r,
		xref: nil,
	}, nil
}

type Document struct {
	r    io.ReadSeeker
	xref *types.XRef
}

type Parser struct {
	r io.ReadSeeker
}

type objectParser struct {
	sourceReader io.ReaderAt
	l            *ps.Lexer
	st           *ds.Stack[ps.Lexem]
}

func (p *objectParser) ParseObject() (Object, error) {
	first := p.l.Next()
	if (first == ps.EOF{}) {
		return nil, io.EOF
	}

	second := p.l.Next()
	if (second == ps.EOF{}) {
		obj, ok := first.(Object)
		if ok {
			return obj, nil
		}

		return nil, io.EOF
	}

	third := p.l.Next()

	if (third == ps.EOF{}) {
		p.unreadLexem(second)
		p.unreadLexem(first)

		return p.parseDirectObject()
	}

	_, firstIsNumber := first.(Integer)
	_, secondIsNumber := second.(Integer)

	p.unreadLexem(third)
	p.unreadLexem(second)
	p.unreadLexem(first)

	if !firstIsNumber || !secondIsNumber || third != ps.Keyword("obj") {
		return p.parseDirectObject()
	}

	return p.parseIndirectObject()
}

func (p *objectParser) unreadLexem(l ps.Lexem) {
	p.st.Push(l)
}

func (p *objectParser) readLexem() ps.Lexem {
	if p.st.Empty() {
		return p.l.Next()
	}

	return p.st.Pop()
}

func (p *objectParser) parseDirectObject() (Object, error) {
	lexems := ds.NewStack[any]()

	for {
		next := p.readLexem()
		if (next == ps.EOF{}) {
			return nil, io.EOF
		}

		if lexems.Empty() {
			if obj, ok := next.(Object); ok {
				return obj, nil
			}
		}

		if next == ps.Delimiter("]") {
			arr := types.Array{}

			for {
				if lexems.Empty() {
					break
				}

				top := lexems.Pop()
				if top == ps.Delimiter("[") {
					break
				}

				obj, ok := top.(Object)
				if !ok {
					return nil, ErrInvalidObject
				}

				arr = append(arr, obj)
			}

			slices.Reverse(arr)

			if lexems.Empty() {
				return arr, nil
			}

			next = arr
		}

		if next == ps.Delimiter(">>") {
			dict := types.Dict{}

			for {
				if lexems.Empty() {
					break
				}

				top := lexems.Pop()
				if top == ps.Delimiter("<<") {
					break
				}

				val, ok := top.(Object)
				if !ok {
					return nil, ErrInvalidObject
				}

				if lexems.Empty() {
					return nil, ErrInvalidObject
				}

				top = lexems.Pop()
				key, ok := top.(Name)
				if !ok {
					return nil, ErrInvalidObject
				}

				dict[key] = val
			}

			if lexems.Empty() {
				return dict, nil
			}

			next = dict
		}

		lexems.Push(next)
	}
}

func (p *objectParser) parseIndirectObject() (Object, error) {
	if _, ok := p.readLexem().(Integer); !ok {
		return nil, ErrInvalidIndirectObject
	}

	if _, ok := p.readLexem().(Integer); !ok {
		return nil, ErrInvalidIndirectObject
	}

	if keyword := p.readLexem(); keyword != ps.Keyword("obj") {
		return nil, ErrInvalidIndirectObject
	}

	obj, err := p.parseDirectObject()
	if err != nil {
		return nil, err
	}

	next := p.readLexem()

	if next == ps.Keyword("endobj") {
		return obj, nil
	}

	if next != ps.Keyword("stream") {
		return nil, ErrInvalidIndirectObject
	}

	obj, err = p.parseDirectObject()
	if err != nil {
		return nil, err
	}

	decodeConfig, err := types.NewDecodeConfig(obj)
	if err != nil {
		return nil, err
	}

	stream, ok := p.readLexem().(ps.Keyword)
	if !ok {
		return nil, ErrInvalidIndirectObject
	}

	var buf bytes.Buffer

	switch decodeConfig.Filter {
	case types.FlateDecode:
		z, err := zlib.NewReader(bytes.NewBufferString(string(stream)))
		if err != nil {
			return nil, err
		}

		defer z.Close()

		if _, err := io.Copy(&buf, z); err != nil {
			return nil, err
		}
	default:
		panic("unsupported decode filter")
	}

	if next != ps.Keyword("endstream") {
		return nil, ErrInvalidIndirectObject
	}

	return p.parseDirectObject()
}

func (p *Parser) parseXRef() ([][]types.XRefRecord, error) {
	_, err := p.r.Seek(128, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	bs, err := io.ReadAll(p.r)
	if err != nil {
		return nil, err
	}

	shift := bytes.Index(bs, []byte("startxref"))
	if shift < 0 {
		return nil, ErrXRefNotFound
	}

	offset := trailerSeekOffset - shift

	if _, err := p.r.Seek(int64(offset), io.SeekEnd); err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *Parser) parseSingleXRef() (*types.XRef, *types.Trailer, error) {
	l := ps.NewLexer(p.r)
	header := l.Next()
	if header != ps.Keyword("xref") {
		return nil, nil, ErrInvalidXRef
	}

	xref := types.NewXRef()

	var cur ps.Lexem

	for {
		var (
			xrefStartObject = 0
			xrefSize        = 0
		)

		cur = l.Next()
		if _, ok := cur.(Integer); !ok {
			break
		}

		xrefStartObject = int(cur.(Integer))

		n := l.Next()
		if _, ok := n.(Integer); !ok {
			return nil, nil, ErrInvalidXRef
		}

		xrefSize = int(cur.(Integer))

		for i := range xrefSize {
			offset := l.Next()
			generation := l.Next()
			status := l.Next()

			if _, ok := offset.(Integer); !ok {
				return nil, nil, ErrInvalidXRef
			}

			if _, ok := generation.(Integer); !ok {
				return nil, nil, ErrInvalidXRef
			}

			if _, ok := status.(ps.Keyword); !ok || !(status == ps.Keyword("n") || status == ps.Keyword("f")) {
				return nil, nil, ErrInvalidXRef
			}

			generationPDF := generation.(Integer)

			xref.Add(generationPDF, Integer(i+xrefStartObject), types.XRefRecord{
				Offset: int64(offset.(Integer)),
				Free:   status == ps.Keyword("f"),
			})
		}
	}

	trailer := &types.Trailer{}

	return xref, trailer, nil
}
