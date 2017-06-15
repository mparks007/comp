package main

import (
	"flag"
	"fmt"

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
			switches1, err := circuit.NewEightSwitchBank(*bitString1)
			if err != nil {
				fmt.Println("Error:" + err.Error())
				return
			}
			switches2, err := circuit.NewEightSwitchBank(*bitString2)
			if err != nil {
				fmt.Println("Error:" + err.Error())
				return
			}

			a8 := circuit.NewEightBitAdder(switches1.AsPwrEmitters(), switches2.AsPwrEmitters(), nil)
			var carry string
			if a8.CarryOutAsBool() {
				carry = "1"
			}
			fmt.Printf("%10s\n+%9s\n=%1s%8s\n\n", *bitString1, *bitString2, carry, a8.AsAnswerString())
		case 16:
			switches1, err := circuit.NewSixteenSwitchBank(*bitString1)
			if err != nil {
				fmt.Println("Error:" + err.Error())
				return
			}
			switches2, err := circuit.NewSixteenSwitchBank(*bitString2)
			if err != nil {
				fmt.Println("Error:" + err.Error())
				return
			}

			a16 := circuit.NewSixteenBitAdder(switches1.AsPwrEmitters(), switches2.AsPwrEmitters(), nil)
			var carry string
			if a16.CarryOutAsBool() {
				carry = "1"
			}
			fmt.Printf("%18s\n+%17s\n=%1s%16s\n\n", *bitString1, *bitString2, carry, a16.AsAnswerString())
		}
	case "comp":
		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		c := circuit.NewOnesComplementer(switches1.AsPwrEmitters(), circuit.NewBattery())
		fmt.Println("   Input String: " + *bitString1)
		fmt.Println("Ones Complement: " + c.AsComplementString())
	}
}
