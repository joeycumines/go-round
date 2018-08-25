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

package round

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func BenchmarkParseString_maxFloat64(b *testing.B) {
	s := String(math.MaxFloat64)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		signbit, integer, fractional, exponential, ok := ParseString(s)
		b.StopTimer()
		v, err := Float64(Runes(signbit, integer, fractional, exponential, ok))
		if err != nil {
			b.Fatal(s, err)
		}
		if v != math.MaxFloat64 {
			b.Fatal(v)
		}
		b.StartTimer()
	}
}

func BenchmarkBigFloatSetString_maxFloat64(b *testing.B) {
	s := String(math.MaxFloat64)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		f, ok := new(big.Float).
			SetString(s)
		b.StopTimer()
		if !ok {
			b.Fatal(f, ok)
		}
		v, _ := f.Float64()
		if v != math.MaxFloat64 {
			b.Fatal(v)
		}
		b.StartTimer()
	}
}

func ExampleParse_valid() {
	printParse := func(v interface{}) {
		signbit, integer, fractional, exponential, ok := Parse(v)
		fmt.Printf("%v,%v,%v,%v,%v\n", signbit, integer, fractional, exponential, ok)
	}

	// integer string
	printParse("214888")

	// integer string with +
	printParse("+ 214888")

	// integer string with -
	printParse("- 214888")

	// negative zero integer string
	printParse("-0")

	// a very small decimal string
	printParse("0.0000000000000000000000000000000023410000000000000000000000000127000000000000000077743")

	// exponential 1
	printParse("24124.2321699 x 10 ^ 51")

	// exponential 2
	printParse("24124.2321699 * 10 ^ 51")

	// exponential 3
	printParse("24124.2321699E51")

	// exponential with negative
	printParse("24124.2321699 x 10 ^ -51")

	// commas
	printParse("25,000,000")

	// a float64 literal that can be stored (but not exactly)
	printParse(float64(241.992))

	// a float64 literal that cannot be stored (in a float64)
	printParse(float64(99141249866.12323500200005000000004124412))

	// max int64 value
	printParse(int64(math.MaxInt64))

	// min int64 value
	printParse(int64(math.MinInt64))

	// leading and trailing zeros
	printParse("00000000021434000.00288800000000000000")

	// empty fractional
	printParse("00000000021434000.00000000000000")

	// scientific notation again
	printParse("-1.234456e+78")

	// max float64
	printParse(float64(math.MaxFloat64))

	// min non-zero float64
	printParse(float64(math.SmallestNonzeroFloat64))

	// max float32
	printParse(float32(math.MaxFloat32))

	// min non-zero float32
	printParse(float32(math.SmallestNonzeroFloat32))

	// Output:
	// false,214888,,0,true
	// false,214888,,0,true
	// true,214888,,0,true
	// false,,,0,true
	// false,,0000000000000000000000000000000023410000000000000000000000000127000000000000000077743,0,true
	// false,24124,2321699,51,true
	// false,24124,2321699,51,true
	// false,24124,2321699,51,true
	// false,24124,2321699,-51,true
	// false,25000000,,0,true
	// false,241,99199999999999,0,true
	// false,99141249866,12323,0,true
	// false,9223372036854775807,,0,true
	// true,9223372036854775808,,0,true
	// false,21434000,002888,0,true
	// false,21434000,,0,true
	// true,1,234456,78,true
	// false,1,7976931348623157,308,true
	// false,4,9406564584124654,-324,true
	// false,3,40282347,38,true
	// false,1,40129846,-45,true
}

func ExampleParse_invalid() {
	printParse := func(v interface{}) {
		signbit, integer, fractional, exponential, ok := Parse(v)
		fmt.Printf("%v,%v,%v,%v,%v\n", signbit, integer, fractional, exponential, ok)
	}

	// omitted integer
	printParse(".00124")

	// exponential that won't fit in an int
	printParse("1 x 10 ^ 9999999999999999999999999999999999999999999999999999999999999999999")

	// things that don't fmt.Sprint as valid numbers
	printParse(nil)

	// Output:
	// false,,,0,false
	// false,,,0,false
	// false,,,0,false
}

func ExampleDecimal_demoNBounds() {
	// round to 2 decimal places
	fmt.Println(Decimal(125.12475212144, 2))

	// round negative to 2 decimal places
	fmt.Println(Decimal(-125.12475212144, 2))

	// string with positive sign to four places
	fmt.Println(Decimal("+125.12475212144", 4))

	// and the same for negative
	fmt.Println(Decimal("-125.12475212144", 4))

	// you can round to an arbitrary negative point
	fmt.Println(Decimal(125.12475212144, -1))

	// and of course to zero
	fmt.Println(Decimal(125.12475212144, 0))

	// Output:
	// 125.12 true
	// -125.12 true
	// 125.1248 true
	// -125.1248 true
	// 130 true
	// 125 true
}

