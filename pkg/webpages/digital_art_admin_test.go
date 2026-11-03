package webpages

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"git.aetherial.dev/aeth/keiji/pkg/storage"
)

func TestDigitalArtAdminTemplates(t *testing.T) {
	tmpl, err := template.ParseFS(content, "html/*.html")
	if err != nil {
		t.Fatal(err)
	}
	image := storage.Image{Ident: "abc", Title: "A <title>", Desc: "A & description", MediaType: "image/png"}
	video := storage.Image{Ident: "vid", Title: "Video", MediaType: "video/webm"}
	cases := []struct {
		name string
		data any
		want []string
	}{
		{"digital_art_admin_list", map[string]any{"Images": []storage.Image{image, video}}, []string{"/admin/digital/abc", "/admin/digital/vid", "<img ", "<video ", "A &lt;title&gt;", "A &amp; description"}},
		{"digital_art_admin_list", map[string]any{"Images": []storage.Image{}}, []string{"No digital art found."}},
		{"digital_art_admin_editor", image, []string{"action=\"/admin/digital/abc\"", "method=\"post\"", "hx-post=\"/admin/digital/abc\"", "name=\"title\"", "name=\"description\"", "A &lt;title&gt;", "A &amp; description", "<img ", "hx-get=\"/admin/digital\"", "digital-art-response", `hx-delete="/admin/digital/abc"`, "hx-confirm", "media remains"}},
		{"digital_art_admin_editor", video, []string{"<video ", "/api/v1/images/vid", "controls"}},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		if err := tmpl.ExecuteTemplate(&out, tc.name, tc.data); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		for _, want := range tc.want {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s: expected %q in output", tc.name, want)
			}
		}
	}
}
