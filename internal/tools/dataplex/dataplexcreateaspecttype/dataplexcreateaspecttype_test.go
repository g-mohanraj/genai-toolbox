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

package dataplexcreateaspecttype_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexcreateaspecttype"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexCreateAspectType(t *testing.T) {
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
				create_aspect_type:
					kind: dataplex-create-aspect-type
					source: my-instance
					description: create an aspect type
			`,
			want: server.ToolConfigs{
				"create_aspect_type": dataplexcreateaspecttype.Config{
					Name:         "create_aspect_type",
					Kind:         "dataplex-create-aspect-type",
					Source:       "my-instance",
					Description:  "create an aspect type",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				create_aspect_type:
					kind: dataplex-create-aspect-type
					source: my-instance
					description: create an aspect type
					parameters:
						- name: location
							type: string
							description: aspect type location
						- name: aspectTypeId
							type: string
							description: aspect type id
						- name: aspectType
							type: map
							description: aspect type payload
							required: true
						- name: validateOnly
							type: boolean
							description: validate only
							default: false
			`,
			want: server.ToolConfigs{
				"create_aspect_type": dataplexcreateaspecttype.Config{
					Name:         "create_aspect_type",
					Kind:         "dataplex-create-aspect-type",
					Source:       "my-instance",
					Description:  "create an aspect type",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("location", "aspect type location"),
						parameters.NewStringParameter("aspectTypeId", "aspect type id"),
						parameters.NewMapParameterWithRequired("aspectType", "aspect type payload", true, ""),
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
