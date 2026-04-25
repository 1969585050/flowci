package ai

import (
	"context"
	"strings"
)

// 构建日志诊断 prompt。
const buildDiagnoseSystem = `你是 Docker 构建与 CI/CD 专家。
用户会贴 docker build 的输出日志，请你按下面格式输出**简洁的中文诊断**（Markdown）：

## 失败原因
（一句话指出根因。如果日志显示构建成功但用户仍来问，回复"日志显示构建成功，未发现失败"）

## 关键证据
（从日志中摘出 1-3 行最关键的错误信息，用代码块包起来）

## 建议修复
1. 第一条具体可操作建议
2. 第二条
3. ...（最多 5 条，按重要性排序）

注意：
- 不要重复用户的日志原文
- 修复建议要给具体命令或代码改动，不要泛泛而谈
- 如果根因不明显，列出 2-3 个最可能的方向`

// DiagnoseBuild 调用 LLM 分析 docker build 日志，返回 Markdown 诊断报告。
// log 过长时会截取末尾 30KB（错误信息通常在尾部）。
func (p *Provider) DiagnoseBuild(ctx context.Context, log string) (string, error) {
	const maxBytes = 30 * 1024
	trimmed := log
	if len(trimmed) > maxBytes {
		trimmed = "...(已截断前 " + lenStr(len(trimmed)-maxBytes) + " 字节)\n\n" + trimmed[len(trimmed)-maxBytes:]
	}
	user := "下面是 docker build 的输出，请诊断:\n\n```\n" + strings.TrimSpace(trimmed) + "\n```"
	return p.Chat(ctx, buildDiagnoseSystem, user)
}

// lenStr int → string 的最简版本（避免引入 strconv 仅为日志拼接）。
func lenStr(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