func ExampleDecimal_formats() {
	// int64 value
	fmt.Println(Decimal(int64(2999694421), -5))

	// scientific notation combined with rounding
	fmt.Println(Decimal("-5.1234567890000 x 10 ^ 4", 3))

	// scientific notation with negative n
	fmt.Println(Decimal("511.1234567890000 x 10 ^ 4", -1))

	// scientific notation, negative exponential, using e notation
	fmt.Println(Decimal("12128882148812.9123124124E-4", 2))

	// number with commas and whitespace
	fmt.Println(Decimal("      9, 214, 501        ", -4))

	// Output:
	// 2999700000 true
	// -51234.568 true
	// 5111230 true
	// 1212888214.88 true
	// 9210000 true
}

func ExampleDecimal_edgeCases() {
	// an empty string
	fmt.Println(Decimal("", 0))

	// an invalid number
	fmt.Println(Decimal(true, 0))

	// an exponential that won't fit in an int
	fmt.Println(Decimal("2 x 10 ^ 99999999999999999999999999999999999999999999999999999999999999999999999", 0))

	// rounding past the actual number (up)
	fmt.Println(Decimal("500", -3))

	// round past the actual number (down)
	fmt.Println(Decimal("499", -3))

	// a small positive number which rounds to zero
	fmt.Println(Decimal("+0.4", 0))

	// a small negative number which rounds to zero
	fmt.Println(Decimal("-0.4", 0))

	// a very large decimal number
	fmt.Println(Decimal("888888888888888888888888888888888888888888888888888888888888888.888888888888888888888888888888888888888888888888888888888888888", 5))

	// a very small decimal number
	fmt.Println(Decimal("5.213 * 10 ^ -50", 52))

	// a massive number using scientific notation
	fmt.Println(Decimal("4 * 10 ^ 1000", 0))

	// a number with trailing zeros
	fmt.Println(Decimal("41242141243135.00213214200020000000000000000000000000000000000000000", 20))

	// leading zeros too
	fmt.Println(Decimal("0000000000000000412421412431350000000000.0000000000000000", 20))

	// many zeros
	fmt.Println(Decimal("0.000000000000000000000E50", 0))

	// Output:
	//  false
	//  false
	//  false
	// 1000 true
	// 0 true
	// 0 true
	// 0 true
	// 888888888888888888888888888888888888888888888888888888888888888.88889 true
	// 0.0000000000000000000000000000000000000000000000000521 true
	// 40000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 true
	// 41242141243135.0021321420002 true
	// 412421412431350000000000 true
	// 0 true
}

