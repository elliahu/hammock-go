package editor

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/bbredesen/go-vk"
	"golang.org/x/sys/windows"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	SW_SHOW             = 5
	SW_USE_DEFAULT      = 0x80000000
	WM_DESTROY          = 0x0002
	WM_CLOSE            = 0x0010
	CS_HREDRAW          = 0x0002
	CS_VREDRAW          = 0x0001
	IDC_ARROW           = 32512
	COLOR_WINDOW        = 5
)

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := procGetModuleHandleW.Call(0)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func registerClassEx(wndClass *WNDCLASSEX) (uint16, error) {
	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wndClass)))
	if ret == 0 {
		return 0, err
	}
	return uint16(ret), nil
}

func createWindowEx(className, windowName *uint16, style uint32, x, y, width, height uint32, parent, menu, instance syscall.Handle) (syscall.Handle, error) {
	ret, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		0,
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func showWindow(hwnd syscall.Handle, cmdShow int32) bool {
	ret, _, _ := procShowWindow.Call(uintptr(hwnd), uintptr(cmdShow))
	return ret != 0
}

func updateWindow(hwnd syscall.Handle) bool {
	ret, _, _ := procUpdateWindow.Call(uintptr(hwnd))
	return ret != 0
}

func getMessage(msg *MSG, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) (bool, error) {
	ret, _, err := procGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	if int32(ret) == -1 {
		return false, err
	}
	return ret != 0, nil
}

func translateMessage(msg *MSG) {
	procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

func dispatchMessage(msg *MSG) {
	procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

func defWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wparam,
		lparam,
	)
	return ret
}

func postQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func loadCursor(instance syscall.Handle, cursorName uintptr) (syscall.Handle, error) {
	ret, _, err := procLoadCursorW.Call(uintptr(instance), cursorName)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

// Window procedure callback
var wndProcCallback uintptr

func wndProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		postQuitMessage(0)
		return 0
	}
	return defWindowProc(hwnd, msg, wparam, lparam)
}

func createWin32WindowInternal(title string, width uint32, height uint32) (syscall.Handle, syscall.Handle, error) {
	// Get module handle
	hInstance, err := getModuleHandle()
	if err != nil {
		return 0, 0, fmt.Errorf("GetModuleHandle failed: %v", err)
	}

	// Convert class name to UTF16
	className, err := syscall.UTF16PtrFromString("MyWindowClass")
	if err != nil {
		return 0, 0, err
	}

	// Load cursor
	cursor, err := loadCursor(0, IDC_ARROW)
	if err != nil {
		return 0, 0, fmt.Errorf("LoadCursor failed: %v", err)
	}

	// Create callback and keep it alive
	wndProcCallback = syscall.NewCallback(wndProc)

	// Register window class
	wndClass := WNDCLASSEX{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:      CS_HREDRAW | CS_VREDRAW,
		WndProc:    wndProcCallback,
		Instance:   hInstance,
		Cursor:     cursor,
		Background: syscall.Handle(COLOR_WINDOW + 1),
		ClassName:  className,
	}

	_, err = registerClassEx(&wndClass)
	if err != nil {
		return 0, 0, fmt.Errorf("RegisterClassEx failed: %v", err)
	}

	// Convert window title to UTF16
	windowName, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return 0, 0, err
	}

	// Create window
	hwnd, err := createWindowEx(
		className,
		windowName,
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		SW_USE_DEFAULT,
		SW_USE_DEFAULT,
		width,
		height,
		0,
		0,
		hInstance,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("CreateWindowEx failed: %v", err)
	}

	// Show and update window
	showWindow(hwnd, SW_SHOW)
	updateWindow(hwnd)

	return hwnd, hInstance, nil
}

func CreateWindow(title string, width uint32, height uint32) (Window, error) {
	hwnd, hinstance, err := createWin32WindowInternal(title, width, height)
	if err != nil {
		return Window{}, err
	}
	return Window{
		hwnd:        windows.HWND(hwnd),
		hinstance:   windows.Handle(hinstance),
		shouldClose: false,
	}, nil
}

type Window struct {
	hwnd        windows.HWND
	hinstance   windows.Handle
	shouldClose bool
}

func NewWindow(hwnd windows.HWND) *Window {
	return &Window{
		hwnd:        hwnd,
		shouldClose: false,
	}
}

func (w *Window) CreateSurface(instance vk.Instance) (vk.SurfaceKHR, error) {
	surfaceInfo := vk.Win32SurfaceCreateInfoKHR{
		Hinstance: w.hinstance,
		Hwnd:      w.hwnd,
	}

	surface, err := vk.CreateWin32SurfaceKHR(instance, &surfaceInfo, nil)
	if err != nil {
		return vk.SurfaceKHR(vk.NULL_HANDLE), err
	}

	return surface, nil
}

func (w *Window) ShouldClose() bool {
	return w.shouldClose
}

func (w *Window) PollEvents() error {
	var msg MSG
	ret, err := getMessage(&msg, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("GetMessage failed: %v", err)
	}
	if !ret {
		w.shouldClose = true
		return nil
	}

	translateMessage(&msg)
	dispatchMessage(&msg)
	return nil
}

func Win32Loop() error {
	// Message loop
	var msg MSG
	for {
		ret, err := getMessage(&msg, 0, 0, 0)
		if err != nil {
			return fmt.Errorf("GetMessage failed: %v", err)
		}
		if !ret {
			break
		}

		translateMessage(&msg)
		dispatchMessage(&msg)
	}

	fmt.Println("Window closed")
	return nil
}
