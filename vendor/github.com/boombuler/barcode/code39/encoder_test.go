package code39

import (
	"image/color"
	"testing"
)

func doTest(t *testing.T, addCS, fullASCII bool, data, testResult string) {
	code, err := Encode(data, addCS, fullASCII)
	if err != nil {
		t.Error(err)
	}
	if len(testResult) != code.Bounds().Max.X {
		t.Errorf("Invalid code size. Expected %d got %d", len(testResult), code.Bounds().Max.X)
	}
	for i, r := range testResult {
		if (code.At(i, 0) == color.Black) != (r == '1') {
			t.Errorf("Failed at position %d", i)
		}
	}
}

func Test_Encode(t *testing.T) {
	doTest(t, false, false, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		"1001011011010110101001011010110100101101101101001010101011001011011010110010101"+
			"011011001010101010011011011010100110101011010011010101011001101011010101001101011010"+
			"100110110110101001010101101001101101011010010101101101001010101011001101101010110010"+
			"101101011001010101101100101100101010110100110101011011001101010101001011010110110010"+
			"110101010011011010101010011011010110100101011010110010101101101100101010101001101011"+
			"011010011010101011001101010101001011011011010010110101011001011010100101101101")
}
