// Package size implements parsing for string values with units.
package size

import (
	"fmt"
	"strconv"
	"unicode"
)

// Sign is the sign associated with a unit's value.
type Sign uint8

const (
	None Sign = iota
	Negative
	Positive
)

// Value is any value that can be represented by a unit.
//
// Value implements flag.Value and flag.Getter.
type Value struct {
	// unit is the associated unit.
	unit *Unit

	// value is the integer value.
	value int64

	// sign is the explicit sign given by the string converted to the
	// integer.
	sign Sign
}

// Unit is a map of unit names to conversion multipliers.
//
// There must be a unit that maps to 1.
type Unit struct {
	mapping map[string]int64
}

func NewUnit(m map[string]int64) (*Unit, error) {
	var found bool
	for _, mult := range m {
		if mult == 1 {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("could not find unit that maps to multiplier 1 for %v", m)
	}
	return &Unit{m}, nil
}

func (u *Unit) NewValue(value int64, sign Sign) *Value {
	return &Value{
		value: value,
		sign:  sign,
		unit:  u,
	}
}

func (u *Unit) ValueFromString(str string) (*Value, error) {
	s := &Value{unit: u}

	if err := s.Set(str); err != nil {
		return nil, err
	}
	return s, nil
}

// String implements flag.Value.String and fmt.Stringer.
func (s Value) String() string {
	var bestName string
	bestMult := int64(1)
	for name, mult := range s.unit.mapping {
		if s.value%mult == 0 && mult >= bestMult {
			bestName = name
			bestMult = mult
		}
	}
	var sign string
	if s.sign == Negative {
		sign = "-"
	} else if s.sign == Positive {
		sign = "+"
	}
	if bestName == "" {
		return fmt.Sprintf("%s%d (no unit)", sign, s.value)
	}
	return fmt.Sprintf("%s%d%s", sign, s.value/bestMult, bestName)
}

// Get implements flag.Getter.Get.
func (s Value) Get() interface{} {
	return s
}

// Set implements flag.Value.Set.
func (s *Value) Set(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("invalid size %q", str)
	}

	start, end := 0, len(str)
	if str[0] == '+' {
		s.sign = Positive
		start++
	} else if str[0] == '-' {
		s.sign = Negative
		start++
	}

	for i, r := range str[start:] {
		if unicode.IsLetter(r) {
			end = start + i
			break
		}
	}

	value, err := strconv.ParseInt(str[:end], 10, 64)
	if err != nil {
		return fmt.Errorf("could not convert %q to size: %v", str, err)
	}

	unitName := str[end:]
	mult, ok := s.unit.mapping[unitName]
	if !ok {
		return fmt.Errorf("unit %q is not valid", unitName)
	}
	s.value = value * mult
	return nil
}
