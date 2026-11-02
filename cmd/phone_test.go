package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/saberapp/cli/internal/client"
)

func TestValidatePhoneIdentityFlags(t *testing.T) {
	for mask := 0; mask < 32; mask++ {
		req := client.FindPhoneRequest{}
		fields := []*string{&req.ContactProfileURL, &req.FirstName, &req.LastName, &req.CompanyName, &req.CompanyDomain}
		for i, field := range fields {
			if mask&(1<<i) != 0 {
				*field = " example "
			} else {
				*field = " \t\n "
			}
		}
		wantValid := mask&1 != 0 || (mask&2 != 0 && mask&4 != 0 && mask&24 != 0)
		t.Run(fmt.Sprintf("combination_%d", mask), func(t *testing.T) {
			got, err := validatePhoneFlags(req, 120, false, false)
			if (err == nil) != wantValid {
				t.Fatalf("request = %+v, error = %v, want valid = %v", req, err, wantValid)
			}
			for _, value := range []string{got.ContactProfileURL, got.FirstName, got.LastName, got.CompanyName, got.CompanyDomain} {
				if value != "" && value != "example" {
					t.Fatalf("untrimmed value: %q", value)
				}
			}
		})
	}
}

func TestValidatePhoneWaitFlags(t *testing.T) {
	for _, tt := range []struct {
		name       string
		maxWait    int
		noWait     bool
		maxWaitSet bool
		wantErr    string
	}{
		{name: "minimum", maxWait: 1},
		{name: "maximum", maxWait: 120, maxWaitSet: true},
		{name: "no wait", maxWait: 120, noWait: true},
		{name: "negative", maxWait: -1, wantErr: "--max-wait must be between 1 and 120"},
		{name: "zero", maxWait: 0, wantErr: "--max-wait must be between 1 and 120"},
		{name: "too high", maxWait: 121, wantErr: "--max-wait must be between 1 and 120"},
		{name: "conflicting explicit default", maxWait: 120, noWait: true, maxWaitSet: true, wantErr: "--no-wait and --max-wait are mutually exclusive"},
		{name: "conflicting explicit minimum", maxWait: 1, noWait: true, maxWaitSet: true, wantErr: "--no-wait and --max-wait are mutually exclusive"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validatePhoneFlags(client.FindPhoneRequest{ContactProfileURL: "profile validated by server"}, tt.maxWait, tt.noWait, tt.maxWaitSet)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestPhoneCommandsRegistered(t *testing.T) {
	contact := newContactCmd()
	phone, remaining, err := contact.Find([]string{"find-phone"})
	if err != nil || phone.Name() != "find-phone" || len(remaining) != 0 {
		t.Fatalf("phone command not found: %v, %v", remaining, err)
	}
	if phone.Args(phone, []string{"gte"}) == nil {
		t.Fatal("find-phone must reject positional arguments")
	}
	get, remaining, err := contact.Find([]string{"find-phone", "get", "example-run"})
	if err != nil || get.Name() != "get" || !reflect.DeepEqual(remaining, []string{"example-run"}) {
		t.Fatalf("get command not found: %v, %v", remaining, err)
	}
	for _, args := range [][]string{nil, {"one"}, {"one", "two"}} {
		if err := get.Args(get, args); (err == nil) != (len(args) == 1) {
			t.Fatalf("get args %v: %v", args, err)
		}
	}
}

func TestPhoneCreditConfirmation(t *testing.T) {
	oldYes, oldQuiet, oldStdin := yes, quiet, os.Stdin
	t.Cleanup(func() { yes, quiet, os.Stdin = oldYes, oldQuiet, oldStdin })
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()
	os.Stdin, yes, quiet = stdin, false, false
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/v1/credits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"remainingCredits":1234.5}`))
	}))
	defer srv.Close()
	c := client.New(srv.URL, "test-key", "test", false, io.Discard)
	err = confirmCreditActionWithMessage(c, context.Background(), phoneCreditMessage)
	want := "This lookup consumes credits only if a mobile number is found (balance: 1,234.5). Proceed? [y/N] use --yes to confirm non-interactively"
	if err == nil || err.Error() != want {
		t.Fatalf("confirmation = %v, want %q", err, want)
	}
	err = confirmCreditAction(c, context.Background())
	want = "This will consume credits (balance: 1,234.5). Proceed? [y/N] use --yes to confirm non-interactively"
	if err == nil || err.Error() != want {
		t.Fatalf("existing confirmation changed: %v", err)
	}
	if requests.Load() != 2 {
		t.Fatalf("credit requests = %d, want 2", requests.Load())
	}
	for _, flags := range [][2]bool{{true, false}, {false, true}} {
		yes, quiet = flags[0], flags[1]
		if err := confirmCreditActionWithMessage(nil, context.Background(), phoneCreditMessage); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPhoneLookupError(t *testing.T) {
	for _, tt := range []struct {
		err  error
		read bool
		want string
	}{
		{&client.APIError{StatusCode: 404}, false, "phone lookup is not enabled for your organization"},
		{&client.APIError{StatusCode: 404}, true, "phone lookup run not found"},
		{&client.APIError{StatusCode: 402}, false, "not enough credits; check your balance with: saber credits"},
		{&client.APIError{StatusCode: 422, Message: "invalid run id"}, true, "API error 422: invalid run id"},
		{&client.APIError{StatusCode: 401, Message: "invalid key"}, false, "API error 401: invalid key"},
		{&client.APIError{StatusCode: 429, Message: "try later", RetryAfter: 1}, false, "API error 429: try later (retry after 1s)"},
		{&client.APIError{StatusCode: 500, Message: "server error"}, false, "API error 500: server error"},
		{fmt.Errorf("request: %w", context.DeadlineExceeded), false, "request: context deadline exceeded; run the same command again; it continues the lookup in progress"},
		{context.DeadlineExceeded, true, "context deadline exceeded"},
		{context.Canceled, false, "context canceled"},
	} {
		if got := phoneLookupError(tt.err, tt.read); got.Error() != tt.want {
			t.Errorf("phoneLookupError(%v, %v) = %q, want %q", tt.err, tt.read, got, tt.want)
		}
	}
}

func TestPhoneCommandOutput(t *testing.T) {
	oldJSON, oldQuiet, oldVerbose, oldYes, oldURL, oldVersion := jsonOutput, quiet, verbose, yes, apiURL, cliVersion
	t.Cleanup(func() {
		jsonOutput, quiet, verbose, yes, apiURL, cliVersion = oldJSON, oldQuiet, oldVerbose, oldYes, oldURL, oldVersion
	})
	t.Setenv("SABER_API_KEY", "test-key")
	t.Setenv("SABER_NO_UPDATE_CHECK", "1")
	for _, tt := range []struct {
		name   string
		args   []string
		body   string
		method string
		header string
		want   string
	}{
		{name: "default wait", args: []string{"--profile", " example-profile "}, method: http.MethodPost, header: "120", body: `{"id":"example-run","type":"phone","status":"processing"}`, want: "Phone lookup still processing  ID: example-run\nRe-run later with: saber contact find-phone get example-run\n"},
		{name: "custom wait", args: []string{"--profile", "example-profile", "--max-wait", "37"}, method: http.MethodPost, header: "37", body: `{"type":"phone","status":"completed","result":{"found":false,"phone":null}}`, want: "No mobile number found.\n"},
		{name: "no wait completed repeat", args: []string{"--profile", "example-profile", "--no-wait"}, method: http.MethodPost, body: `{"type":"phone","status":"completed","result":{"found":false,"phone":null}}`, want: "No mobile number found.\n"},
		{name: "raw json", args: []string{"--profile", "example-profile", "--no-wait", "--json"}, method: http.MethodPost, body: "{\n\"type\":\"phone\",\"status\":\"processing\"\n}\n", want: "{\n\"type\":\"phone\",\"status\":\"processing\"\n}\n"},
		{name: "quiet", args: []string{"--profile", "example-profile", "--quiet"}, method: http.MethodPost, header: "120", body: `{"type":"phone","status":"completed"}`},
		{name: "quiet json", args: []string{"--profile", "example-profile", "--quiet", "--json"}, method: http.MethodPost, header: "120", body: `{"type":"phone","status":"completed"}`},
		{name: "get failed", args: []string{"get", "example-run"}, method: http.MethodGet, body: `{"type":"phone","status":"failed","error":"lookup unavailable","errorCode":"PHONE_LOOKUP_UNAVAILABLE","retryable":true}`, want: "Phone lookup failed: lookup unavailable (PHONE_LOOKUP_UNAVAILABLE)\nYou can retry.\n"},
		{name: "get raw", args: []string{"get", "example-run", "--json"}, method: http.MethodGet, body: `{"type":"phone","status":"completed"}`, want: `{"type":"phone","status":"completed"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				path := "/v1/contacts/enrich/phone"
				if tt.method == http.MethodGet {
					path = "/v1/enrichment-runs/example-run"
				}
				if r.Method != tt.method || r.URL.Path != path || r.Header.Get("X-Sbr-Timeout-Sec") != tt.header {
					t.Errorf("unexpected request: %s %s, wait %q", r.Method, r.URL.Path, r.Header.Get("X-Sbr-Timeout-Sec"))
				}
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			stdout, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			oldStdout := os.Stdout
			os.Stdout = stdout
			t.Cleanup(func() { os.Stdout = oldStdout })
			root := NewRootCmd("dev", "test", "test")
			args := []string{"contact", "find-phone", "--api-url", srv.URL}
			if tt.method == http.MethodPost {
				args = append(args, "--yes")
			}
			root.SetArgs(append(args, tt.args...))
			err = root.Execute()
			_ = stdout.Close()
			if err != nil {
				t.Fatal(err)
			}
			output, err := os.ReadFile(stdout.Name())
			if err != nil {
				t.Fatal(err)
			}
			if string(output) != tt.want || requests.Load() != 1 {
				t.Fatalf("output = %q, want %q; requests = %d, want 1", output, tt.want, requests.Load())
			}
		})
	}
}

func TestPhoneValidationBeforeClient(t *testing.T) {
	for _, args := range [][]string{
		{}, {"--profile", " "}, {"--first-name", "Jane", "--domain", "example.com"},
		{"--profile", "example-profile", "--max-wait", "0"},
		{"--profile", "example-profile", "--no-wait", "--max-wait", "120"},
		{"gte"}, {"get"}, {"get", "one", "two"},
	} {
		cmd := newContactFindPhoneCmd()
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("expected validation error for %s", strings.Join(args, " "))
		}
	}
}
