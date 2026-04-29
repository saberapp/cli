package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newScoringCmd() *cobra.Command {
	scoring := &cobra.Command{
		Use:   "scoring",
		Short: "Manage scoring profiles, rules, assignments, and scores",
		Long: `Configure how Saber turns signal results into structured fit and urgency scores
(0–100) for companies and contacts.

A profile groups rules for a single object type. A rule maps one signal template
to point values for a dimension (fit or urgency). Assignments link a profile to
specific companies or contacts and trigger an immediate recompute.`,
	}
	scoring.AddCommand(newScoringProfileCmd())
	scoring.AddCommand(newScoringRuleCmd())
	scoring.AddCommand(newScoringAssignmentCmd())
	scoring.AddCommand(newScoringScoresCmd())
	scoring.AddCommand(newScoringComputeCmd())
	return scoring
}

// ── Profiles ──────────────────────────────────────────────────────────────────

func newScoringProfileCmd() *cobra.Command {
	profile := &cobra.Command{
		Use:   "profile",
		Short: "Manage scoring profiles",
	}
	profile.AddCommand(newScoringProfileCreateCmd())
	profile.AddCommand(newScoringProfileListCmd())
	profile.AddCommand(newScoringProfileGetCmd())
	profile.AddCommand(newScoringProfileUpdateCmd())
	profile.AddCommand(newScoringProfileDeleteCmd())
	return profile
}

func newScoringProfileCreateCmd() *cobra.Command {
	var (
		profileType string
		name        string
		description string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scoring profile",
		Example: `  saber scoring profile create --type company --name "EMEA Enterprise"
  saber scoring profile create --type contact --name "Decision makers" --description "Senior engineering buyers"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(profileType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			req := client.CreateScoringProfileRequest{
				ProfileType: t,
				Name:        name,
			}
			if description != "" {
				d := description
				req.Description = &d
			}
			if jsonOutput {
				_, err := c.CreateScoringProfile(ctx, req, os.Stdout)
				return err
			}
			p, err := c.CreateScoringProfile(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringProfile(os.Stdout, p)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileType, "type", "", "Object type the profile scores: company or contact")
	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	cmd.Flags().StringVar(&description, "description", "", "Optional description")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newScoringProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scoring profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListScoringProfiles(ctx, os.Stdout)
				return err
			}
			profiles, err := c.ListScoringProfiles(ctx, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringProfiles(os.Stdout, profiles)
			}
			return nil
		},
	}
}

func newScoringProfileGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <profileId>",
		Short: "Get a scoring profile by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetScoringProfile(ctx, args[0], os.Stdout)
				return err
			}
			p, err := c.GetScoringProfile(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringProfile(os.Stdout, p)
			}
			return nil
		},
	}
}

func newScoringProfileUpdateCmd() *cobra.Command {
	var (
		name        string
		description string
	)
	cmd := &cobra.Command{
		Use:   "update <profileId>",
		Short: "Update a scoring profile (rename or re-describe)",
		Long:  `Profile type is immutable — only name and description can be changed.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.UpdateScoringProfileRequest{Name: name}
			if cmd.Flags().Changed("description") {
				d := description
				req.Description = &d
			}
			if jsonOutput {
				_, err := c.UpdateScoringProfile(ctx, args[0], req, os.Stdout)
				return err
			}
			p, err := c.UpdateScoringProfile(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringProfile(os.Stdout, p)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New profile name")
	cmd.Flags().StringVar(&description, "description", "", "New description (passing an empty string sets it to empty)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newScoringProfileDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profileId>",
		Short: "Delete a scoring profile (cascades to rules, assignments, and scores)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteScoringProfile(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted profile %s\n", args[0])
			}
			return nil
		},
	}
}

// ── Rules ─────────────────────────────────────────────────────────────────────

func newScoringRuleCmd() *cobra.Command {
	rule := &cobra.Command{
		Use:   "rule",
		Short: "Manage scoring rules within a profile",
	}
	rule.AddCommand(newScoringRuleUpsertCmd())
	rule.AddCommand(newScoringRuleListCmd())
	rule.AddCommand(newScoringRuleDeleteCmd())
	return rule
}

