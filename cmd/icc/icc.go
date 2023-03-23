package main

import "github.com/dekarrin/ictiobus/fishi"

func main() {

	err := fishi.ReadFishiMdFile("fishi.md", false)
	if err != nil {
		panic(err)
	}

}
