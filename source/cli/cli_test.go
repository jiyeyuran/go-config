package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jiyeyuran/go-config"
	"github.com/jiyeyuran/go-config/source"
	"github.com/urfave/cli/v2"
)

func TestCliSourceDefault(t *testing.T) {
	const expVal string = "flagvalue"

	var cliSrc source.Source

	app := &cli.App{
		Flags: []cli.Flag{
			// to be able to run inside go test
			&cli.StringFlag{
				Name: "test.timeout",
			},
			&cli.BoolFlag{
				Name: "test.v",
			},
			&cli.StringFlag{
				Name: "test.run",
			},
			&cli.StringFlag{
				Name: "test.testlogfile",
			},
			&cli.StringFlag{
				Name: "test.paniconexit0",
			},
			&cli.StringFlag{
				Name:    "flag",
				Usage:   "It changes something",
				EnvVars: []string{"flag"},
				Value:   expVal,
			},
		},
		Action: func(c *cli.Context) error {
			cliSrc = NewSource(c)
			return nil
		},
	}

	// run app
	app.Run(os.Args)

	config.Load(cliSrc)
	if fval := config.Get("flag").String("default"); fval != expVal {
		t.Fatalf("default flag value not loaded %v != %v", fval, expVal)
	}
}

func TestCliSource(t *testing.T) {
	var src source.Source

	// setup app
	app := &cli.App{
		Name: "testapp",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "db-host",
				EnvVars: []string{"db-host"},
				Value:   "myval",
			},
		},
		Action: func(c *cli.Context) error {
			src = NewSource(c)
			return nil
		},
	}

	// run app
	app.Run([]string{"run", "-db-host", "localhost"})

	// test config
	c, err := src.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data, &actual); err != nil {
		t.Error(err)
	}

	actualDB := actual["db"].(map[string]interface{})
	if actualDB["host"] != "localhost" {
		t.Errorf("expected localhost, got %v", actualDB["name"])
	}
}
