# FlowCI Design System v1

> Single source of truth for FlowCI 桌面端的视觉与交互。所有 Vue 组件 / 样式必须消费这里定义的 token，不允许 hard-code 颜色 / 字体 / 间距。
>
> **文件实现**：
> - [`frontend/src/styles/theme.css`](../../frontend/src/styles/theme.css) — design tokens（颜色、字体、间距、圆角、阴影、动效）
> - [`frontend/src/styles/components.css`](../../frontend/src/styles/components.css) — 通用组件 class（`.btn` / `.card` / `.badge` / `.table-row` / `.input`）
> - 样板页：[`frontend/src/views/ProjectsView.vue`](../../frontend/src/views/ProjectsView.vue)

---

## 1. 调性

**冷静技术感** — 参考 Linear / Vercel / Raycast。

- **暗色优先**：默认 `[data-theme="dark"]`，亮色作为可切换 override
- **品牌色**：冷青蓝单色（`#3b82f6` → `#2563eb` → `#1d4ed8`），不用渐变
- **状态色**：饱和度压低，避免亮蓝紫粉的"消费级"感
- **等宽字体**：用于 commit hash / 镜像 tag / 容器名 / 路径 / 日志 — 工具属性的灵魂
- **8-grid 间距**：`--space-2` (8px) 为基线；`--space-1` (4px) 仅作微调
- **稳定 hover**：用 background / border / color 过渡，不用 `transform` 抖动（避免布局抖动）

## 2. Design Tokens（速查）

完整定义在 `frontend/src/styles/theme.css`。下表只列每类 token 的命名与典型用途。

### 2.1 颜色

| 类别 | 变量 | 暗色 | 用法 |
|------|------|------|------|
| Surface | `--bg-canvas`   | `#0b0d12` | 最底层背景（content 区） |
| Surface | `--bg-surface`  | `#11141b` | 卡片表面（最常用） |
| Surface | `--bg-elevated` | `#161a23` | 弹窗、菜单（浮起元素） |
| Surface | `--bg-sunken`   | `#0a0c11` | 输入框、日志块（凹陷元素） |
| Surface | `--bg-hover`    | `#1a1f2a` | 表面 hover 态 |
| Text | `--text-primary`   | `#e6e8ef` | 主文本（对比 12.4:1） |
| Text | `--text-secondary` | `#a4abbd` | 次文本 |
| Text | `--text-muted`    | `#6b7280` | 淡文本（AA 4.5:1） |
| Text | `--text-ghost`    | `#4a5061` | 装饰，不能放正文 |
| Border | `--border-subtle`  | `#222632` | 默认细线 |
| Border | `--border-default` | `#2a2f3d` | 卡片边框 |
| Border | `--border-strong`  | `#3a4050` | 强调边框 |
| Brand | `--brand-500` | `#3b82f6` | 主品牌色（按钮主色 / focus） |
| Brand | `--brand-600` | `#2563eb` | 按钮 hover |
| Brand | `--brand-700` | `#1d4ed8` | 按钮 active |
| Brand | `--brand-100` | `rgba(59,130,246,0.12)` | 透明品牌背景 |
| Status | `--success-fg/bg/bd` | `#56d364` 系 | 成功 / 完成 |
| Status | `--danger-fg/bg/bd`  | `#f85149` 系 | 失败 / 删除 |
| Status | `--warning-fg/bg/bd` | `#e3b341` 系 | 警告 / 进行中 |
| Status | `--info-fg/bg/bd`    | `#58a6ff` 系 | 提示 / 中性信息 |

亮色色板见 `theme.css` 的 `[data-theme="light"]` 定义。

### 2.2 字体

| 变量 | 值 | 用途 |
|------|----|----|
| `--font-sans` | system-ui 栈 + Inter + 中文 fallback | 默认全局 |
| `--font-mono` | JetBrains Mono / SF Mono / Consolas | 技术性内容 |
| `--text-xs`  | 11px | 微辅助标签 |
| `--text-sm`  | 12px | 次文本 / 表格内 |
| `--text-base`| 13px | **body 默认** |
| `--text-md`  | 14px | 卡片标题 / 按钮内 |
| `--text-lg`  | 16px | 子标题 |
| `--text-xl`  | 18px | 模态标题 |
| `--text-2xl` | 22px | 区块标题 |
| `--text-3xl` | 28px | 页面 H1 |
| `--weight-regular/medium/semibold` | 400/500/600 | 不用 700 |
| `--leading-tight/base/relaxed` | 1.3 / 1.5 / 1.7 | — |

