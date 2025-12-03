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

package dataplexupdateentry_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexupdateentry"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexUpdateEntry(t *testing.T) {
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
				example_tool:
					kind: dataplex-update-entry
					source: my-instance
					description: some description
			`,
			want: server.ToolConfigs{
				"example_tool": dataplexupdateentry.Config{
					Name:         "example_tool",
					Kind:         "dataplex-update-entry",
					Source:       "my-instance",
					Description:  "some description",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				example_tool:
					kind: dataplex-update-entry
					source: my-instance
					description: some description
					parameters:
						- name: entry
							type: string
							description: entry description
						- name: description
							type: string
							description: entry source description
							default: ""
						- name: contacts
							type: map
							description: contacts payload
							required: false
						- name: contactsAspectType
							type: string
							description: contacts aspect type
							default: projects/dataplex-types/locations/global/aspectTypes/contacts
						- name: path
							type: string
							description: aspect path
							default: ""
						- name: allowMissingEntry
							type: boolean
							description: allow missing entry
							default: false
						- name: deleteMissingAspects
							type: boolean
							description: delete missing aspects
							default: false
			`,
			want: server.ToolConfigs{
				"example_tool": dataplexupdateentry.Config{
					Name:         "example_tool",
					Kind:         "dataplex-update-entry",
					Source:       "my-instance",
					Description:  "some description",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("entry", "entry description"),
						parameters.NewStringParameterWithDefault("description", "", "entry source description"),
						parameters.NewMapParameterWithRequired("contacts", "contacts payload", false, ""),
						parameters.NewStringParameterWithDefault("contactsAspectType", "projects/dataplex-types/locations/global/aspectTypes/contacts", "contacts aspect type"),
						parameters.NewStringParameterWithDefault("path", "", "aspect path"),
						parameters.NewBooleanParameterWithDefault("allowMissingEntry", false, "allow missing entry"),
						parameters.NewBooleanParameterWithDefault("deleteMissingAspects", false, "delete missing aspects"),
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
			err := yaml.UnmarshalContext(ctx, testutils.FormatYaml(tc.in), &got)
			if err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if diff := cmp.Diff(tc.want, got.Tools); diff != "" {
				t.Fatalf("incorrect parse: diff %v", diff)
			}
		})
	}
}
