package main

import (
	"hammock-go/editor"
	"runtime"
)

func main() {
	// Lock to OS thread for Vulkan and Win32
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var editor editor.Editor
	err := editor.Create()
	if err != nil {
		panic(err)
	}

	editor.Run()

	defer editor.Destroy()

}
