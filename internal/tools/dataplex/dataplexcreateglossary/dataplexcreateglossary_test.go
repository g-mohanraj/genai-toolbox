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

package dataplexcreateglossary_test

import (
	"strings"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/sources"
	dataplexds "github.com/googleapis/genai-toolbox/internal/sources/dataplex"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools/dataplex/dataplexcreateglossary"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

func TestParseFromYamlDataplexCreateGlossary(t *testing.T) {
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
				create_glossary:
					kind: dataplex-create-glossary
					source: my-instance
					description: create a glossary
			`,
			want: server.ToolConfigs{
				"create_glossary": dataplexcreateglossary.Config{
					Name:         "create_glossary",
					Kind:         "dataplex-create-glossary",
					Source:       "my-instance",
					Description:  "create a glossary",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "advanced example",
			in: `
			tools:
				create_glossary:
					kind: dataplex-create-glossary
					source: my-instance
					description: create a glossary
					parameters:
						- name: location
							type: string
							description: glossary location
						- name: glossaryId
							type: string
							description: glossary id
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
						- name: validateOnly
							type: boolean
							description: validate only
							default: false
			`,
			want: server.ToolConfigs{
				"create_glossary": dataplexcreateglossary.Config{
					Name:         "create_glossary",
					Kind:         "dataplex-create-glossary",
					Source:       "my-instance",
					Description:  "create a glossary",
					AuthRequired: []string{},
					Parameters: []parameters.Parameter{
						parameters.NewStringParameter("location", "glossary location"),
						parameters.NewStringParameter("glossaryId", "glossary id"),
						parameters.NewStringParameterWithDefault("displayName", "", "display name"),
						parameters.NewStringParameterWithDefault("description", "", "glossary description"),
						parameters.NewMapParameterWithRequired("labels", "glossary labels", false, "string"),
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

func TestParseFromYamlDataplexCreateGlossaryMissingKind(t *testing.T) {
	ctx, err := testutils.ContextWithNewLogger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	in := `
	tools:
		create_glossary:
			source: my-instance
			description: create a glossary
	`
	got := struct {
		Tools server.ToolConfigs `yaml:"tools"`
	}{}
	err = yaml.UnmarshalContext(ctx, testutils.FormatYaml(in), &got)
	if err == nil {
		t.Fatal("expected error for missing kind, got nil")
	}
	if !strings.Contains(err.Error(), "missing 'kind' field") {
		t.Fatalf("expected missing kind error, got %q", err.Error())
	}
}

func TestConfigInitializeValidSource(t *testing.T) {
	cfg := dataplexcreateglossary.Config{
		Name:        "create_glossary",
		Kind:        "dataplex-create-glossary",
		Source:      "dataplex-source",
		Description: "create a glossary",
	}

	source := &dataplexds.Source{
		Config: dataplexds.Config{
			Name:    "dataplex-source",
			Kind:    dataplexds.SourceKind,
			Project: "test-project",
		},
	}

	tool, err := cfg.Initialize(map[string]sources.Source{
		"dataplex-source": source,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	dataplexTool, ok := tool.(dataplexcreateglossary.Tool)
	if !ok {
		t.Fatalf("expected Tool type, got %T", tool)
	}
	if dataplexTool.ProjectID != "test-project" {
		t.Errorf("expected project ID %q, got %q", "test-project", dataplexTool.ProjectID)
	}
	if dataplexTool.Name != "create_glossary" {
		t.Errorf("expected tool name %q, got %q", "create_glossary", dataplexTool.Name)
	}
}

func TestConfigInitializeMissingSource(t *testing.T) {
	cfg := dataplexcreateglossary.Config{
		Name:        "create_glossary",
		Kind:        "dataplex-create-glossary",
		Source:      "missing-source",
		Description: "create a glossary",
	}

	_, err := cfg.Initialize(map[string]sources.Source{})
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}

	expectedErr := `no source named "missing-source" configured`
	if err.Error() != expectedErr {
		t.Fatalf("expected error %q, got %q", expectedErr, err.Error())
	}
}

type mockSource struct{}

func (m *mockSource) SourceKind() string {
	return "mock"
}

func (m *mockSource) ToConfig() sources.SourceConfig {
	return nil
}

func TestConfigInitializeIncompatibleSource(t *testing.T) {
	cfg := dataplexcreateglossary.Config{
		Name:        "create_glossary",
		Kind:        "dataplex-create-glossary",
		Source:      "bad-source",
		Description: "create a glossary",
	}

	_, err := cfg.Initialize(map[string]sources.Source{
		"bad-source": &mockSource{},
	})
	if err == nil {
		t.Fatal("expected error for incompatible source, got nil")
	}
	if !strings.Contains(err.Error(), "invalid source for") {
		t.Fatalf("expected incompatible source error, got %q", err.Error())
	}
}
