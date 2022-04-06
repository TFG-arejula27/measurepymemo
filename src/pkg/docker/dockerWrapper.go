package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func pullImage(image string, ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	_, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	return nil

}

func RunContainer(image string) error {
	ctx := context.Background()
	if !checkImage(image, ctx) {
		pullImage(image, ctx)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "arejula27/pymemo:test",
		Cmd:   []string{"./do_run.sh"},
		Tty:   false,
	}, &container.HostConfig{AutoRemove: true}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func checkImage(image string, ctx context.Context) bool {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return false
	}

	for _, img := range images {

		if len(img.RepoTags) > 0 && img.RepoTags[0] == image {
			return true
		}
	}

	return false

}
