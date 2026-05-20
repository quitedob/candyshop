# CandyPro OEM B2B Platform — 设计思想

## 一句话核心理念

**以业务边界为第一性原理的分层架构：按门户（Portal）切面隔离关注点，按领域（Domain）纵向组织代码，三层（Handler→Service→Repository）保证可测试性，AI 作为横切能力以工具图模式注入。**

---

## 一、后端架构设计

### 1.1 三层 + 切面（Handler → Service → Repository × Scope）

不是简单的"三层架构"，而是在三层之上叠加了**五大门户切面**（Public / Auth / UserPortal / AdminPortal / System），形成矩阵式组织结构：

```
                 Public   Auth   UserPortal   AdminPortal   System
Handler 层       ✓        ✓      ✓            ✓             ✓
Service 层       ✓        ✓      ✓            ✓             ✓
Repository 层    ✓        ✓      ✓            ✓             ✓
```

**为什么需要切面？**
- 每个门户面向不同的用户角色，有不同的鉴权策略、限流策略、数据可见性
- 同一个"订单"在 AdminPortal 可以看到全部，在 UserPortal 只能看自己的
- 切面隔离防止代码腐化——Admin 的 handler 不可能意外调用 Public 的 repository

**实现方式：**
每个切面在 `repository/scopes/`、`services/scopes/`、`handlers/` 下各有一个聚合体（Aggregator），由 `main.go` 在启动时统一装配：

```
DB → Repos(5个切面) → Services(5个切面) → Handlers(5个切面) → Router(5个路由组)
```

每个切面的 GORM 实例通过 `db.Session(&gorm.Session{NewDB: true})` 创建，标记 `portal_scope` 上下文，为未来的分库分表预留入口。

### 1.2 接口定义在消费方（Consumer-side Interface）

遵循 Go 的惯用法：**接口由使用方定义，而非实现方导出**。

```go
// service/product/product.go
type productRepository interface {
    FindAll(ctx context.Context, page, limit int, ...) ([]Product, int64, error)
    FindBySlug(ctx context.Context, slug string) (*Product, error)
}

type ProductService struct {
    repo productRepository  // 私有接口，非具体类型
}
```

**收益：** Service 不依赖 Repository 的具体实现，测试时无需 mock 框架，手写一个满足接口的 struct 即可。这也有力约束了"上层不能访问下层未暴露的方法"。

### 1.3 AI 横切：工具图模式（Graph as Tool）

AI 不被封装在某个切面内部，而是作为**横切能力**存在于 `pkg/eino/` 中：

- **ChatModelAgent（TradeAssistant）：** 13 个工具，处理贸易文档生成、合规检查、翻译
- **DeepAgent（B2B Coordinator）：** 3 个子 Agent（ProductExpert / PricingExpert / LogisticsExpert）协同
- **Plan-Execute-Replan Agent：** 订单处理的多步骤规划与执行

**核心设计决策——Graph as Tool：**
确定性业务流程（如文档生成）不交给 LLM 自由发挥，而是编译为 Eino 的 `compose.Graph`（有向无环图），再将整张图包装为一个 Tool 暴露给 Agent。这实现了**确定性与概率性的混合架构**——LLM 负责意图理解与决策路由，Graph 负责精确执行。

从 `main.go` 的启动流程看得很清楚：
```
InitAgent() → AttachTradeAgent(AdminHandler) → AttachAgentToAIService(SystemHandler)
```
Agent 实例在 System 和 AdminPortal 两个切面中共享，但各自有不同的调用入口（SSE 实时对话 vs API 批量任务）。

### 1.4 事件驱动的最终一致性

订单→贸易单证的转换不是同步调用，而是通过 **EventOutbox 模式**异步完成：

```
Order.Confirm() → 写入 EventOutbox 表（与订单事务同库）
Background Worker 轮询 → 创建 TradeTransaction → 触发 AI 文档生成
```

这样做的原因：
- 订单确认不需要等待 AI 文档生成完成（降低响应延迟）
- 即使 AI 服务暂时不可用，订单仍然可以确认
- 重试机制天然内建（outbox 有 retry_attempts 字段）

同类模式还用于：订单草稿过期清理、Agent Checkpoint 定期清理。

### 1.5 状态机守卫（State Machine Guard）

核心业务对象的状态转换不是靠 if-else 判断，而是通过**显式的状态转换矩阵**校验：

```go
// order/order.go
var ValidOrderStatusTransitions = map[string][]string{
    "pending":              {"pending_confirmation", "cancelled", "expired"},
    "pending_confirmation": {"confirmed", "cancelled", "expired"},
    "confirmed":            {"production", "cancelled"},
    // ... 11 个状态，50+ 条合法转换路径
}
```

Return/RMA（5 状态）、Payment（5 状态）、OEMProject（8 状态）同理。**任何状态变更都必须经过矩阵校验**，杜绝非法状态转换。

