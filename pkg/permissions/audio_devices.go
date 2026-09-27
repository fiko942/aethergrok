package permissions

// AudioDeviceInfo represents a hardware or virtual audio input device
type AudioDeviceInfo struct {
	Name         string `json:"name"`
	IsDefault    bool   `json:"isDefault"`
	Transport    string `json:"transport"`
	Manufacturer string `json:"manufacturer"`
}
