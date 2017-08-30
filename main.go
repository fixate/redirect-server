package main

import (
	"fmt"
	"log"
	"os"

	mfst "github.com/fixate/redirect-server/manifest"
	srv "github.com/fixate/redirect-server/server"

	"github.com/urfave/cli"
)

const version string = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "redirect server"
	app.Version = version
	app.Usage = "Redirect according to yaml config"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "m, manifest",
			Usage:  "path to the redirect manifest",
			EnvVar: "REDIRECT_MANIFEST_PATH",
		},
		cli.StringFlag{
			Name:   "b, bind",
			Usage:  "bind address",
			EnvVar: "REDIRECT_BIND_ADDRESS",
			Value:  "0.0.0.0:3000",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	manifest := mfst.Manifest{}
	manifestPath := c.String("manifest")
	if len(manifestPath) == 0 {
		return fmt.Errorf("Manifest path not specified")
	}

	if err := mfst.Load(manifestPath, &manifest); err != nil {
		return err
	}
	bind := c.String("bind")
	log.Printf("Starting server on %s\n", bind)
	server := srv.NewServer(&srv.ServerOptions{
		Manifest: &manifest,
		Bind:     bind,
	})

	panic(server.ListenAndServe())
	return nil
}
