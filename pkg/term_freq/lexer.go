package termfreq

import (
	"strings"
	"unicode"
)

type lexer struct {
	content []rune
	index   int
}

func NewLexer(c string) *lexer {
	return &lexer{content: []rune(c), index: 0}
}

func (l *lexer) trimLeftSpace() {
	for len(l.content) != l.index && unicode.IsSpace(l.content[l.index]) {
		l.index++
	}
}

func (l *lexer) chopFrom(start, end int) *string {
	token := string(l.content[start:end])
	l.index = end
	return &token
}

func (l *lexer) NextToken() *string {
	l.trimLeftSpace()
	if l.index == len(l.content) {
		return nil
	}

	if unicode.IsLetter(l.content[l.index]) {
		return l.chopLetters()
	}

	if unicode.IsNumber(l.content[l.index]) {
		return l.chopNumbers()
	}

	return l.chopFrom(l.index, l.index+1)
}

func (l *lexer) chopLetters() *string {
	i := l.index
	for i != len(l.content) && unicode.IsLetter(l.content[i]) {
		i += 1
	}
	token := strings.ToLower(*l.chopFrom(l.index, i))
	return &token
}

func (l *lexer) chopNumbers() *string {
	i := l.index
	for i != len(l.content) && unicode.IsNumber(l.content[i]) {
		i += 1
	}
	return l.chopFrom(l.index, i)
}
