package docker

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func pullImage(image string) error {
	ctx := context.Background()
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
	if !checkImage(image) {
		pullImage(image)
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

	log.Println("Contenedor finalizado")
	return nil
}

func checkImage(image string) bool {

	ctx := context.Background()
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
