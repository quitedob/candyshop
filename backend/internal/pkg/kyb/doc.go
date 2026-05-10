// Package kyb 封装客户激活（KYB）与小额免审档位等业务规则。
//
// 与路由的配合方式（见 internal/api/routes/userportal/register.go）：
//
//   - RequireActiveUser：询盘创建/更新、贸易创建、OEM 项目创建、KYB 证照上传、订单付款凭证上传等写入须账号已 active。
//
//   - Handler 内 ensureActiveOrKYBBypassForAmount（及 system 侧等价逻辑）：下单、购物车结算、部分 AI 辅助下单在 pending 时
//     可按 config.KYB（KYB_BYPASS_MAX_ORDER_USD、样品档位等）放行小额免审。
//
// 两类入口并存是刻意设计：高敏感上传与「对外承诺型」写入走 active；订单金额类走可配置的免审额度。
package kyb
