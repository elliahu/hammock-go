package main

import (
	"hammock-go/editor"
	"runtime"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	editor, err := editor.CreateEditor()
	if err != nil {
		panic(err)
	}

	editor.Run()

	defer editor.Destroy()

}
