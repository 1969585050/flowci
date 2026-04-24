package docker

import (
	"reflect"
	"testing"
)

func TestParseImages(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect []Image
	}{
		{
			name:   "空输入",
			input:  "",
			expect: []Image{},
		},
		{
			name:   "只有空白",
			input:  "   \n\n\t",
			expect: []Image{},
		},
		{
			name:  "单行标准",
			input: "abc123|nginx|latest|142MB|2 hours ago",
			expect: []Image{
				{ID: "abc123", Repository: "nginx", Tag: "latest", Size: "142MB", CreatedAt: "2 hours ago"},
			},
		},
		{
			name:  "多行",
			input: "a1|nginx|latest|142MB|2h\nb2|redis|7|110MB|1d\nc3|myapp|v1.2|80MB|5m",
			expect: []Image{
				{ID: "a1", Repository: "nginx", Tag: "latest", Size: "142MB", CreatedAt: "2h"},
				{ID: "b2", Repository: "redis", Tag: "7", Size: "110MB", CreatedAt: "1d"},
				{ID: "c3", Repository: "myapp", Tag: "v1.2", Size: "80MB", CreatedAt: "5m"},
			},
		},
		{
			name:  "中间空行要跳过",
			input: "a1|x|y|1MB|now\n\nb2|p|q|2MB|now",
			expect: []Image{
				{ID: "a1", Repository: "x", Tag: "y", Size: "1MB", CreatedAt: "now"},
				{ID: "b2", Repository: "p", Tag: "q", Size: "2MB", CreatedAt: "now"},
			},
		},
		{
			name:   "字段不足 5 段整行跳过",
			input:  "id|repo|tag|size",
			expect: []Image{},
		},
		{
			name:  "合法行与残缺行混合",
			input: "a1|nginx|latest|142MB|2h\nbad|line\nb2|redis|7|110MB|1d",
			expect: []Image{
				{ID: "a1", Repository: "nginx", Tag: "latest", Size: "142MB", CreatedAt: "2h"},
				{ID: "b2", Repository: "redis", Tag: "7", Size: "110MB", CreatedAt: "1d"},
			},
		},
		{
			name:  "Repository 为 <none>",
			input: "sha256:abc|<none>|<none>|0B|5m",
			expect: []Image{
				{ID: "sha256:abc", Repository: "<none>", Tag: "<none>", Size: "0B", CreatedAt: "5m"},
			},
		},
		{
			name:  "Size 字段含空格不影响解析",
			input: "a1|nginx|latest|1.44 GB|2 hours ago",
			expect: []Image{
				{ID: "a1", Repository: "nginx", Tag: "latest", Size: "1.44 GB", CreatedAt: "2 hours ago"},
			},
		},
		{
			name:  "6 段输入（docker 未来格式兼容）只取前 5 段",
			input: "a1|nginx|latest|142MB|2h|extra",
			expect: []Image{
				{ID: "a1", Repository: "nginx", Tag: "latest", Size: "142MB", CreatedAt: "2h"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseImages(tc.input)
			if !reflect.DeepEqual(got, tc.expect) {
				t.Errorf("parseImages(%q):\n got = %+v\nwant = %+v", tc.input, got, tc.expect)
			}
		})
	}
}
