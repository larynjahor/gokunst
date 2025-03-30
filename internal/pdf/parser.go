package pdf

import (
	"bytes"
	"fmt"
	"io"
	"pdf/internal/ps"
	"pdf/internal/types"
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

func (p *Parser) parseObject() (Object, error) {
	return nil, nil
}

func (p *Parser) parseIndirectObject() (Object, error) {
	l := ps.NewLexer(p.r)

	if _, ok := l.Next().(Integer); !ok {
		return nil, ErrInvalidIndirectObject
	}

	if _, ok := l.Next().(Integer); !ok {
		return nil, ErrInvalidIndirectObject
	}

	if keyword := l.Next(); keyword != ps.Keyword("obj") {
		return nil, ErrInvalidIndirectObject
	}

	obj, err := p.parseObject()
	if err != nil {
		return nil, err
	}

	next := l.Next()

	if next == ps.Keyword("endobj") {
		return obj, nil
	}

	if next != ps.Keyword("stream") {
		return nil, ErrInvalidIndirectObject
	}

	// TODO parse stream

	if next != ps.Keyword("endstream") {
		return nil, ErrInvalidIndirectObject
	}

	return nil, nil
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

	for {

	}

	return xref, trailer, nil
}
