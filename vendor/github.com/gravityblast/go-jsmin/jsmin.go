package jsmin

import (
	"bufio"
	"errors"
	"io"
	"log"
)

const eof = -1

type minifier struct {
	src          *bufio.Reader
	dest         *bufio.Writer
	theA         int
	theB         int
	theX         int
	theY         int
	theLookahead int
}

var (
	errorUnterminatedComment            = errors.New("unterminated comment")
	errorUnterminatedStringLiteral      = errors.New("unterminated string literal")
	errorUnterminatedSetInRegexpLiteral = errors.New("unterminated set in regular expression literal")
	errorUnterminatedRegexpLiteral      = errors.New("unterminated regular expression literal")
)

func (m *minifier) putc(c int) {
	m.dest.WriteByte(byte(c))
	m.dest.Flush()
}

// isAlphanum -- return true if the character is a letter, digit, underscore,
// dollar sign, or non-ASCII character.
func (m *minifier) isAlphanum(c int) bool {
	return ((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') || c == '_' || c == '$' || c == '\\' ||
		c > 126)
}

// get -- return the next character from stdin. Watch out for lookahead. If
// the character is a control character, translate it to a space or
// linefeed.
func (m *minifier) get() int {
	c := m.theLookahead
	m.theLookahead = eof
	if c == eof {
		b, err := m.src.ReadByte()
		if err == io.EOF {
			c = eof
		} else if err != nil {
			log.Fatal(err)
		} else {
			c = int(b)
		}
	}
	if c >= ' ' || c == '\n' || c == eof {
		return c
	}
	if c == '\r' {
		return '\n'
	}

	return ' '
}

// peek -- get the next character without getting it.
func (m *minifier) peek() int {
	m.theLookahead = m.get()
	return m.theLookahead
}

// next -- get the next character, excluding comments. peek() is used to see
// if a '/' is followed by a '/' or '*'.
func (m *minifier) next() (int, error) {
	c := m.get()
	if c == '/' {
		switch m.peek() {
		case '/':
			for {
				c = m.get()
				if c <= '\n' {
					break
				}
			}
		case '*':
			m.get()
			for c != ' ' {
				switch m.get() {
				case '*':
					if m.peek() == '/' {
						m.get()
						c = ' '
					}
				case eof:
					return -1, errorUnterminatedComment
				}
			}
		}
	}
	m.theY = m.theX
	m.theX = c
	return c, nil

}

// action -- do something! What you do is determined by the argument:
// 	1   Output A. Copy B to A. Get the next B.
// 	2   Copy B to A. Get the next B. (Delete A).
// 	3   Get the next B. (Delete B).
// action treats a string as a single character. Wow!
// action recognizes a regular expression if it is preceded by ( or , or =.
func (m *minifier) action(d int) error {
	switch d {
	case 1:
		m.putc(m.theA)
		if (m.theY == '\n' || m.theY == ' ') &&
			(m.theA == '+' || m.theA == '-' || m.theA == '*' || m.theA == '/') &&
			(m.theB == '+' || m.theB == '-' || m.theB == '*' || m.theB == '/') {
			m.putc(m.theY)
		}
		fallthrough
	case 2:
		m.theA = m.theB
		if m.theA == '\'' || m.theA == '"' || m.theA == '`' {
			for {
				m.putc(m.theA)
				m.theA = m.get()
				if m.theA == m.theB {
					break
				}
				if m.theA == '\\' {
					m.putc(m.theA)
					m.theA = m.get()
				}
				if m.theA == eof {
					return errorUnterminatedStringLiteral
				}
			}
		}
		fallthrough
	case 3:
		var err error
		if m.theB, err = m.next(); err != nil {
			return err
		}
		if m.theB == '/' && (m.theA == '(' || m.theA == ',' || m.theA == '=' || m.theA == ':' ||
			m.theA == '[' || m.theA == '!' || m.theA == '&' || m.theA == '|' ||
			m.theA == '?' || m.theA == '+' || m.theA == '-' || m.theA == '~' ||
			m.theA == '*' || m.theA == '/' || m.theA == '{' || m.theA == '\n') {
			m.putc(m.theA)
			if m.theA == '/' || m.theA == '*' {
				m.putc(' ')
			}
			m.putc(m.theB)
			for {
				m.theA = m.get()
				if m.theA == '[' {
					for {
						m.putc(m.theA)
						m.theA = m.get()
						if m.theA == ']' {
							break
						}
						if m.theA == '\\' {
							m.putc(m.theA)
							m.theA = m.get()
						}
						if m.theA == eof {
							return errorUnterminatedSetInRegexpLiteral
						}
					}
				} else if m.theA == '/' {
					switch m.peek() {
					case '/', '*':
						return errorUnterminatedSetInRegexpLiteral
					}
					break
				} else if m.theA == '\\' {
					m.putc(m.theA)
					m.theA = m.get()
				}
				if m.theA == eof {
					return errorUnterminatedRegexpLiteral
				}
				m.putc(m.theA)
			}
			var err error
			if m.theB, err = m.next(); err != nil {
				return err
			}
		}
	}

	return nil
}

// jsmin -- Copy the input to the output, deleting the characters which are
// insignificant to JavaScript. Comments will be removed. Tabs will be
// replaced with spaces. Carriage returns will be replaced with linefeeds.
// Most spaces and linefeeds will be removed.
func (m *minifier) min() error {
	if m.peek() == 0xEF {
		m.get()
		m.get()
		m.get()
	}
	m.theA = '\n'
	if err := m.action(3); err != nil {
		return err
	}

	for m.theA != eof {
		switch m.theA {
		case ' ':
			a := 2
			if m.isAlphanum(m.theB) {
				a = 1
			}
			if err := m.action(a); err != nil {
				return err
			}
		case '\n':
			switch m.theB {
			case '{', '[', '(', '+', '-', '!', '~':
				if err := m.action(1); err != nil {
					return err
				}
			case ' ':
				if err := m.action(3); err != nil {
					return err
				}
			default:
				a := 2
				if m.isAlphanum(m.theB) {
					a = 1
				}
				if err := m.action(a); err != nil {
					return err
				}
			}
		default:
			switch m.theB {
			case ' ':
				a := 3
				if m.isAlphanum(m.theA) {
					a = 1
				}
				if err := m.action(a); err != nil {
					return err
				}
			case '\n':
				switch m.theA {
				case '}', ']', ')', '+', '-', '"', '\'', '`':
					if err := m.action(1); err != nil {
						return err
					}
				default:
					a := 3
					if m.isAlphanum(m.theA) {
						a = 1
					}
					if err := m.action(a); err != nil {
						return err
					}
				}
			default:
				if err := m.action(1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func newMinifier(src io.Reader, dest io.Writer) *minifier {
	return &minifier{
		src:  bufio.NewReader(src),
		dest: bufio.NewWriter(dest),
		theX: eof,
		theY: eof,
	}
}

// Min minifies javascript reading from src and writing to dest
func Min(src io.Reader, dest io.Writer) error {
	m := newMinifier(src, dest)
	return m.min()
}
