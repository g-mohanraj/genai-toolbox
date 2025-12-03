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

package dataplexcreateaspecttype

import (
	"context"
	"encoding/json"
	"fmt"

	dataplexapi "cloud.google.com/go/dataplex/apiv1"
	dataplexpb "cloud.google.com/go/dataplex/apiv1/dataplexpb"
	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	dataplexds "github.com/googleapis/genai-toolbox/internal/sources/dataplex"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
	"google.golang.org/protobuf/encoding/protojson"
)

const kind string = "dataplex-create-aspect-type"

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

	location := parameters.NewStringParameter("location", "The location where the aspect type should be created (for example, us-central1).")
	aspectTypeID := parameters.NewStringParameter("aspectTypeId", "Identifier for the aspect type to create. Becomes the final segment of the resource name.")
	aspectType := parameters.NewMapParameterWithRequired("aspectType", "Aspect type definition payload. Must follow Dataplex aspect type schema (for example, metadataTemplate).", true, "")
	validateOnly := parameters.NewBooleanParameterWithDefault("validateOnly", false, "When true, validates the request without creating the aspect type.")

	params := parameters.Parameters{location, aspectTypeID, aspectType, validateOnly}
	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, params)

	t := Tool{
		Config:      cfg,
		Parameters:  params,
		Catalog:     s.CatalogClient(),
		ProjectID:   s.ProjectID(),
		manifest:    tools.Manifest{Description: cfg.Description, Parameters: params.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest: mcpManifest,
	}
	return t, nil
}

type Tool struct {
	Config
	Parameters  parameters.Parameters
	Catalog     *dataplexapi.CatalogClient
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
	aspectTypeID, _ := paramsMap["aspectTypeId"].(string)
	aspectMap, ok := paramsMap["aspectType"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("aspectType must be an object")
	}

	body, err := json.Marshal(aspectMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal aspectType payload: %w", err)
	}

	var aspect dataplexpb.AspectType
	if err := protojson.Unmarshal(body, &aspect); err != nil {
		return nil, fmt.Errorf("invalid aspectType payload: %w", err)
	}
	if aspect.MetadataTemplate == nil {
		return nil, fmt.Errorf("aspectType.metadataTemplate is required")
	}

	validateOnly, _ := paramsMap["validateOnly"].(bool)

	parent := fmt.Sprintf("projects/%s/locations/%s", t.ProjectID, location)
	req := &dataplexpb.CreateAspectTypeRequest{
		Parent:       parent,
		AspectTypeId: aspectTypeID,
		AspectType:   &aspect,
		ValidateOnly: validateOnly,
	}

	op, err := t.Catalog.CreateAspectType(ctx, req)
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
