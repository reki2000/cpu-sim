package sim

import (
	"fmt"
	"os"
	"testing"
)

func (e *env) assertVol(t *testing.T, idx nodeIndex, v voltage) {
	if vol, ok := e.getVol(idx); ok {
		if vol != v {
			fmt.Fprintln(os.Stderr, "node:", idx, "[", e.name[idx], "] expected: ", v, " actual:", vol)
			t.Fail()
		}
		return
	}
	fmt.Fprintln(os.Stderr, "voltage is not set at node:", idx, " ")
	t.Fail()
	return
}

func TestNot(t *testing.T) {
	e := newEnv()

	not1 := e.makeNot("NOT1")

	sw := e.makeSwitch("SW1")
	e.connect(not1.in, sw.out)

	e.simulate()
	e.assertVol(t, not1.out, L)

	sw.on = false
	e.simulate()
	e.assertVol(t, not1.out, H)

}

func TestNand(t *testing.T) {
	e := newEnv()

	nand1 := e.makeNand("NAND1")

	sw0 := e.makeSwitch("SW0")
	e.connect(nand1.in0, sw0.out)

	sw1 := e.makeSwitch("SW1")
	e.connect(nand1.in1, sw1.out)

	sw0.on = false
	sw1.on = false
	e.simulate()
	e.assertVol(t, nand1.out, H)

	sw0.on = true
	sw1.on = false
	e.simulate()
	e.assertVol(t, nand1.out, H)

	sw0.on = false
	sw1.on = true
	e.simulate()
	e.assertVol(t, nand1.out, H)

	sw0.on = true
	sw1.on = true
	e.simulate()
	e.assertVol(t, nand1.out, L)
}
