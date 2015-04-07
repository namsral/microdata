// Copyright 2015 Lars Wiegman

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/namsral/microdata"
)

func main() {
	var data *microdata.Microdata
	var err error

	baseURL := flag.String("base-url", "http://example.com", "base url to use for the data in the stdin stream.")
	contentType := flag.String("content-type", "", "content type of the data in the stdin stream.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] [url]:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\nExtract the HTML Microdata from a HTML5 document.")
		fmt.Fprint(os.Stderr, " Provide an URL to a valid HTML5 document or stream a valid HTML5 document through stdin.\n")
	}

	flag.Parse()

	// Args
	if args := flag.Args(); len(args) > 0 {
		urlStr := args[0]
		data, err = microdata.ParseURL(urlStr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		printResult(os.Stdout, data)
		return
	}

	// Stdin
	r := os.Stdin
	u, _ := url.Parse(*baseURL)
	data, err = microdata.ParseHTML(r, *contentType, u)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printResult(os.Stdout, data)
}

// printResult pretty formats and prints the given items in a JSON object.
func printResult(w io.Writer, data *microdata.Microdata) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	w.Write(b)
	w.Write([]byte("\n"))
}
