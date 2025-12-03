// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dataplexlistglossaries_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexlistglossaries"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexListGlossaries(t *testing.T) {
	ctx, err := testutils.ContextWithNewLogger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	tcs := []struct {
		desc string
		in   string
		want server.ToolConfigs
	}{
		{
			desc: "basic example",
			in: `
			tools:
				list_glossaries:
					kind: dataplex-list-glossaries
					source: my-instance
					description: list glossaries
			`,
			want: server.ToolConfigs{
				"list_glossaries": dataplexlistglossaries.Config{
					Name:         "list_glossaries",
					Kind:         "dataplex-list-glossaries",
					Source:       "my-instance",
					Description:  "list glossaries",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				list_glossaries:
					kind: dataplex-list-glossaries
					source: my-instance
					description: list glossaries
					parameters:
						- name: location
							type: string
							description: glossary location
						- name: pageSize
							type: integer
							description: page size
							default: 50
						- name: pageToken
							type: string
							description: page token
							default: ""
						- name: filter
							type: string
							description: filter
							default: ""
						- name: orderBy
							type: string
							description: order by
							default: ""
			`,
			want: server.ToolConfigs{
				"list_glossaries": dataplexlistglossaries.Config{
					Name:         "list_glossaries",
					Kind:         "dataplex-list-glossaries",
					Source:       "my-instance",
					Description:  "list glossaries",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("location", "glossary location"),
						parameters.NewIntParameterWithDefault("pageSize", 50, "page size"),
						parameters.NewStringParameterWithDefault("pageToken", "", "page token"),
						parameters.NewStringParameterWithDefault("filter", "", "filter"),
						parameters.NewStringParameterWithDefault("orderBy", "", "order by"),
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Tools server.ToolConfigs `yaml:"tools"`
			}{}
			if err := yaml.UnmarshalContext(ctx, testutils.FormatYaml(tc.in), &got); err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if diff := cmp.Diff(tc.want, got.Tools); diff != "" {
				t.Fatalf("incorrect parse: diff %v", diff)
			}
		})
	}
}
