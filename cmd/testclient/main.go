package main

import (
	"flag"
	"fmt"

	"strings"

	"strconv"

	"context"

	"time"

	"github.concur.com/mparks/adder/circuit"
)

var actionType = flag.String("action", "", "Type of action to take (battery/switch/relay/and/or/nand/nor/xor/invert/osc/select/halfadd/fulladd/add/3add/sub/comp/flipflop/levellatch/nbitlatch/freqdiv)")
var aIn = flag.String("aIn", "", "Power/bit indicator at pin A (1 or 0)")
var bIn = flag.String("bIn", "", "Power/bit indicator at pin B (1 or 0)")
var cIn = flag.String("cIn", "", "Power/bit indicator at pin C (1 or 0)")
var bitString1 = flag.String("bits1", "", "First string of bits in an action (e.g. 11110000)")
var bitString2 = flag.String("bits2", "", "Second string of bits in an action (e.g. 00001111)")
var bitString3 = flag.String("bits3", "", "Third string of bits in an action (e.g. 10101010)")

var logger circuit.MySqlLogger

func main() {
	flag.Parse()

	l, err := circuit.NewMySqlLogger("mparks:dbadmin@/circuit", context.Background())
	if err != nil {
		fmt.Println("Error creating MySqlLogger:", err.Error())
		return
	}

	logger = *l
	execute()
}