func newScoringRuleUpsertCmd() *cobra.Command {
	var (
		signalTemplate string
		dimension      string
		answerType     string
		pointsJSON     string
		pointsFile     string
		boolTrue       float64
		boolFalse      float64
		ranges         []string
		choices        []string
	)
	cmd := &cobra.Command{
		Use:   "upsert <profileId>",
		Short: "Create or replace a scoring rule",
		Long: `Upsert a rule for (profileId, signalTemplateId, dimension). Triggers a recompute
of every object assigned to the profile so existing scores reflect the change.

--answer-type must match the referenced signal template's answer type. The
server validates the point-values shape against it and returns 422 on mismatch
rather than silently failing at compute time.

Provide point values via exactly one of:

  Boolean signal:
    --answer-type boolean --true 20 --false -5

  Number / percentage / currency signal:
    --answer-type number --range "0:500:5" --range "500:5000:15"
    (min:max:points; max is exclusive)

  List signal:
    --answer-type list --choice "Salesforce:10" --choice "HubSpot:8"

  Or provide raw JSON matching the ScoringPointValues schema:
    --answer-type number --points '{"ranges":[{"min":0,"max":500,"points":5}]}'
    --answer-type list --points-file rules.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dim, err := parseDimension(dimension)
			if err != nil {
				return err
			}

			at, err := parseAnswerType(answerType)
			if err != nil {
				return err
			}

			pv, err := buildPointValues(cmd, pointsJSON, pointsFile, boolTrue, boolFalse, ranges, choices)
			if err != nil {
				return err
			}

			c, ctx := mustClient()
			req := client.UpsertScoringRuleRequest{
				SignalTemplateID: signalTemplate,
				Dimension:        dim,
				AnswerType:       at,
				PointValues:      pv,
			}
			if jsonOutput {
				_, err := c.UpsertScoringRule(ctx, args[0], req, os.Stdout)
				return err
			}
			r, err := c.UpsertScoringRule(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringRule(os.Stdout, r)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&signalTemplate, "signal-template", "", "Signal template ID this rule scores")
	cmd.Flags().StringVar(&dimension, "dimension", "", "Score dimension: fit or urgency")
	cmd.Flags().StringVar(&answerType, "answer-type", "", "Signal answer type: boolean, number, percentage, currency, or list")
	cmd.Flags().StringVar(&pointsJSON, "points", "", "Point values as JSON (matches the ScoringPointValues schema)")
	cmd.Flags().StringVar(&pointsFile, "points-file", "", "Path to a JSON file containing point values")
	cmd.Flags().Float64Var(&boolTrue, "true", 0, "Points awarded when boolean signal is true")
	cmd.Flags().Float64Var(&boolFalse, "false", 0, "Points awarded when boolean signal is false")
	cmd.Flags().StringArrayVar(&ranges, "range", nil, "Numeric range as min:max:points (repeatable, max is exclusive)")
	cmd.Flags().StringArrayVar(&choices, "choice", nil, "List choice as value:points (repeatable)")
	_ = cmd.MarkFlagRequired("signal-template")
	_ = cmd.MarkFlagRequired("dimension")
	_ = cmd.MarkFlagRequired("answer-type")
	return cmd
}

func newScoringRuleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <profileId>",
		Short: "List scoring rules for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListScoringRules(ctx, args[0], os.Stdout)
				return err
			}
			rules, err := c.ListScoringRules(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintScoringRules(os.Stdout, rules)
			}
			return nil
		},
	}
}

func newScoringRuleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profileId> <ruleId>",
		Short: "Delete a scoring rule (triggers a recompute for all assigned objects)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteScoringRule(ctx, args[0], args[1]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted rule %s from profile %s\n", args[1], args[0])
			}
			return nil
		},
	}
}

// ── Assignments ───────────────────────────────────────────────────────────────

func newScoringAssignmentCmd() *cobra.Command {
	a := &cobra.Command{
		Use:     "assignment",
		Aliases: []string{"assignments"},
		Short:   "Manage profile assignments",
	}
	a.AddCommand(newScoringAssignmentCreateCmd())
	a.AddCommand(newScoringAssignmentBulkCmd())
	a.AddCommand(newScoringAssignmentListCmd())
	a.AddCommand(newScoringAssignmentDeleteCmd())
	return a
}

func newScoringAssignmentCreateCmd() *cobra.Command {
	var (
		profileID  string
		objectType string
		objectID   string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Assign a profile to a single company or contact",
		Long: `Links one company (by domain) or contact (by LinkedIn profile URL) to a scoring
profile. Triggers immediate score computation.`,
		Example: `  saber scoring assignment create --profile <id> --type company --object acme.com
  saber scoring assignment create --profile <id> --type contact --object https://linkedin.com/in/jane`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(objectType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			req := client.CreateProfileAssignmentRequest{
				ProfileID:  profileID,
				ObjectType: t,
				ObjectID:   objectID,
			}
			if jsonOutput {
				_, err := c.CreateProfileAssignment(ctx, req, os.Stdout)
				return err
			}
			a, err := c.CreateProfileAssignment(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintProfileAssignment(os.Stdout, a)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileID, "profile", "", "Scoring profile ID")
	cmd.Flags().StringVar(&objectType, "type", "", "Object type: company or contact")
	cmd.Flags().StringVar(&objectID, "object", "", "Domain (for company) or LinkedIn profile URL (for contact)")
	_ = cmd.MarkFlagRequired("profile")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

func newScoringAssignmentBulkCmd() *cobra.Command {
	var (
		profileID  string
		objectType string
		objectIDs  []string
	)
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Assign a profile to many objects at once",
		Long: `Links many objects to a single profile in one call. Duplicates are skipped (only
newly created assignments are returned).`,
		Example: `  saber scoring assignment bulk --profile <id> --type company --object acme.com --object stripe.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(objectType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			req := client.BulkCreateProfileAssignmentsRequest{
				ProfileID:  profileID,
				ObjectType: t,
				ObjectIDs:  objectIDs,
			}
			if jsonOutput {
				_, err := c.BulkCreateProfileAssignments(ctx, req, os.Stdout)
				return err
			}
			as, err := c.BulkCreateProfileAssignments(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintProfileAssignments(os.Stdout, as)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileID, "profile", "", "Scoring profile ID")
	cmd.Flags().StringVar(&objectType, "type", "", "Object type: company or contact")
	cmd.Flags().StringArrayVar(&objectIDs, "object", nil, "Object ID (repeatable). Domains for company, LinkedIn URLs for contact.")
	_ = cmd.MarkFlagRequired("profile")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

func newScoringAssignmentListCmd() *cobra.Command {
	var (
		objectType string
		objectID   string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profile assignments for a single object",
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(objectType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListProfileAssignments(ctx, t, objectID, os.Stdout)
				return err
			}
			as, err := c.ListProfileAssignments(ctx, t, objectID, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintProfileAssignments(os.Stdout, as)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&objectType, "type", "", "Object type: company or contact")
	cmd.Flags().StringVar(&objectID, "object", "", "Domain (company) or LinkedIn profile URL (contact)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

func newScoringAssignmentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <assignmentId>",
		Short: "Remove a profile assignment (clears its scores)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteProfileAssignment(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted assignment %s\n", args[0])
			}
			return nil
		},
	}
}

// ── Scores ────────────────────────────────────────────────────────────────────

func newScoringScoresCmd() *cobra.Command {
	var (
		objectType string
		objectIDs  []string
		detailed   bool
	)
	cmd := &cobra.Command{
		Use:   "scores",
		Short: "Read scores for one or more objects",
		Long: `Returns one row per (profile, object, dimension) triple. Pass --object multiple
times to read scores for several objects in a single call.

Use --detailed to render each score with its per-rule contribution breakdown.`,
		Example: `  saber scoring scores --type company --object acme.com
  saber scoring scores --type company --object acme.com --object stripe.com --detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(objectType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetScores(ctx, t, objectIDs, os.Stdout)
				return err
			}
			scores, err := c.GetScores(ctx, t, objectIDs, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if detailed {
				for i := range scores {
					if i > 0 {
						fmt.Fprintln(os.Stdout)
					}
					format.PrintScoreResult(os.Stdout, &scores[i])
				}
				return nil
			}
			format.PrintScoreResults(os.Stdout, scores)
			return nil
		},
	}
	cmd.Flags().StringVar(&objectType, "type", "", "Object type: company or contact")
	cmd.Flags().StringArrayVar(&objectIDs, "object", nil, "Object ID (repeatable). Domains for company, LinkedIn URLs for contact.")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show per-rule contribution breakdown for each score")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

// ── Compute ───────────────────────────────────────────────────────────────────

func newScoringComputeCmd() *cobra.Command {
	var (
		objectType string
		objectIDs  []string
	)
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Trigger an async score recomputation",
		Long: `Queues a Temporal workflow per object to recompute scores against the latest
signal data. Idempotent — duplicate triggers attach to the running workflow.

Returns immediately. Read results with: saber scoring scores --type ... --object ...`,
		Example: `  saber scoring compute --type company --object acme.com
  saber scoring compute --type contact --object https://linkedin.com/in/jane`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseObjectType(objectType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			req := client.ComputeScoresRequest{ObjectType: t, ObjectIDs: objectIDs}
			resp, err := c.TriggerScoreCompute(ctx, req)
			if err != nil {
				return err
			}
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(resp)
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Recomputation queued: %d %s object(s)", resp.Queued, t)
				if resp.Failed > 0 {
					// Partial failure — surface so the user can retry the failed slice
					// rather than assume everything went through. The server already
					// logged per-object reasons; we just relay the count.
					fmt.Fprintf(os.Stdout, " (%d dispatch failure(s); retry to pick those up)", resp.Failed)
				}
				fmt.Fprintln(os.Stdout)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&objectType, "type", "", "Object type: company or contact")
	cmd.Flags().StringArrayVar(&objectIDs, "object", nil, "Object ID (repeatable)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildPointValues assembles a ScoringPointValues from the rule upsert flags. Exactly
// one shape (boolean / ranges / choices / raw JSON) must be specified.
func buildPointValues(cmd *cobra.Command, pointsJSON, pointsFile string, boolTrue, boolFalse float64, ranges, choices []string) (client.ScoringPointValues, error) {
	var pv client.ScoringPointValues
	shapes := 0

	if pointsJSON != "" || pointsFile != "" {
		if pointsJSON != "" && pointsFile != "" {
			return pv, fmt.Errorf("--points and --points-file are mutually exclusive")
		}
		raw := []byte(pointsJSON)
		if pointsFile != "" {
			b, err := os.ReadFile(pointsFile)
			if err != nil {
				return pv, fmt.Errorf("read --points-file: %w", err)
			}
			raw = b
		}
		if err := json.Unmarshal(raw, &pv); err != nil {
			return pv, fmt.Errorf("parse points JSON: %w", err)
		}
		return pv, nil
	}

	if cmd.Flags().Changed("true") || cmd.Flags().Changed("false") {
		shapes++
		if cmd.Flags().Changed("true") {
			t := boolTrue
			pv.True = &t
		}
		if cmd.Flags().Changed("false") {
			f := boolFalse
			pv.False = &f
		}
	}

	if len(ranges) > 0 {
		shapes++
		for _, r := range ranges {
			parsed, err := parseRange(r)
			if err != nil {
				return pv, err
			}
			pv.Ranges = append(pv.Ranges, parsed)
		}
	}

	if len(choices) > 0 {
		shapes++
		pv.Choices = make(map[string]float64, len(choices))
		for _, c := range choices {
			key, points, err := parseChoice(c)
			if err != nil {
				return pv, err
			}
			pv.Choices[key] = points
		}
	}

	if shapes == 0 {
		return pv, fmt.Errorf("must provide point values via --true/--false, --range, --choice, --points, or --points-file")
	}
	if shapes > 1 {
		return pv, fmt.Errorf("--true/--false, --range, and --choice are mutually exclusive — pick one shape")
	}
	return pv, nil
}

func parseRange(s string) (client.ScoringPointValueRange, error) {
	var r client.ScoringPointValueRange
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return r, fmt.Errorf("invalid --range %q (expected min:max:points)", s)
	}
	min, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return r, fmt.Errorf("invalid range min in %q: %w", s, err)
	}
	max, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return r, fmt.Errorf("invalid range max in %q: %w", s, err)
	}
	points, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return r, fmt.Errorf("invalid range points in %q: %w", s, err)
	}
	return client.ScoringPointValueRange{Min: min, Max: max, Points: points}, nil
}

