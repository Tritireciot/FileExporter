package config

var MessAppSrv struct {
	HttpPort int `json:"port"`
	WsPort   int `json:"wsPort"`
}
