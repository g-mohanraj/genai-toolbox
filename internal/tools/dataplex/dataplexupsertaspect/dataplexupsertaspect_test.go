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

package dataplexupsertaspect_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexupsertaspect"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexUpsertAspect(t *testing.T) {
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
					kind: dataplex-upsert-aspect
					source: my-instance
					description: some description
			`,
			want: server.ToolConfigs{
				"example_tool": dataplexupsertaspect.Config{
					Name:         "example_tool",
					Kind:         "dataplex-upsert-aspect",
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
					kind: dataplex-upsert-aspect
					source: my-instance
					description: some description
					parameters:
						- name: entry
							type: string
							description: entry description
						- name: aspectType
							type: string
							description: aspect type description
						- name: path
							type: string
							description: path description
							default: ""
						- name: data
							type: map
							description: aspect payload
							required: true
						- name: deleteMissingAspects
							type: boolean
							description: delete missing description
							default: false
						- name: allowMissingEntry
							type: boolean
							description: allow missing description
							default: false
			`,
			want: server.ToolConfigs{
				"example_tool": dataplexupsertaspect.Config{
					Name:         "example_tool",
					Kind:         "dataplex-upsert-aspect",
					Source:       "my-instance",
					Description:  "some description",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("entry", "entry description"),
						parameters.NewStringParameter("aspectType", "aspect type description"),
						parameters.NewStringParameterWithDefault("path", "", "path description"),
						parameters.NewMapParameterWithRequired("data", "aspect payload", true, ""),
						parameters.NewBooleanParameterWithDefault("deleteMissingAspects", false, "delete missing description"),
						parameters.NewBooleanParameterWithDefault("allowMissingEntry", false, "allow missing description"),
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

func TestBuildAspectKey(t *testing.T) {
	tcs := []struct {
		name      string
		aspect    string
		path      string
		want      string
		expectErr bool
	}{
		{
			name:   "dotted reference without path",
			aspect: "dataplex-types.global.schema",
			want:   "dataplex-types.global.schema",
		},
		{
			name:   "resource name with path",
			aspect: "projects/foo/locations/us/aspectTypes/bar",
			path:   "Schema.column",
			want:   "foo.us.bar@Schema.column",
		},
		{
			name:      "invalid format",
			aspect:    "justType",
			expectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := dataplexupsertaspect.BuildAspectKeyForTest(tc.aspect, tc.path)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("unexpected aspect key (-want +got): %s", diff)
			}
		})
	}
}
