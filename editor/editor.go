package editor

import (
	"hammock-go/core"
	"hammock-go/renderer"

	"github.com/bbredesen/go-vk"
)

// TODO remove glfw dependency

type Editor struct {
	surface   vk.SurfaceKHR
	instance  vk.Instance
	context   core.Context
	renderer  renderer.Renderer
	swapchain core.SwapChain
}

func (edit *Editor) mainLoop() {

}

func createWindowAndSurface(edit *Editor) error {
	// Create window
	hInstance, hwnd, err := CreateWin32Window("HammockGo Editor", 1920, 1080)
	if err != nil {
		return err
	}

	// Create surface
	surfaceInfo := vk.Win32SurfaceCreateInfoKHR{
		Hinstance: hInstance,
		Hwnd:      hwnd,
	}

	surface, err := vk.CreateWin32SurfaceKHR(edit.instance, &surfaceInfo, nil)
	if err != nil {
		return err
	}

	edit.surface = surface

	return nil
}

func (editor *Editor) Create() error {

	// Create instance
	instance, err := core.CreateInstance()
	if err != nil {
		return err
	}
	editor.instance = instance

	// Create window and surface
	err = createWindowAndSurface(editor)
	if err != nil {
		return err
	}

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
	Win32Loop()
}

func (edit *Editor) Destroy() {

}
