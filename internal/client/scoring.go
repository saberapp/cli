package client

import (
	"context"
	"io"
	"net/url"
	"time"
)

// --- Scoring Profiles ---

// ScoringProfile is a named, org-scoped scoring configuration bound to an object type.
type ScoringProfile struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Type           string    `json:"type"` // company | contact
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateScoringProfileRequest is the payload for POST /v1/scoring/profiles.
type CreateScoringProfileRequest struct {
	ProfileType string  `json:"profileType"` // company | contact
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateScoringProfileRequest is the payload for PUT /v1/scoring/profiles/{id}.
type UpdateScoringProfileRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (c *Client) CreateScoringProfile(ctx context.Context, req CreateScoringProfileRequest, rawDst io.Writer) (*ScoringProfile, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/scoring/profiles", req, nil, rawDst)
	}
	var p ScoringProfile
	if err := c.Post(ctx, "/v1/scoring/profiles", req, &p, nil); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) ListScoringProfiles(ctx context.Context, rawDst io.Writer) ([]ScoringProfile, error) {
	if rawDst != nil {
		return nil, c.Get(ctx, "/v1/scoring/profiles", nil, rawDst)
	}
	var ps []ScoringProfile
	if err := c.Get(ctx, "/v1/scoring/profiles", &ps, nil); err != nil {
		return nil, err
	}
	return ps, nil
}

func (c *Client) GetScoringProfile(ctx context.Context, id string, rawDst io.Writer) (*ScoringProfile, error) {
	path := "/v1/scoring/profiles/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var p ScoringProfile
	if err := c.Get(ctx, path, &p, nil); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) UpdateScoringProfile(ctx context.Context, id string, req UpdateScoringProfileRequest, rawDst io.Writer) (*ScoringProfile, error) {
	path := "/v1/scoring/profiles/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Put(ctx, path, req, nil, rawDst)
	}
	var p ScoringProfile
	if err := c.Put(ctx, path, req, &p, nil); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) DeleteScoringProfile(ctx context.Context, id string) error {
	return c.Delete(ctx, "/v1/scoring/profiles/"+url.PathEscape(id))
}

// --- Scoring Rules ---

// ScoringPointValueRange is one bucket in a numeric ranges definition. Upper bound is exclusive.
type ScoringPointValueRange struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Points float64 `json:"points"`
}

// ScoringPointValues encodes the points awarded for a rule. Exactly one shape is
// populated based on the answer type of the referenced signal template.
//
//   - boolean      → True / False
//   - number / percentage / currency → Ranges
//   - list         → Choices (map of value → points)
type ScoringPointValues struct {
	True    *float64                 `json:"true,omitempty"`
	False   *float64                 `json:"false,omitempty"`
	Ranges  []ScoringPointValueRange `json:"ranges,omitempty"`
	Choices map[string]float64       `json:"choices,omitempty"`
}

// ScoringRule maps one signal template to point values for a dimension within a profile.
type ScoringRule struct {
	ID               string             `json:"id"`
	ProfileID        string             `json:"profileId"`
	SignalTemplateID string             `json:"signalTemplateId"`
	Dimension        string             `json:"dimension"` // fit | urgency
	PointValues      ScoringPointValues `json:"pointValues"`
	CreatedAt        time.Time          `json:"createdAt"`
}

// UpsertScoringRuleRequest is the payload for PUT /v1/scoring/profiles/{profileId}/rules.
//
// AnswerType drives server-side validation of PointValues' shape: a mismatch
// (e.g. ranges for a boolean signal) returns 422 INVALID_POINT_VALUES rather
// than a silent compute failure later. Required by the API.
type UpsertScoringRuleRequest struct {
	SignalTemplateID string             `json:"signalTemplateId"`
	Dimension        string             `json:"dimension"`
	AnswerType       string             `json:"answerType"`
	PointValues      ScoringPointValues `json:"pointValues"`
}

func (c *Client) UpsertScoringRule(ctx context.Context, profileID string, req UpsertScoringRuleRequest, rawDst io.Writer) (*ScoringRule, error) {
	path := "/v1/scoring/profiles/" + url.PathEscape(profileID) + "/rules"
	if rawDst != nil {
		return nil, c.Put(ctx, path, req, nil, rawDst)
	}
	var r ScoringRule
	if err := c.Put(ctx, path, req, &r, nil); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) ListScoringRules(ctx context.Context, profileID string, rawDst io.Writer) ([]ScoringRule, error) {
	path := "/v1/scoring/profiles/" + url.PathEscape(profileID) + "/rules"
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var rs []ScoringRule
	if err := c.Get(ctx, path, &rs, nil); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) DeleteScoringRule(ctx context.Context, profileID, ruleID string) error {
	return c.Delete(ctx, "/v1/scoring/profiles/"+url.PathEscape(profileID)+"/rules/"+url.PathEscape(ruleID))
}

// --- Profile Assignments ---

