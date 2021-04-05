package codec

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

func TestGobCodec_Write(t *testing.T) {
	data := "ABC"
	buf := new(bytes.Buffer)

	//glob encoding
	enc := gob.NewEncoder(buf)
	enc.Encode(data)
	fmt.Println("Encoded:", data)  //Encoded: ABC

	//glob decoding
	d := gob.NewDecoder(buf)
	d.Decode(data)
	fmt.Println("Decoded: ", data) //Decoded:  ABC
}
