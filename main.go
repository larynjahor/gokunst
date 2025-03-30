package main

import "log"

func main() {
	if err := wrapMain(); err != nil {
		log.Fatalln(err)
	}
}

func wrapMain() error {
	return nil
}
