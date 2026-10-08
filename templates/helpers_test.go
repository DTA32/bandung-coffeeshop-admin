package templates

import "testing"

func TestDisplayValue(t *testing.T) {
	cases := []struct {
		field, raw, want string
	}{
		{"hours", `{"open":"08:00","close":"22:00"}`, "08:00 – 22:00"},
		{"hours", `{"open":null,"close":null}`, "—"},
		{"phone", `"+62227777"`, "+62227777"},
		{"temp_closed", `true`, "yes"},
		{"menu_content", `{"url":"https://x/menu.pdf","hash":"abc"}`, "https://x/menu.pdf"},
		{"phone", `null`, "—"},
		{"phone", ``, "—"},
	}
	for _, tc := range cases {
		if got := displayValue(tc.field, []byte(tc.raw)); got != tc.want {
			t.Errorf("displayValue(%s, %s) = %q, want %q", tc.field, tc.raw, got, tc.want)
		}
	}
}

func TestHoursDays(t *testing.T) {
	got := hoursDays([]byte(`{"open":"08:00","close":"22:00","days":[{"day":"Senin","hours":"08.00–22.00"},{"day":"Selasa","hours":"Tutup","special":true}]}`))
	if len(got) != 2 || got[1] != "Selasa: Tutup (special hours)" {
		t.Fatalf("%v", got)
	}
}

func TestThumb(t *testing.T) {
	ImageBase = "https://image.bdgcafe.com"
	if got := thumb("https://image.bdgcafe.com/ig/a.jpg", 320); got != "https://image.bdgcafe.com/ig/a.jpg?w=320" {
		t.Fatal(got)
	}
	if got := thumb("https://elsewhere/a.jpg", 320); got != "https://elsewhere/a.jpg" {
		t.Fatal(got)
	}
}
