package main

import (
	"hammock-go/core"
)

func main() {
	// Create Vulkan instance
	instance, err := core.CreateInstance()
	if err != nil {
		panic(err)
	}
	defer core.DestroyInstance(instance)

	// Create debug messenger
	debugCallback, err := core.CreateDebugCallback(instance)
	if err != nil {
		panic(err)
	}
	defer core.DestroyDebugCallback(instance, debugCallback)

	// Physical device
	physicalDevice, err := core.PickPhysicalDevice(instance)
	if err != nil {
		panic(err)
	}

	// Find queue families
	presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex := core.FindQueueFamilies(physicalDevice, nil)

	// Create logical device
	device, _, _, _, _, err := core.CreateDevice(physicalDevice, presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex)
	if err != nil {
		panic(err)
	}

	defer core.DestroyDevice(device)

	// Create command pools
	graphicsCommandPool, computeCommandPool, transferCommandPool, err := core.CreateCommandPools(device, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex)
	if err != nil {
		panic(err)
	}

	defer core.DestroyCommandPools(device, graphicsCommandPool, computeCommandPool, transferCommandPool)
}
