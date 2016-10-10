// Copyright 2015 Lars Wiegman

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"text/template"

	"github.com/namsral/microdata"
)

var fnmap = template.FuncMap{
	"jsonMarshal": jsonMarshal,
}

func main() {
	var data *microdata.Microdata
	var err error

	baseURL := flag.String("base-url", "http://example.com", "base url to use for the data in the stdin stream.")
	contentType := flag.String("content-type", "", "content type of the data in the stdin stream.")
	format := flag.String("format", "{{. |jsonMarshal }}", `alternate format for the output of the
	microdata, using the syntax of package html/template. The default output is
	equivalent to -f '{{. |jsonMarshal }}'. The struct being passed to the
	template is:
		
		type Microdata struct
			Items []*Item 'json:"items"'
		}

		type Item struct {
			Types      []string    'json:"type"'
			Properties PropertyMap 'json:"properties"'
			Id         string      'json:"id,omitempty"'
		}

		type PropertyMap map[string]ValueList
		
		type ValueList []interface{}

	The template function "jsonMarshal" calls json.Marshal
`)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] [url]:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\nExtract the HTML Microdata from a HTML5 document. Format to JSON or using the syntax of package html/template.")
		fmt.Fprint(os.Stderr, " Provide an URL to a valid HTML5 document or stream a valid HTML5 document through stdin.\n")
	}

	flag.Parse()

	// Fetch and parse microdata
	switch len(flag.Args()) {
	case 0:
		u, err := url.Parse(*baseURL)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data, err = microdata.ParseHTML(os.Stdin, *contentType, u)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		data, err = microdata.ParseURL(flag.Args()[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	t := template.Must(template.New("format").Funcs(fnmap).Parse(*format))
	if err := t.Execute(os.Stdout, data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// jsonMarshal encodes the given data to JSON.
func jsonMarshal(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
