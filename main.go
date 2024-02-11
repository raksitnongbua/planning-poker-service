package main

import (
	"github.com/raksitnongbua/planning-poker-service/configs"
	"github.com/raksitnongbua/planning-poker-service/protocol"
)

func main() {
	configs.Init()
	protocol.ServeREST()
}
