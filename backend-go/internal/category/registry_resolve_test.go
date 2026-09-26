package category

import "testing"

// Pins the T-CAT flexibility rule: one category can hold named answer formats,
// the requested one wins, and an unknown/empty name falls back to the default
// then to the legacy single template — never an error and never flat text by
// accident while a template exists.
func TestResolveSections(t *testing.T) {
	legacy := TemplateSection{Key: "pattern", Label: "Pattern", Type: SectionTypeProse}
	codeSec := TemplateSection{Key: "code", Label: "Code", Type: SectionTypeCode}
	approachSec := TemplateSection{Key: "approach", Label: "Approach", Type: SectionTypeProse}

	cases := []struct {
		name      string
		cat       *Category
		requested string
		wantLen   int
		wantName  string
		wantKey   string
	}{
		{
			name:      "legacy single template is unchanged",
			cat:       &Category{TemplateSchema: TemplateSchema{Sections: []TemplateSection{legacy}}},
			requested: "",
			wantLen:   1, wantName: "", wantKey: "pattern",
		},
		{
			name: "requested variant wins",
			cat: &Category{TemplateSchema: TemplateSchema{
				Templates: []Template{{Name: "Code", Sections: []TemplateSection{codeSec}}, {Name: "Approach", Sections: []TemplateSection{approachSec}}},
				Default:   "Approach",
			}},
			requested: "code", // case-insensitive
			wantLen:   1, wantName: "Code", wantKey: "code",
		},
		{
			name: "empty request uses the default variant",
			cat: &Category{TemplateSchema: TemplateSchema{
				Templates: []Template{{Name: "Code", Sections: []TemplateSection{codeSec}}, {Name: "Approach", Sections: []TemplateSection{approachSec}}},
				Default:   "Approach",
			}},
			requested: "",
			wantLen:   1, wantName: "Approach", wantKey: "approach",
		},
		{
			name: "unknown request falls back to first variant, not flat text",
			cat: &Category{TemplateSchema: TemplateSchema{
				Templates: []Template{{Name: "Code", Sections: []TemplateSection{codeSec}}},
			}},
			requested: "nonsense",
			wantLen:   1, wantName: "Code", wantKey: "code",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sections, name := tc.cat.ResolveSections(tc.requested)
			if len(sections) != tc.wantLen || name != tc.wantName || sections[0].Key != tc.wantKey {
				t.Fatalf("got len=%d name=%q key=%q, want len=%d name=%q key=%q",
					len(sections), name, sections[0].Key, tc.wantLen, tc.wantName, tc.wantKey)
			}
		})
	}

	// A nil category must not panic and must yield flat text.
	if secs, _ := (*Category)(nil).ResolveSections("code"); len(secs) != 0 {
		t.Fatal("nil category must resolve to no sections")
	}
}
