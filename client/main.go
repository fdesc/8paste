package main

import (
	"flag"
	"fmt"
)

func main() {
	var id   string
	var path string
	var password string
	var title string
	var duration string
	var data string
	var target string

	flag.Usage = func() {
		fmt.Println("pasteur-client: Usage")
		fmt.Println("  -u <content>:      Upload the specified content")
		fmt.Println("  -g <ID>:           Get the content from ID")
		fmt.Println("  -f <path>:         Specify a path to a file for uploading")
		fmt.Println("  -p <password>:     Set a password for uploaded content")
		fmt.Println("  -t <title>:        Set title for the content")
		fmt.Println("  -d x(h) y(m) z(s): Set duration for the uploaded content as in x hours y minutes z seconds")
		fmt.Println("  -source <URL>:     Set the source to get/upload contents from")
	}


	flag.StringVar(&data,"u","<content>","Upload the specified content")
	flag.StringVar(&id,"g","<ID>","Get the content from ID")
	flag.StringVar(&path,"f","<path>","Upload a file from specified path")
	flag.StringVar(&password,"p","<password>","Set password for the uploaded content")
	flag.StringVar(&title,"t","<title>","Set title for the uploaded content")
	flag.StringVar(&duration,"d","x(h) y(m) z(s)","Set duration for the uploaded content as in x hours, y minutes and z seconds")
	flag.StringVar(&target,"source","<URL>","Set the source to get/upload contents from")

	flag.Parse()
}
