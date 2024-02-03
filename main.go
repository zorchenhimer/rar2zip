package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nwaples/rardecode"
)

var verbose bool = false

func main() {
	// Default to all rar files in the current directory.
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "*.rar")
	}

	files := []string{}
	for _, arg := range os.Args[1:len(os.Args)] {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
			fmt.Println("Verbose turned on")
			continue
		}

		matches, err := filepath.Glob(arg)
		if err != nil {
			fmt.Printf("Glob failed: %s\n", err)
			continue
		}
		if verbose {
			for _, f := range matches {
				fmt.Printf("Adding file %q\n", f)
			}
		}
		files = append(files, matches...)

		if len(matches) == 0 && exists(arg) {
			if verbose {
				fmt.Printf("Adding file: %q\n", arg)
			}
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Println("No files to process")
		return
	}
	fmt.Printf("Number of files to process: %d\n", len(files))

	for _, f := range files {
		err := convert(f)
		if err != nil {
			fmt.Printf("Error converting %s: %s", f, err.Error())
		}
	}
}

func convert(inputname string) error {
	outname := inputname
	if strings.HasSuffix(outname, ".rar") {
		outname = outname[:len(outname)-4] + ".zip"
	} else {
		outname = outname + ".zip"
	}

	rawin, err := os.Open(inputname)
	if err != nil {
		return err
	}
	defer rawin.Close()

	reader, err := rardecode.NewReader(rawin, "")
	if err != nil {
		return err
	}

	rawout, err := os.Create(outname)
	if err != nil {
		return err
	}
	defer rawout.Close()

	writer := zip.NewWriter(rawout)
	defer writer.Close()

	for {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if header.IsDir {
			continue
		}

		fmt.Println(header.Name)
		zhead := &zip.FileHeader{
			Name:     header.Name,
			Method:   zip.Deflate,
			Modified: header.ModificationTime,
		}

		w, err := writer.CreateHeader(zhead)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, reader)
		if err != nil {
			return err
		}
	}

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
