package main

import (
	"hammock-go/core"
)

func main() {
	instance, err := core.CreateInstance()
	if err != nil {
		panic(err)
	}
	defer core.DestroyInstance(instance)

	debugCallback, err := core.CreateDebugCallback(instance)
	if err != nil {
		panic(err)
	}
	defer core.DestroyDebugCallback(instance, debugCallback)

	println("Vulkan instance and debug callback created successfully!")

	physicalDevice, err := core.PickPhysicalDevice(instance)
	if err != nil {
		panic(err)
	}
	presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex := core.FindQueueFamilies(physicalDevice, nil)

	device, _, _, _, _, err := core.CreateDevice(physicalDevice, presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex)
	if err != nil {
		panic(err)
	}

	defer core.DestroyDevice(device)

	println("Logical device and graphics queue created successfully!")

}
