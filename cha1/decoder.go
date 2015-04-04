package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrShortRead = errors.New("short read")
	Header       = [6]byte{'S', 'P', 'L', 'I', 'C', 'E'}
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = checkSpliceHeader(f)
	if err != nil {
		return nil, err
	}

	var size uint64
	err = binary.Read(f, binary.BigEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("drum: error reading record size (%v)", err)
	}

	dec := newDecoder(f, int64(size))

	dec.decodeVersion(&p.Version)
	dec.decodeTempo(&p.Tempo)

	for dec.r.N > 0 && dec.err == nil {
		var track Track
		dec.decodeTrack(&track)
		if dec.err == nil {
			p.Tracks = append(p.Tracks, track)
		}
	}

	return p, dec.err
}

func checkSpliceHeader(r io.Reader) error {
	var magic [6]byte
	nb, err := r.Read(magic[:])
	if err != nil {
		return fmt.Errorf("drum: error reading SPLICE header (err=%v)", err)
	}

	if nb != len(magic) {
		return fmt.Errorf("drum: error reading SPLICE header (err=%v)", ErrShortRead)
	}

	if magic != Header {
		return fmt.Errorf(
			"drum: invalid SPLICE header (got=%q, want=%q)",
			string(magic[:]), string(Header[:]),
		)
	}

	return nil
}

type decoder struct {
	r   *io.LimitedReader
	err error
}

func newDecoder(r io.Reader, n int64) *decoder {
	return &decoder{
		r: &io.LimitedReader{
			R: r,
			N: n,
		},
	}
}

func (dec *decoder) decodeVersion(version *string) error {
	if dec.err != nil {
		return dec.err
	}
	var buf [32]byte
	var nb int
	nb, dec.err = dec.r.Read(buf[:])
	if dec.err != nil {
		dec.err = fmt.Errorf("drum: error decoding version (%v)", dec.err)
		return dec.err
	}
	if nb != len(buf) {
		dec.err = fmt.Errorf("drum: error decoding version (%v)", ErrShortRead)
		return dec.err
	}

	i := bytes.Index(buf[:], []byte{0})
	*version = string(buf[:i])
	return dec.err
}

func (dec *decoder) decodeTempo(tempo *float32) error {
	if dec.err != nil {
		return dec.err
	}

	dec.err = binary.Read(dec.r, binary.LittleEndian, tempo)
	if dec.err != nil {
		dec.err = fmt.Errorf("drum: error reading tempo value (%v)", dec.err)
	}

	return dec.err
}

func (dec *decoder) decodeTrack(track *Track) error {
	if dec.err != nil {
		return dec.err
	}

	order := binary.BigEndian
	err := binary.Read(dec.r, order, &track.ID)
	if err != nil {
		dec.err = fmt.Errorf("drum: error reading track id (%v)", err)
		return dec.err
	}

	size := uint32(0)
	err = binary.Read(dec.r, order, &size)
	if err != nil {
		return fmt.Errorf("drum: error reading track name header (%v)", err)
	}
	name := make([]byte, int(size))
	_, err = dec.r.Read(name)
	if err != nil {
		dec.err = fmt.Errorf("drum: error reading track name (%v)", err)
		return dec.err
	}

	track.Name = string(name)
	err = binary.Read(dec.r, order, &track.Steps)
	if err != nil {
		dec.err = fmt.Errorf("drum: error reading track steps (%v)", err)
		return dec.err
	}

	return dec.err
}
