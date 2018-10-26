package engineer

import (
	"fmt"
	"testing"
	"time"
)

func TestARFile(t *testing.T) {
	f, e := newARFile("/path/to/test.log", RotateTypeSecond)

	if e != nil {
		fmt.Println("new failed : ", e)
	}

	for i := 0; i < 100; i++ {
		time.Sleep(time.Second * 1)
		fmt.Println(f.Write([]byte("hello")))
	}

}
