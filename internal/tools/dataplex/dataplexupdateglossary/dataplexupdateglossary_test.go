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

package dataplexupdateglossary_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexupdateglossary"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexUpdateGlossary(t *testing.T) {
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
				update_glossary:
					kind: dataplex-update-glossary
					source: my-instance
					description: update a glossary
			`,
			want: server.ToolConfigs{
				"update_glossary": dataplexupdateglossary.Config{
					Name:         "update_glossary",
					Kind:         "dataplex-update-glossary",
					Source:       "my-instance",
					Description:  "update a glossary",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				update_glossary:
					kind: dataplex-update-glossary
					source: my-instance
					description: update a glossary
					parameters:
						- name: name
							type: string
							description: glossary name
						- name: displayName
							type: string
							description: display name
							default: ""
						- name: description
							type: string
							description: glossary description
							default: ""
						- name: labels
							type: map
							description: glossary labels
							required: false
						- name: etag
							type: string
							description: glossary etag
							default: ""
						- name: validateOnly
							type: boolean
							description: validate only
							default: false
			`,
			want: server.ToolConfigs{
				"update_glossary": dataplexupdateglossary.Config{
					Name:         "update_glossary",
					Kind:         "dataplex-update-glossary",
					Source:       "my-instance",
					Description:  "update a glossary",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("name", "glossary name"),
						parameters.NewStringParameterWithDefault("displayName", "", "display name"),
						parameters.NewStringParameterWithDefault("description", "", "glossary description"),
						parameters.NewMapParameterWithRequired("labels", "glossary labels", false, "string"),
						parameters.NewStringParameterWithDefault("etag", "", "glossary etag"),
						parameters.NewBooleanParameterWithDefault("validateOnly", false, "validate only"),
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
