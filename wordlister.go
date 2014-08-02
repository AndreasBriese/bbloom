package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	templ := "2014/%02d/%02d %02d:%02d:%02d /info.html\n"
	file, err := os.OpenFile("words.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for M := 1; M < 13; M++ {
		for D := 1; D < 31; D++ {
			for h := 0; h < 24; h++ {
				for m := 0; m < 60; m++ {
					for s := 0; s < 60; s++ {
						fmt.Fprintf(file, templ, M, D, h, m, s)
					}
				}
			}
		}
	}
	file.Close()
}
