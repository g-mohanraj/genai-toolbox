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

package dataplexcreateglossary

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

const kind string = "dataplex-create-glossary"

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
	CatalogClient() *dataplexapi.CatalogClient
	BusinessGlossaryClient() *dataplexapi.BusinessGlossaryClient
	ProjectID() string
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

	location := parameters.NewStringParameter("location", "The location where the glossary should be created (for example, us-central1).")
	glossaryID := parameters.NewStringParameter("glossaryId", "Identifier for the glossary to create. Becomes part of the resource name.")
	displayName := parameters.NewStringParameterWithDefault("displayName", "", "Optional user-friendly display name for the glossary.")
	description := parameters.NewStringParameterWithDefault("description", "", "Optional description for the glossary.")
	labels := parameters.NewMapParameterWithRequired("labels", "Optional labels to set on the glossary. Values must be strings.", false, "string")
	validateOnly := parameters.NewBooleanParameterWithDefault("validateOnly", false, "When true, validates the request without creating the glossary.")

	params := parameters.Parameters{location, glossaryID, displayName, description, labels, validateOnly}
	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, params)

	t := Tool{
		Config:      cfg,
		Parameters:  params,
		Glossary:    s.BusinessGlossaryClient(),
		ProjectID:   s.ProjectID(),
		manifest:    tools.Manifest{Description: cfg.Description, Parameters: params.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest: mcpManifest,
	}
	return t, nil
}

type Tool struct {
	Config
	Parameters  parameters.Parameters
	Glossary    *dataplexapi.BusinessGlossaryClient
	ProjectID   string
	manifest    tools.Manifest
	mcpManifest tools.McpManifest
}

func (t Tool) ToConfig() tools.ToolConfig {
	return t.Config
}

func (t Tool) Invoke(ctx context.Context, params parameters.ParamValues, accessToken tools.AccessToken) (any, error) {
	paramsMap := params.AsMap()

	location, _ := paramsMap["location"].(string)
	glossaryID, _ := paramsMap["glossaryId"].(string)
	displayName, _ := paramsMap["displayName"].(string)
	description, _ := paramsMap["description"].(string)

	labelsRaw, labelsProvided := paramsMap["labels"]
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

	validateOnly, _ := paramsMap["validateOnly"].(bool)

	parent := fmt.Sprintf("projects/%s/locations/%s", t.ProjectID, location)
	glossary := &dataplexpb.Glossary{
		DisplayName: displayName,
		Description: description,
	}
	if labelsProvided {
		glossary.Labels = labels
	}

	req := &dataplexpb.CreateGlossaryRequest{
		Parent:       parent,
		GlossaryId:   glossaryID,
		Glossary:     glossary,
		ValidateOnly: validateOnly,
	}

	op, err := t.Glossary.CreateGlossary(ctx, req)
	if err != nil {
		return nil, err
	}
	return op.Wait(ctx)
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
