/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package home

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/cloudfoundry/libcfbuildpack/v2/build"
	"github.com/cloudfoundry/libcfbuildpack/v2/helper"
	"github.com/cloudfoundry/libcfbuildpack/v2/layers"
	"github.com/cloudfoundry/tomcat-cnb/internal"
)

// TomcatDependency indicates that Tomcat is required for the web application.
const TomcatDependency = "tomcat"

type Home struct {
	layer  layers.DependencyLayer
	layers layers.Layers
}

func (h Home) Contribute() error {
	if err := h.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Extracting to %s", layer.Root)

		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := modifyCatalinaStart(layer); err != nil {
			return err
		}

		return layer.OverrideLaunchEnv("CATALINA_HOME", layer.Root)
	}, layers.Launch); err != nil {
		return err
	}

	command := "catalina.sh run"

	return h.layers.WriteApplicationMetadata(layers.Metadata{
		Processes: layers.Processes{
			{Type: "task", Command: command},
			{Type: "tomcat", Command: command},
			{Type: "web", Command: command},
		},
	})
}

// NewHome creates a new CATALINA_HOME instance.
func NewHome(build build.Build) (Home, error) {
	p, _, err := build.Plans.GetShallowMerged(TomcatDependency)
	if err != nil {
		return Home{}, err
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return Home{}, err
	}

	version, err := internal.Version(TomcatDependency, p, build.Buildpack)
	if err != nil {
		return Home{}, err
	}

	dep, err := deps.Best(TomcatDependency, version, build.Stack)
	if err != nil {
		return Home{}, err
	}

	return Home{
		build.Layers.DependencyLayer(dep),
		build.Layers,
	}, nil
}

var pattern = regexp.MustCompile(`\n\s*CLASSPATH=\s*\n`)

func modifyCatalinaStart(layer layers.DependencyLayer) error {
	layer.Logger.Body("Modifying catalina.sh")

	filename := filepath.Join(layer.Root, "bin", "catalina.sh")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	hit := pattern.Find(content)
	if hit == nil {
		layer.Logger.Body("No replacements needed")
		err = nil
	} else {
		layer.Logger.Body("Replacements needed")
		modifiedContent := pattern.ReplaceAll(content, []byte("\n#CLASSPATH=\n"))
		err = ioutil.WriteFile(filename, modifiedContent, 0755)
	}

	return err
}