### 1.6 三阶段库存（Reserve → Deduct → Release）

```
下单确认 → Reserve（预留库存，不可再售）
发货履行 → Deduct（实际扣减）  
取消/过期 → Release（释放预留，恢复可售）
```

采用 PostgreSQL `SELECT ... FOR UPDATE SKIP LOCKED` 实现 FEFO（先到期先出）批次扣减，避免并发超卖。

### 1.7 优雅降级（Graceful Degradation）

应用可以在数据库不可用时**启动但不提供数据服务**：

```go
func (h *Handler) AdminGetProducts(c *gin.Context) {
    if h.services == nil {
        response.ServiceUnavailableResp(c)  // 503 + i18n 消息
        return
    }
    // ...
}
```

所有 Handler 方法首行都做 nil check。这保证了健康检查端点（/health, /ready）在数据库故障时仍然可以正常响应。

---

## 二、前端架构设计

### 2.1 条件式 Chrome 模式

不像传统 SPA 每个路由嵌入自己的导航，`app.vue` 通过一个计算属性 `showMarketingChrome` 统一决策是否显示公共 Header/Footer：

```
路径以 /admin /customer /auth 开头 → 隐藏营销外壳 → 使用专属 Layout
其他路径 → 显示营销外壳（Header + Footer + 询盘悬浮按钮）
```

**四个 Layout 区域：**
| Layout | 用途 | 视觉主题 |
|--------|------|---------|
| default | 公开营销页 | 品牌橙/暖色调 |
| auth | 认证页面 | 羊皮纸/陶土编辑风 |
| customer | 客户门户 | 侧边导航 + 顶栏 |
| admin | 管理后台 | 数据密集型仪表盘 |

### 2.2 单体 API 客户端

所有 HTTP 通信汇聚在**一个** `useApi.ts` 组合函数（~930 行），而非分散在多个 service 文件中：

- `fetchApi<T>()` — 需认证的请求（自动附加 Bearer Token、Accept-Language）
- `fetchPublicApi<T>()` — 公开接口（自动前缀 `/public`）
- 按业务域组织方法：`products.xxx()`、`customer.xxx()`、`admin.xxx()`、`search.xxx()`

**为什么不用多个 service 文件？**
- 单一入口便于统一处理 Token 刷新、错误转换、请求取消
- 项目的 API 消费方只有浏览器端一个，无需多端适配
- 减少导入路径的心智负担

### 2.3 Cookie JWT + 客户端解码

认证状态通过两个 Cookie 存储（`auth_token` + `refresh_token`），而非 localStorage：

- 前端通过 `jwt-decode` 在客户端解码 payload 判断过期，无需先发请求试错
- `initAuth()` 在插件层（`init-auth.client.ts`）和中间件层（`middleware/auth.ts`）双重调用
- "记住我"关闭时 Cookie 不设过期时间（会话级）

### 2.4 懒加载 i18n + 自定义语言检测

9 种语言 × 15 个命名空间 = 135 个 JSON 文件，全部 `lazy: true` 按需加载：

**语言检测优先级：**
```
?lang= 查询参数 → user-locale Cookie → Accept-Language 请求头 → 默认 zh
```

中文（zh）为默认语言且不携带 URL 前缀（`prefix_except_default`），其他 8 种语言各带 `/<code>/` 前缀。阿拉伯语（ar）启用 RTL 布局。

### 2.5 CSS 自定义属性设计系统

不使用 Tailwind 原子类的任意组合，而是定义了一套**语义化 CSS 自定义属性**（~1400 行 `main.css`）：

```css
--color-primary: #1c1917;     /* 暖墨色 */
--color-accent: #fb923c;      /* 暖橙 */
--color-highlight: #b83509;   /* 烧橙 CTA */
--spacing-md: 1rem;
--radius-lg: 12px;
--shadow-md: 0 4px 6px ...;
--transition-base: 250ms cubic-bezier(0.4, 0, 0.2, 1);
```

**四层组织：** Base（reset/字体/变量）→ Utilities（`.sr-only`/`.container`）→ Components（`.btn`/`.card`/`.badge`）→ Animations（`@keyframes`）

暗色模式通过 `.dark` 选择器覆盖全套变量，仅需切换根元素 class 即可全局生效。

### 2.6 三态 UI 模式

每个数据驱动的组件必须处理三种状态：

```
加载中 → SkeletonLoader（骨架屏，避免布局跳动）
错误   → ErrorState（错误描述 + 重试按钮）
空数据 → EmptyState（友好提示 + 引导操作）
```

这是**强制约定而非可选**——代码审查会检查每个数据组件是否具备三态处理。

### 2.7 结构化数据无处不在

