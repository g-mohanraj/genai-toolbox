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

package dataplexcreateglossaryterm_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexcreateglossaryterm"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexCreateGlossaryTerm(t *testing.T) {
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
				create_glossary_term:
					kind: dataplex-create-glossary-term
					source: my-instance
					description: create a glossary term
			`,
			want: server.ToolConfigs{
				"create_glossary_term": dataplexcreateglossaryterm.Config{
					Name:         "create_glossary_term",
					Kind:         "dataplex-create-glossary-term",
					Source:       "my-instance",
					Description:  "create a glossary term",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				create_glossary_term:
					kind: dataplex-create-glossary-term
					source: my-instance
					description: create a glossary term
					parameters:
						- name: glossaryProject
							type: string
							description: glossary project
						- name: glossaryLocation
							type: string
							description: glossary location
						- name: glossaryId
							type: string
							description: glossary id
						- name: termId
							type: string
							description: term id
						- name: displayName
							type: string
							description: display name
							default: ""
						- name: description
							type: string
							description: description
							default: ""
						- name: labels
							type: map
							description: labels
							required: false
			`,
			want: server.ToolConfigs{
				"create_glossary_term": dataplexcreateglossaryterm.Config{
					Name:         "create_glossary_term",
					Kind:         "dataplex-create-glossary-term",
					Source:       "my-instance",
					Description:  "create a glossary term",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("glossaryProject", "glossary project"),
						parameters.NewStringParameter("glossaryLocation", "glossary location"),
						parameters.NewStringParameter("glossaryId", "glossary id"),
						parameters.NewStringParameter("termId", "term id"),
						parameters.NewStringParameterWithDefault("displayName", "", "display name"),
						parameters.NewStringParameterWithDefault("description", "", "description"),
						parameters.NewMapParameterWithRequired("labels", "labels", false, "string"),
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
