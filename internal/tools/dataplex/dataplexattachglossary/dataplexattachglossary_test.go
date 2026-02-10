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

package dataplexattachglossary_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexattachglossary"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexAttachGlossary(t *testing.T) {
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
				attach_glossary:
					kind: dataplex-attach-glossary
					source: my-instance
					description: attach glossary term
			`,
			want: server.ToolConfigs{
				"attach_glossary": dataplexattachglossary.Config{
					Name:         "attach_glossary",
					Kind:         "dataplex-attach-glossary",
					Source:       "my-instance",
					Description:  "attach glossary term",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				attach_glossary:
					kind: dataplex-attach-glossary
					source: my-instance
					description: attach glossary term
					parameters:
						- name: parent
							type: string
							description: parent path
						- name: entryLinkId
							type: string
							description: link id
						- name: sourceEntry
							type: string
							description: source entry
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
						- name: entryLinkType
							type: string
							description: entry link type
							default: projects/dataplex-types/locations/global/entryLinkTypes/definition
			`,
			want: server.ToolConfigs{
				"attach_glossary": dataplexattachglossary.Config{
					Name:         "attach_glossary",
					Kind:         "dataplex-attach-glossary",
					Source:       "my-instance",
					Description:  "attach glossary term",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("parent", "parent path"),
						parameters.NewStringParameter("entryLinkId", "link id"),
						parameters.NewStringParameter("sourceEntry", "source entry"),
						parameters.NewStringParameter("glossaryProject", "glossary project"),
						parameters.NewStringParameter("glossaryLocation", "glossary location"),
						parameters.NewStringParameter("glossaryId", "glossary id"),
						parameters.NewStringParameter("termId", "term id"),
						parameters.NewStringParameterWithDefault("entryLinkType", "projects/dataplex-types/locations/global/entryLinkTypes/definition", "entry link type"),
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
