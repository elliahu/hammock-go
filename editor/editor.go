package editor

import (
	"hammock-go/core"
	"hammock-go/renderer"

	"github.com/bbredesen/go-vk"
)

type Editor struct {
	window    Window
	surface   vk.SurfaceKHR
	instance  vk.Instance
	context   core.Context
	renderer  renderer.Renderer
	swapchain core.SwapChain
}

func (edit *Editor) mainLoop() {

}

func (editor *Editor) Create() error {

	// Create instance
	instance, err := core.CreateInstance()
	if err != nil {
		return err
	}
	editor.instance = instance

	// Create window
	window, err := CreateWindow("HammockGo Editor", 1920, 1080)
	if err != nil {
		return err
	}

	editor.window = window

	// Create surface
	surface, err := window.CreateSurface(editor.instance)
	if err != nil {
		return err
	}

	editor.surface = surface

	// Create context
	editor.context, err = core.CreateContext(editor.instance, editor.surface)
	if err != nil {
		return err
	}

	// Create renderer
	editor.renderer = renderer.CreateRenderer(&editor.context)

	// Create swapchain
	editor.swapchain.Create(instance, editor.context.GetPhysicalDevice(), editor.surface, editor.context.GetDevice(), 1920, 1080, false)

	return nil
}

func (edit *Editor) Run() {
	for !edit.window.ShouldClose() {
		edit.mainLoop()
		edit.window.PollEvents()
	}
}

func (edit *Editor) Destroy() {

}
