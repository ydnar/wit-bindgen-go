package cm

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestListMethods(t *testing.T) {
	want := []byte("hello world")
	type myList List[uint8]
	l := myList(ToList(want))
	got := l.Slice()
	if !bytes.Equal(want, got) {
		t.Errorf("got (%s) != want (%s)", string(got), string(want))
	}
}

func TestListJSON(t *testing.T) {
	simpleList := []string{"one", "two", "three"}
	simpleJSON, err := json.Marshal(simpleList)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("encoding", func(t *testing.T) {
		l := ToList(simpleList)
		data, err := json.Marshal(l)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := data, simpleJSON; !bytes.Equal(got, want) {
			t.Errorf("got (%s) != want (%s)", string(got), string(want))
		}

		var emptyList List[string]
		data, err = json.Marshal(emptyList)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := data, nullLiteral; !bytes.Equal(got, want) {
			t.Errorf(" got (%s) when should have got nil", string(data))
		}
	})

	t.Run("decoding", func(t *testing.T) {
		var decodedList List[string]
		if err := json.Unmarshal(simpleJSON, &decodedList); err != nil {
			t.Fatal(err)
		}

		if got, want := decodedList.Slice(), simpleList; !reflect.DeepEqual(got, want) {
			t.Errorf("got (%s) != want (%s)", got, want)
		}

		var emptyList List[string]
		if err := json.Unmarshal(nullLiteral, &emptyList); err != nil {
			t.Fatal(err)
		}

		if got, want := emptyList.Slice(), []string(nil); !reflect.DeepEqual(got, want) {
			t.Errorf("got (%+v) != want (%+v)", got, want)
		}
	})
}
