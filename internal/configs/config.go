package configs

type Config struct {
	Cameras       []Cameras
	Password      string
	Onnxlibpath   string
	Onnxmodelpath string
}

type Cameras struct {
	IP       string
	User     string
	Password string
	Endpoint string
	Port     int
	Name     string
	Type     string
	Url      string
}
