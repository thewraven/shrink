package main

import (
	"flag"
	"fmt"
	"image"
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
)

var overrride *bool

func init() {
	formatCheck = regexp.MustCompile("(?i)\\.(jpg|png)$")
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
	overrride = flag.Bool("override", true, `
	 define if the output file must be overriden.`)
	flag.Parse()
}

func main() {

	var paths []string

	filepath.Walk(dir.value, func(path string, info os.FileInfo, err error) error {
		if !formatCheck.MatchString(strings.ToLower(path)) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})

	if outdir.set {
		os.MkdirAll(outdir.value, 0766)
	}

	wg := throttler.New(goroutines.value, len(paths))
	fmt.Println(len(paths))
	for _, path := range paths {
		go func(path string) {
			err := compress(path)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("ok", path)
			}
			wg.Done(nil)
		}(path)

		wg.Throttle()
	}
	_ = wg.Err()
}

func compress(route string) error {
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
		path := filepath.Join(outdir.value, filepath.Base(route))
		if *overrride {
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
		fmt.Println("optimizer")
		fmt.Println(err)
	}
	err = file.Close()
	return err
}
