package main

//Key тип данных ключа доступа
type Key struct {
	Key      string
	isEnable string
}

//Config основные параметры контроллера
type Config struct {
	SerialPort           string `json:"serialPort"`
	HTTPPort             string `json:"httpPort"`
	NormalModeEndpoint   string `json:"normalModeEndpoint"`
	HardLockModeEndpoint string `json:"hardLockModeEndpoint"`
	CloseEndpoint        string `json:"closeEndpoint"`
	OpenEndpoint         string `json:"openEndpoint"`
	AddKeyEndpoint       string `json:"addKeyEndpoint"`
	DeleteKeyEndpoint    string `json:"deleteKeyEndpoint"`
	ReadKeysEndpoint     string `json:"readKeysEndpoint"`
	LogFilePath          string `json:"logFilePath"`
}
