package getitems

import (
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name      string
		csvText   string
		errCreate bool
		reqIDList []string
		code      int
		body      string
	}{
		{
			name: "ok 1",
			csvText: `#,id,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,7079,абвгдеёж
`,
			reqIDList: []string{"фыва", "872", "фыва"},
			code:      200,
			body: `[{"#":"2","id":"фыва","uid":"S-1-5-21-3686381713-1037878038-1682765610-2544"},` +
				`{"#":"1","id":"872","uid":"S-1-5-21-3686381713-1037878038-1682765610-1877"},` +
				`{"#":"2","id":"фыва","uid":"S-1-5-21-3686381713-1037878038-1682765610-2544"}]`,
		},
		{
			name: "ok 2",
			csvText: `id,fffid,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,7079,абвгдеёж
`,
			reqIDList: []string{"2", "1", "2"},
			code:      200,
			body: `[{"id":"2","fffid":"фыва","uid":"S-1-5-21-3686381713-1037878038-1682765610-2544"},` +
				`{"id":"1","fffid":"872","uid":"S-1-5-21-3686381713-1037878038-1682765610-1877"},` +
				`{"id":"2","fffid":"фыва","uid":"S-1-5-21-3686381713-1037878038-1682765610-2544"}]`,
		},
		{
			name: "ok empty",
			csvText: `id,fffid,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,7079,абвгдеёж
`,
			reqIDList: []string{},
			code:      200,
			body:      `[]`,
		},
		{
			name: "duplicate",
			csvText: `#,id,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,872,абвгдеёж
`,
			errCreate: true,
		},
		{
			name: "wrong row length",
			csvText: `#,id,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877,
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,7079,абвгдеёж
`,
			errCreate: true,
		},
		{
			name: "incorrect csv",
			csvText: `;lkjasdf
`,
			errCreate: true,
		},
		{
			name: "not found",
			csvText: `#,id,uid
1,872,S-1-5-21-3686381713-1037878038-1682765610-1877
2,фыва,S-1-5-21-3686381713-1037878038-1682765610-2544
3,7079,абвгдеёж
`,
			reqIDList: []string{"фыва", "8720", "фыва"},
			code:      404,
			body:      "id is absent in data",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "go_test_handler_")
			if err != nil {
				t.Fatalf("error creating temp file: %s", err)
			}
			name := f.Name()
			defer func() {
				err := os.Remove(name)
				if err != nil {
					t.Fatalf("error removing temp file: %s", err)
				}
			}()

			if _, err := f.WriteString(tt.csvText); err != nil {
				t.Fatalf("write to temp file: %s", err)
			}

			err = f.Close()
			if err != nil {
				t.Fatalf("error closing temp file: %s", err)
			}

			h, err := NewHandler(name)
			assert.Equal(t, tt.errCreate, err != nil, "wrong error status, err: %#v", err)

			if tt.errCreate || err != nil {
				return
			}

			var u url.URL
			q := u.Query()
			q["id"] = tt.reqIDList

			req := httptest.NewRequest("GET", "/get-items?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			resp := w.Result()
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					t.Fatalf("closing body: %s", err)
				}
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("ReadAll from response body: %s", err)
			}

			assert.Equal(t, tt.body, string(body), "bad body")
			assert.Equal(t, tt.code, resp.StatusCode, "bad response code")
		})
	}
}
