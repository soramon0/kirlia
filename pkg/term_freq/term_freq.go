package termfreq

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
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

var outputFormats = map[string]string{
	"json":    "json",
	"msgpack": "msgpack",
}

func GenerateIndex(args IndexArgs) (TermFreqIndex, error) {
	if args.InputFile == "" {
		return nil, fmt.Errorf("error: file name is required")
	}

	format, supported := outputFormats[args.OutputFormat]
	if !supported {
		return nil, fmt.Errorf("error: %q not supported", args.OutputFormat)
	}

	tfIndex, err := newIndex(args.InputFile, args.ReportSkipped)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("index.%s", format)
	if err := saveFile(&tfIndex, filename, format); err != nil {
		return nil, err
	}
	fmt.Println("Saved index in", filename)

	return tfIndex, nil
}

func newIndex(inputFile string, reportSkipped bool) (TermFreqIndex, error) {
	root, err := getAbsRootPath(inputFile)
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
			if reportSkipped {
				skippedCount++
				fmt.Printf("Skiping %s\n", path)
			}
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("error: could not open file.", err)
			return nil
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("error: could not read file stats.", err)
			return nil
		}

		buf := make([]byte, fileInfo.Size())
		_, err = file.Read(buf)
		if err != nil {
			fmt.Println("error: could read file.", err)
			return nil
		}

		tf, err := createTermFreq(bytes.NewReader(buf))
		if err != nil {
			fmt.Printf("error: failed to index file %s. %s\n", file.Name(), err)
			return nil
		}

		key := strings.Builder{}
		subpaths := strings.Split(path, "/")
		targetPaths := strings.Split(inputFile, "/")
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
		return nil, err
	}

	if reportSkipped {
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
			l := NewLexer(string(t))
			term := l.NextToken()
			for {
				if term == nil {
					break
				}
				_, ok := tf[*term]
				if ok {
					tf[*term] += 1
				} else {
					tf[*term] = 1
				}

				term = l.NextToken()
			}
		}
	}

	return tf, nil
}

func ReadIndexFile(format string) (TermFreqIndex, error) {
	if format == "" {
		return nil, fmt.Errorf("error: format is required")
	}

	encodingFormat, supported := outputFormats[format]
	if !supported {
		return nil, fmt.Errorf("error: %q not supported", format)
	}

	filename := fmt.Sprintf("index.%s", encodingFormat)
	root, err := getAbsRootPath(filename)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(
				"error: %s file not found. index your folder/file first",
				filename,
			)
		}
		return nil, fmt.Errorf("error: reading %s failed. %s", filename, err)
	}

	buf := bytes.NewReader(data)
	tfIndex := make(TermFreqIndex)

	if format == "json" {
		err = json.NewDecoder(buf).Decode(&tfIndex)
	}

	if format == "msgpack" {
		err = msgpack.NewDecoder(buf).Decode(&tfIndex)
	}

	if err != nil {
		return nil, fmt.Errorf("error: decoding %s failed. %s", filename, err)
	}

	return tfIndex, nil
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

func saveFile(tfIndex *TermFreqIndex, filename, format string) error {
	var buf bytes.Buffer
	var err error

	if format == "json" {
		err = json.NewEncoder(&buf).Encode(tfIndex)
	}

	if format == "msgpack" {
		err = msgpack.NewEncoder(&buf).Encode(tfIndex)
	}

	if err != nil {
		return fmt.Errorf("error: %s encoding failed. %s", format, err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0666); err != nil {
		return fmt.Errorf("error: creating %s failed. %s", filename, err)
	}

	return nil
}
