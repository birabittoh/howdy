package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/nishanths/go-xkcd/v2"
)

type Config struct {
	directory string
	file      string
	id        int
	braille   bool
	symbols   string
	grayscale bool
	width     int
	height    int
	help      bool
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.directory, "d", "", "serve a random image from the specified directory")
	flag.StringVar(&config.directory, "directory", "", "serve a random image from the specified directory")

	flag.StringVar(&config.file, "f", "", "convert a given file")
	flag.StringVar(&config.file, "file", "", "convert a given file")

	flag.IntVar(&config.id, "i", 0, "fetch and convert XKCD comic with the given ID")
	flag.IntVar(&config.id, "id", 0, "fetch and convert XKCD comic with the given ID")

	flag.BoolVar(&config.braille, "b", false, "use braille characters for higher resolution")
	flag.BoolVar(&config.braille, "braille", false, "use braille characters for higher resolution")

	flag.StringVar(&config.symbols, "s", "", "string with allowed symbols")
	flag.StringVar(&config.symbols, "symbols", "", "string with allowed symbols")

	flag.BoolVar(&config.grayscale, "g", false, "disable color output")
	flag.BoolVar(&config.grayscale, "grayscale", false, "disable color output")

	flag.IntVar(&config.width, "w", 0, "set the image width")
	flag.IntVar(&config.width, "width", 0, "set the image width")

	flag.IntVar(&config.height, "h", 0, "set the image height")
	flag.IntVar(&config.height, "height", 0, "set the image height")

	flag.BoolVar(&config.help, "help", false, "print this help text")

	flag.Parse()

	return config
}

func printHelp() {
	fmt.Println("howdy - Display XKCD comics or your own images as ASCII art in the terminal")
	fmt.Println("Usage: howdy [OPTIONS]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -d, --directory DIR    Serve a random image from the specified directory")
	fmt.Println("  -f, --file FILE        Convert a given file")
	fmt.Println("  -i, --id ID            Fetch and convert XKCD comic with the given ID")
	fmt.Println("  -b, --braille          Use braille characters for higher resolution")
	fmt.Println("  -s, --symbols STRING   String with allowed symbols")
	fmt.Println("  -g, --grayscale        Disable color output")
	fmt.Println("  -w, --width WIDTH      Set the image width")
	fmt.Println("  -h, --height HEIGHT    Set the image height")
	fmt.Println("      --help             Print this help text")
	fmt.Println()
	fmt.Println("If no options are provided, howdy fetches the latest XKCD comic by default.")
}

func getLatestComic() (path string, err error) {
	client := xkcd.NewClient()
	ctx := context.Background()
	comic, err := client.Latest(ctx)
	if err != nil {
		return "", err
	}
	return comic.ImageURL, nil
}

func getComic(id int) (path string, err error) {
	client := xkcd.NewClient()
	ctx := context.Background()
	comic, err := client.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return comic.ImageURL, nil
}

func getRandomImageFromDir(dir string) (string, error) {
	// Support common image formats
	patterns := []string{"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.webp"}
	var images []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		images = append(images, matches...)
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no images found in %s", dir)
	}

	randomIndex := make([]byte, 1)
	_, err := rand.Read(randomIndex)
	if err != nil {
		return "", err
	}

	return images[int(randomIndex[0])%len(images)], nil
}

func convertToASCII(imagePath string, config *Config) (string, error) {
	flags := aic_package.DefaultFlags()
	flags.Dither = true

	// Set dimensions
	if config.width > 0 {
		flags.Width = config.width
	}
	if config.height > 0 {
		flags.Height = config.height
	}

	if config.width == 0 && config.height == 0 {
		flags.Full = true
	}

	// Set color/grayscale
	flags.Colored = !config.grayscale
	flags.Grayscale = config.grayscale
	flags.Dither = true
	flags.Braille = false
	flags.Complex = true

	// Set symbols vs braille
	if config.symbols != "" {
		flags.CustomMap = config.symbols
	} else if config.braille {
		flags.Braille = true
	}

	return aic_package.Convert(imagePath, flags)
}

func main() {
	config := parseFlags()

	if config.help {
		printHelp()
		return
	}

	var imagePath string
	var err error

	// Determine image source based on flags
	switch {
	case config.file != "":
		imagePath = config.file

	case config.directory != "":
		imagePath, err = getRandomImageFromDir(config.directory)
		if err != nil {
			fmt.Printf("Error getting random image from directory: %v\n", err)
			os.Exit(1)
		}

	case config.id != 0:
		imagePath, err = getComic(config.id)
		if err != nil {
			fmt.Printf("Error fetching comic with ID %d: %v\n", config.id, err)
			os.Exit(1)
		}

	default:
		imagePath, err = getLatestComic()
		if err != nil {
			fmt.Printf("Error fetching comic: %v\n", err)
			os.Exit(1)
		}
	}

	asciiArt, err := convertToASCII(imagePath, config)
	if err != nil {
		fmt.Printf("Error converting image to ASCII: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n\n%s\n", asciiArt, imagePath)
}