func execute() {

	switch *actionType {
	case "battery":

		charge, err := parseBool(*aIn, "aIn", "battery power state")
		if err != nil {
			return
		}

		fmt.Println("Creating battery")
		batt := circuit.NewBattery()
		batt.WireUp(func(state bool) {
			fmt.Printf("Is Charged?: %v\n", state)
		})

		if charge {
			fmt.Println("Recharging charged battery (no affect)")
			batt.Charge()
		} else {
			fmt.Println("Discharging battery")
			batt.Discharge()
		}

	case "switch":

		state, err := parseBool(*aIn, "aIn", "initial switch state")
		if err != nil {
			return
		}

		fmt.Println("Creating switch")
		sw := circuit.NewSwitch(state)
		sw.WireUp(func(state bool) {
			fmt.Printf("Is On?: %v\n", state)
		})

	case "relay":

		upperIn, err := parseBool(*aIn, "aIn", "relay upper pin")
		if err != nil {
			return
		}
		lowerIn, err := parseBool(*bIn, "bIn", "relay lower/electromagnet pin")
		if err != nil {
			return
		}

		relay := circuit.NewRelay(circuit.NewSwitch(upperIn), circuit.NewSwitch(lowerIn))

		fmt.Printf("Open circuit power?: %v\n", relay.OpenOut.GetIsPowered())
		fmt.Printf("Closed circuit power?: %v\n", relay.ClosedOut.GetIsPowered())

	case "and":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)

		var and *circuit.ANDGate = nil
		if len(*cIn) > 0 {

			cPin, err := parseBool(*cIn, "cIn", "gate input pin")
			if err != nil {
				return
			}
			cSwitch := circuit.NewSwitch(cPin)
			and = circuit.NewANDGate(aSwitch, bSwitch, cSwitch)
		} else {
			and = circuit.NewANDGate(aSwitch, bSwitch)
		}

		fmt.Printf("AND Gate outputting power?: %v\n", and.GetIsPowered())

	case "or":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)

		var or *circuit.ORGate = nil
		if len(*cIn) > 0 {

			cPin, err := parseBool(*cIn, "cIn", "gate input pin")
			if err != nil {
				return
			}
			cSwitch := circuit.NewSwitch(cPin)
			or = circuit.NewORGate(aSwitch, bSwitch, cSwitch)
		} else {
			or = circuit.NewORGate(aSwitch, bSwitch)
		}

		fmt.Printf("OR Gate outputting power?: %v\n", or.GetIsPowered())

	case "nand":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)

		var nand *circuit.NANDGate = nil
		if len(*cIn) > 0 {

			cPin, err := parseBool(*cIn, "cIn", "gate input pin")
			if err != nil {
				return
			}
			cSwitch := circuit.NewSwitch(cPin)
			nand = circuit.NewNANDGate(aSwitch, bSwitch, cSwitch)
		} else {
			nand = circuit.NewNANDGate(aSwitch, bSwitch)
		}

		fmt.Printf("NAND Gate outputting power?: %v\n", nand.GetIsPowered())

	case "nor":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)

		var nor *circuit.NORGate = nil
		if len(*cIn) > 0 {

			cPin, err := parseBool(*cIn, "cIn", "gate input pin")
			if err != nil {
				return
			}
			cSwitch := circuit.NewSwitch(cPin)
			nor = circuit.NewNORGate(aSwitch, bSwitch, cSwitch)
		} else {
			nor = circuit.NewNORGate(aSwitch, bSwitch)
		}

		fmt.Printf("NOR Gate outputting power?: %v\n", nor.GetIsPowered())

	case "xor":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)
		xor := circuit.NewXORGate(aSwitch, bSwitch)

		fmt.Printf("XOR Gate outputting power?: %v\n", xor.GetIsPowered())

	case "xnor":

		aPin, err := parseBool(*aIn, "aIn", "gate input pin")
		if err != nil {
			return
		}
		bPin, err := parseBool(*bIn, "bIn", "gate input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		bSwitch := circuit.NewSwitch(bPin)
		xnor := circuit.NewXNORGate(aSwitch, bSwitch)

		fmt.Printf("XNOR Gate outputting power?: %v\n", xnor.GetIsPowered())

	case "invert":

		aPin, err := parseBool(*aIn, "aIn", "inverter input pin")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aPin)
		inv := circuit.NewInverter(aSwitch)

		fmt.Printf("Inverted outputting power?: %v\n", inv.GetIsPowered())

	case "select":

		selState, err := parseBool(*aIn, "aIn", "selector state")
		if err != nil {
			return
		}
		aSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "selector input")
		if err != nil {
			return
		}
		bSwitches, err := parseStringToSwitchbank(*bitString2, "bits2", "selector input")
		if err != nil {
			return
		}
		selectB := circuit.NewSwitch(selState)
		sel, err := circuit.NewTwoToOneSelector(selectB, aSwitches.AsPwrEmitters(), bSwitches.AsPwrEmitters())
		if err != nil {
			fmt.Println("Error: " + err.Error())
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

		aBit, err := parseBool(*aIn, "aIn", "bit value")
		if err != nil {
			return
		}
		bBit, err := parseBool(*bIn, "bIn", "bit value")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aBit)
		bSwitch := circuit.NewSwitch(bBit)
		halfAdd := circuit.NewHalfAdder(aSwitch, bSwitch)

		sum := "0"
		if halfAdd.Sum.(*circuit.XORGate).GetIsPowered() {
			sum = "1"
		}
		carry := "0"
		if halfAdd.Carry.(*circuit.ANDGate).GetIsPowered() {
			carry = "1"
		}
		fmt.Printf("Sum: %s\n", sum)
		fmt.Printf("Carry: %s\n", carry)

	case "fulladd":

		aBit, err := parseBool(*aIn, "aIn", "bit value")
		if err != nil {
			return
		}
		bBit, err := parseBool(*bIn, "bIn", "bit value")
		if err != nil {
			return
		}
		cBit, err := parseBool(*cIn, "cIn", "carry-in bit value")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(aBit)
		bSwitch := circuit.NewSwitch(bBit)
		cSwitch := circuit.NewSwitch(cBit)
		fullAdd := circuit.NewFullAdder(aSwitch, bSwitch, cSwitch)

		sum := "0"
		if fullAdd.Sum.(*circuit.XORGate).GetIsPowered() {
			sum = "1"
		}
		carry := "0"
		if fullAdd.Carry.(*circuit.ORGate).GetIsPowered() {
			carry = "1"
		}
		fmt.Printf("Sum: %s\n", sum)
		fmt.Printf("Carry: %s\n", carry)

	case "add":

		aSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "addend bits")
		if err != nil {
			return
		}
		bSwitches, err := parseStringToSwitchbank(*bitString2, "bits2", "addend bits")
		if err != nil {
			return
		}
		carryIn := false
		if len(*aIn) > 0 {
			carryIn, err = parseBool(*aIn, "aIn", "carry-in bit value")
			if err != nil {
				return
			}
		}
		cSwitch := circuit.NewSwitch(carryIn)

		addr, err := circuit.NewNBitAdder(aSwitches.AsPwrEmitters(), bSwitches.AsPwrEmitters(), cSwitch)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return
		}

		var carry string
		if addr.CarryOutAsBool() {
			carry = "1"
		}
		fmt.Printf("  %s\n+ %s\n=%1s%s\n", *bitString1, *bitString2, carry, addr.AsAnswerString())
		logger.Log("testcat", "testdata", context.Background())

	case "3add":

		aSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "addend bits")
		if err != nil {
			return
		}
		bSwitches, err := parseStringToSwitchbank(*bitString2, "bits2", "addend bits")
		if err != nil {
			return
		}
		_, err = parseStringToSwitchbank(*bitString3, "bits3", "addend bits")
		if err != nil {
			return
		}
		if len(*bitString1) != len(*bitString3) {
			fmt.Printf("Third addend's length is invalid: %s\n", *bitString3)
			return
		}

		addr, err := circuit.NewThreeNumberAdder(aSwitches, bSwitches)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return
		}

		addr.SaveToLatch.Set(true)
		addr.SaveToLatch.Set(false)
		addr.ReadFromLatch.Set(true)

		for i, b := range *bitString3 {
			aSwitches.Switches[i].Set(b == '1')
		}

		var carry string
		if addr.CarryOutAsBool() {
			carry = "1"
		}

		fmt.Printf("  %s\n+ %s\n+ %s\n=%1s%s\n", *bitString1, *bitString2, *bitString3, carry, addr.AsAnswerString())

	case "sub":

		aSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "minuend bits")
		if err != nil {
			return
		}
		bSwitches, err := parseStringToSwitchbank(*bitString2, "bits2", "subtrahend bits")
		if err != nil {
			return
		}

		subtr, err := circuit.NewNBitSubtractor(aSwitches.AsPwrEmitters(), bSwitches.AsPwrEmitters())
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return
		}

		fmt.Printf("  %s\n- %s\n= %s\n", *bitString1, *bitString2, subtr.AsAnswerString())

	case "comp":

		aSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "bits")
		if err != nil {
			return
		}

		onesComp := circuit.NewOnesComplementer(aSwitches.AsPwrEmitters(), circuit.NewBattery())
		fmt.Println("   Input String: " + *bitString1)
		fmt.Println("Ones Complement: " + onesComp.AsComplementString())

		switches, _ := circuit.NewNSwitchBank(strings.Repeat("0", len(onesComp.Complements)-1) + "1")
		twosComp, _ := circuit.NewNBitAdder(onesComp.Complements, switches.AsPwrEmitters(), nil)
		fmt.Println("Twos Complement: " + twosComp.AsAnswerString())

	case "flipflop":
		rBit, err := parseBool(*aIn, "aIn", "Reset bit value")
		if err != nil {
			return
		}
		sBit, err := parseBool(*bIn, "bIn", "Set bit value")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(rBit)
		bSwitch := circuit.NewSwitch(sBit)
		flipflop := circuit.NewRSFlipFLop(aSwitch, bSwitch)

		fmt.Printf("Q:    %v\n", flipflop.Q.GetIsPowered())
		fmt.Printf("QBar: %v\n", flipflop.QBar.GetIsPowered())

	case "levellatch":
		clockBit, err := parseBool(*aIn, "aIn", "Clock bit value")
		if err != nil {
			return
		}
		dataBit, err := parseBool(*bIn, "bIn", "Data bit value")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(clockBit)
		bSwitch := circuit.NewSwitch(dataBit)
		latch := circuit.NewLevelTriggeredDTypeLatch(aSwitch, bSwitch)

		fmt.Printf("Q:    %v\n", latch.Q.GetIsPowered())
		fmt.Printf("QBar: %v\n", latch.QBar.GetIsPowered())

	case "nbitlatch":
		clockBit, err := parseBool(*aIn, "aIn", "Clock bit value")
		if err != nil {
			return
		}
		dataSwitches, err := parseStringToSwitchbank(*bitString1, "bits1", "data bits")
		if err != nil {
			return
		}
		aSwitch := circuit.NewSwitch(clockBit)
		latch := circuit.NewNBitLatch(aSwitch, dataSwitches.AsPwrEmitters())

		output := ""
		for _, q := range latch.Qs {

			if q.(*circuit.NORGate).GetIsPowered() {
				output += "1"
			} else {
				output += "0"
			}
		}

		fmt.Printf("Latch state: %s\n", output)

	case "freqdiv":
		hertz, err := parseInt(*aIn, "aIn", "hertz value")
		if err != nil {
			return
		}
		secs, err := parseInt(*bIn, "aIn", "duration (seconds) value")
		if err != nil {
			return
		}
		var init = false
		if len(*cIn) > 0 {

			init, err = parseBool(*cIn, "cIn", "initial oscillator state")
			if err != nil {
				return
			}
		}

		osc := circuit.NewOscillator(init)
		if osc.GetIsPowered() {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}

		freqDiv := circuit.NewFrequencyDivider(osc)

		freqDiv.QBar.WireUp(func(state bool) {
			if state {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		})

		osc.Oscillate(hertz)

		time.Sleep(time.Second * time.Duration(secs))

		osc.Stop()
	}
}

func parseBool(paramVal, paramName, context string) (bool, error) {

	retVal, err := strconv.ParseBool(paramVal)
	if err != nil {
		fmt.Printf("Value \"%s\" for parameter %s is invalid for %s\n", paramVal, paramName, context)
		return false, err
	}
	return retVal, nil
}

func parseInt(paramVal, paramName, context string) (int, error) {

	retVal, err := strconv.Atoi(paramVal)
	if err != nil {
		fmt.Printf("Value \"%s\" for parameter %s is invalid for %s\n", paramVal, paramName, context)
		return 0, err
	}
	return retVal, nil
}

func parseStringToSwitchbank(paramVal, paramName, context string) (*circuit.NSwitchBank, error) {

	switches, err := circuit.NewNSwitchBank(paramVal)
	if err != nil {
		fmt.Printf("Error parsing \"%s\" from parameter %s into %s.  Error details: %s\n", paramVal, paramName, context, err.Error())
		return nil, err
	}

	return switches, nil
}
