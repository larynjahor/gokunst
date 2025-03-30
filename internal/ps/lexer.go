package ps

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

const (
	bytesBufferSize = 128
)

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r: bufio.NewReader(r),
	}
}

type Lexer struct {
	r io.ByteScanner
}

func (l *Lexer) Next() Lexem {
	for {
		b, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		switch b {
		case '%':
			for b != '\n' {
				b, err = l.r.ReadByte()
				if err != nil {
					return eof
				}
			}

		case '\x00', '\t', '\n', '\f', '\r', ' ':
			continue
		}

		if err := l.r.UnreadByte(); err != nil {
			return eof
		}

		break
	}

	b, err := l.r.ReadByte()
	if err != nil {
		return eof
	}

	switch b {
	case '/':
		return l.readName()
	case '<':
		next, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		if next == '<' {
			return Delimiter("<<")
		}

		if err := l.r.UnreadByte(); err != nil {
			return eof
		}

		return l.readHexString()
	case '(':
		return l.readLiteralString()
	case '[', ']', '{', '}':
		return Delimiter(b)
	case '>':
		next, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		if next == '>' {
			return Delimiter(">>")
		}

		if err := l.r.UnreadByte(); err != nil {
			return eof
		}

		fallthrough
	default:
		if isDelim(b) {
			return null
		}

		if err := l.r.UnreadByte(); err != nil {
			return eof
		}

		return l.readKeyword()
	}
}

func (l *Lexer) readName() Lexem {
	tmp := []byte{'/'}

	for {
		c, err := l.r.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return eof
		}

		if isDelim(c) || isSpace(c) {
			if err := l.r.UnreadByte(); err != nil {
				return eof
			}

			break
		}

		if c == '#' {
			b1, err := l.r.ReadByte()
			if err != nil {
				return eof
			}

			b2, err := l.r.ReadByte()
			if err != nil {
				return eof
			}

			x := l.unhexByte(b1)<<4 | l.unhexByte(b2)
			if x < 0 {
			}

			tmp = append(tmp, byte(x))

			continue
		}
		tmp = append(tmp, c)
	}

	return Name(string(tmp))
}

func (l *Lexer) readKeyword() Lexem {
	var tmp []byte

	for {
		c, err := l.r.ReadByte()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return eof
		}

		if isDelim(c) || isSpace(c) {
			if err := l.r.UnreadByte(); err != nil {
				return eof
			}

			break
		}

		tmp = append(tmp, c)
	}

	s := string(tmp)

	switch {
	case s == "true":
		return Boolean(true)
	case s == "false":
		return Boolean(false)
	case isInteger(s):
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return null
		}

		return Integer(x)
	case isReal(s):
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return null
		}

		return Real(x)
	}

	if len(s) == 0 {
		return eof
	}

	return Keyword(s)
}

func (l *Lexer) readHexString() Lexem {
	var tmp []byte

	for {
	Loop:
		c, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		if c == '>' {
			break
		}

		if isSpace(c) {
			goto Loop
		}
	Loop2:
		c2, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		if isSpace(c2) {
			goto Loop2
		}

		x := l.unhexByte(c)<<4 | l.unhexByte(c2)
		if x < 0 {
			break
		}

		tmp = append(tmp, byte(x))
	}

	return String(tmp)
}

func (l *Lexer) readLiteralString() Lexem {
	var tmp []byte

	depth := 1
Loop:
	for {
		c, err := l.r.ReadByte()
		if err != nil {
			return eof
		}

		switch c {
		default:
			tmp = append(tmp, c)
		case '(':
			depth++
			tmp = append(tmp, c)
		case ')':
			if depth--; depth == 0 {
				break Loop
			}
			tmp = append(tmp, c)
		case '\\':
			c, err := l.r.ReadByte()
			if err != nil {
				return eof
			}

			switch c {
			default:
				tmp = append(tmp, '\\', c)
			case 'n':
				tmp = append(tmp, '\n')
			case 'r':
				tmp = append(tmp, '\r')
			case 'b':
				tmp = append(tmp, '\b')
			case 't':
				tmp = append(tmp, '\t')
			case 'f':
				tmp = append(tmp, '\f')
			case '(', ')', '\\':
				tmp = append(tmp, c)
			case '\r':
				ch, err := l.r.ReadByte()
				if err != nil {
					return eof
				}

				if ch != '\n' {
					if err := l.r.UnreadByte(); err != nil {
						return eof
					}
				}

				fallthrough
			case '\n':
				// no append
			case '0', '1', '2', '3', '4', '5', '6', '7':
				x := int(c - '0')

				for i := 0; i < 2; i++ {
					ch, err := l.r.ReadByte()
					if err != nil {
						return eof
					}

					if ch < '0' || ch > '7' {
						if err := l.r.UnreadByte(); err != nil {
							return eof
						}

						break
					}

					x = x*8 + int(ch-'0')
				}

				if x > 255 {
					return eof
				}

				tmp = append(tmp, byte(x))
			}
		}
	}

	return String(tmp)
}

func isInteger(s string) bool {
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || '9' < c {
			return false
		}
	}
	return true
}

func isReal(s string) bool {
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	if len(s) == 0 {
		return false
	}
	ndot := 0
	for _, c := range s {
		if c == '.' {
			ndot++
			continue
		}
		if c < '0' || '9' < c {
			return false
		}
	}
	return ndot == 1
}

func (l *Lexer) unhexStringBytes(s []byte) String {
	s = s[:len(s)-1]
	ret := make([]byte, 0, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		ret = append(ret, byte(l.unhexByte(s[i])<<4|l.unhexByte(s[i+1])))
	}

	return String(ret)
}

func (l *Lexer) unhexByte(b byte) int {
	switch {
	case '0' <= b && b <= '9':
		return int(b) - '0'
	case 'a' <= b && b <= 'f':
		return int(b) - 'a' + 10
	case 'A' <= b && b <= 'F':
		return int(b) - 'A' + 10
	}

	return -1
}

func (l *Lexer) unescapeStringBytes(s []byte) String {
	s = s[:len(s)-1]
	ret := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i == len(s)-1 {
			ret = append(ret, s[i])
			continue
		}

		switch s[i+1] {
		case 'n', 'r', 't', 'f', '(', ')', 'b', '\\':
			i += 1
			ret = append(ret, s[i+1])
		}

		var (
			o    byte
			base byte = 1
		)

		for j := range 3 {
			d := s[i+j+1] - '0'

			if 0 < d || d > 9 {
				o = 0
				break
			}

			o += d * base
			base *= 8
		}

		if o > 0 {
			i += 3
			ret = append(ret, o)
		}
	}

	return String(ret)
}

func isSpace(b byte) bool {
	switch b {
	case '\x00', '\t', '\n', '\f', '\r', ' ':
		return true
	}
	return false
}

func isDelim(b byte) bool {
	switch b {
	case '<', '>', '(', ')', '[', ']', '{', '}', '/', '%':
		return true
	}
	return false
}
