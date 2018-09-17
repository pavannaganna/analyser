package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pavannaganna/analyser/pkg/filesystem"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "analyser"
	app.Usage = "Analyse filesystem to fix space issues"
	app.Commands = []cli.Command{
		cli.Command{
			Name:        "filesystem",
			Category:    "filesystem",
			Description: "Analyse filestystem to fix space issues and errors",
			Subcommands: cli.Commands{
				cli.Command{
					Name:   "space",
					Action: filesystemSpace,
					Flags: []cli.Flag{
						cli.StringFlag{Name: "path, p"},
						cli.StringFlag{Name: "filter, f"},
					},
				},
				cli.Command{
					Name: "volume",
					Subcommands: cli.Commands{
						cli.Command{
							Name:   "scan",
							Action: fileSystemVolScanner,
							Flags: []cli.Flag{
								cli.StringFlag{
									Name: "path, p",
								},
							},
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("ERROR:", err)
	}

}

func filesystemSpace(c *cli.Context) error {
	filesystemPath := c.String("path")
	rootPath := c.String("root-path")
	filter := c.String("filter")
	var err error
	if filesystemPath == "" {
		err = errors.New("Path is not defined")
		return err
	}
	inputs := filesystem.SpaceInputs{
		FilesystemPath: filesystemPath,
		RootFilesystem: rootPath,
		Filter:         filter,
	}
	return filesystem.Space(inputs)
}

func fileSystemVolScanner(c *cli.Context) error {
	filesystemPath := c.String("path")
	var err error
	if filesystemPath == "" {
		err = errors.New("Path is not defined")
		return err
	}
	return filesystem.VolumeScanner(filesystemPath)
}
