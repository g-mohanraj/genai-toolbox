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

package dataplexupdateentry

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

const kind string = "dataplex-update-entry"

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
	description := parameters.NewStringParameterWithDefault("description", "", "Optional description to set on the entry (stored in entry_source.description). Leave empty to skip updating.")
	contacts := parameters.NewMapParameterWithRequired("contacts", "Optional contacts payload to attach using the contacts aspect type.", false, "")
	contactsAspectType := parameters.NewStringParameterWithDefault("contactsAspectType", "projects/dataplex-types/locations/global/aspectTypes/contacts", "Aspect type to use for contacts. Accepts full resource name or dotted reference.")
	path := parameters.NewStringParameterWithDefault("path", "", "Optional path for the contacts aspect (e.g., Schema.column_name). Leave empty to attach to the entry.")
	allowMissingEntry := parameters.NewBooleanParameterWithDefault("allowMissingEntry", false, "If true, creates the entry when it does not exist.")
	deleteMissingAspects := parameters.NewBooleanParameterWithDefault("deleteMissingAspects", false, "If true, delete missing aspects in the specified keys range.")

	params := parameters.Parameters{entry, description, contacts, contactsAspectType, path, allowMissingEntry, deleteMissingAspects}
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

	entry := &dataplexpb.Entry{Name: entryName}
	updatePaths := []string{}

	if desc, ok := paramsMap["description"].(string); ok && desc != "" {
		entry.EntrySource = &dataplexpb.EntrySource{Description: desc}
		updatePaths = append(updatePaths, "entry_source")
	}

	contactsRaw, contactsProvided := paramsMap["contacts"]
	var aspectKeys []string
	if contactsProvided && contactsRaw != nil {
		contactsMap, ok := contactsRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("contacts must be an object when provided")
		}
		contactsStruct, err := structpb.NewStruct(contactsMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert contacts into struct: %w", err)
		}

		aspectTypeRef, _ := paramsMap["contactsAspectType"].(string)
		path, _ := paramsMap["path"].(string)
		aspectKey, err := buildAspectKey(aspectTypeRef, path)
		if err != nil {
			return nil, err
		}
		entry.Aspects = map[string]*dataplexpb.Aspect{
			aspectKey: {Data: contactsStruct},
		}
		updatePaths = append(updatePaths, "aspects")
		aspectKeys = []string{aspectKey}
	}

	if len(updatePaths) == 0 {
		return nil, fmt.Errorf("no updates provided; set description and/or contacts")
	}

	allowMissing, _ := paramsMap["allowMissingEntry"].(bool)
	deleteMissing, _ := paramsMap["deleteMissingAspects"].(bool)

	req := &dataplexpb.UpdateEntryRequest{
		Entry:                entry,
		UpdateMask:           &fieldmaskpb.FieldMask{Paths: updatePaths},
		AllowMissing:         allowMissing,
		DeleteMissingAspects: deleteMissing,
		AspectKeys:           aspectKeys,
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
		return "", fmt.Errorf("contactsAspectType is required when contacts are provided")
	}
	var aspectRef string
	if strings.Contains(aspectTypeRef, "/aspectTypes/") {
		split := strings.Split(aspectTypeRef, "/")
		if len(split) < 6 {
			return "", fmt.Errorf("invalid aspect type format, expected projects/{project}/locations/{location}/aspectTypes/{aspectTypeId}")
		}
		aspectRef = fmt.Sprintf("%s.%s.%s", split[1], split[3], split[5])
	} else {
		aspectRef = aspectTypeRef
	}
	if strings.Count(aspectRef, ".") != 2 {
		return "", fmt.Errorf("contactsAspectType must be full resource name or dotted project.location.aspectTypeId")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return aspectRef, nil
	}
	return fmt.Sprintf("%s@%s", aspectRef, path), nil
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
