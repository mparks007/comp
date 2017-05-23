package main

import (
	"fmt"

	"flag"

	"github.concur.com/mparks/adder/circuit"
)

var actionType = flag.String("action", "add", "Type of action to take against input(s) (e.g. add/sub/comp)")
var bitLength = flag.Int("bitLen", 8, "The number of bits in math actions (8 or 16)")
var bitString1 = flag.String("bits1", "00000000", "First string of bits in an action (e.g. 11110000)")
var bitString2 = flag.String("bits2", "00000000", "Second string of bits in an action that takes two inputs (e.g. 00001111)")

func main() {
	flag.Parse()

	executeAdder()
}

func executeAdder() {
	switch *actionType {
	case "add":
		switch *bitLength {
		case 8:
			a8, err := circuit.NewEightBitAdder(*bitString1, *bitString2, nil)
			if err != nil {
				fmt.Println("Error:" + err.Error())
			} else {
				fmt.Printf("%10s\n+%9s\n=%9s\n\n", *bitString1, *bitString2, a8)
			}
		case 16:
			a16, err := circuit.NewSixteenBitAdder(*bitString1, *bitString2, nil)
			if err != nil {
				fmt.Println("Error:" + err.Error())
			} else {
				fmt.Printf("%18s\n+%17s\n=%17s\n\n", *bitString1, *bitString2, a16)
			}
		}
	case "comp":
		c, err := circuit.NewOnesComplementer([]byte(*bitString1), &circuit.Battery{})
		if err != nil {
			fmt.Println("Error:" + err.Error())
		} else {
			fmt.Println("   Input String: " + *bitString1)
			fmt.Println("Ones Complement: " + c.Complement())
		}
	}
}
