package main

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// imageHolder is an object that can contain a name and an image object
// the name is used to construct a resized filename
type imageHolder struct {
	name string
	data image.Image
}

// takes filename, opens the file and decodes jpeg into an image.Image
func decoder(filenames []string, out chan<- imageHolder) {
	for _, fh := range filenames {
		file, err := os.Open(fh)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}

		// decode contents (jpeg)
		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err) // perhaps continue instead if decode fails?
		}
		imageH := imageHolder{name: fh, data: img}
		out <- imageH // write decoded image to channel
	}
	close(out)
}

// takes image and resizes to a thumbnail
// interpolation method can be adjusted (see documentation)
func thumbnailer(in <-chan imageHolder, out chan<- imageHolder) {
	for imageH := range in {
		imageH.data = resize.Thumbnail(150, 150, imageH.data, resize.Lanczos3)
		out <- imageH
	}
	close(out)
}

// renames a file to indicate it is a thumbnail of the original
func renameFile(filename string) string {
	extension := filepath.Ext(filename)
	newFname := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_thumb" + extension // remove extension
	return newFname
}

// encodes the image objects and writes to file
func writer(in <-chan imageHolder, out chan<- int) { // read from channel and write to disk
	for imageH := range in {
		// creating new filename adding the thumb keyword to indicate it is a thumbnaile
		out, err := os.Create(renameFile(imageH.name))
		defer out.Close()
		if err != nil {
			log.Fatal(err)
		}
		jpeg.Encode(out, imageH.data, nil) // write with default parameters (nil)
	}
	close(out)
}

// initialization
func init() {

	// in newest versions of golang it is no longer needed to indicate the maximum
	// number of cpu's that should be used
	// runtime.GOMAXPROCS(runtime.NumCPU()) // runs with max number of cpus available

}

// test pipeline without writing to file
func main() {
	// channels for the pipeline
	decoded := make(chan imageHolder)
	thumbnailed := make(chan imageHolder)
	done := make(chan int)

	// initiate with list of filenames
	filenames, err := filepath.Glob("./pictures/*")
	if err != nil {
		log.Fatal(err)
	}
	// filenames := []string{"./pictures/green1.jpg", "./pictures/stained_green.jpg"} // test slice, should be taken from argv

	go decoder(filenames, decoded)
	go thumbnailer(decoded, thumbnailed)
	go writer(thumbnailed, done)
	<-done // wait until done is finished
}
