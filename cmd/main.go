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
)

type (
	TermFreq      = map[string]uint
	TermFreqIndex = map[string]TermFreq
)

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

func main() {
	dirPath, dir, err := getDirEntries("docs.gl/gl4")
	if err != nil {
		log.Fatal(err)
	}

	// contents := []string{}
	tfIndex := make(TermFreqIndex)
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		filename := filepath.Join(dirPath, entry.Name())
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		decoder := xml.NewTokenDecoder(xml.NewDecoder(file))
		tf := make(TermFreq)

		for {
			token, err := decoder.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
			}

			switch t := token.(type) {
			case xml.CharData:
				data := strings.Split(string(t), " ")
				if len(data) == 0 {
					continue
				}

				for _, d := range data {
					term := strings.Trim(d, " ")
					if len(term) == 0 {
						continue
					}

					v, ok := tf[term]
					if ok {
						tf[term] = v + 1
					} else {
						tf[term] = 1
					}
				}
			}
		}

		tfIndex[filename] = tf
	}

	printIndex(&tfIndex, 1)
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
