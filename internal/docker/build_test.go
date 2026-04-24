package docker

import "testing"

func TestSplitTag(t *testing.T) {
	cases := []struct {
		input    string
		wantName string
		wantTag  string
	}{
		{"nginx:latest", "nginx", "latest"},
		{"nginx", "nginx", "latest"},                           // 无冒号 → 默认 latest
		{"my/app:v1.2.3", "my/app", "v1.2.3"},                  // 路径前缀
		{"registry.example.com:5000/my/app:v1", "registry.example.com", "5000/my/app:v1"}, // SplitN 2 段：端口被当 tag 一部分（已知弱点）
		{"", "", "latest"},                                     // 空串
		{":tag", "", "tag"},                                    // 只有 tag
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			gotName, gotTag := splitTag(tc.input)
			if gotName != tc.wantName || gotTag != tc.wantTag {
				t.Errorf("splitTag(%q) = (%q, %q); want (%q, %q)",
					tc.input, gotName, gotTag, tc.wantName, tc.wantTag)
			}
		})
	}
}
