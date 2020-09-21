package imgtools

import (
	"context"
	"os"

	"github.com/marcelriegr/draide/pkg/ui"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// BuildOptions tbd
type BuildOptions struct {
	Dockerfile string
	Tags       []string
	Labels     map[string]string
	BuildArgs  map[string]string
	NoCache    bool
}

// Build a docker image
func Build(contextDir string, opts BuildOptions) {
	cli, err := client.NewEnvClient()
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed establishing connection to Docker engine")
	}

	contextDirTar, err := archive.TarWithOptions(contextDir, &archive.TarOptions{})
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed reading context directory")
	}

	response, err := cli.ImageBuild(context.Background(), contextDirTar, types.ImageBuildOptions{
		Dockerfile:     opts.Dockerfile,
		Tags:           opts.Tags,
		BuildArgs:      opts.BuildArgs,
		Labels:         opts.Labels,
		NoCache:        opts.NoCache,
		SuppressOutput: !ui.IsVerbose(),
	})
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed building image")
	}
	defer response.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, termFd, isTerm, nil)
}
