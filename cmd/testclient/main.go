package main

import (
	"flag"
	"fmt"

	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"strconv"

	"github.concur.com/mparks/adder/circuit"
)

var actionType = flag.String("action", "", "Type of action to take (e.g. battery/switch/relay/and/or/nand/nor/xor/invert/osc/select/halfadd/fulladd/add/3add/sub/comp/flipflop/levellatch/nbitlatch/edgelatch/freqdiv)")
var aIn = flag.String("aIn", "", "Power indicator at pin A (1 or 0)")
var bIn = flag.String("bIn", "", "Power indicator at pin B (1 or 0)")
var cIn = flag.String("cIn", "", "Power indicator at pin C (1 or 0)")
var bitString1 = flag.String("bits1", "00000000", "First string of bits in an action (e.g. 11110000)")
var bitString2 = flag.String("bits2", "00000000", "Second string of bits in an action (e.g. 00001111)")
var bitString3 = flag.String("bits3", "00000000", "Third string of bits in an action (e.g. 10101010)")

func main() {
	flag.Parse()

	_, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		fmt.Println("Error:" + err.Error())
		return
	}

	execute()
}

func execute() {

	switch *actionType {
	case "battery":

		fmt.Println("Creating battery")
		batt := circuit.NewBattery()
		batt.WireUp(func(state bool) {
			fmt.Printf("Is Charged?: %v\n", batt.GetIsPowered())
		})
		fmt.Println("Discharging battery")
		batt.Discharge()
		fmt.Println("Recharging battery")
		batt.Charge()
		fmt.Println("Recharging charged battery (no affect)")
		batt.Charge()

	case "switch":

		state, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}

		fmt.Println("Creating switch")
		sw := circuit.NewSwitch(state)
		sw.WireUp(func(state bool) {
			fmt.Printf("Is On?: %v\n", sw.GetIsPowered())
		})
		if sw.GetIsPowered() {
			fmt.Println("Turning switch off")
			sw.Set(false)
		}

		fmt.Println("Turning switch on")
		sw.Set(true)
		fmt.Println("Turning switch on again (no affect)")
		sw.Set(true)

	case "relay":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}

		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		relay := circuit.NewRelay(sw1, sw2)

		fmt.Printf("Open circuit power?: %v\n", relay.OpenOut.GetIsPowered())
		fmt.Printf("Closed circuit power?: %v\n", relay.ClosedOut.GetIsPowered())

	case "and":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)

		var and *circuit.ANDGate = nil
		if len(*cIn) > 0 {

			c, err := strconv.ParseBool(*cIn)
			if err != nil {
				fmt.Println("Invalid value for boolean power state (cIn):" + *cIn)
				return
			}
			sw3 := circuit.NewSwitch(c)

			and = circuit.NewANDGate(sw1, sw2, sw3)
		} else {
			and = circuit.NewANDGate(sw1, sw2)
		}

		fmt.Printf("AND Gate outputting power?: %v\n", and.GetIsPowered())

	case "or":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)

		var or *circuit.ORGate = nil
		if len(*cIn) > 0 {

			c, err := strconv.ParseBool(*cIn)
			if err != nil {
				fmt.Println("Invalid value for boolean power state (cIn):" + *cIn)
				return
			}
			sw3 := circuit.NewSwitch(c)

			or = circuit.NewORGate(sw1, sw2, sw3)
		} else {
			or = circuit.NewORGate(sw1, sw2)
		}

		fmt.Printf("OR Gate outputting power?: %v\n", or.GetIsPowered())

	case "nand":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)

		var nand *circuit.NANDGate = nil
		if len(*cIn) > 0 {

			c, err := strconv.ParseBool(*cIn)
			if err != nil {
				fmt.Println("Invalid value for boolean power state (cIn):" + *cIn)
				return
			}
			sw3 := circuit.NewSwitch(c)

			nand = circuit.NewNANDGate(sw1, sw2, sw3)
		} else {
			nand = circuit.NewNANDGate(sw1, sw2)
		}

		fmt.Printf("NAND Gate outputting power?: %v\n", nand.GetIsPowered())

	case "nor":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)

		var nor *circuit.NORGate = nil
		if len(*cIn) > 0 {

			c, err := strconv.ParseBool(*cIn)
			if err != nil {
				fmt.Println("Invalid value for boolean power state (cIn):" + *cIn)
				return
			}
			sw3 := circuit.NewSwitch(c)

			nor = circuit.NewNORGate(sw1, sw2, sw3)
		} else {
			nor = circuit.NewNORGate(sw1, sw2)
		}

		fmt.Printf("NOR Gate outputting power?: %v\n", nor.GetIsPowered())

	case "xor":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		xor := circuit.NewXORGate(sw1, sw2)

		fmt.Printf("XOR Gate outputting power?: %v\n", xor.GetIsPowered())

	case "xnor":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		xnor := circuit.NewXNORGate(sw1, sw2)

		fmt.Printf("XNOR Gate outputting power?: %v\n", xnor.GetIsPowered())

	case "invert":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		inv := circuit.NewInverter(sw1)

		fmt.Printf("Inverted output power: %v\n", inv.GetIsPowered())

	case "select":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for boolean power state (aIn): " + *aIn)
			return
		}
		switchesA, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		switchesB, err := circuit.NewNSwitchBank(*bitString2)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		selB := circuit.NewSwitch(a)
		sel, err := circuit.NewTwoToOneSelector(selB, switchesA.AsPwrEmitters(), switchesB.AsPwrEmitters())
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		answer := ""
		for _, o := range sel.Outs {

			if o.(*circuit.ORGate).GetIsPowered() {
				answer += "1"
			} else {
				answer += "0"
			}
		}

		fmt.Printf("Selector outputting: %s\n", answer)

	case "halfadd":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for bit (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for bit (bIn): " + *bIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		hAdd := circuit.NewHalfAdder(sw1, sw2)

		sum := "0"
		if hAdd.Sum.(*circuit.XORGate).GetIsPowered() {
			sum = "1"
		}
		carry := "0"
		if hAdd.Carry.(*circuit.ANDGate).GetIsPowered() {
			carry = "1"
		}
		fmt.Printf("Sum: %s\n", sum)
		fmt.Printf("Carry: %s\n", carry)

	case "fulladd":

		a, err := strconv.ParseBool(*aIn)
		if err != nil {
			fmt.Println("Invalid value for bit (aIn): " + *aIn)
			return
		}
		b, err := strconv.ParseBool(*bIn)
		if err != nil {
			fmt.Println("Invalid value for bit (bIn): " + *bIn)
			return
		}
		c, err := strconv.ParseBool(*cIn)
		if err != nil {
			fmt.Println("Invalid value for carry-in bit (cIn):" + *cIn)
			return
		}
		sw1 := circuit.NewSwitch(a)
		sw2 := circuit.NewSwitch(b)
		sw3 := circuit.NewSwitch(c)
		fAdd := circuit.NewFullAdder(sw1, sw2, sw3)

		sum := "0"
		if fAdd.Sum.(*circuit.XORGate).GetIsPowered() {
			sum = "1"
		}
		carry := "0"
		if fAdd.Carry.(*circuit.ORGate).GetIsPowered() {
			carry = "1"
		}
		fmt.Printf("Sum: %s\n", sum)
		fmt.Printf("Carry: %s\n", carry)

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
		fmt.Printf("  %s\n+ %s\n=%1s%s\n\n", *bitString1, *bitString2, carry, addr.AsAnswerString())

	case "3add":

		switchesA, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		switchesB, err := circuit.NewNSwitchBank(*bitString2)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}
		if len(*bitString3) != len(*bitString2) {
			fmt.Printf("Final addend length doesn't match other addends: %s\n", *bitString3)
			return
		}
		_, err = circuit.NewNSwitchBank(*bitString3)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		addr, err := circuit.NewThreeNumberAdder(switchesA, switchesB)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		addr.SaveToLatch.Set(true)
		addr.SaveToLatch.Set(false)
		addr.ReadFromLatch.Set(true)

		for i, b := range *bitString3 {
			switchesA.Switches[i].Set(b == '1')
		}

		var carry string
		if addr.CarryOutAsBool() {
			carry = "1"
		}

		fmt.Printf("  %s\n+ %s\n+ %s\n=%1s%s\n\n", *bitString1, *bitString2, *bitString3, carry, addr.AsAnswerString())

	case "sub":

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

		subtr, err := circuit.NewNBitSubtractor(switches1.AsPwrEmitters(), switches2.AsPwrEmitters())
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		fmt.Printf("  %s\n- %s\n= %s\n\n", *bitString1, *bitString2, subtr.AsAnswerString())

	case "comp":

		switches1, err := circuit.NewNSwitchBank(*bitString1)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		}

		c := circuit.NewOnesComplementer(switches1.AsPwrEmitters(), circuit.NewBattery())
		fmt.Println("   Input String: " + *bitString1)
		fmt.Println("Ones Complement: " + c.AsComplementString())

		switches, _ := circuit.NewNSwitchBank(strings.Repeat("0", len(c.Complements)-1) + "1")
		twosC, err := circuit.NewNBitAdder(c.Complements, switches.AsPwrEmitters(), nil)
		if err != nil {
			fmt.Println("Error:" + err.Error())
			return
		} else {
			fmt.Println("Twos Complement: " + twosC.AsAnswerString())
		}

	}
}
