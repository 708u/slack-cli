package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	slackgo "github.com/slack-go/slack"
)

// CanvasOps groups canvas-related API calls. It stores the raw token
// because canvases.sections.lookup returns richer data than what the
// slack-go library exposes in its CanvasSection type.
type CanvasOps struct {
	api        *slackgo.Client
	token      string
	baseURL    string // defaults to slackgo.APIURL
	channelOps *ChannelOps
}

func newCanvasOps(api *slackgo.Client, token string, channelOps *ChannelOps) *CanvasOps {
	return &CanvasOps{
		api:        api,
		token:      token,
		baseURL:    slackgo.APIURL,
		channelOps: channelOps,
	}
}

// canvasSectionsLookupResponse is the full API response for
// canvases.sections.lookup including section elements.
type canvasSectionsLookupResponse struct {
	OK       bool                   `json:"ok"`
	Error    string                 `json:"error,omitempty"`
	Sections []canvasSectionPayload `json:"sections"`
}

type canvasSectionPayload struct {
	ID       string                        `json:"id"`
	Elements []canvasSectionElementPayload `json:"elements"`
}

type canvasSectionElementPayload struct {
	Type     string                        `json:"type"`
	Text     string                        `json:"text"`
	Elements []canvasSectionElementPayload `json:"elements"`
}

// ReadCanvas returns the sections of a canvas document by looking up
// all header sections.
func (c *CanvasOps) ReadCanvas(canvasID string) ([]CanvasSection, error) {
	criteria, err := json.Marshal(map[string]any{
		"section_types": []string{"any_header"},
	})
	if err != nil {
		return nil, fmt.Errorf("read canvas: marshal criteria: %w", err)
	}

	vals := url.Values{
		"token":     {c.token},
		"canvas_id": {canvasID},
		"criteria":  {string(criteria)},
	}

	resp, err := http.PostForm(c.baseURL+"canvases.sections.lookup", vals)
	if err != nil {
		return nil, fmt.Errorf("read canvas: %w", err)
	}
	defer resp.Body.Close()

	var result canvasSectionsLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("read canvas: decode response: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("read canvas: %s", result.Error)
	}

	sections := make([]CanvasSection, 0, len(result.Sections))
	for _, s := range result.Sections {
		sections = append(sections, CanvasSection{
			ID:       s.ID,
			Elements: convertElements(s.Elements),
		})
	}
	return sections, nil
}

// convertElements recursively converts the API element payloads to
// internal CanvasSectionElement values.
func convertElements(elems []canvasSectionElementPayload) []CanvasSectionElement {
	if len(elems) == 0 {
		return nil
	}
	out := make([]CanvasSectionElement, 0, len(elems))
	for _, e := range elems {
		out = append(out, CanvasSectionElement{
			Type:     e.Type,
			Text:     e.Text,
			Elements: convertElements(e.Elements),
		})
	}
	return out
}

// ListCanvases returns canvas files attached to the specified channel.
// Canvases are Slack files with the type "spaces". The channel
// parameter accepts a name or ID.
func (c *CanvasOps) ListCanvases(channel string) ([]CanvasFile, error) {
	channelID, err := c.channelOps.ResolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	files, _, err := c.api.GetFiles(slackgo.GetFilesParameters{
		Channel: channelID,
		Types:   "spaces",
		Count:   100,
		Page:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("list canvases: %w", err)
	}

	canvases := make([]CanvasFile, 0, len(files))
	for _, f := range files {
		canvases = append(canvases, CanvasFile{
			ID:       f.ID,
			Name:     f.Name,
			Created:  int64(f.Created),
			FileType: f.Filetype,
		})
	}
	return canvases, nil
}
