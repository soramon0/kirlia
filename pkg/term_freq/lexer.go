package termfreq

import (
	"strings"
	"unicode"
)

type lexer struct {
	content []rune
}

func newLexer(c string) *lexer {
	return &lexer{content: []rune(c)}
}

func (l *lexer) trimLeftSpace() {
	for len(l.content) > 0 && unicode.IsSpace(l.content[0]) {
		l.content = l.content[1:]
	}
}

func (l *lexer) chop(n int) *string {
	token := string(l.content[0:n])
	l.content = l.content[n:]
	return &token
}

func (l *lexer) nextToken() *string {
	l.trimLeftSpace()
	if len(l.content) == 0 {
		return nil
	}

	if unicode.IsLetter(l.content[0]) {
		return l.chopLetters()
	}

	if unicode.IsNumber(l.content[0]) {
		return l.chopNumbers()
	}

	return l.chop(1)
}

func (l *lexer) chopLetters() *string {
	i := 0
	for i < len(l.content) && unicode.IsLetter(l.content[i]) {
		i += 1
	}
	token := strings.ToUpper(*l.chop(i))
	return &token
}

func (l *lexer) chopNumbers() *string {
	i := 0
	for i < len(l.content) && unicode.IsNumber(l.content[i]) {
		i += 1
	}
	return l.chop(i)
}
