package deployer

import (
	"testing"

	"github.com/docker/docker/api/types"
)

func TestEnvMapToSlice(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantLen int
	}{
		{
			name:    "empty map",
			env:     {},
			wantLen: 0,
		},
		{
			name:    "single entry",
			env:     map[string]string{"KEY": "value"},
			wantLen: 1,
		},
		{
			name:    "multiple entries",
			env:     map[string]string{"KEY1": "value1", "KEY2": "value2"},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := envMapToSlice(tt.env)
			if len(got) != tt.wantLen {
				t.Errorf("envMapToSlice() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestExposedPorts(t *testing.T) {
	ports := []PortMapping{
		{HostPort: 8080, ContainerPort: 8080, Protocol: "tcp"},
		{HostPort: 443, ContainerPort: 443, Protocol: "tcp"},
	}

	result := exposedPorts(ports)

	if len(result) != 2 {
		t.Errorf("exposedPorts() returned %d ports, want 2", len(result))
	}

	expected := []types.Port{
		"8080/tcp",
		"443/tcp",
	}

	for _, exp := range expected {
		if _, ok := result[exp]; !ok {
			t.Errorf("exposedPorts() missing expected port %s", exp)
		}
	}
}

func TestPortBindings(t *testing.T) {
	ports := []PortMapping{
		{HostPort: 8080, ContainerPort: 8080, Protocol: "tcp"},
	}

	result := portBindings(ports)

	if len(result) != 1 {
		t.Errorf("portBindings() returned %d bindings, want 1", len(result))
	}

	binding := result["8080/tcp"]
	if len(binding) != 1 {
		t.Errorf("portBindings() binding count = %d, want 1", len(binding))
	}

	if binding[0].HostPort != "8080" {
		t.Errorf("portBindings() HostPort = %q, want %q", binding[0].HostPort, "8080")
	}

	if binding[0].HostIP != "0.0.0.0" {
		t.Errorf("portBindings() HostIP = %q, want %q", binding[0].HostIP, "0.0.0.0")
	}
}

func TestSplitVolume(t *testing.T) {
	tests := []struct {
		name     string
		volume   string
		wantLen  int
		wantSrc  string
		wantTgt  string
	}{
		{
			name:    "with colon separator",
			volume:  "/host/path:/container/path",
			wantLen: 2,
			wantSrc: "/host/path",
			wantTgt: "/container/path",
		},
		{
			name:    "without separator",
			volume:  "/single/path",
			wantLen: 1,
			wantSrc: "/single/path",
			wantTgt: "",
		},
		{
			name:    "empty string",
			volume:  "",
			wantLen: 1,
			wantSrc: "",
			wantTgt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitVolume(tt.volume)
			if len(got) != tt.wantLen {
				t.Errorf("splitVolume(%q) len = %d, want %d", tt.volume, len(got), tt.wantLen)
			}
			if len(got) >= 1 && got[0] != tt.wantSrc {
				t.Errorf("splitVolume(%q)[0] = %q, want %q", tt.volume, got[0], tt.wantSrc)
			}
			if len(got) >= 2 && got[1] != tt.wantTgt {
				t.Errorf("splitVolume(%q)[1] = %q, want %q", tt.volume, got[1], tt.wantTgt)
			}
		})
	}
}

func TestTrimSlash(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"with leading slash", "/container", "container"},
		{"without leading slash", "container", "container"},
		{"empty string", "", ""},
		{"single slash", "/", ""},
		{"multiple slashes", "//container", "/container"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimSlash(tt.arg); got != tt.want {
				t.Errorf("trimSlash(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

func TestGenerateDeployID(t *testing.T) {
	id1 := generateDeployID()
	id2 := generateDeployID()

	if id1 == "" {
		t.Error("generateDeployID() returned empty string")
	}

	if id1 == id2 {
		t.Errorf("generateDeployID() returned same ID twice: %q", id1)
	}

	if len(id1) < 10 {
		t.Errorf("generateDeployID() ID too short: %q", id1)
	}
}

func TestDeployStatusConstants(t *testing.T) {
	tests := []struct {
		status  DeployStatus
		want    string
	}{
		{DeployStatusPending, "pending"},
		{DeployStatusRunning, "running"},
		{DeployStatusStopped, "stopped"},
		{DeployStatusFailed, "failed"},
		{DeployStatusRolling, "rolling_update"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Errorf("DeployStatus = %v, want %v", got, tt.want)
			}
		})
	}
}
