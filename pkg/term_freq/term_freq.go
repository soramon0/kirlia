package termfreq

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type (
	TermFreq      = map[string]uint
	TermFreqIndex = map[string]TermFreq
)

type IndexArgs struct {
	InputFile     string
	ReportSkipped bool
	OutputFormat  string
}

func NewIndex(args IndexArgs) (TermFreqIndex, error) {
	if args.InputFile == "" {
		return nil, fmt.Errorf("error: file name is required")
	}

	root, err := getAbsRootPath(args.InputFile)
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
			if args.ReportSkipped {
				skippedCount++
				fmt.Printf("Skiping %s\n", path)
			}
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("error: could not read file.", err)
			return nil
		}
		defer file.Close()
		tf, err := createTermFreq(file)
		if err != nil {
			fmt.Printf("error: failed to index file %s. %s\n", file.Name(), err)
			return nil
		}

		key := strings.Builder{}
		subpaths := strings.Split(path, "/")
		targetPaths := strings.Split(args.InputFile, "/")
		parentPath := targetPaths[len(targetPaths)-1]
		for i, p := range subpaths {
			if strings.EqualFold(p, parentPath) {
				rest := strings.Join(subpaths[i:], "/")
				key.WriteString(rest)
				break
			}
		}

		tfIndex[key.String()] = tf
		return nil
	})

	if err != nil {
		return tfIndex, err
	}

	if args.ReportSkipped {
		fmt.Printf("Skipped %d files\n", skippedCount)
	}

	return tfIndex, nil
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
			l := newLexer(string(t))
			term := l.nextToken()
			for {
				if term == nil {
					break
				}
				v, ok := tf[*term]
				if ok {
					tf[*term] = v + 1
				} else {
					tf[*term] = 1
				}

				term = l.nextToken()
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
