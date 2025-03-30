package types

import "fmt"

var (
	ErrInvalidDecodeConfig = fmt.Errorf("invalid decode config")
)

func NewDecodeConfig(o Object) (*DecodeConfig, error) {
	if o.Kind() != DictKind {
		return nil, ErrInvalidDecodeConfig
	}

	dict := o.(Dict)

	filter, ok := dict[Name("/Filter")]
	if !ok {
		return nil, ErrInvalidDecodeConfig
	}

	if filter.Kind() != NameKind {
		return nil, ErrInvalidDecodeConfig
	}

	var fk FilterKind

	switch filter.(Name) {
	case Name("/FlateDecode"):
		fk = FlateDecode
	case Name("/ASCIIHexDecode"):
		panic("TODO")
	case Name("/ASCII85Decode"):
		panic("TODO")
	case Name("/LZWDecode"):
		panic("TODO")
	case Name("/RunLengthDecode"):
		panic("TODO")
	case Name("/CCITTFaxDecode"):
		panic("TODO")
	case Name("/DCTDecode"):
		panic("TODO")
	default:
		return nil, ErrInvalidDecodeConfig
	}

	return &DecodeConfig{
		Filter: fk,
	}, nil
}

type FilterKind int

const (
	UnknownFilterKind = iota
	FlateDecode
)

type DecodeConfig struct {
	Filter FilterKind
}

type Trailer struct {
	Prev *int64
	Root Ref
}

type Ref struct {
	Generation Integer
	ID         Integer
}

func (Ref) Kind() ObjectKind {
	return Reference
}

func NewXRef() *XRef {
	return &XRef{
		records: map[Ref]XRefRecord{},
	}
}

type XRef struct {
	records map[Ref]XRefRecord
}

func (x *XRef) Get(gen Integer, id Integer) XRefRecord {
	return x.records[Ref{
		Generation: gen,
		ID:         id,
	}]
}

func (x *XRef) Add(gen Integer, id Integer, rec XRefRecord) {
	x.records[Ref{
		Generation: gen,
		ID:         id,
	}] = rec
}

type XRefRecord struct {
	Offset int64
	Free   bool
}

type ObjectKind int

const (
	NullKind ObjectKind = iota
	IntegerKind
	RealKind
	BooleanKind
	StringKind
	NameKind
	DictKind
	ArrayKind
	Reference
)

type Object interface {
	Kind() ObjectKind
}

type Null struct{}

func (Null) Kind() ObjectKind {
	return NullKind
}

type Integer int

func (Integer) Kind() ObjectKind {
	return IntegerKind
}

type Real float64

func (Real) Kind() ObjectKind {
	return RealKind
}

type Boolean bool

func (Boolean) Kind() ObjectKind {
	return BooleanKind
}

type String string

func (String) Kind() ObjectKind {
	return StringKind
}

type Name string

func (Name) Kind() ObjectKind {
	return NameKind
}

type Dict map[Name]Object

func (Dict) Kind() ObjectKind {
	return DictKind
}

type Array []Object

func (Array) Kind() ObjectKind {
	return ArrayKind
}

type Stream struct {
	DecodeConfig Dict
}
