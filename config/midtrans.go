package config

import (
	midtrans "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"os"
)

type MidtransConfig struct {
	SnapClient *snap.Client
}

func NewMidtrans() *MidtransConfig {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	env := midtrans.Sandbox
	if os.Getenv("MIDTRANS_ENV") == "production" {
		env = midtrans.Production
	}
	snapClient := snap.Client{}
	snapClient.New(serverKey, env)
	return &MidtransConfig{
		SnapClient: &snapClient,
	}
}