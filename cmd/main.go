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
	targetPath := "docs.gl"

	tfIndex, err := createTermFreqIndex(targetPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Indexed %d files in %s ...\n", len(tfIndex), targetPath)
}

func createTermFreqIndex(targetPath string) (TermFreqIndex, error) {
	root, err := getAbsRootPath(targetPath)
	if err != nil {
		return nil, err
	}

	tfIndex := make(TermFreqIndex)
	skippedCount := 0

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dirName := d.Name()
		if d.IsDir() || dirName[0] == '.' {
			return nil
		}

		ext := filepath.Ext(dirName)
		if ext == "" || ext != ".xhtml" && ext != ".xml" {
			// TODO: show using a command line flag
			// fmt.Printf("Skiping %s\n", path)
			skippedCount++
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error: could not read file.", err)
			return nil
		}
		defer file.Close()
		tf, err := createTermFreq(file)
		if err != nil {
			fmt.Printf("Error: failed to index file %s. %s\n", file.Name(), err)
			return nil
		}

		key := strings.Builder{}
		subpaths := strings.Split(path, "/")
		for i, p := range subpaths {
			if strings.EqualFold(p, targetPath) {
				rest := strings.Join(subpaths[i:], "/")
				key.WriteString(rest)
				break
			}
		}

		tfIndex[key.String()] = tf
		return nil
	})

	if err == nil && skippedCount > 0 {
		fmt.Printf("Skipped %d files\n", skippedCount)
	}

	return tfIndex, err
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

func getAbsRootPath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if wd[len(wd)-3:] == "cmd" {
		wd = wd[0 : len(wd)-4]
	}

	return fmt.Sprintf("%s/%s", wd, path), nil
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
