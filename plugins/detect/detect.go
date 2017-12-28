package detect

type Detector interface {
	Detect(key string) (responseBytes []byte, ok bool)
	Debug(action string)
	Info(keys map[string]string) (details map[string]interface{}, ok bool)
}

type detectors map[string]string
