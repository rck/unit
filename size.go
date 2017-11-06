package size

import (
	"flag"
	"fmt"
	"strconv"
	"unicode"
)

// type size int64
type Mapping map[string]int64

type Size struct {
	Size         int64 // Parsed size including sign
	ExplicitSign int   // -1 if size string started with a '-', +1 if size string started with a '+', otherwise 0
	Unit         string
	str          string
	mapping      Mapping
}

func NewSize(size int64, unit string, mapping Mapping) Size {
	s := Size{
		Unit:    unit,
		mapping: mapping,
		str:     fmt.Sprintf("%d%s", size, unit),
	}

	if sz, err := calcSize(size, unit, mapping); err != nil {
		panic(err)
	} else {
		s.Size = sz
	}
	return s
}

func (s Size) String() string {
	return s.str
}

func calcSize(size int64, unit string, mapping Mapping) (int64, error) {
	if unit == "" {
		return size, nil
	}

	if m, ok := mapping[unit]; ok {
		return size * int64(m), nil
	} else {
		return 0, fmt.Errorf("Unit %q is not valid", unit)
	}
}

type sizeFlag struct{ Size }

func (sf *sizeFlag) Set(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("invalid size %q", s)
	}

	start, end := 0, len(s)
	if s[0] == '+' {
		sf.ExplicitSign = 1
		start++
	} else if s[0] == '-' {
		sf.ExplicitSign = -1
		start++
	}

	sf.Size.Unit = "" // reset the default one
	for i, r := range s[start:] {
		if unicode.IsLetter(r) {
			sf.Size.Unit = s[start+i:]
			break
		}
	}
	end -= len(sf.Size.Unit)

	if end < start { // this really should not happen...
		return fmt.Errorf("Internal parse error, end (%d) < start (%d)", end, start)
	}

	var parsed int64
	var err error
	if parsed, err = strconv.ParseInt(s[:end], 10, 64); err != nil { // do not use start, we want the sign
		return fmt.Errorf("Could not convert %q to size: %v", s, err)
	}
	if sz, err := calcSize(parsed, sf.Size.Unit, sf.Size.mapping); err != nil {
		return fmt.Errorf("Could not calculate size for %q: %v", s, err)
	} else {
		sf.Size.Size = sz
	}

	sf.str = s
	return nil
}

func Flag(name string, value Size, usage string) *Size {
	f := sizeFlag{value}
	flag.CommandLine.Var(&f, name, usage)
	return &f.Size
}
