package imgtools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// AuthConfig tbd
type AuthConfig struct {
	Username string
	Password string
}

// PushOptions tbd
type PushOptions struct {
	Auth AuthConfig
}

// Push a docker image
func Push(imageName string, opts PushOptions) {
	cli, err := client.NewEnvClient()
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed establishing connection to Docker engine")
	}

	authConfigAsBytes, err := json.Marshal(types.AuthConfig{
		Username: opts.Auth.Username,
		Password: opts.Auth.Password,
	})
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed encoding credentials")
	}

	response, err := cli.ImagePush(context.Background(), imageName, types.ImagePushOptions{
		RegistryAuth: base64.URLEncoding.EncodeToString(authConfigAsBytes),
	})
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed pushing image")
	}
	defer response.Close()

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(response, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		ui.ErrorAndExit(1, err.Error())
	}
}