### 2.3 间距 / 圆角 / 阴影 / 动效

| 类别 | 变量 | 值 | 用法 |
|------|------|----|----|
| Space | `--space-1`..`--space-16` | 4px..64px | 8-grid（4/12/20 为微调） |
| Radius | `--radius-xs/sm/md/lg/xl/pill` | 3/4/6/8/12/999px | md 默认 |
| Shadow | `--shadow-sm/md/lg` | 暗色重 / 亮色轻 | 浮起元素 |
| Motion | `--duration-fast/base/slow` | 120/180/280ms | 动效 |
| Motion | `--easing-out`, `--easing-in-out` | cubic-bezier | — |
| Z-index | `--z-dropdown`..`--z-toast` | 10..60 | 显式分层 |

### 2.4 向后兼容 alias

旧变量名（`--bg-primary`, `--card-bg`, `--brand-start/end/soft`, `--shadow-sm` 等）全部映射到新 token。其他 13 个 view 暂未迁移仍可正常渲染，**新代码请直接用新变量**。

## 3. 四类核心组件 class

完整定义见 `frontend/src/styles/components.css`。

### 3.1 `.btn`

```html
<button class="btn btn-primary">主操作</button>
<button class="btn btn-secondary">次操作</button>
<button class="btn btn-ghost">无 chrome</button>
<button class="btn btn-danger">删除</button>
<button class="btn btn-link">链接型</button>
<button class="btn btn-icon" aria-label="更多">⋯</button>

<button class="btn btn-primary btn-sm">小号</button>
<button class="btn btn-secondary btn-lg">大号</button>
```

### 3.2 `.card`

```html
<div class="card">静态卡片</div>
<div class="card card-interactive">可点击卡片（hover 边框加重，无 transform）</div>
<div class="card card-pinned">置顶 / 选中状态（品牌色描边）</div>
<div class="card card-sunken">凹陷子面板（如 build 状态条）</div>
```

### 3.3 `.badge`

```html
<span class="badge badge-neutral">Node.js</span>
<span class="badge badge-brand">Beta</span>
<span class="badge badge-success">success</span>
<span class="badge badge-danger">failed</span>
<span class="badge badge-warning">building</span>
<span class="badge badge-info">info</span>

<!-- 等宽 badge：commit hash / 镜像 tag -->
<span class="badge badge-mono badge-neutral">a3f2e1c</span>
<!-- 状态点：runner online / agent idle -->
<span class="badge badge-dot badge-success">online</span>
```

### 3.4 `.table-row`

```html
<div class="table-row table-row-header">
  <div>Name</div><div>Status</div><div>Updated</div>
</div>
<div class="table-row table-row-clickable" @click="...">
  <div>my-image</div><div>...</div><div>3 分钟前</div>
</div>
<div class="table-row table-row-selected">...</div>
```

### 3.5 `.input` / `.select` / `.textarea`

```html
<input class="input" v-model="x" placeholder="..." />
<select class="select" v-model="y">...</select>
<textarea class="textarea" v-model="z"></textarea>
```

聚焦时边框变 `--brand-500`，背景从 `--bg-sunken` 提升到 `--bg-surface`。

## 4. Do / Don't

### 颜色

| ✅ Do | ❌ Don't |
|------|---------|
| `background: var(--brand-500);` | `background: #3b82f6;` |
| `color: var(--text-primary);` | `color: #e6e8ef;` |
| `border: 1px solid var(--border-default);` | `border: 1px solid #2a2f3d;` |
| 用 `--brand-500` 单色 | 用 `linear-gradient(135deg, #667eea, #764ba2)` |

### 字体

