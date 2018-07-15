/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

// Package round implements string-based decimal number rounding, parsing, and normalisation, supporting scientific
// notation and malformed input.
package round

import (
	"strings"
	"unicode"
	"strconv"
	"regexp"
	"fmt"
	"errors"
)

const (
	// FormatFloat32 is a format template that can be used to print the minimum precision to guarantee that a program
	// (obeying the same standard) can read the formatted float32 back in exactly.
	// https://en.wikipedia.org/wiki/Single-precision_floating-point_format#IEEE_754_single-precision_binary_floating-point_format:_binary32
	FormatFloat32 = `%.9g`

	// FormatFloat64 is a format template that can be used to print the minimum precision to guarantee that a program
	// (obeying the same standard) can read the formatted float64 back in exactly.
	// https://en.wikipedia.org/wiki/Double-precision_floating-point_format#IEEE_754_double-precision_binary_floating-point_format:_binary64
	FormatFloat64 = `%.17g`
)

// String converts a value to a string, handling special cases for floating points in order to apply FormatFloat32 and
// FormatFloat64, otherwise by default just using fmt.Sprint.
func String(v interface{}) string {
	switch value := v.(type) {
	case float32:
		return fmt.Sprintf(FormatFloat32, value)
	case float64:
		return fmt.Sprintf(FormatFloat64, value)
	default:
		return fmt.Sprint(v)
	}
}

// Runes converts the strings in the output of Parse to rune slices.
func Runes(signbit bool, integer string, fractional string, exponential int, ok bool) (bool, []rune, []rune, int, bool) {
	return signbit, []rune(integer), []rune(fractional), exponential, ok
}

// Apply can be used to round the output of Runes(Parse(...)) to n decimal places, note it may adjust the
// exponential, and will return all zero values if ok was false, see Decimal for more info.
func Apply(signbit bool, integer []rune, fractional []rune, exponential int, ok bool) func(n int) (signbit bool, integer []rune, fractional []rune, exponential int, ok bool) {
	return func(n int) (bool, []rune, []rune, int, bool) {
		if !ok {
			return false, nil, nil, 0, false
		}

		// adjust the n decimal arg by the exponential, so we round to the actual point we want
		// e.g. if we want to round to two decimal places, and have (false, "12", "1456", 1, true), then since the
		// actual number is 121.456 (=12.1456 x 10 ^ 1), we want to use 3 digits from fractional, instead of 2
		n += exponential

		// shift digits between fractional and integer until n is 0
		// NOTE: we also adjust the exponential to keep track of the actual number
		for {
			if n > 0 {
				n--
				exponential--
				integer, fractional = moveLeft(integer, fractional)
			} else if n < 0 {
				n++
				exponential++
				integer, fractional = moveRight(integer, fractional)
			} else {
				break
			}
		}

		// if fractional starts with 5 or above then add 1 to the uint that integer represents (round part 1)
		if roundFractional(fractional) {
			integer = incrementInteger(integer)
		}

		// discard anything left in fractional (round part 2)
		fractional = nil

		// signbit and ok are unchanged, integer, fractional and exponential may have been modified
		return signbit, integer, fractional, exponential, true
	}
}

// Join can be used with the output of Runes(Parse(...)) to build a sane decimal string, it returns false if parse did.
func Join(signbit bool, integer []rune, fractional []rune, exponential int, ok bool) (string, bool) {
	if !ok {
		return "", false
	}

	// bring exponential to 0 by moving digits between integer and fractional
	for {
		if exponential > 0 {
			exponential--
			integer, fractional = moveLeft(integer, fractional)
		} else if exponential < 0 {
			exponential++
			integer, fractional = moveRight(integer, fractional)
		} else {
			break
		}
	}

	// trim any leading zeros from integer
	for len(integer) > 0 {
		if integer[0] != '0' {
			break
		}
		integer = integer[1:]
	}

	// trim any trailing zeros from fractional
	for i := len(fractional) - 1; i >= 0; i-- {
		if fractional[i] != '0' {
			break
		}
		fractional = fractional[:i]
	}

	// before we ensure integer has at least '0' in it, check we won't end up with -0
	if signbit && len(integer) == 0 && len(fractional) == 0 {
		signbit = false
	}

	// ensure integer has at least one digit ('0' if none)
	if len(integer) == 0 {
		integer = append(integer, '0')
	}

	// build the output (init with capacity of all digits + 2, to account for potential sign and decimal)
	result := make([]rune, 0, len(integer)+len(fractional)+2)

	// append any negative sign first
	if signbit {
		result = append(result, '-')
	}

	// append all of integer
	result = append(result, integer...)

	// and any fractional (only adding the period if there was any)
	if len(fractional) != 0 {
		result = append(result, '.')
		result = append(result, fractional...)
	}

	// building complete, return result as a string
	return string(result), true
}

// Float32 can be used with Runes(Parse(...)) to parse and convert to float32 in one step, note it will return an
// error if ok is false, and will pass through errors from strconv.ParseFloat without modification.
func Float32(signbit bool, integer []rune, fractional []rune, exponential int, ok bool) (float32, error) {
	s, ok := Join(signbit, integer, fractional, exponential, ok)
	if !ok {
		return 0, errors.New("round.Float32 failed to parse string")
	}

	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}

	return float32(f), nil
}

