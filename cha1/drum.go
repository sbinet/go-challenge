// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "Saved with HW Version: %s\n", p.Version)
	fmt.Fprintf(out, "Tempo: %g\n", p.Tempo)
	for _, t := range p.Tracks {
		fmt.Fprintf(out, "(%d) %s\t%v\n", t.ID, t.Name, t.Steps)
	}
	return out.String()
}

type Track struct {
	ID    uint8
	Name  string
	Steps Steps
}

type Step byte

func (s Step) String() string {
	switch s {
	case 0:
		return "-"
	case 1:
		return "x"
	}
	panic("impossible")
}

type Steps [16]Step

func (s Steps) String() string {
	o := make([]byte, 1, 5*4)
	o[0] = '|'
	for i, v := range s {
		switch v {
		case 0:
			v = '-'
		case 1:
			v = 'x'
		}

		o = append(o, byte(v))
		if i%4 == 3 {
			o = append(o, '|')
		}
	}

	return string(o)
}