// ProfileAssignment links a scoring profile to one company or contact.
type ProfileAssignment struct {
	ID             string    `json:"id"`
	ProfileID      string    `json:"profileId"`
	OrganizationID string    `json:"organizationId"`
	ObjectType     string    `json:"objectType"` // company | contact
	ObjectID       string    `json:"objectId"`
	AssignedAt     time.Time `json:"assignedAt"`
}

// CreateProfileAssignmentRequest is the payload for POST /v1/scoring/assignments.
type CreateProfileAssignmentRequest struct {
	ProfileID  string `json:"profileId"`
	ObjectType string `json:"objectType"`
	ObjectID   string `json:"objectId"`
}

// BulkCreateProfileAssignmentsRequest is the payload for POST /v1/scoring/assignments/bulk.
type BulkCreateProfileAssignmentsRequest struct {
	ProfileID  string   `json:"profileId"`
	ObjectType string   `json:"objectType"`
	ObjectIDs  []string `json:"objectIds"`
}

func (c *Client) CreateProfileAssignment(ctx context.Context, req CreateProfileAssignmentRequest, rawDst io.Writer) (*ProfileAssignment, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/scoring/assignments", req, nil, rawDst)
	}
	var a ProfileAssignment
	if err := c.Post(ctx, "/v1/scoring/assignments", req, &a, nil); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) ListProfileAssignments(ctx context.Context, objectType, objectID string, rawDst io.Writer) ([]ProfileAssignment, error) {
	q := url.Values{}
	q.Set("objectType", objectType)
	q.Set("objectId", objectID)
	path := "/v1/scoring/assignments?" + q.Encode()
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var as []ProfileAssignment
	if err := c.Get(ctx, path, &as, nil); err != nil {
		return nil, err
	}
	return as, nil
}

func (c *Client) BulkCreateProfileAssignments(ctx context.Context, req BulkCreateProfileAssignmentsRequest, rawDst io.Writer) ([]ProfileAssignment, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/scoring/assignments/bulk", req, nil, rawDst)
	}
	var as []ProfileAssignment
	if err := c.Post(ctx, "/v1/scoring/assignments/bulk", req, &as, nil); err != nil {
		return nil, err
	}
	return as, nil
}

func (c *Client) DeleteProfileAssignment(ctx context.Context, id string) error {
	return c.Delete(ctx, "/v1/scoring/assignments/"+url.PathEscape(id))
}

// --- Scores ---

// ScoreContribution describes how a single rule contributed to a dimension score.
type ScoreContribution struct {
	RuleID           string  `json:"ruleId"`
	SignalTemplateID string  `json:"signalTemplateId"`
	MatchedValue     string  `json:"matchedValue"`
	PointsEarned     float64 `json:"pointsEarned"`
	MaxPoints        float64 `json:"maxPoints"`
}

// ScoreResult is the latest computed score for one (profile, object, dimension) triple.
type ScoreResult struct {
	ID                    string              `json:"id"`
	ProfileID             string              `json:"profileId"`
	OrganizationID        string              `json:"organizationId"`
	ObjectType            string              `json:"objectType"`
	ObjectID              string              `json:"objectId"`
	Dimension             string              `json:"dimension"`
	Score                 float64             `json:"score"`
	PreviousScore         *float64            `json:"previousScore"`
	Contributions         []ScoreContribution `json:"contributions"`
	PreviousContributions []ScoreContribution `json:"previousContributions"`
	SignalCoverage        int                 `json:"signalCoverage"`
	TotalRules            int                 `json:"totalRules"`
	ComputedAt            time.Time           `json:"computedAt"`
	Version               int                 `json:"version"`
}

// ComputeScoresRequest is the payload for POST /v1/scoring/compute.
type ComputeScoresRequest struct {
	ObjectType string   `json:"objectType"`
	ObjectIDs  []string `json:"objectIds"`
}

// GetScores reads scores for one or more objects. objectIDs is repeated as
// `objectId` in the query string (style=form, explode=true).
func (c *Client) GetScores(ctx context.Context, objectType string, objectIDs []string, rawDst io.Writer) ([]ScoreResult, error) {
	q := url.Values{}
	q.Set("objectType", objectType)
	for _, id := range objectIDs {
		q.Add("objectId", id)
	}
	path := "/v1/scoring/scores?" + q.Encode()
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var rs []ScoreResult
	if err := c.Get(ctx, path, &rs, nil); err != nil {
		return nil, err
	}
	return rs, nil
}

// ComputeScoresResponse is the body of a 202 Accepted from POST /v1/scoring/compute.
// `failed > 0` means some object dispatches failed (typically Temporal hiccups
// for those specific calls); the request as a whole is "accepted" because at
// least one workflow was queued. If every dispatch fails the API returns 502
// instead of 202.
type ComputeScoresResponse struct {
	Queued int `json:"queued"`
	Failed int `json:"failed"`
}

// TriggerScoreCompute queues async recomputation and returns the
// {queued, failed} counts from the 202 response.
func (c *Client) TriggerScoreCompute(ctx context.Context, req ComputeScoresRequest) (*ComputeScoresResponse, error) {
	var resp ComputeScoresResponse
	if err := c.Post(ctx, "/v1/scoring/compute", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
