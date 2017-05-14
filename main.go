package main

import (
    "archive/zip"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/nwaples/rardecode"
)

var verbose bool = false

type Archive struct {
    Name    string
    Files   []File
}

type File struct {
    Name    string
    Data    []byte
}

func (f *File) String() string {
    return fmt.Sprintf("<File Name:%q len(Data):%d>", f.Name, len(f.Data))
}

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
        archive, err := process(f)
        if err != nil {
            fmt.Printf("Unable to read rar archive: %s\n", err)
            continue
        }

        if err = create(archive); err != nil {
            fmt.Printf("Unable to write zip archive: %s\n", err)
        }
    }
}

func create(archive *Archive) error {
    file, err := os.Create(archive.Name)
    if err != nil {
        return fmt.Errorf("Unable to create file: %s", err)
    }
    defer file.Close()

    writer := zip.NewWriter(file)
    defer writer.Close()
    for _, f := range archive.Files {
        w, err := writer.CreateHeader(&zip.FileHeader{
            Name: f.Name,
        })
        if err != nil {
            return fmt.Errorf("Unable to create file %q in archive: %s", f.Name, err)
        }

        fmt.Fprintf(w, "%s", f.Data)
    }
    return nil
}

func process(filename string) (*Archive, error) {
    fmt.Printf("Processing %s\n", filename)
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("Error opening %q for reading: %s\n", filename, err)
    }
    defer file.Close()

    reader, err := rardecode.NewReader(file, "")
    if err != nil {
        return nil, fmt.Errorf("Unable to open RAR file %q: %s\n", filename, err)
    }

    var header *rardecode.FileHeader
    header, err = reader.Next()
    if err != nil {
        fmt.Printf("Bad first read: %s\n", err)
        header, err = reader.Next()
        if err != nil {
            fmt.Printf("Bad second read: %s\n", err)
            return nil, fmt.Errorf("First two reads bad")
        }
    }
    filelist := []File{}
    for err == nil {
        if !header.IsDir {
            if strings.ToLower(filepath.Base(header.Name)) == "thumbs.db" {
                header, err = reader.Next()
                continue
            }
            data, err := ioutil.ReadAll(reader)
            if err != nil {
                fmt.Println("Unable to ReadAll(): ", err)
            }
            if verbose {
                fmt.Printf(" %s - len(data): %d\n", header.Name, len(data))
            }
            filelist = append(filelist, File{Name: header.Name, Data: data})
        }
        header, err = reader.Next()
    }
    if err != nil && err != io.EOF {
        return nil, fmt.Errorf("Bad read: %s\n", err)
    }

    newName := filename[0:strings.LastIndex(filename, ".")] + ".zip"
    return &Archive{Name: newName, Files: filelist}, nil
}

func exists(path string) bool {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    return true
}
