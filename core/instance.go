package core

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/vulkan-go/vulkan"
)

// Debug callback function using the older DebugReport API
func debugCallback(
	flags vulkan.DebugReportFlags,
	objectType vulkan.DebugReportObjectType,
	object uint64,
	location uint,
	messageCode int32,
	pLayerPrefix string,
	pMessage string,
	pUserData unsafe.Pointer,
) vulkan.Bool32 {

	// Format severity
	var severity string
	if flags&vulkan.DebugReportFlags(vulkan.DebugReportErrorBit) != 0 {
		severity = "ERROR"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportWarningBit) != 0 {
		severity = "WARNING"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportPerformanceWarningBit) != 0 {
		severity = "PERFORMANCE"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportInformationBit) != 0 {
		severity = "INFO"
	} else if flags&vulkan.DebugReportFlags(vulkan.DebugReportDebugBit) != 0 {
		severity = "DEBUG"
	} else {
		severity = "UNKNOWN"
	}

	log.Printf("[%s] %s: %s\n", severity, pLayerPrefix, pMessage)

	// Return false to continue execution
	return vulkan.False
}

func CreateDebugCallback(instance vulkan.Instance) (vulkan.DebugReportCallback, error) {
	createInfo := vulkan.DebugReportCallbackCreateInfo{
		SType: vulkan.StructureTypeDebugReportCallbackCreateInfo,
		Flags: vulkan.DebugReportFlags(
			vulkan.DebugReportErrorBit |
				vulkan.DebugReportWarningBit |
				vulkan.DebugReportPerformanceWarningBit,
		),
		PfnCallback: debugCallback,
	}

	var debugCallback vulkan.DebugReportCallback
	res := vulkan.CreateDebugReportCallback(instance, &createInfo, nil, &debugCallback)
	if res != vulkan.Success {
		return nil, fmt.Errorf("failed to create debug callback: %v", res)
	}

	return debugCallback, nil
}

func DestroyDebugCallback(instance vulkan.Instance, callback vulkan.DebugReportCallback) {
	vulkan.DestroyDebugReportCallback(instance, callback, nil)
}

func CreateInstance() (vulkan.Instance, error) {
	// Load Vulkan library
	if err := vulkan.SetDefaultGetInstanceProcAddr(); err != nil {
		return nil, fmt.Errorf("failed to load Vulkan: %w", err)
	}
	vulkan.Init()

	appName := "Hammock app\x00"
	engName := "HammockGo\x00"

	appInfo := vulkan.ApplicationInfo{
		SType:              vulkan.StructureTypeApplicationInfo,
		PApplicationName:   appName,
		ApplicationVersion: vulkan.MakeVersion(1, 0, 0),
		PEngineName:        engName,
		EngineVersion:      vulkan.MakeVersion(1, 0, 0),
		ApiVersion:         vulkan.ApiVersion10,
	}

	extensions := []string{
		"VK_KHR_surface\x00",
		"VK_KHR_win32_surface\x00",
		"VK_EXT_debug_report\x00",
	}

	layers := []string{
		"VK_LAYER_KHRONOS_validation\x00",
	}

	instanceCreateInfo := vulkan.InstanceCreateInfo{
		SType:                   vulkan.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        &appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
	}

	var instance vulkan.Instance
	res := vulkan.CreateInstance(&instanceCreateInfo, nil, &instance)
	if res != vulkan.Success {
		return nil, fmt.Errorf("failed to create Vulkan instance: %v", res)
	}

	vulkan.InitInstance(instance)

	return instance, nil
}

func DestroyInstance(instance vulkan.Instance) {
	vulkan.DestroyInstance(instance, nil)
}
