package main

import (
	"flag"
	"fmt"

	"github.concur.com/mparks/adder/circuit"
)

var actionType = flag.String("action", "add", "Type of action to take against input(s) (e.g. add/sub/comp)")
var bitString1 = flag.String("bits1", "00000000", "First string of bits in an action (e.g. 11110000)")
var bitString2 = flag.String("bits2", "00000000", "Second string of bits in an action that takes two inputs (e.g. 00001111)")

func main() {
	flag.Parse()

	executeAdder()
}

func executeAdder() {

	switch *actionType {
	case "add":
		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		switches2, err := circuit.NewNSwitchBank(*bitString2)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		addr, err := circuit.NewNBitAdder(switches1.AsPwrEmitters(), switches2.AsPwrEmitters(), nil)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		var carry string
		if addr.CarryOutAsBool() {
			carry = "1"
		}
		fmt.Printf("%18s\n+%17s\n=%1s%16s\n\n", *bitString1, *bitString2, carry, addr.AsAnswerString())
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
