package manifest

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Redirect struct {
	Host   string `yaml:host`
	Path   string `yaml:path`
	Target string `yaml:target`
}

type ManifestOptions struct {
	EnforceHttps bool   `yaml:enforcehttps`
	HealthCheck  string `yaml:healthCheck`
}

type Manifest struct {
	Redirects []Redirect      `yaml:redirects`
	Options   ManifestOptions `yaml:options`
}

func Load(path string, manifest *Manifest) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &manifest)
	if err != nil {
		return err
	}

	return nil
}
