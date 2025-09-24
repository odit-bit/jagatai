package driver

type Config struct {

	//Optional. It can be different for each driver.
	Endpoint    string
	TopK        *float32 ``
	TopP        *float32
	Temperature *float32
	MinP        *float32
}
