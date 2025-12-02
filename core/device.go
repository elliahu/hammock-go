package core

import (
	"fmt"

	"github.com/vulkan-go/vulkan"
)

type QueueFamilyIndex struct {
	hasValue bool
	index    uint32
}

func PickPhysicalDevice(instance vulkan.Instance) (vulkan.PhysicalDevice, error) {
	var deviceCount uint32
	vulkan.EnumeratePhysicalDevices(instance, &deviceCount, nil)

	if deviceCount == 0 {
		return nil, fmt.Errorf("failed to find GPUs with Vulkan support")
	}

	devices := make([]vulkan.PhysicalDevice, deviceCount)
	vulkan.EnumeratePhysicalDevices(instance, &deviceCount, devices)

	// Just pick the first device for simplicity
	physicalDevice := devices[0]

	var props vulkan.PhysicalDeviceProperties
	vulkan.GetPhysicalDeviceProperties(physicalDevice, &props)
	props.Deref()

	fmt.Printf("Selected GPU: %s\n", vulkan.ToString(props.DeviceName[:]))

	return physicalDevice, nil
}

// Get present, graphics, compute, transfer queue family indices in this order
func FindQueueFamilies(physicalDevice vulkan.PhysicalDevice, surface vulkan.Surface) (QueueFamilyIndex, QueueFamilyIndex, QueueFamilyIndex, QueueFamilyIndex) {
	var queueFamilyCount uint32
	vulkan.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &queueFamilyCount, nil)

	queueFamilies := make([]vulkan.QueueFamilyProperties, queueFamilyCount)
	vulkan.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &queueFamilyCount, queueFamilies)

	presentQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	graphicsQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	computeQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	transferQueueFamilyIndex := QueueFamilyIndex{hasValue: false}

	// Find a queue family that supports graphics
	for i, queueFamily := range queueFamilies {
		queueFamily.Deref()

		isGraphics := queueFamily.QueueFlags&vulkan.QueueFlags(vulkan.QueueGraphicsBit) != 0
		isCompute := queueFamily.QueueFlags&vulkan.QueueFlags(vulkan.QueueComputeBit) != 0
		isTransfer := queueFamily.QueueFlags&vulkan.QueueFlags(vulkan.QueueTransferBit) != 0

		// Find a queue that supports graphics
		if isGraphics {
			graphicsQueueFamilyIndex = QueueFamilyIndex{hasValue: true, index: uint32(i)}
		}
		// Find a dedicated compute queue that is separate from graphics
		if isCompute && !isGraphics {
			computeQueueFamilyIndex = QueueFamilyIndex{hasValue: true, index: uint32(i)}
		}
		// Find a dedicated transfer queue that is separate from graphics and compute
		if isTransfer && !isGraphics && !isCompute {
			transferQueueFamilyIndex = QueueFamilyIndex{hasValue: true, index: uint32(i)}
		}

		var presentSupport vulkan.Bool32 = vulkan.False
		if surface != nil {
			vulkan.GetPhysicalDeviceSurfaceSupport(physicalDevice, uint32(i), nil, &presentSupport)
			if presentSupport == vulkan.True {
				presentQueueFamilyIndex = QueueFamilyIndex{hasValue: true, index: uint32(i)}
			}
		}
	}

	// Compute fallback - use graphics if no dedicated compute queue
	if !computeQueueFamilyIndex.hasValue {
		computeQueueFamilyIndex = graphicsQueueFamilyIndex
		fmt.Println("No dedicated compute queue found; using graphics queue for compute.")
	}

	// Transfer fallback - use compute or graphics if no dedicated transfer queue
	if !transferQueueFamilyIndex.hasValue {
		if computeQueueFamilyIndex.hasValue && computeQueueFamilyIndex.index != graphicsQueueFamilyIndex.index {
			transferQueueFamilyIndex = computeQueueFamilyIndex
			fmt.Println("No dedicated transfer queue found; using compute queue for transfer.")
		} else {
			transferQueueFamilyIndex = graphicsQueueFamilyIndex
			fmt.Println("No dedicated transfer queue found; using graphics queue for transfer.")
		}
	}

	return presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex
}

func CreateDevice(
	physicalDevice vulkan.PhysicalDevice,
	presentQueueFamilyIndex QueueFamilyIndex,
	graphicsQueueFamilyIndex QueueFamilyIndex,
	computeQueueFamilyIndex QueueFamilyIndex,
	transferQueueFamilyIndex QueueFamilyIndex) (vulkan.Device, vulkan.Queue, vulkan.Queue, vulkan.Queue, vulkan.Queue, error) {

	queuePriority := float32(1.0)

	// Collect unique queue family indices
	uniqueQueueFamilies := make(map[uint32]bool)
	if presentQueueFamilyIndex.hasValue {
		uniqueQueueFamilies[presentQueueFamilyIndex.index] = true
	}
	if graphicsQueueFamilyIndex.hasValue {
		uniqueQueueFamilies[graphicsQueueFamilyIndex.index] = true
	}
	if computeQueueFamilyIndex.hasValue {
		uniqueQueueFamilies[computeQueueFamilyIndex.index] = true
	}
	if transferQueueFamilyIndex.hasValue {
		uniqueQueueFamilies[transferQueueFamilyIndex.index] = true
	}

	// Create queue create infos for each unique queue family
	var queueCreateInfos []vulkan.DeviceQueueCreateInfo
	for queueFamily := range uniqueQueueFamilies {
		queueCreateInfo := vulkan.DeviceQueueCreateInfo{
			SType:            vulkan.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: queueFamily,
			QueueCount:       1,
			PQueuePriorities: []float32{queuePriority},
		}
		queueCreateInfos = append(queueCreateInfos, queueCreateInfo)
	}

	// Device extensions
	deviceExtensions := []string{
		"VK_KHR_swapchain\x00",
	}

	// Basic device features
	deviceFeatures := vulkan.PhysicalDeviceFeatures{}
	deviceFeatures.SamplerAnisotropy = vulkan.True

	// Device create info
	deviceCreateInfo := vulkan.DeviceCreateInfo{
		SType:                   vulkan.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueCreateInfos)),
		PQueueCreateInfos:       queueCreateInfos,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
		PEnabledFeatures:        []vulkan.PhysicalDeviceFeatures{deviceFeatures},
	}

	var device vulkan.Device
	res := vulkan.CreateDevice(physicalDevice, &deviceCreateInfo, nil, &device)
	if res != vulkan.Success {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create logical device: %v", res)
	}

	// Get the queues
	var presentQueue, graphicsQueue, computeQueue, transferQueue vulkan.Queue

	if presentQueueFamilyIndex.hasValue {
		vulkan.GetDeviceQueue(device, presentQueueFamilyIndex.index, 0, &presentQueue)
	}
	if graphicsQueueFamilyIndex.hasValue {
		vulkan.GetDeviceQueue(device, graphicsQueueFamilyIndex.index, 0, &graphicsQueue)
	}
	if computeQueueFamilyIndex.hasValue {
		vulkan.GetDeviceQueue(device, computeQueueFamilyIndex.index, 0, &computeQueue)
	}
	if transferQueueFamilyIndex.hasValue {
		vulkan.GetDeviceQueue(device, transferQueueFamilyIndex.index, 0, &transferQueue)
	}

	return device, presentQueue, graphicsQueue, computeQueue, transferQueue, nil
}

func DestroyDevice(device vulkan.Device) {
	vulkan.DestroyDevice(device, nil)
}
