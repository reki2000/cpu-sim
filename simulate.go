package sim

import "os"

type nodeIndex int
type voltage int

const (
	// H means Vdd
	H voltage = iota
	// L means Gnd
	L
)

// Vdd is a reserved nodeIndex for Vdd
const Vdd nodeIndex = 0

// Gnd is a reserved nodeIndex for Gnd
const Gnd nodeIndex = 1

type transistor struct {
	g nodeIndex
	s nodeIndex
	d nodeIndex
}

type Switch struct {
	on  bool
	out nodeIndex
}

type env struct {
	vol    map[nodeIndex]voltage
	parent []nodeIndex
	name   map[nodeIndex]string
	pmos   []transistor
	nmos   []transistor
	sw     []*Switch
}

func newEnv() *env {
	e := &env{map[nodeIndex]voltage{}, []nodeIndex{}, map[nodeIndex]string{}, []transistor{}, []transistor{}, []*Switch{}}
	e.node() // Vdd
	e.node() // Gnd
	e.node() // Clk
	e.node() // ~Clk
	return e
}

func (e *env) node() nodeIndex {
	idx := nodeIndex(len(e.parent))
	e.parent = append(e.parent, idx)
	return idx
}

func (e *env) setName(idx nodeIndex, name string) {
	e.name[idx] = name
}

func (e *env) root(i nodeIndex) nodeIndex {
	for e.parent[i] != i {
		i = e.parent[i]
	}
	return i
}

func (e *env) getVol(i nodeIndex) (voltage, bool) {
	v, ok := e.vol[e.root(i)]
	return v, ok
}

func (e *env) setVol(i nodeIndex, v voltage) bool {
	root := e.root(i)
	prev, ok := e.vol[root]
	if !ok {
		e.vol[root] = v
		return true
	} else if prev == v {
		return false
	}
	println("conflict at root:", root, "prev:", prev, " new:", v)
	os.Exit(-1)
	return false
}

func (e *env) connect(from nodeIndex, to nodeIndex) {
	e.parent[e.root(from)] = e.root(to)
}

func (e *env) dump() {
	for idx, name := range e.name {
		vol := "?"
		v, ok := e.getVol(idx)
		if ok {
			if v == L {
				vol = "L"
			} else {
				vol = "H"
			}
		}
		println("root:", e.root(idx), " vol:", vol, " name:", name)
	}
}

func (e *env) debug() {
	for idx, v := range e.vol {
		vol := "?"
		if v == L {
			vol = "L"
		} else {
			vol = "H"
		}
		println("root:", e.root(idx), " vol:", vol)
	}

	for i, p := range e.parent {
		println("parent[", i, "] = ", p)
	}
}

func (e *env) simulate() {
	prev := e.vol
	e.vol = map[nodeIndex]voltage{}
	e.setVol(Vdd, H)
	e.setVol(Gnd, L)
	for _, sw := range e.sw {
		if sw.on {
			e.setVol(sw.out, H)
		} else {
			e.setVol(sw.out, L)
		}
	}
	for updateCount := 0; updateCount < 10; updateCount++ {
		updated := false
		for _, tr := range e.nmos {
			if v, ok := e.getVol(tr.g); ok && v == H {
				if v, ok := e.getVol(tr.s); ok && v == L {
					updated = updated || e.setVol(tr.d, L)
				}
			}
		}
		for _, tr := range e.pmos {
			if v, ok := e.getVol(tr.g); ok && v == L {
				if v, ok := e.getVol(tr.s); ok && v == H {
					updated = updated || e.setVol(tr.d, H)
				}
			}
		}
		if !updated {
			break
		}
	}

	for updateCount := 0; updateCount < 10; updateCount++ {
		updated := false
		for _, tr := range e.nmos {
			if _, ok := e.getVol(tr.g); !ok {
				if v, ok := prev[e.root(tr.g)]; ok && v == H {
					if v, ok := e.getVol(tr.s); ok && v == L {
						updated = updated || e.setVol(tr.d, L)
					}
				}
			}
		}
		for _, tr := range e.pmos {
			if _, ok := e.getVol(tr.g); !ok {
				if v, ok := prev[e.root(tr.g)]; ok && v == L {
					if v, ok := e.getVol(tr.s); ok && v == H {
						updated = updated || e.setVol(tr.d, H)
					}
				}
			}
		}
		if !updated {
			break
		}
	}
}

func (e *env) makePmos() transistor {
	g, s, d := e.node(), e.node(), e.node()
	t := transistor{g: g, s: s, d: d}
	e.pmos = append(e.pmos, t)
	return t
}

func (e *env) makeNmos() transistor {
	g, s, d := e.node(), e.node(), e.node()
	t := transistor{g: g, s: s, d: d}
	e.nmos = append(e.nmos, t)
	return t
}

type Not struct {
	in, out nodeIndex
}

type Nand struct {
	in0, in1, out nodeIndex
}

func (e *env) makeSwitch(name string) *Switch {
	sw0 := &Switch{on: true, out: e.node()}
	e.sw = append(e.sw, sw0)
	return sw0
}

func (e *env) makeNot(name string) Not {
	p := e.makePmos()
	n := e.makeNmos()
	e.connect(p.d, n.d)
	e.setName(p.d, name+" out")
	e.connect(p.s, Vdd)
	e.connect(n.s, Gnd)
	e.connect(p.g, n.g)
	return Not{n.g, n.d}
}

func (e *env) makeNand(name string) Nand {
	p0, p1 := e.makePmos(), e.makePmos()
	n0, n1 := e.makeNmos(), e.makeNmos()

	e.connect(p0.s, Vdd)
	e.connect(p1.s, Vdd)
	e.connect(p0.g, n0.g)
	e.connect(p1.g, n1.g)

	e.connect(n1.s, Gnd)
	e.connect(n0.s, n1.d)

	e.connect(p0.d, p1.d)
	e.connect(n0.d, p1.d)

	return Nand{n0.g, n1.g, p1.d}
}
