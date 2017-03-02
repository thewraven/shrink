package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bradberger/optimizer"
	"github.com/nozzle/throttler"
)

type checkedStringFlag struct {
	set   bool
	value string
}

func (c *checkedStringFlag) Set(v string) error {
	c.value, c.set = v, true
	return nil
}

func (c *checkedStringFlag) String() string {
	return c.value
}

type checkedIntFlag struct {
	set   bool
	value int
}

func (c *checkedIntFlag) Set(v string) error {
	var err error
	c.value, err = strconv.Atoi(v)
	if err != nil {
		c.set = true
	}
	return err
}

func (c *checkedIntFlag) String() string {
	return strconv.Itoa(c.value)
}

var (
	outdir      = checkedStringFlag{}
	quality     = checkedIntFlag{value: 60}
	dir         = checkedStringFlag{value: "."}
	formatCheck *regexp.Regexp
	goroutines  = checkedIntFlag{value: 50}

	override          *bool
	preserveHierarchy *bool
)

func init() {
	formatCheck = regexp.MustCompile("(?i)\\.(jpg|png|jpeg|gif)$")
	flag.Var(&outdir, "output", `
    output folder for the compressed images.
    If not set, it will override the original image`)
	flag.Var(&quality, "quality", `
   quality of the compressed images.
   `)
	flag.Var(&dir, "dir", `
   directory to lookup for images to be compressed. It accepts jpg,jpeg,png formats.
   `)
	flag.Var(&goroutines, "workers", `
   number of the workers that can be spawned concurrently.
   `)
	override = flag.Bool("override", false, `
	 define if the output file must be overriden.`)
	preserveHierarchy = flag.Bool("hierarchy", false, `
	 describes whether the structure of the inner folders must be preserved or not`)
	flag.Parse()
}

func main() {

	var paths []string
	var subfolders []string
	filepath.Walk(dir.value, func(path string, info os.FileInfo, err error) error {
		if !formatCheck.MatchString(strings.ToLower(path)) {
			return nil
		}
		paths = append(paths, path)
		subfolder, _ := filepath.Rel(dir.value, filepath.Dir(path))
		subfolders = append(subfolders, subfolder)
		return nil
	})

	if outdir.set {
		os.MkdirAll(outdir.value, 0766)
	}

	wg := throttler.New(goroutines.value, len(paths))
	for i := range paths {
		go func(path, subdir string) {
			err := compress(path, subdir)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("ok", path)
			}
			wg.Done(nil)
		}(paths[i], subfolders[i])

		wg.Throttle()
	}
	_ = wg.Err()
}

func compress(route, subfolder string) error {
	file, err := os.Open(route)
	if err != nil {
		return err
	}
	img, mime, err := image.Decode(file)
	file.Close()
	if err != nil {
		return err
	}
	if !outdir.set {
		os.Remove(route)
		file, err = os.Create(route)
	} else {
		var path string
		if *preserveHierarchy {
			path = filepath.Join(outdir.value, subfolder, filepath.Base(route))
			os.MkdirAll(filepath.Dir(path), 0766)
		} else {
			path = filepath.Join(outdir.value, filepath.Base(route))
		}
		if *override {
			os.Remove(path)
		}
		file, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	err = optimizer.Encode(file, img, optimizer.Options{
		Mime:    "image/" + mime,
		Quality: quality.value,
	})
	if err != nil {
		fmt.Println(err)
	}
	err = file.Close()
	return err
}
