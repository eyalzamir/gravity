/*
Copyright 2019 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"fmt"
	"strings"

	"github.com/gravitational/trace"
	"helm.sh/helm/v3/pkg/repo"
)

// HasVars takes a slice of values and value files and returns true
// if there is a variable with the provided name among them.
func HasVar(name string, valueFiles valueFiles, values []string) (bool, error) {
	allVals, err := merge(valueFiles, values, nil, nil, "", "", "")
	if err != nil {
		return false, trace.Wrap(err)
	}
	return hasVar(strings.Split(name, "."), allVals), nil
}

func hasVar(name []string, vals map[string]interface{}) bool {
	if len(name) == 0 {
		return true
	}
	v, ok := vals[name[0]]
	if !ok {
		return false
	}
	if len(name) == 1 {
		return true
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	return hasVar(name[1:], m)
}

// ParseChartFilename returns chart name and version from the provided chart
// package filename generated by ToChartFilename function below.
func ParseChartFilename(filename string) (name, version string, err error) {
	parts := strings.Split(strings.TrimSuffix(filename, ".tgz"), "-")
	if len(parts) < 2 {
		return "", "", trace.BadParameter("bad chart filename: %v", filename)
	}
	return strings.Join(parts[:len(parts)-1], "-"), parts[len(parts)-1], nil
}

// ToChartFilename returns a chart archive filename for the provided name/version.
func ToChartFilename(name, version string) string {
	return fmt.Sprintf("%v-%v.tgz", name, version)
}

// CopyIndexFile returns a deep copy of the provided index file.
func CopyIndexFile(indexFile repo.IndexFile) *repo.IndexFile {
	newIndex := &repo.IndexFile{
		APIVersion: indexFile.APIVersion,
		Generated:  indexFile.Generated,
		Entries:    make(map[string]repo.ChartVersions),
		PublicKeys: indexFile.PublicKeys,
	}
	for chartName, chartVersions := range indexFile.Entries {
		for _, chartVersion := range chartVersions {
			chartVersionCopy := *chartVersion
			newIndex.Entries[chartName] = append(newIndex.Entries[chartName],
				&chartVersionCopy)
		}
	}
	return newIndex
}