| ✅ Do | ❌ Don't |
|------|---------|
| `font-family: var(--font-mono);` 用于 hash / tag / 路径 | 用 sans-serif 显示 commit hash |
| `font-size: var(--text-sm);` (12px) | `font-size: 12.5px;`（破坏 scale） |
| `font-weight: var(--weight-medium);` (500) | `font-weight: 700;`（除 H1 不需要） |
| 把 mono 字体写在 `class="mono"` 工具类 | 在每个组件里重复 `'JetBrains Mono', monospace` |

### 间距

| ✅ Do | ❌ Don't |
|------|---------|
| `padding: var(--space-4);` (16px) | `padding: 15px;`（破坏 8-grid） |
| `gap: var(--space-2);` (8px) | `gap: 7px;` |
| `--space-1` 用于 icon-text gap、badge inner padding | `--space-1` 用于卡片间距（应该 `--space-4` 起步） |

### Hover 与交互

| ✅ Do | ❌ Don't |
|------|---------|
| `:hover { background: var(--bg-hover); }` | `:hover { transform: translateY(-2px); }`（页面抖动） |
| 所有可点击元素加 `cursor: pointer;` | 让用户猜哪里能点 |
| icon-only 按钮加 `aria-label="..."` | `<button>⋯</button>` 没有可读标签 |
| `:focus-visible` 显示 outline | `:focus { outline: none; }` 不补任何指示 |
| 过渡时长 `var(--duration-fast)` (120ms) | 过渡 500ms+（感觉迟钝） |

### Z-index

显式用 `--z-dropdown` / `--z-modal` / `--z-toast`，不要 `z-index: 9999;` / `z-index: 1;` 这样的魔数。

### Emoji 与 icon

当前阶段：模板内的 emoji（🌿 📦 ⚙️ 等）作为占位**保留**。

