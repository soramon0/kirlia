package termfreq

import (
	"unicode"
)

type lexer struct {
	content []rune
}

func (l *lexer) trimLeftSpace() {
	for len(l.content) > 0 && unicode.IsSpace(l.content[0]) {
		l.content = l.content[1:]
	}
}

func (l *lexer) chop(n int) []rune {
	token := l.content[0:n]
	l.content = l.content[n:]
	return token
}

func (l *lexer) nextToken() []rune {
	l.trimLeftSpace()
	if len(l.content) == 0 {
		return nil
	}

	if unicode.IsLetter(l.content[0]) {
		i := 0
		for i < len(l.content) && unicode.IsLetter(l.content[i]) {
			i += 1
		}
		return l.chop(i)
	}

	if unicode.IsNumber(l.content[0]) {
		i := 0
		for i < len(l.content) && unicode.IsNumber(l.content[i]) {
			i += 1
		}
		return l.chop(i)
	}

	return l.chop(1)
}
