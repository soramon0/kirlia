package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type (
	TermFreq      = map[string]uint
	TermFreqIndex = map[string]TermFreq
)

func main() {
	targetPath := "docs.gl/gl4"
	dirPath, dir, err := getDirEntries(targetPath)
	if err != nil {
		log.Fatal(err)
	}

	tfIndex := make(TermFreqIndex)
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == "" || ext != ".xhtml" && ext != ".xml" {
			fmt.Printf("Skiping %s\n", fmt.Sprintf("%s/%s", targetPath, entry.Name()))
			continue
		}

		filename := filepath.Join(dirPath, entry.Name())
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		tf, err := createTermFreq(file)
		if err != nil {
			fmt.Printf("Error: failed to index file %s. %s\n", file.Name(), err)
			continue
		}
		tfIndex[filename] = tf
	}

	fmt.Printf("Indexed %d files in %s ...\n", len(tfIndex), targetPath)
}

func createTermFreq(r io.Reader) (TermFreq, error) {
	decoder := xml.NewTokenDecoder(xml.NewDecoder(r))
	tf := make(TermFreq)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return tf, err
		}

		switch t := token.(type) {
		case xml.CharData:
			lexer := &Lexer{content: []rune(string(t))}
			lt := lexer.nextToken()
			for {
				if lt == nil {
					break
				}
				term := strings.ToUpper(string(lt))
				v, ok := tf[term]
				if ok {
					tf[term] = v + 1
				} else {
					tf[term] = 1
				}

				lt = lexer.nextToken()
			}
		}
	}

	return tf, nil
}

func getDirEntries(path string) (string, []fs.DirEntry, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	if wd[len(wd)-3:] == "cmd" {
		wd = wd[0 : len(wd)-3]
	}
	dirPath := fmt.Sprintf("%s/%s", wd, path)
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return "", nil, err
	}

	return dirPath, dir, nil
}

type Lexer struct {
	content []rune
}

func (l *Lexer) trimLeftSpace() {
	for len(l.content) > 0 && unicode.IsSpace(l.content[0]) {
		l.content = l.content[1:]
	}
}

func (l *Lexer) chop(n int) []rune {
	token := l.content[0:n]
	l.content = l.content[n:]
	return token
}

func (l *Lexer) nextToken() []rune {
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

func printIndex(tfi *TermFreqIndex, depth int) {
	if tfi == nil {
		return
	}

	for filename, tf := range *tfi {
		if depth == 0 {
			break
		}

		fmt.Printf("Index %s:\n", filename)
		for term, freq := range tf {
			fmt.Printf("Term: %q -> %d\n", term, freq)
		}
		depth--
	}
}