func TestIncrementInteger(t *testing.T) {
	type TestCase struct {
		Input  string
		Output string
	}

	testCases := []TestCase{
		{
			Input:  "",
			Output: "1",
		},
		{
			Input:  "1",
			Output: "2",
		},
		{
			Input:  "0",
			Output: "1",
		},
		{
			Input:  "6",
			Output: "7",
		},
		{
			Input:  "8",
			Output: "9",
		},
		{
			Input:  "9",
			Output: "10",
		},
		{
			Input:  "100231321",
			Output: "100231322",
		},
		{
			Input:  "999999999",
			Output: "1000000000",
		},
		{
			Input:  "998999999",
			Output: "999000000",
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestIncrementInteger_#%d", i+1)

		output := string(incrementInteger([]rune(testCase.Input)))

		if output != testCase.Output {
			t.Error(name, "output", output, "!= expected", testCase.Output)
		}
	}
}

func TestRoundFractional(t *testing.T) {
	type TestCase struct {
		Input  string
		Output bool
	}

	testCases := []TestCase{
		{
			Input:  "",
			Output: false,
		},
		{
			Input:  "1",
			Output: false,
		},
		{
			Input:  "0",
			Output: false,
		},
		{
			Input:  "4",
			Output: false,
		},
		{
			Input:  "5",
			Output: true,
		},
		{
			Input:  "6",
			Output: true,
		},
		{
			Input:  "8",
			Output: true,
		},
		{
			Input:  "9",
			Output: true,
		},
		{
			Input:  "4999999999999",
			Output: false,
		},
		{
			Input:  "50000000000000000",
			Output: true,
		},
		{
			Input:  "100231321",
			Output: false,
		},
		{
			Input:  "999999999",
			Output: true,
		},
		{
			Input:  "998999999",
			Output: true,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestIncrementInteger_#%d", i+1)

		output := roundFractional([]rune(testCase.Input))

		if output != testCase.Output {
			t.Error(name, "output", output, "!= expected", testCase.Output)
		}
	}
}

func TestMoveLeft(t *testing.T) {
	type TestCase struct {
		AIn, BIn, AOut, BOut string
	}

	testCases := []TestCase{
		{
			AOut: "0",
		},
		{
			AIn:  "",
			BIn:  "3",
			AOut: "3",
			BOut: "",
		},
		{
			AIn:  "1241924",
			AOut: "12419240",
		},
		{
			AIn:  "",
			BIn:  "7251",
			AOut: "7",
			BOut: "251",
		},
		{
			AIn:  "8",
			BIn:  "7251",
			AOut: "87",
			BOut: "251",
		},
		{
			AIn:  "8",
			BIn:  "",
			AOut: "80",
			BOut: "",
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestMoveLeft_#%d", i+1)

		l, r := moveLeft([]rune(testCase.AIn), []rune(testCase.BIn))
		a, b := string(l), string(r)

		if a != testCase.AOut {
			t.Error(name, "a", a, "!= expected", testCase.AOut)
		}

		if b != testCase.BOut {
			t.Error(name, "b", b, "!= expected", testCase.BOut)
		}
	}
}

func TestMoveRight(t *testing.T) {
	type TestCase struct {
		AIn, BIn, AOut, BOut string
	}

	testCases := []TestCase{
		{
			BOut: "0",
		},
		{
			AIn:  "",
			BIn:  "124",
			AOut: "",
			BOut: "0124",
		},
		{
			AIn:  "7",
			BIn:  "",
			AOut: "",
			BOut: "7",
		},
		{
			AIn:  "",
			BIn:  "7",
			AOut: "",
			BOut: "07",
		},
		{
			AIn:  "3",
			BIn:  "4",
			AOut: "",
			BOut: "34",
		},
		{
			AIn:  "124567",
			BIn:  "93121",
			AOut: "12456",
			BOut: "793121",
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestMoveLeft_#%d", i+1)

		l, r := moveRight([]rune(testCase.AIn), []rune(testCase.BIn))
		a, b := string(l), string(r)

		if a != testCase.AOut {
			t.Error(name, "a", a, "!= expected", testCase.AOut)
		}

		if b != testCase.BOut {
			t.Error(name, "b", b, "!= expected", testCase.BOut)
		}
	}
}

func TestFloat32_success(t *testing.T) {
	run := func(f float32) {
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return
		}
		//t.Log("INPUT:", f)
		p, err := Float32(Runes(Parse(f)))
		if err != nil {
			t.Error("failed to parse", f, "with error", err)
			return
		}
		//t.Log("OUTPUT:", p)
		if p != f {
			t.Error("output", p, "!= input", f)
		}
	}

	run(math.MaxFloat32)
	run(math.SmallestNonzeroFloat32)

	for x := 0; x < 10000; x++ {
		run(math.Float32frombits(rand.Uint32()))
	}
}

func TestFloat64_success(t *testing.T) {
	run := func(f float64) {
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return
		}
		//t.Log("INPUT:", f)
		p, err := Float64(Runes(Parse(f)))
		if err != nil {
			t.Error("failed to parse", f, "with error", err)
			return
		}
		//t.Log("OUTPUT:", p)
		if p != f {
			t.Error("output", p, "!= input", f)
		}
	}

	run(math.MaxFloat64)
	run(math.SmallestNonzeroFloat64)

	for x := 0; x < 10000; x++ {
		run(math.Float64frombits(rand.Uint64()))
	}
}

func TestFloat64_errors(t *testing.T) {
	if f, err := Float64(false, []rune("9"), []rune(""), 1000, true); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float64(Runes(Parse(math.NaN()))); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float64(Runes(Parse(math.Inf(1)))); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float64(Runes(Parse(math.Inf(-1)))); f != 0 || err == nil {
		t.Error(f, err)
	}
}

func TestFloat32_errors(t *testing.T) {
	if f, err := Float32(false, []rune("9"), []rune(""), 1000, true); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float32(Runes(Parse(math.NaN()))); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float32(Runes(Parse(math.Inf(1)))); f != 0 || err == nil {
		t.Error(f, err)
	}

	if f, err := Float32(Runes(Parse(math.Inf(-1)))); f != 0 || err == nil {
		t.Error(f, err)
	}
}

func TestEnsureExponentFloat64(t *testing.T) {
	type TestCase struct {
		E int
		O bool
	}

	testCases := []TestCase{
		{
			E: 0,
			O: true,
		},
		{
			E: -1022,
			O: true,
		},
		{
			E: -1023,
			O: false,
		},
		{
			E: 1023,
			O: true,
		},
		{
			E: 1024,
			O: false,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestEnsureExponentFloat64_#%d", i+1)

		signbit, integer, fractional, exponential, ok := EnsureExponentFloat64(true, "2141", "7452", testCase.E, true)

		if !signbit || integer != "2141" || fractional != "7452" || exponential != testCase.E {
			t.Error(name, "bad passthrough", signbit, integer, fractional, exponential, ok)
		}

		if testCase.O != ok {
			t.Error(name, "bad ok", signbit, integer, fractional, exponential, ok)
		}
	}
}

func TestEnsureExponentFloat64_bounds(t *testing.T) {
	d, err := Float64(Apply(Runes(EnsureExponentFloat64(ParseString(String(math.SmallestNonzeroFloat64)))))(2000))
	if err != nil || d != math.SmallestNonzeroFloat64 {
		t.Fatal(d, err)
	}
	d, err = Float64(Apply(Runes(EnsureExponentFloat64(ParseString(String(math.MaxFloat64)))))(0))
	if err != nil || d != math.MaxFloat64 {
		t.Fatal(d, err)
	}
}
