package main

import (
	"fmt"

	"flag"

	"github.concur.com/mparks/adder/circuit"
)

var bitLength = flag.Int("bitLen", 8, "The number of bits to parse (8 or 16)")
var bitString1 = flag.String("bits1", "00000000", "First string of bits in the calculation (e.g. 11110000)")
var bitString2 = flag.String("bits2", "00000000", "Second string of bits in the calculation (e.g. 00001111)")
var calcType = flag.String("calc", "add", "Type of calculation to perform (e.g. add")

func main() {
	flag.Parse()

	executeAdder()
}

func executeAdder() {
	switch *bitLength {
	case 8:
		switch *calcType {
		case "add":
			a8, err := circuit.NewEightBitAdder(*bitString1, *bitString2, nil)
			if err != nil {
				fmt.Println("Error:" + err.Error())
			} else {
				fmt.Printf("%10s\n+%9s\n=%9s\n\n", *bitString1, *bitString2, a8)
			}
		}
	case 16:
		switch *calcType {
		case "add":
			a16, err := circuit.NewSixteenBitAdder(*bitString1, *bitString2, nil)
			if err != nil {
				fmt.Println("Error:" + err.Error())
			} else {
				fmt.Printf("%18s\n+%17s\n=%17s\n\n", *bitString1, *bitString2, a16)
			}
		}
	}
}