长期方向：迁到 SVG icon set（推荐 [Lucide](https://lucide.dev) — 跟 Linear/Vercel 同源），建立 `frontend/src/components/Icon.vue`。这不在本次 design system v1 范围内。

## 5. 主题切换

通过 `useSettings` composable 切换：`dark` / `light` / `system`。

```ts
const { theme, isDark, setTheme } = useSettings()
setTheme('dark')  // 或 'light' / 'system'
```

底层：在 `<html>` 上设 `data-theme="dark"|"light"`。
- `system` 模式跟随 `prefers-color-scheme`
- 切换时所有 `var(--xxx)` 自动重新解析

侧边栏永远是暗色（无论主题），用 `--bg-sidebar` 等独立变量驱动。

## 6. 迁移路线

| 阶段 | 状态 | 内容 |
|------|------|------|
| v1（当前） | ✅ 已落地 | tokens + 4 类核心 class + ProjectsView 样板 |
| v1.1 | TODO | 把另外 13 个 view 的 hard-code 颜色 / `linear-gradient(...)` 全部替换为 token |
| v1.2 | TODO | 抽 `<AppButton>` / `<AppCard>` / `<AppBadge>` 等 Vue 组件，渐进替换 class 写法 |
| v2 | TODO | SVG icon set（Lucide）+ `<Icon>` 组件 |
| v2.1 | TODO | 字体本地化（打包 JetBrains Mono Variable，不依赖系统是否安装） |
| v2.2 | TODO | 主题色可定制（用户在设置里选品牌色 - 蓝/绿/紫） |

## 7. 审查清单

提交触动样式的 PR 前自查：

- [ ] 没有 hard-code 颜色（hex / rgb / hsl）— 全部用 `var(--xxx)`
- [ ] 没有紫蓝渐变 `linear-gradient(135deg, #667eea ...)` 残留
- [ ] 字号用 token（`--text-xs`..`--text-3xl`），不出现 `font-size: 11.5px`
- [ ] 间距是 8 的倍数（4 仅微调）— 不出现 `padding: 15px`
- [ ] 等宽内容（hash / tag / 路径）用 `var(--font-mono)` 或 `.mono` class
- [ ] hover 不用 `transform: translateY(...)` — 用 background / border 过渡
- [ ] 可点击元素有 `cursor: pointer;` 和 `:hover` 反馈
- [ ] icon-only 按钮有 `aria-label` 或 `title`
- [ ] z-index 用 `--z-xxx`，不出现 `z-index: 9999`
- [ ] 暗色 + 亮色都肉眼检查过对比度（不能只看一种主题）
- [ ] 文件保存为 UTF-8 无 BOM（中文密集文件特别注意 — 见 `~/.claude/CLAUDE.md` 中文编码规则）

---

## 8. v1 / v1.1 落地复盘 & 陷阱

> 2026-04-27 ~ 2026-04-28 v1 + v1.1 落地过程踩过的坑。新做 UI 改造前先扫一遍。

### 8.1 Sticky 三大坑

#### 坑 1 — 祖先 `overflow: hidden / auto / scroll` 阻断 sticky

CSS sticky 的"定位基准"是**最近的 overflow != visible 的祖先**（即使是 `hidden` 也算）。如果该祖先不滚动，sticky `top: 0` 就死死贴在它的 padding edge 顶，**永远不脱离自然位置**，看起来像"没生效"。

**症状**：sticky 元素不贴顶，跟着内容一起滚走。

**典型场景**：父卡片用 `overflow: hidden` 来 clip `border-radius` 子元素 — 这是经典坑。

```css
/* ❌ 父容器有 overflow: hidden，内部 sticky 全失效 */
.repos-pane {
  border-radius: 8px;
  overflow: hidden;   /* sticky 杀手 */
}
.repos-toolbar {
  position: sticky;
  top: 0;             /* 永远在 .repos-pane 内顶部，不贴 viewport */
}

/* ✅ 去掉 overflow: hidden，给 sticky 子元素自己加圆角 */
.repos-pane {
  border-radius: 8px;
  /* 不能用 overflow: hidden — 见 design-system §8.1 */
}
.repos-toolbar {
  position: sticky;
  top: 0;
  border-radius: calc(var(--radius-lg) - 1px) calc(var(--radius-lg) - 1px) 0 0;
}
```

#### 坑 2 — Scroll container 的 padding 让 sticky `top: 0` 透出内容

Sticky `top: 0` 是相对 scroll port 的 **padding edge** 顶，不是 border edge。如果 scroll container 有 `padding-top: Xpx`，sticky 元素停在距 viewport 顶 Xpx 处，那 Xpx 区域**仍可见且会透出滚动中的内容**。

**症状**：sticky toolbar 顶部漏出半行内容（看起来像 toolbar 半透明）。

```css
/* main.content 提供全 view 的 padding */
.content { overflow-y: auto; padding: var(--space-6); /* 24px */ }

/* ❌ 留 24px 透明带 */
.repos-toolbar { position: sticky; top: 0; }

/* ✅ 用负 top 抵消 padding，sticky 真正贴 viewport 顶 */
.repos-toolbar { position: sticky; top: calc(-1 * var(--space-6)); }
```

#### 坑 3 — Containing block 不接力（双层 sticky）

Sticky 元素在它的**直接父 block**（containing block）内独立 stick。如果两个 sticky 元素分属不同父 block（如分组列表中每个 `.org-block` 一个 `.org-row`），它们**不会接力推走前一个**，会同时叠在一起。

**症状**：双层 sticky 两个组头堆叠在同一位置。

**修法**：DOM flat 重构 — 把 sticky siblings 全部提到同一父容器。

```vue
<!-- ❌ 每个 .org-row 在自己的 .org-block 内，sticky 不接力 -->
<div class="repos-tree">
  <div v-for="g in groups" class="org-block">
    <div class="org-row" sticky />
    <ul><li v-for="r in g.repos" /></ul>
  </div>
</div>

<!-- ✅ flat：所有 .org-row + .repo-row 共享 .repos-tree 这一个 containing block -->
<div class="repos-tree">
  <template v-for="g in groups">
    <div class="org-row" sticky />
    <div v-for="r in g.repos" class="repo-row" />
  </template>
</div>
```

### 8.2 动态高度 sticky — ResizeObserver + CSS var

Hard-code sticky 元素后续 sibling 的 `top` 偏移（如 `top: 64px`）脆弱：toolbar padding / 字号 / 内容变化都会让像素失准。用 ResizeObserver 测量并写入 CSS var：

```vue
<section ref="paneRef" class="pane">
  <div ref="toolbarRef" class="toolbar">...</div>
  <div class="next-sticky-row">...</div>
</section>

<script setup>
const paneRef = ref(null), toolbarRef = ref(null)
let ro = null
onMounted(() => {
  ro = new ResizeObserver(() => {
    paneRef.value?.style.setProperty('--toolbar-h', toolbarRef.value.offsetHeight + 'px')
  })
  ro.observe(toolbarRef.value)
})
onUnmounted(() => ro?.disconnect())
</script>

<style>
.next-sticky-row {
  position: sticky;
  /* 抵消 main.content padding + 动态 toolbar 高度 */
  top: calc(-1 * var(--space-6) + var(--toolbar-h, 64px));
}
</style>
```

### 8.3 渐进式 token 化策略

一次性把所有 view 切到新 token 风险大（一夜全坏）。正确做法：**新增 token 同时保留旧变量名 alias**，让未迁移的 view 自动跟着主题刷。

```css
[data-theme="dark"] {
  /* 新 token */
  --bg-canvas: #0b0d12;
  --bg-surface: #11141b;

  /* Backward-compat alias — 旧 view 还在用这些名字 */
  --bg-primary:    var(--bg-canvas);
  --bg-secondary:  var(--bg-elevated);
  --card-bg:       var(--bg-surface);
  --brand-start:   var(--brand-500);   /* 旧渐变两端点指向同色 → 视觉单色 */
  --brand-end:     var(--brand-600);
}
```

效果：v1 只重写 ProjectsView，但其他 13 个 view 通过 alias 继续渲染（颜色跟着新主题切换，只是布局风格还停在旧版）。v1.1 / v2 再逐页迁。

### 8.4 调性升级路径

工具类应用从"消费级"向"技术冷静"升级时的关键决策：

| 维度 | 旧（消费级） | 新（Linear / Vercel / Raycast 风） |
|------|-------------|-----------------------------------|
| 品牌色 | 紫蓝渐变 `#667eea → #764ba2` | 冷青蓝单色 `#3b82f6` |
| 状态色 | 高饱和（亮蓝紫粉） | 降饱和 + 透明背景 `rgba(..., 0.10)` |
| 字体 | 仅 sans-serif | sans + **mono**（hash / tag / 路径） |
| 图标 | Emoji（彩色卡通，OS 渲染不一致） | Lucide SVG（统一 stroke 1.75） |
| 卡片 hover | `transform: translateY(-2px)` 抖动 | `background` / `border-color` 过渡 |
| 侧边栏 | 永远暗（Slack 风） | **跟主题切换**（亮色下也变浅，避免黑 L 形） |

最大反直觉点：**侧边栏永远暗**在亮色主题下会形成"黑色 L 形"框住亮内容，反差极大 — 必须让 sidebar / titlebar 也跟主题切换。

### 8.5 中文文件 Edit 安全

Edit 工具替换大段含中文文本时，`old_string` 边界必须落在**纯 ASCII 行**（大括号、import、`---` 分隔符、`<style scoped>` 标签等），否则可能切断 UTF-8 多字节字符尾字节，产生 U+FFFD `?` 替换字符。

```
✅ old_string 起止：     <style scoped>  ...  </style>
✅ old_string 起止：     :root {  ...  }
❌ old_string 起止：     /* 配置卡 */  ...  /* 列表 */  （注释含中文，易切字节）
```

每次大段编辑后用 `Grep` 搜 U+FFFD 替换字符验证文件无损坏（替换字符长得像菱形问号）。

### 8.6 协作教训

1. **`npm run build` 干净 ≠ 实际可用** — UI 改动必须 `wails dev` 在窗口里点过才算完成。本次 sticky / 间距 / 透出 bug 都是 build 通过但 dev 看才发现。
2. **改 UI 前必读 baseline** — `theme.css` / 主组件 / 关键 view，避免跟既有约定冲突（如 useSettings 永远设 `data-theme` attribute，不存在"无 attribute"状态）。
3. **DEBUG outline** — sticky / 嵌套布局 / 透出 bug 调试时给元素加 `outline: 2px dashed magenta;`（不影响 layout，比 border 安全），用户截图就能精确量像素。调好删除。
4. **小步迭代** — sticky 这种一次成功率低的特性，分多次改 + 用户截图反馈比一次性大改稳得多。
