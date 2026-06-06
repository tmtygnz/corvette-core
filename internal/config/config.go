package config

type Config struct {
	AiScalingSize int          `toml:"ai_scaling_size"`
	Cameras       []CameraInfo `toml:"cameras"`
}

type CameraInfo struct {
	URL     string `toml:"url"`
	Type    string `toml:"type"`
	CamName string `toml:"cam_name"`
}
