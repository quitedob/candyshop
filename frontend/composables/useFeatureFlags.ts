/** 功能开关 — 单仓业务，多仓库 UI/路由默认关闭 */
export function useFeatureFlags() {
  const config = useRuntimeConfig()
  const enableMultiWarehouse = config.public.enableMultiWarehouse === true
  return { enableMultiWarehouse }
}
