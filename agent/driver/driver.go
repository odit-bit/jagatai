package driver

type Config struct {
	Endpoint    string
	TopK        *float32
	TopP        *float32
	Temperature *float32
	MinP        *float32
}
