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

package dataplexcreateglossaryterm

import (
	"context"
	"fmt"

	dataplexapi "cloud.google.com/go/dataplex/apiv1"
	dataplexpb "cloud.google.com/go/dataplex/apiv1/dataplexpb"
	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	dataplexds "github.com/googleapis/genai-toolbox/internal/sources/dataplex"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

const kind string = "dataplex-create-glossary-term"

func init() {
	if !tools.Register(kind, newConfig) {
		panic(fmt.Sprintf("tool kind %q already registered", kind))
	}
}

func newConfig(ctx context.Context, name string, decoder *yaml.Decoder) (tools.ToolConfig, error) {
	actual := Config{Name: name}
	if err := decoder.DecodeContext(ctx, &actual); err != nil {
		return nil, err
	}
	return actual, nil
}

type compatibleSource interface {
	BusinessGlossaryClient() *dataplexapi.BusinessGlossaryClient
}

// validate compatible sources are still compatible
var _ compatibleSource = &dataplexds.Source{}

var compatibleSources = [...]string{dataplexds.SourceKind}

type Config struct {
	Name         string                `yaml:"name" validate:"required"`
	Kind         string                `yaml:"kind" validate:"required"`
	Source       string                `yaml:"source" validate:"required"`
	Description  string                `yaml:"description"`
	AuthRequired []string              `yaml:"authRequired"`
	Parameters   parameters.Parameters `yaml:"parameters"`
}

// validate interface
var _ tools.ToolConfig = Config{}

func (cfg Config) ToolConfigKind() string {
	return kind
}

func (cfg Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	rawS, ok := srcs[cfg.Source]
	if !ok {
		return nil, fmt.Errorf("no source named %q configured", cfg.Source)
	}

	s, ok := rawS.(compatibleSource)
	if !ok {
		return nil, fmt.Errorf("invalid source for %q tool: source kind must be one of %q", kind, compatibleSources)
	}

	glossaryProject := parameters.NewStringParameter("glossaryProject", "Project containing the glossary.")
	glossaryLocation := parameters.NewStringParameter("glossaryLocation", "Location of the glossary (often global).")
	glossaryId := parameters.NewStringParameter("glossaryId", "Identifier of the glossary that will own the term.")
	termId := parameters.NewStringParameter("termId", "Identifier for the term to create.")
	displayName := parameters.NewStringParameterWithDefault("displayName", "", "Optional display name for the term.")
	description := parameters.NewStringParameterWithDefault("description", "", "Optional description for the term.")
	labels := parameters.NewMapParameterWithRequired("labels", "Optional labels map (string values) to set on the term.", false, "string")

	params := parameters.Parameters{glossaryProject, glossaryLocation, glossaryId, termId, displayName, description, labels}
	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, params)

	t := Tool{
		Config:      cfg,
		Parameters:  params,
		Glossary:    s.BusinessGlossaryClient(),
		manifest:    tools.Manifest{Description: cfg.Description, Parameters: params.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest: mcpManifest,
	}
	return t, nil
}

type Tool struct {
	Config
	Parameters  parameters.Parameters
	Glossary    *dataplexapi.BusinessGlossaryClient
	manifest    tools.Manifest
	mcpManifest tools.McpManifest
}

func (t Tool) ToConfig() tools.ToolConfig {
	return t.Config
}

func (t Tool) Invoke(ctx context.Context, params parameters.ParamValues, accessToken tools.AccessToken) (any, error) {
	m := params.AsMap()

	gProj, _ := m["glossaryProject"].(string)
	gLoc, _ := m["glossaryLocation"].(string)
	gID, _ := m["glossaryId"].(string)
	termID, _ := m["termId"].(string)

	displayName, _ := m["displayName"].(string)
	description, _ := m["description"].(string)

	labelsRaw, labelsProvided := m["labels"]
	var labels map[string]string
	if labelsProvided && labelsRaw != nil {
		rawMap, ok := labelsRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("labels must be an object with string values")
		}
		labels = make(map[string]string, len(rawMap))
		for k, v := range rawMap {
			str, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("label value for key %q must be a string", k)
			}
			labels[k] = str
		}
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/glossaries/%s", gProj, gLoc, gID)
	term := &dataplexpb.GlossaryTerm{
		DisplayName: displayName,
		Description: description,
		Parent:      parent,
	}
	if labelsProvided {
		term.Labels = labels
	}

	req := &dataplexpb.CreateGlossaryTermRequest{
		Parent: parent,
		TermId: termID,
		Term:   term,
	}

	return t.Glossary.CreateGlossaryTerm(ctx, req)
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (parameters.ParamValues, error) {
	return parameters.ParseParams(t.Parameters, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t Tool) Authorized(verifiedAuthServices []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthServices)
}

func (t Tool) RequiresClientAuthorization() bool {
	return false
}

func (t Tool) GetAuthTokenHeaderName() string {
	return "Authorization"
}
