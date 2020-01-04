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

package home_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/tomcat-cnb/home"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestHome(t *testing.T) {
	spec.Run(t, "Home", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("returns true with WEB-INF", func() {
			f.AddDependency("tomcat", filepath.Join("testdata", "stub-tomcat.tar.gz"))
			if err := os.MkdirAll(filepath.Join(f.Build.Application.Root, "WEB-INF"), 0755); err != nil {
				t.Fatal(err)
			}

			h, err := home.NewHome(f.Build)
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(h.Contribute()).To(gomega.Succeed())

			layer := f.Build.Layers.Layer("tomcat")
			g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			g.Expect(filepath.Join(layer.Root, "fixture-marker")).To(gomega.BeARegularFile())
			g.Expect(layer).To(test.HaveOverrideLaunchEnvironment("CATALINA_HOME", layer.Root))

			command := "catalina.sh run"
			g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Processes: []layers.Process{
					{Type: "task", Command: command},
					{Type: "tomcat", Command: command},
					{Type: "web", Command: command},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
