package docker

import (
	"reflect"
	"testing"
)

func TestParseContainers(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect []Container
	}{
		{
			name:   "空输入",
			input:  "",
			expect: []Container{},
		},
		{
			name:   "空白",
			input:  "\n\t \n",
			expect: []Container{},
		},
		{
			name:  "单行完整 6 段",
			input: "abc123|web|nginx|running|Up 5m|0.0.0.0:80->80/tcp",
			expect: []Container{
				{ID: "abc123", Names: []string{"web"}, Image: "nginx", State: "running", Status: "Up 5m", Ports: "0.0.0.0:80->80/tcp"},
			},
		},
		{
			name:  "Ports 字段为空（容器无端口映射）",
			input: "id1|bg|busybox|exited|Exited 5m|",
			expect: []Container{
				{ID: "id1", Names: []string{"bg"}, Image: "busybox", State: "exited", Status: "Exited 5m", Ports: ""},
			},
		},
		{
			name:  "只有 5 段（docker 漏 Ports 字段）应补空",
			input: "id1|bg|busybox|exited|Exited 5m",
			expect: []Container{
				{ID: "id1", Names: []string{"bg"}, Image: "busybox", State: "exited", Status: "Exited 5m", Ports: ""},
			},
		},
		{
			name:  "多行",
			input: "a|n1|nginx|running|Up 1m|80/tcp\nb|n2|redis|running|Up 2m|6379/tcp\nc|n3|app|exited|Exited|",
			expect: []Container{
				{ID: "a", Names: []string{"n1"}, Image: "nginx", State: "running", Status: "Up 1m", Ports: "80/tcp"},
				{ID: "b", Names: []string{"n2"}, Image: "redis", State: "running", Status: "Up 2m", Ports: "6379/tcp"},
				{ID: "c", Names: []string{"n3"}, Image: "app", State: "exited", Status: "Exited", Ports: ""},
			},
		},
		{
			name:  "空行跳过",
			input: "a|n1|nginx|running|Up|80/tcp\n\nb|n2|redis|running|Up|6379/tcp",
			expect: []Container{
				{ID: "a", Names: []string{"n1"}, Image: "nginx", State: "running", Status: "Up", Ports: "80/tcp"},
				{ID: "b", Names: []string{"n2"}, Image: "redis", State: "running", Status: "Up", Ports: "6379/tcp"},
			},
		},
		{
			name:  "Status 含空格（最常见）",
			input: "id|web|nginx|running|Up 3 days (healthy)|0.0.0.0:80->80/tcp",
			expect: []Container{
				{ID: "id", Names: []string{"web"}, Image: "nginx", State: "running", Status: "Up 3 days (healthy)", Ports: "0.0.0.0:80->80/tcp"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseContainers(tc.input)
			if !reflect.DeepEqual(got, tc.expect) {
				t.Errorf("parseContainers(%q):\n got = %+v\nwant = %+v", tc.input, got, tc.expect)
			}
		})
	}
}
