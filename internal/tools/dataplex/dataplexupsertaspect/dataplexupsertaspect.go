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

package dataplexupsertaspect

import (
	"context"
	"fmt"
	"strings"

	dataplexapi "cloud.google.com/go/dataplex/apiv1"
	dataplexpb "cloud.google.com/go/dataplex/apiv1/dataplexpb"
	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	dataplexds "github.com/googleapis/genai-toolbox/internal/sources/dataplex"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

const kind string = "dataplex-upsert-aspect"

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

	entry := parameters.NewStringParameter("entry", "The resource name of the Entry to update in the format projects/{project}/locations/{location}/entryGroups/{entryGroup}/entries/{entry}.")
	aspectType := parameters.NewStringParameter("aspectType", "Aspect type to attach to the entry. Accepts the full resource name projects/{project}/locations/{location}/aspectTypes/{aspectTypeId} or the aspect reference {project}.{location}.{aspectTypeId}.")
	path := parameters.NewStringParameterWithDefault("path", "", "Optional aspect path when attaching to a nested resource (for example, Schema.my_column). Leave empty to attach the aspect directly to the entry.")
	data := parameters.NewMapParameterWithRequired("data", "JSON payload for the aspect content. Must follow the schema expected by the aspect type.", true, "")
	deleteMissingAspects := parameters.NewBooleanParameterWithDefault("deleteMissingAspects", false, "If true, removes other aspects that match the provided aspect type and path range but were not supplied in this request.")
	allowMissingEntry := parameters.NewBooleanParameterWithDefault("allowMissingEntry", false, "If true, creates the entry when it does not already exist.")

	params := parameters.Parameters{entry, aspectType, path, data, deleteMissingAspects, allowMissingEntry}

	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, params)

	t := Tool{
		Config:        cfg,
		Parameters:    params,
		CatalogClient: s.CatalogClient(),
		manifest: tools.Manifest{
			Description:  cfg.Description,
			Parameters:   params.Manifest(),
			AuthRequired: cfg.AuthRequired,
		},
		mcpManifest: mcpManifest,
	}
	return t, nil
}

type Tool struct {
	Config
	Parameters    parameters.Parameters
	CatalogClient *dataplexapi.CatalogClient
	manifest      tools.Manifest
	mcpManifest   tools.McpManifest
}

func (t Tool) ToConfig() tools.ToolConfig {
	return t.Config
}

func (t Tool) Invoke(ctx context.Context, params parameters.ParamValues, accessToken tools.AccessToken) (any, error) {
	paramsMap := params.AsMap()

	entryName, _ := paramsMap["entry"].(string)
	aspectTypeInput, _ := paramsMap["aspectType"].(string)

	aspectData, ok := paramsMap["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data must be an object")
	}
	aspectStruct, err := structpb.NewStruct(aspectData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert aspect data into struct: %w", err)
	}

	path, _ := paramsMap["path"].(string)
	aspectKey, err := buildAspectKey(aspectTypeInput, path)
	if err != nil {
		return nil, err
	}

	deleteMissing, _ := paramsMap["deleteMissingAspects"].(bool)
	allowMissing, _ := paramsMap["allowMissingEntry"].(bool)

	req := &dataplexpb.UpdateEntryRequest{
		Entry: &dataplexpb.Entry{
			Name: entryName,
			Aspects: map[string]*dataplexpb.Aspect{
				aspectKey: {
					Data: aspectStruct,
				},
			},
		},
		UpdateMask:           &fieldmaskpb.FieldMask{Paths: []string{"aspects"}},
		AllowMissing:         allowMissing,
		DeleteMissingAspects: deleteMissing,
		AspectKeys:           []string{aspectKey},
	}

	result, err := t.CatalogClient.UpdateEntry(ctx, req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func buildAspectKey(aspectTypeRef, path string) (string, error) {
	aspectTypeRef = strings.TrimSpace(aspectTypeRef)
	if aspectTypeRef == "" {
		return "", fmt.Errorf("aspectType is required")
	}
	var aspectRef string
	if strings.Contains(aspectTypeRef, "/aspectTypes/") {
		split := strings.Split(aspectTypeRef, "/")
		if len(split) < 6 {
			return "", fmt.Errorf("invalid aspectType format, expected projects/{project}/locations/{location}/aspectTypes/{aspectTypeId}")
		}
		// projects/{project}/locations/{location}/aspectTypes/{aspectTypeId}
		aspectRef = fmt.Sprintf("%s.%s.%s", split[1], split[3], split[5])
	} else {
		aspectRef = aspectTypeRef
	}

	if strings.Count(aspectRef, ".") != 2 {
		return "", fmt.Errorf("aspectType must be in the form project.location.aspectTypeId or full resource name")
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return aspectRef, nil
	}
	return fmt.Sprintf("%s@%s", aspectRef, path), nil
}

// BuildAspectKeyForTest exposes buildAspectKey for unit tests.
func BuildAspectKeyForTest(aspectTypeRef, path string) (string, error) {
	return buildAspectKey(aspectTypeRef, path)
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