func parseChoice(s string) (string, float64, error) {
	// Split on the LAST colon so values containing ":" still work.
	idx := strings.LastIndex(s, ":")
	if idx <= 0 || idx == len(s)-1 {
		return "", 0, fmt.Errorf("invalid --choice %q (expected value:points)", s)
	}
	key := strings.TrimSpace(s[:idx])
	pointsStr := strings.TrimSpace(s[idx+1:])
	if key == "" {
		return "", 0, fmt.Errorf("invalid --choice %q (empty value)", s)
	}
	points, err := strconv.ParseFloat(pointsStr, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid choice points in %q: %w", s, err)
	}
	return key, points, nil
}

// parseDimension validates a --dimension flag value.
func parseDimension(s string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if v != "fit" && v != "urgency" {
		return "", fmt.Errorf("dimension must be 'fit' or 'urgency', got %q", s)
	}
	return v, nil
}

// parseObjectType validates a --type flag value (objectType / profileType).
func parseObjectType(s string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if v != "company" && v != "contact" {
		return "", fmt.Errorf("type must be 'company' or 'contact', got %q", s)
	}
	return v, nil
}

// parseAnswerType validates a --answer-type flag value. Mirrors the API's
// supported answer types — must match the signal template's answer_type.
func parseAnswerType(s string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "boolean", "number", "percentage", "currency", "list":
		return v, nil
	}
	return "", fmt.Errorf("answer-type must be one of boolean|number|percentage|currency|list, got %q", s)
}
