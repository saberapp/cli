package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestFindPhone(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted} {
		for _, wait := range []int{-1, 0, 120} {
			t.Run(strconv.Itoa(status)+"/"+strconv.Itoa(wait), func(t *testing.T) {
				req := FindPhoneRequest{
					ContactProfileURL: "https://www.linkedin.com/in/jane-doe-example",
					FirstName:         "Jane", LastName: "Doe", CompanyName: "Example", CompanyDomain: "example.com",
				}
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost || r.URL.Path != "/v1/contacts/enrich/phone" {
						t.Errorf("request = %s %s", r.Method, r.URL.Path)
					}
					wantHeader := ""
					if wait > 0 {
						wantHeader = strconv.Itoa(wait)
					}
					if got := r.Header.Get("X-Sbr-Timeout-Sec"); got != wantHeader {
						t.Errorf("wait header = %q, want %q", got, wantHeader)
					}
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					want := map[string]string{"contactProfileUrl": req.ContactProfileURL, "firstName": "Jane", "lastName": "Doe", "companyName": "Example", "companyDomain": "example.com"}
					if !reflect.DeepEqual(body, want) {
						t.Errorf("body = %v, want %v", body, want)
					}
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"id":"example-run","type":"phone","status":"processing","result":null,"credits":{"state":"reserved"}}`))
				}))
				defer srv.Close()
				c := New(srv.URL, "test-key", "test", false, io.Discard)
				run, err := c.FindPhone(context.Background(), req, wait, nil)
				if err != nil {
					t.Fatal(err)
				}
				if run.ID != "example-run" || run.Status != "processing" || run.Result != nil || run.Credits.State != "reserved" || run.Credits.Amount != nil {
					t.Fatalf("unexpected run: %+v", run)
				}
			})
		}
	}
}

func TestPhoneRunDecode(t *testing.T) {
	for _, status := range []string{"completed", "failed", "future-status"} {
		t.Run(status, func(t *testing.T) {
			body := map[string]any{
				"id": "example-run", "object": "enrichment_run", "entityType": "contact", "type": "phone", "status": status,
				"entity":              map[string]any{"contactProfileUrl": "https://www.linkedin.com/in/jane-doe-example", "domain": "example.com"},
				"result":              map[string]any{"found": true, "phone": map[string]any{"e164": "+447700900123", "display": "+44 7700 900123", "lineType": "mobile", "lineStatus": "active", "country": "GB"}},
				"resultSchemaVersion": "1", "credits": map[string]any{"state": "future-state", "amount": math.Pi, "priceVersion": "example-version"},
				"createdAt": "2026-10-29T10:00:00Z", "completedAt": "2026-10-29T10:00:42Z",
				"error": "lookup unavailable", "errorCode": "PHONE_LOOKUP_UNAVAILABLE", "retryable": true, "futureField": true,
			}
			if status != "completed" {
				body["result"] = nil
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(body); err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()
			c := New(srv.URL, "test-key", "test", false, io.Discard)
			run, err := c.GetPhoneRun(context.Background(), "example-run", nil)
			if err != nil {
				t.Fatal(err)
			}
			if run.Status != status || run.Object != "enrichment_run" || run.EntityType != "contact" || run.Entity.Domain != "example.com" || run.Entity.ContactProfileURL == "" || run.ResultSchemaVersion != "1" || run.CreatedAt == "" || run.CompletedAt == "" {
				t.Fatalf("unexpected run: %+v", run)
			}
			if status == "completed" {
				if run.Result == nil || !run.Result.Found || run.Result.Phone.E164 != "+447700900123" || run.Result.Phone.Display != "+44 7700 900123" || run.Result.Phone.LineType != "mobile" || run.Result.Phone.LineStatus != "active" || run.Result.Phone.Country != "GB" {
					t.Fatalf("unexpected result: %+v", run.Result)
				}
			} else if run.Result != nil {
				t.Fatalf("unfinished run has a result: %+v", run.Result)
			}
			if run.Credits.State != "future-state" || run.Credits.Amount == nil || *run.Credits.Amount != math.Pi || run.Credits.PriceVersion != "example-version" || !run.Retryable || run.Error != "lookup unavailable" || run.ErrorCode != "PHONE_LOOKUP_UNAVAILABLE" {
				t.Fatalf("unexpected credits or error: %+v", run)
			}
		})
	}
}

func TestPhoneRawResponses(t *testing.T) {
	for _, get := range []bool{false, true} {
		for _, status := range []int{http.StatusOK, http.StatusAccepted} {
			t.Run(strconv.FormatBool(get)+"/"+strconv.Itoa(status), func(t *testing.T) {
				body := "{\n  \"id\":\"example-run\",\"type\":\"phone\",\"status\":\"processing\",\"unknown\":true\n}\n"
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(status)
					_, _ = io.WriteString(w, body)
				}))
				defer srv.Close()
				c := New(srv.URL, "test-key", "test", false, io.Discard)
				var out bytes.Buffer
				var err error
				if get {
					_, err = c.GetPhoneRun(context.Background(), "example-run", &out)
				} else {
					_, err = c.FindPhone(context.Background(), FindPhoneRequest{}, 0, &out)
				}
				if err != nil || out.String() != body {
					t.Fatalf("error = %v, output = %q, want %q", err, out.String(), body)
				}
			})
		}
	}
}

func TestGetPhoneRunRejectsOtherTypes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"example-run","type":"signal","status":"completed"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "test-key", "test", false, io.Discard)
	for _, raw := range []bool{false, true} {
		var out bytes.Buffer
		var dst io.Writer
		if raw {
			dst = &out
		}
		_, err := c.GetPhoneRun(context.Background(), "example-run", dst)
		if err == nil || err.Error() != "run example-run is not a phone lookup" || out.Len() != 0 {
			t.Fatalf("error = %v, output = %q", err, out.String())
		}
	}
}

func TestGetPhoneRunEscapesID(t *testing.T) {
	id := "example/run?key=value#fragment"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.EscapedPath() != "/v1/enrichment-runs/"+url.PathEscape(id) || r.URL.RawQuery != "" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		_, _ = w.Write([]byte(`{"type":"phone"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "test-key", "test", false, io.Discard)
	if _, err := c.GetPhoneRun(context.Background(), id, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPhoneAPIErrors(t *testing.T) {
	for _, status := range []int{401, 402, 404, 422, 429, 500} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				if status == 429 {
					_, _ = w.Write([]byte(`{"statusCode":429,"message":"try later","retryAfter":1}`))
				} else {
					_, _ = w.Write([]byte(`{"error":{"type":"validation","code":"EXAMPLE","message":"example error"}}`))
				}
			}))
			defer srv.Close()
			c := New(srv.URL, "test-key", "test", false, io.Discard)
			var out bytes.Buffer
			_, err := c.FindPhone(context.Background(), FindPhoneRequest{}, 0, &out)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status || out.Len() != 0 {
				t.Fatalf("error = %v, output = %q", err, out.String())
			}
		})
	}
}
