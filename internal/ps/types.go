package ps

import "github.com/larynjahor/gokunst/internal/types"

type Real = types.Real
type Integer = types.Integer
type Boolean = types.Boolean
type Name = types.Name
type Null = types.Null
type String = types.String

type Keyword string
type Delimiter string
type EOF struct{}

var (
	eof  = EOF{}
	null = Null{}
)

type Lexem any
