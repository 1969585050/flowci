package builder

import (
	"testing"
)

func TestLanguageConstants(t *testing.T) {
	tests := []struct {
		lang    Language
		want    string
	}{
		{LangJavaMaven, "java-maven"},
		{LangJavaGradle, "java-gradle"},
		{LangNodeJS, "nodejs"},
		{LangPython, "python"},
		{LangGo, "go"},
		{LangPHP, "php"},
		{LangRuby, "ruby"},
		{LangDotnet, "dotnet"},
		{LangCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			if got := string(tt.lang); got != tt.want {
				t.Errorf("Language = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildStatusConstants(t *testing.T) {
	tests := []struct {
		status  BuildStatus
		want    string
	}{
		{BuildStatusPending, "pending"},
		{BuildStatusRunning, "running"},
		{BuildStatusSuccess, "success"},
		{BuildStatusFailed, "failed"},
		{BuildStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Errorf("BuildStatus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	tests := []struct {
		lang    string
		want    bool
	}{
		{"java-maven", true},
		{"java-gradle", true},
		{"nodejs", true},
		{"python", true},
		{"go", true},
		{"php", true},
		{"ruby", true},
		{"dotnet", true},
		{"custom", true},
		{"unknown", false},
		{"rust", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			if got := IsSupportedLanguage(tt.lang); got != tt.want {
				t.Errorf("IsSupportedLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		lang    string
		wantErr bool
	}{
		{"go", false},
		{"python", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			err := ValidateLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguage(%q) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateBuildID(t *testing.T) {
	id1 := generateBuildID()
	id2 := generateBuildID()

	if id1 == "" {
		t.Error("generateBuildID() returned empty string")
	}

	if id1 == id2 {
		t.Errorf("generateBuildID() returned same ID twice: %q", id1)
	}

	if len(id1) < 10 {
		t.Errorf("generateBuildID() ID too short: %q", id1)
	}
}

func TestGenerateDockerfile(t *testing.T) {
	tests := []struct {
		lang    Language
		wantLen int
	}{
		{LangGo, 100},
		{LangPython, 100},
		{LangNodeJS, 100},
		{LangJavaMaven, 100},
		{LangJavaGradle, 100},
		{LangPHP, 100},
		{LangRuby, 100},
		{LangDotnet, 100},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			got := GenerateDockerfile(tt.lang)
			if len(got) < tt.wantLen {
				t.Errorf("GenerateDockerfile(%v) too short, got %d bytes, want at least %d",
					tt.lang, len(got), tt.wantLen)
			}
		})
	}
}

func TestGenerateDockerfileUnknownLanguage(t *testing.T) {
	got := GenerateDockerfile(Language("unknown"))
	if got != GenerateDockerfile(LangCustom) {
		t.Error("Unknown language should return custom template")
	}
}
