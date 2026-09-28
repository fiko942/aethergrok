package system

// SystemVolumeState captures the volume level and mute status before ducking/muting.
type SystemVolumeState struct {
	OriginalVolume int  `json:"originalVolume"`
	WasMuted       bool `json:"wasMuted"`
}
