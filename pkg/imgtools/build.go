package imgtools

import (
	"bytes"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/marcelriegr/draide/pkg/ui"
)

// BuildOptions tbd
type BuildOptions struct {
	Dockerfile string
	Tags       []string
	Labels     map[string]string
	BuildArgs  map[string]string
}

// Build a docker image
func Build(imageName string, contextDir string, opts BuildOptions) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed establishing connection to Docker engine")
	}

	var buildArgs []docker.BuildArg
	if len(opts.BuildArgs) > 0 {
		buildArgs = make([]docker.BuildArg, len(opts.BuildArgs))
		i := 0
		for k, v := range opts.BuildArgs {
			buildArgs[i] = docker.BuildArg{
				Name:  k,
				Value: v,
			}
			i++
		}
	}

	var buf bytes.Buffer
	err = client.BuildImage(docker.BuildImageOptions{
		Name:         imageName,
		NoCache:      true,
		ContextDir:   contextDir,
		OutputStream: &buf,
		Dockerfile:   opts.Dockerfile,
		Labels:       opts.Labels,
		BuildArgs:    buildArgs,
	})
	if err != nil {
		ui.Log(err.Error())
		ui.ErrorAndExit(1, "Failed building image")
	}
	ui.Logf(buf.String())
	ui.Success("Image built successfully")

	ui.Info("Tagging image...")
	for _, tag := range opts.Tags {
		err = client.TagImage(imageName, docker.TagImageOptions{
			Repo: imageName,
			Tag:  tag,
		})
		if err != nil {
			ui.Log(err.Error())
			ui.ErrorAndExit(1, "Failed tagging image as %s:%s", imageName, tag)
		}
		ui.Success("Image tagged as %s:%s", imageName, tag)
	}

}
