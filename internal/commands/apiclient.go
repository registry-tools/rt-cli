package commands

import (
	"log"
	"os"

	userconfig "github.com/registry-tools/rt-cli/internal/userconfig"
	sdk "github.com/registry-tools/rt-sdk"
)

func GetSDK() (sdk.SDK, error) {
	host := os.Getenv("REGISTRY_TOOLS_HOSTNAME")
	if host == "" {
		host = DefaultHostname
	}

	configuredByUserConfig := false
	var token string
	userconfig, err := userconfig.LoadFromUserConfigDirectory()
	if err != nil {
		return nil, err
	}

	token, configuredByUserConfig = userconfig.GetHostToken(host)
	envClientID := os.Getenv("REGISTRY_TOOLS_CLIENT_ID")
	envClientSecret := os.Getenv("REGISTRY_TOOLS_CLIENT_SECRET")

	if envClientID != "" && envClientSecret != "" {
		log.Printf("[TRACE] Initializing SDK using client ID and secret from environment")
		return sdk.NewSDK(host, envClientID, envClientSecret)
	} else if configuredByUserConfig {
		log.Printf("[TRACE] Initializing SDK using token from user config")
		return sdk.NewSDKWithAccessToken(host, token)
	}

	return nil, ErrLoginRequired
}
