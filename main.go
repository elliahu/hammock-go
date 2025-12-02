package main

import (
	"hammock-go/core"
)

func main() {
	context, err := core.CreateContext(nil)
	if err != nil {
		panic(err)
	}
	defer context.Destroy()

}