`useSeo()` 组合函数为每个页面生成 JSON-LD 结构化数据：
- Organization + LocalBusiness（全站）
- Product（产品详情页，含价格/库存/评价）
- Article（博客文章）
- BreadcrumbList（面包屑导航）
- FAQ（问答页）
- Video（视频内容）

同时自动生成 hreflang 交替链接、Open Graph 标签、Twitter Card。**seo: false 关闭 Nuxt 自动 SEO，全部手动接管以确保多语言正确性。**

### 2.8 SSR 关闭

`ssr: false` 是有意为之：
- 平台是 B2B 后台管理系统 + 营销官网，不需要 SEO 驱动的 SSR
- 避免 Node.js 服务端渲染的状态同步复杂度
- Nitro 仅用于 dev proxy（/api → 后端）和静态资源服务

---

## 三、业务领域设计

### 3.1 核心商业流程

```
产品目录 → 询盘 → AI 分析 → 双向议价（NegotiationOffer）
→ 转订单 → 信用/额度校验 → 确认 → 生产 → 
分批履行（Fulfillment）→ 发货 → 贸易单证生成（AI）→ 结算
```

### 3.2 七大业务域

| 领域 | 核心模型 | 关键设计决策 |
|------|---------|------------|
| **产品** | Product(122字段), Category, Variant, PriceList, PriceRule, Supplier, Warehouse, Channel | B2B 定价链：合同价→客户组价目表→默认价目表→产品基价；MOQ 阶梯定价 |
| **订单** | Order(11状态), Cart, NegotiationOffer, Fulfillment, ReturnRequest, Payment, Invoice, Coupon | 状态矩阵守卫；三阶段库存；部分履约；RMA 退货退款 |
| **贸易** | TradeTransaction, 8种专业单证(PI/CI/SC/PL/COO/HC/BL), ComplianceRequirement, ShipmentEvent | 事件驱动异步生成；AI 混合工作流 |
| **用户** | User, Company, BuyerOrganization, OrgMember, ApprovalAction | KYB 认证；三层信用策略（国家→公司→额度） |
| **认证** | Role(含权限JSONB), RefreshToken | JWT 双 Token 机制；角色权限数组化 |
| **内容** | BlogPost, CaseStudy, Translation(DB后端i18n), Notification, ActivityLog | DB 后端翻译引擎支持运行时修改 |
| **系统** | Eino Agent, SSE, WebhookConfig, SystemSetting | 横切 AI 能力；Webhook 生命周期管理 |

### 3.3 定价与信用三层模型

```
第一层 CountryPaymentPolicy  → 按目的国强制规则（如印度/巴基斯坦全预付）
第二层 Company.PaymentTerms   → 客户级别条款（NET_15/30/60, COD, PREPAID）
第三层 Company.CreditLimit    → 额度上限，订单金额超出则拒绝
```

三层逐级收紧，而非取最宽松者——**安全优先**。

### 3.4 MOQ 执行点

最小起订量（MOQ）在产品目录展示、加入购物车、提交询盘、创建订单、议价环节**五个触点**全部校验，而非仅在最终下单时检查。

---

## 四、技术选型思想

| 选择 | 核心理由 |
|------|---------|
| **Go + Gin** | 高并发 B2B 后台，Gin 的中间件链模型天然契合切面架构 |
| **GORM + PostgreSQL** | AutoMigrate 支持快速迭代；pgvector 做语义搜索（产品匹配、合规检索） |
| **Nuxt 3 SPA 模式** | Vue 3 生态 + 文件路由 + 自动导入 = 低心智负担；SSR 关闭因 B2B 后台不需要 |
| **Cloudwego Eino** | 字节跳动开源的 Agent 框架；Graph as Tool 模式让 AI 行为可预测、可测试 |
| **@nuxtjs/i18n v10** | 9 语言的懒加载 + RTL + SEO hreflang 一体化 |
| **Tailwind + CSS 变量混合** | Tailwind 处理布局/间距，CSS 变量统一品牌色/动效/暗色模式 |
| **MiniMax M2.7** | OpenAI 兼容协议，成本优化，中文能力更强 |

---

## 五、代码组织原则

1. **就近原则**：属于同一个业务的代码（handler + service + repo + model + route）在各自目录中按域名组织，而非按层拆分到完全隔离的顶级目录
2. **私有接口**：Service 层定义自己需要的 Repository 接口（小接口、按需定义），而非依赖全局导出的巨型接口
3. **聚合体装配**：`main.go` 是唯一的依赖装配点，不存在服务定位器（Service Locator）或隐式依赖
4. **显式 > 隐式**：GORM Tag 明确标注列名/约束；Swagger 注解在 Handler 方法上；错误码全部显式传入 `response.ErrorResp()`
5. **i18n 第一公民**：所有面向用户的字符串（错误、状态、枚举值、邮件）都经过翻译引擎，不留硬编码文案
