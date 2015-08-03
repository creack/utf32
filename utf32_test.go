package utf32

import "testing"

type testData struct {
	str      string
	utf8len  int
	utf32len int
}

func TestRoundTrip(t *testing.T) {
	var strs = []testData{
		{str: "hello world", utf8len: 11, utf32len: 11},
		{str: "éééééЈ", utf8len: 12, utf32len: 6},
		{str: "a", utf8len: 1, utf32len: 1},
		{str: "Ј", utf8len: 2, utf32len: 1},
		{str: "झ", utf8len: 3, utf32len: 1},
		{str: "𒔊", utf8len: 4, utf32len: 1},
	}
	for _, elem := range strs {
		utf32, err := ConvertUTF8toUTF32(elem.str)
		if err != nil {
			t.Fatal(err)
		}
		str, err := ConvertUTF32toUTF8(utf32)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := elem.str, str; expect != got {
			t.Fatalf("Unexpected result.\nExpect:\t%s\nGot:\t%s\n", expect, got)
		}
		if expect, got := elem.utf8len, len(str); expect != got {
			t.Fatalf("Unexpected utf8 length.\nExpect:\t%d\nGot:\t%d\n", expect, got)
		}
		if expect, got := elem.utf32len, len(utf32); expect != got {
			t.Fatalf("Unexpected utf32 length.\nExpect:\t%d\nGot:\t%d\n", expect, got)
		}
	}
}
