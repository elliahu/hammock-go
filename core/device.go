package core

import (
	"fmt"
	"unsafe"

	"github.com/bbredesen/go-vk"
)

// QueueFamilyIndex is used to group existence of queue family and its index
type QueueFamilyIndex struct {
	hasValue bool
	index    uint32
}

// Picks an available physical device (GPU)
func PickPhysicalDevice(instance vk.Instance) (vk.PhysicalDevice, error) {

	devices, err := vk.EnumeratePhysicalDevices(instance)
	if err != nil {
		return vk.PhysicalDevice(vk.NULL_HANDLE), fmt.Errorf("failed to find GPUs with Vulkan support")
	}

	// Just pick the first device for simplicity
	physicalDevice := devices[0]

	// Get physical device properties
	props := vk.GetPhysicalDeviceProperties(physicalDevice)

	fmt.Printf("Selected GPU: %s\n", props.DeviceName)

	return physicalDevice, nil
}

// Get present, graphics, compute, transfer queue family indices in this order
func FindQueueFamilies(physicalDevice vk.PhysicalDevice, surface vk.SurfaceKHR) (QueueFamilyIndex, QueueFamilyIndex, QueueFamilyIndex, QueueFamilyIndex, error) {
	// Physical device queue family properties
	queueFamilies := vk.GetPhysicalDeviceQueueFamilyProperties(physicalDevice)

	// Don't have value by default
	presentQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	graphicsQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	computeQueueFamilyIndex := QueueFamilyIndex{hasValue: false}
	transferQueueFamilyIndex := QueueFamilyIndex{hasValue: false}

	// Find a queue family that supports graphics
	for i, queueFamily := range queueFamilies {
		isGraphics := (queueFamily.QueueFlags & vk.QUEUE_GRAPHICS_BIT) != 0
		isCompute := (queueFamily.QueueFlags & vk.QUEUE_COMPUTE_BIT) != 0
		isTransfer := (queueFamily.QueueFlags & vk.QUEUE_TRANSFER_BIT) != 0

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

		presentSupport, err := vk.GetPhysicalDeviceSurfaceSupportKHR(physicalDevice, uint32(i), surface)
		if err != nil {
			return presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex, fmt.Errorf("failed to query present support")
		}
		if presentSupport {
			presentQueueFamilyIndex = QueueFamilyIndex{hasValue: true, index: uint32(i)}
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

	return presentQueueFamilyIndex, graphicsQueueFamilyIndex, computeQueueFamilyIndex, transferQueueFamilyIndex, nil
}

// Creates a logical vulkan device
func CreateDevice(
	physicalDevice vk.PhysicalDevice,
	presentQueueFamilyIndex QueueFamilyIndex,
	graphicsQueueFamilyIndex QueueFamilyIndex,
	computeQueueFamilyIndex QueueFamilyIndex,
	transferQueueFamilyIndex QueueFamilyIndex) (vk.Device, vk.Queue, vk.Queue, vk.Queue, vk.Queue, error) {

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
	var queueCreateInfos []vk.DeviceQueueCreateInfo
	for queueFamily := range uniqueQueueFamilies {
		queueCreateInfo := vk.DeviceQueueCreateInfo{
			QueueFamilyIndex: queueFamily,
			PQueuePriorities: []float32{queuePriority},
		}
		queueCreateInfos = append(queueCreateInfos, queueCreateInfo)
	}

	// Device extensions
	deviceExtensions := []string{
		"VK_KHR_swapchain",
		"VK_KHR_synchronization2",
	}

	// Basic device features
	deviceFeatures := vk.PhysicalDeviceFeatures{}
	deviceFeatures.SamplerAnisotropy = true
	deviceFeatures.FillModeNonSolid = true

	// Sync 2 feature are core in 1.3

	// Descriptor indexing features
	descIndexFeatures := vk.PhysicalDeviceDescriptorIndexingFeatures{
		ShaderSampledImageArrayNonUniformIndexing:     true,
		ShaderUniformBufferArrayNonUniformIndexing:    true,
		RuntimeDescriptorArray:                        true,
		DescriptorBindingVariableDescriptorCount:      true,
		DescriptorBindingPartiallyBound:               true,
		DescriptorBindingSampledImageUpdateAfterBind:  true,
		DescriptorBindingUniformBufferUpdateAfterBind: true,
		ShaderStorageBufferArrayNonUniformIndexing:    true,
		DescriptorBindingStorageBufferUpdateAfterBind: true,
	}

	// Dynamic rendering features are core in 1.3

	// Features 2
	deviceFeatures2 := vk.PhysicalDeviceFeatures2{
		Features: deviceFeatures,
		PNext:    unsafe.Pointer(descIndexFeatures.Vulkanize()),
	}

	// Device create info
	deviceCreateInfo := vk.DeviceCreateInfo{
		PQueueCreateInfos:       queueCreateInfos,
		PpEnabledExtensionNames: deviceExtensions,
		PEnabledFeatures:        nil,
		PNext:                   unsafe.Pointer(deviceFeatures2.Vulkanize()),
	}

	device, err := vk.CreateDevice(physicalDevice, &deviceCreateInfo, nil)
	if err != nil {
		return vk.Device(vk.NULL_HANDLE),
			vk.Queue(vk.NULL_HANDLE),
			vk.Queue(vk.NULL_HANDLE),
			vk.Queue(vk.NULL_HANDLE),
			vk.Queue(vk.NULL_HANDLE),
			fmt.Errorf("failed to create logical device")
	}

	// Get the queues
	var presentQueue, graphicsQueue, computeQueue, transferQueue vk.Queue

	if presentQueueFamilyIndex.hasValue {
		presentQueue = vk.GetDeviceQueue(device, presentQueueFamilyIndex.index, 0)
	}
	if graphicsQueueFamilyIndex.hasValue {
		graphicsQueue = vk.GetDeviceQueue(device, graphicsQueueFamilyIndex.index, 0)
	}
	if computeQueueFamilyIndex.hasValue {
		computeQueue = vk.GetDeviceQueue(device, computeQueueFamilyIndex.index, 0)
	}
	if transferQueueFamilyIndex.hasValue {
		transferQueue = vk.GetDeviceQueue(device, transferQueueFamilyIndex.index, 0)
	}

	return device, presentQueue, graphicsQueue, computeQueue, transferQueue, nil
}

// Destroy logical device
func DestroyDevice(device vk.Device) {
	vk.DestroyDevice(device, nil)
}

// From list of candidate formats, this function picks first supported one
func PickSupportedFormat(physicalDevice vk.PhysicalDevice, candidates []vk.Format, tiling vk.ImageTiling, features vk.FormatFeatureFlags) (vk.Format, error) {
	for _, format := range candidates {
		props := vk.GetPhysicalDeviceFormatProperties(physicalDevice, format)

		if tiling == vk.IMAGE_TILING_LINEAR && (props.LinearTilingFeatures&features) == features {
			return format, nil
		}
		if tiling == vk.IMAGE_TILING_OPTIMAL && (props.OptimalTilingFeatures&features) == features {
			return format, nil
		}
	}

	return vk.FORMAT_UNDEFINED, fmt.Errorf("failed to find supported format")
}

// Find memory type that supports required properties
func FindMemoryType(physicalDevice vk.PhysicalDevice, typeFilter uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	memProperties := vk.GetPhysicalDeviceMemoryProperties(physicalDevice)
	for i := uint32(0); i < memProperties.MemoryTypeCount; i++ {
		if (typeFilter&(1<<i)) != 0 && (memProperties.MemoryTypes[i].PropertyFlags&properties) == properties {
			return i, nil
		}
	}
	return 0, fmt.Errorf("failed to find suitable memory type")
}

// Create command pool for each queue in this order: graphics, compute, transfer
func CreateCommandPools(
	device vk.Device,
	graphicsQueueFamilyIndex QueueFamilyIndex,
	computeQueueFamilyIndex QueueFamilyIndex,
	transferQueueFamilyIndex QueueFamilyIndex) (vk.CommandPool, vk.CommandPool, vk.CommandPool, error) {

	commandPoolCreateInfo := vk.CommandPoolCreateInfo{
		Flags:            vk.COMMAND_POOL_CREATE_TRANSIENT_BIT | vk.COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
		QueueFamilyIndex: graphicsQueueFamilyIndex.index,
	}

	graphicsCommandPool, err := vk.CreateCommandPool(device, &commandPoolCreateInfo, nil)
	if err != nil {
		return vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), fmt.Errorf("failed to create graphics command pool: %v")
	}

	commandPoolCreateInfo.QueueFamilyIndex = computeQueueFamilyIndex.index
	commandPoolCreateInfo.Flags = vk.COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT
	computeCommandPool, err := vk.CreateCommandPool(device, &commandPoolCreateInfo, nil)
	if err != nil {
		return vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), fmt.Errorf("failed to create compute command pool: %v")
	}

	commandPoolCreateInfo.QueueFamilyIndex = transferQueueFamilyIndex.index
	transferCommandPool, err := vk.CreateCommandPool(device, &commandPoolCreateInfo, nil)
	if err != nil {
		return vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), vk.CommandPool(vk.NULL_HANDLE), fmt.Errorf("failed to create transfer command pool: %v")
	}

	return graphicsCommandPool, computeCommandPool, transferCommandPool, nil
}

// Destroy command pools for queue families
func DestroyCommandPools(device vk.Device, graphicsCommandPool vk.CommandPool, computeCommandPool vk.CommandPool, transferCommandPool vk.CommandPool) {
	vk.DestroyCommandPool(device, graphicsCommandPool, nil)
	vk.DestroyCommandPool(device, computeCommandPool, nil)
	vk.DestroyCommandPool(device, transferCommandPool, nil)
}