// Float64 can be used with Runes(Parse(...)) to parse and convert to float64 in one step, note it will return an
// error if ok is false, and will pass through errors from strconv.ParseFloat without modification.
func Float64(signbit bool, integer []rune, fractional []rune, exponential int, ok bool) (float64, error) {
	s, ok := Join(signbit, integer, fractional, exponential, ok)
	if !ok {
		return 0, errors.New("round.Float64 failed to parse string")
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

// Parse parses a numeric value, which will be converted to a string using String, supporting scientific notation,
// separating out like integer.fractional x 10 ^ exponential, where signbit will be true if the number evaluates to
// a negative, integer and fractional will contain all meaningful digits (an empty string representing zero),
// or ok will be false if parsing failed, e.g. it did not match the expected format, or the exponential component
// couldn't fit in an int.
//
// NOTES:
// - scientific notation like (x10^, e, *10^) is supported (case insensitive), which works with String(float64)
// - it will strip all commas and whitespace prior to parsing, so strings like "  2,000,000  " etc are supported
// - integer will be a string of digits of 0-n length, with ALL leading zeros stripped
// - fractional will be a string of digits of 0-n length, with ALL trailing zeros stripped
// - signbit will be true for negatives (like math.Signbit) unless integer.fractional x 10^exponential would evaluate
//   to zero (note that this effectively means all cases matching integer="" and fractional="")
// - any exponential component must be well-formed enough to be parsed by strconv.Atoi
func Parse(v interface{}) (signbit bool, integer string, fractional string, exponential int, ok bool) {
	return ParseString(String(v))
}

// ParseString is the implementation of Parse after string conversion has been applied.
func ParseString(s string) (signbit bool, integer string, fractional string, exponential int, ok bool) {
	// strip whitespace and commas
	s = strings.Map(
		func(r rune) rune {
			if unicode.IsSpace(r) || r == ',' {
				return -1
			}
			return r
		},
		s,
	)

	// use a regex to split out the initial components
	sm := parseRegex.FindStringSubmatch(s)

	smLen := len(sm)
	if smLen == 0 {
		// no match, we can just return (false, "", "", 0, false)
		return
	}

	// parsing success, any failures below must return directly or set ok back to false
	ok = true

	if smLen > 1 && sm[1] == `-` {
		// there was a negative sign present, set the flag
		// NOTE: we may have to clear it again if the rest of the expression evaluates to zero
		signbit = true
	}

	if smLen > 2 {
		// parsed an integer component, trim all leading zeros
		integer = strings.TrimLeftFunc(
			sm[2],
			func(r rune) bool {
				return r == '0'
			},
		)
	}

	if smLen > 3 {
		// parsed a fractional component, trim all trailing zeros
		fractional = strings.TrimRightFunc(
			sm[3],
			func(r rune) bool {
				return r == '0'
			},
		)
	}

	if signbit && integer == "" && fractional == "" {
		// we parsed a negative sign, but we then parsed an expression that evaluates to 0, remove the negative
		signbit = false
	}

	if smLen > 4 && sm[4] != "" {
		// parsed an exponential component, convert it to an integer, note it must be well-formed, and must fit
		if v, err := strconv.Atoi(sm[4]); err != nil {
			// bail out, directly return all zero values
			return false, "", "", 0, false
		} else {
			// update the exponential to return with the parsed int
			exponential = v
		}
	}

	// we are done!
	return
}

// Decimal rounds a value to n decimal places, supporting any value that can be parsed using a call
// like Parse(String(value)), and returns it as a string, or false if parsing failed, normalising
// the output to the format [-]INTEGER_COMPONENT[.FRACTIONAL_COMPONENT], with unnecessary trailing or leading
// zeros stripped, and the sign only present for negatives that don't evaluate as -0.
//
// NOTE: the implementation is effectively Join(Apply(Runes(ParseString(String(v))))(n))
func Decimal(v interface{}, n int) (string, bool) {
	return DecimalString(String(v), n)
}

// DecimalString is the Decimal implementation after converting the value to a string using String.
func DecimalString(s string, n int) (string, bool) {
	return Join(Apply(Runes(ParseString(s)))(n))
}

// moveLeft moves the first digit of fractional (default to 0) to the end of integer
func moveLeft(integer, fractional []rune) ([]rune, []rune) {
	digit := '0'
	if len(fractional) != 0 {
		digit = fractional[0]
		fractional = fractional[1:]
	}
	integer = append(integer, digit)
	return integer, fractional
}

// moveRight moves the last digit of integer (default to 0) to the start of fractional
func moveRight(integer, fractional []rune) ([]rune, []rune) {
	digit := '0'
	if l := len(integer); l != 0 {
		digit = integer[l-1]
		integer = integer[:l-1]
	}
	fractional = append(append(make([]rune, 0, len(fractional)+1), digit), fractional...)
	return integer, fractional
}

// incrementInteger increments an integer expressed as a slice of runes (digits) by 1
func incrementInteger(integer []rune) []rune {
	done := false
	for i := len(integer) - 1; i >= 0; i-- {
		if integer[i] == '9' {
			integer[i] = '0'
		} else {
			integer[i]++
			done = true
			break
		}
	}
	if !done {
		integer = append(append(make([]rune, 0, len(integer)+1), '1'), integer...)
	}
	return integer
}

// roundFractional returns true if the fractional component (all digits after any period, or an empty slice) will
// cause rounding to result in different behavior to just truncating it
func roundFractional(fractional []rune) bool {
	if len(fractional) == 0 {
		return false
	}
	return fractional[0] >= '5'
}

var (
	parseRegex = regexp.MustCompile(`(?i)^((?:)|(?:\+)|(?:-))(\d+)(?:(?:)|(?:\.(\d+)))(?:(?:)|(?:(?:(?:x10\^)|(?:\*10\^)|(?:e))((?:(?:)|(?:\+)|(?:-))\d+)))$`)
)
