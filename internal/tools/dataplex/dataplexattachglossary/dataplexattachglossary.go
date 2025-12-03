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

package dataplexattachglossary

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
	"github.com/googleapis/genai-toolbox/internal/util"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
)

const kind string = "dataplex-attach-glossary"
const defaultEntryLinkType = "projects/dataplex-types/locations/global/entryLinkTypes/definition"

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

	parent := parameters.NewStringParameter("parent", "Parent entry group in the form projects/{project}/locations/{location}/entryGroups/{entryGroup}.")
	entryLinkId := parameters.NewStringParameter("entryLinkId", "Identifier for the entry link to create.")
	sourceEntry := parameters.NewStringParameter("sourceEntry", "Resource name of the source entry to link from.")
	glossaryProject := parameters.NewStringParameter("glossaryProject", "Project for the glossary term.")
	glossaryLocation := parameters.NewStringParameter("glossaryLocation", "Location for the glossary term (often 'global').")
	glossaryId := parameters.NewStringParameter("glossaryId", "Glossary ID containing the term.")
	termId := parameters.NewStringParameter("termId", "Glossary term ID to link to.")
	entryLinkType := parameters.NewStringParameterWithDefault("entryLinkType", defaultEntryLinkType, "Entry link type to use for the association.")

	params := parameters.Parameters{parent, entryLinkId, sourceEntry, glossaryProject, glossaryLocation, glossaryId, termId, entryLinkType}
	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, params)

	t := Tool{
		Config:      cfg,
		Parameters:  params,
		Catalog:     s.CatalogClient(),
		manifest:    tools.Manifest{Description: cfg.Description, Parameters: params.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest: mcpManifest,
	}
	return t, nil
}

type Tool struct {
	Config
	Parameters  parameters.Parameters
	Catalog     *dataplexapi.CatalogClient
	manifest    tools.Manifest
	mcpManifest tools.McpManifest
}

func (t Tool) ToConfig() tools.ToolConfig {
	return t.Config
}

func (t Tool) Invoke(ctx context.Context, params parameters.ParamValues, accessToken tools.AccessToken) (any, error) {
	m := params.AsMap()

	parent, _ := m["parent"].(string)
	entryLinkId, _ := m["entryLinkId"].(string)
	sourceEntry, _ := m["sourceEntry"].(string)
	glossaryProject, _ := m["glossaryProject"].(string)
	glossaryLocation, _ := m["glossaryLocation"].(string)
	glossaryId, _ := m["glossaryId"].(string)
	termId, _ := m["termId"].(string)
	entryLinkType, _ := m["entryLinkType"].(string)
	if entryLinkType == "" {
		entryLinkType = defaultEntryLinkType
	}
	if entryLinkType != defaultEntryLinkType {
		return nil, fmt.Errorf("entryLinkType must be %q; got %q", defaultEntryLinkType, entryLinkType)
	}

	parentProject, err := parseProjectFromParent(parent)
	if err != nil {
		return nil, err
	}

	if glossaryProject != parentProject {
		return nil, fmt.Errorf("glossaryProject must match the parent project (%s); create the glossary/term in that project", parentProject)
	}

	targetEntry := buildTermEntryName(ctx, parentProject, glossaryProject, glossaryLocation, glossaryId, termId)

	req := &dataplexpb.CreateEntryLinkRequest{
		Parent:      parent,
		EntryLinkId: entryLinkId,
		EntryLink: &dataplexpb.EntryLink{
			EntryLinkType: entryLinkType,
			EntryReferences: []*dataplexpb.EntryLink_EntryReference{
				{
					Name: sourceEntry,
					Type: dataplexpb.EntryLink_EntryReference_SOURCE,
				},
				{
					Name: targetEntry,
					Type: dataplexpb.EntryLink_EntryReference_TARGET,
				},
			},
		},
	}

	resp, err := t.Catalog.CreateEntryLink(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(
			"v2 to create entry link: entryLinkType=%s\n  entryLinkId=%s\n  parent=%s\n  sourceEntry=%s\n  targetEntry=%s\n  glossaryProject=%s\n  glossaryLocation=%s\n  glossaryId=%s\n  termId=%s\n  underlying error: %v",
			entryLinkType, entryLinkId, parent, sourceEntry, targetEntry, glossaryProject, glossaryLocation, glossaryId, termId, err,
		)
	}
	return resp, nil
}

func buildTermEntryName(ctx context.Context, entryGroupProject, glossaryProject, location, glossaryId, termId string) string {
	targetEntry := fmt.Sprintf(
		"projects/%s/locations/%s/entryGroups/@dataplex/entries/projects/%s/locations/%s/glossaries/%s/terms/%s",
		entryGroupProject, location, glossaryProject, location, glossaryId, termId,
	)

	if logger, err := util.LoggerFromContext(ctx); err == nil {
		logger.DebugContext(
			ctx,
			"building glossary term entry name",
			"entryGroupProject", entryGroupProject,
			"glossaryProject", glossaryProject,
			"location", location,
			"glossaryId", glossaryId,
			"termId", termId,
			"targetEntry", targetEntry,
		)
	}

	return targetEntry
}

func parseProjectFromParent(parent string) (string, error) {
	// Expected: projects/{project}/locations/{location}/entryGroups/{entryGroup}
	parts := strings.Split(parent, "/")
	if len(parts) < 2 || parts[0] != "projects" {
		return "", fmt.Errorf("invalid parent format; expected projects/{project}/locations/{location}/entryGroups/{entryGroup}")
	}
	return parts[1], nil
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
