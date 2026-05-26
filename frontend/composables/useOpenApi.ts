/**
 * OpenAPI 生成客户端入口（orval）
 * 用法：const { getProducts, getProductsFeatured } = useOpenApi()
 * 与手写 useApi 并存，可逐步迁移。
 */
import * as generated from '~/generated/api/endpoints'

export const useOpenApi = () => generated

export type * from '~/generated/api/models'
