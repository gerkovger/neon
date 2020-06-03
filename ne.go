package main

import (
	"flag"
	"fmt"
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var extensions = [...]string{".jpeg", ".jpg"}

var width, height *uint
var prefix *string
var force, silent *bool

func main() {
	width = flag.Uint("w", 0, "width of resized image")
	height = flag.Uint("h", 0, "height of resized image")
	prefix = flag.String("pre", "", "output file prefix")
	force = flag.Bool("f", false, "force resizing files with any extension")
	silent = flag.Bool("s", false, "print only fatal errors")

	flag.Parse()
	infile := flag.Args()[0]

	setPrefix()

	files, err := listFiles(infile)
	logFatal("Listing files failed.", err)

	for _, file := range *files {
		run(file)
	}

}

func run(infile string) {
	if !*force && !acceptExtension(infile) {
		return
	}
	start := time.Now()

	fmt.Printf("Resizing %s ", infile)

	file, err := os.Open(infile)
	logFatal("Error opening file.", err)

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)

	if logError("Error decoding file.", err) {
		return
	}

	maxO := img.Bounds().Max
	fmt.Printf("%dx%d -> ", maxO.X, maxO.Y)

	file.Close()

	m := resize.Resize(*width, *height, img, resize.Lanczos3)

	max := m.Bounds().Max
	fmt.Printf("%dx%d in ", max.X, max.Y)

	out, err := os.Create(*prefix + "_" + infile)
	logFatal("Creating output file failed.", err)

	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
	fmt.Printf("%s\n", time.Since(start))
}

func setPrefix() {
	if *prefix == "" {
		if *width == 0 {
			*prefix = fmt.Sprint(*height)
		} else {
			*prefix = fmt.Sprint(*width)
		}
	}
}

func isDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		return true, nil
	case mode.IsRegular():
		// do file stuff
		return false, nil
	}
	return false, nil
}

func listFiles(root string) (*[]string, error) {
	var files []string

	dir, err := isDir(root)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if !dir {
		files = append(files, root)
		return &files, nil
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &files, nil
}

func acceptExtension(infile string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(infile), ext) {
			return true
		}
	}
	return false
}

func logFatal(msg string, err error) {
	if err != nil {
		log.Fatal(msg, err)
	}
}

func logError(msg string, err error) bool {
	if err != nil {
		if !silent {
			log.Println(msg, err)
		}
		return true
	}
	return false
}
