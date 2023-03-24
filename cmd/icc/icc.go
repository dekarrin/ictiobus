package main

import "github.com/dekarrin/ictiobus/fishi"

func main() {

	err := fishi.ExecuteMarkdownFile("fishi.md")
	if err != nil {
		panic(err)
	}

}
