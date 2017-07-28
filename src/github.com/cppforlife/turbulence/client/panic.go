package client

import (
	"fmt"
)

func panicIfErr(err error, desc string) {
	if err != nil {
		panic(fmt.Sprintf("Failed to %s: %s", desc, err))
	}
}
