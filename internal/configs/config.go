package configs

type Config struct {
	Cameras         []Camera `toml:"cameras"`
	Password        string   `toml:"password"`
	OnnxLibPath     string   `toml:"onnx_lib_path"`
	OnnxModelPath   string   `toml:"onnx_model_path"`
	AiScalingWidth  int      `toml:"ai_scaling_width"`
	AiScalingHeight int      `toml:"ai_scaling_height"`
}

type Camera struct {
	IP       string `toml:"ip"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Endpoint string `toml:"endpoint"`
	Port     int    `toml:"port"`
	Name     string `toml:"name"`
	Type     string `toml:"type"`
	URL      string `toml:"url"`
}
