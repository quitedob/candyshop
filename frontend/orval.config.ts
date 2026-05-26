import { defineConfig } from 'orval'

/** 从 backend/docs/swagger.json（Swagger 2.0）生成 TypeScript 客户端与模型 */
export default defineConfig({
  candypro: {
    input: {
      target: '../backend/docs/swagger.json',
      validation: false,
    },
    output: {
      mode: 'tags-split',
      target: './generated/api/endpoints',
      schemas: './generated/api/models',
      client: 'fetch',
      clean: true,
      prettier: false,
      override: {
        mutator: {
          path: './lib/openapi/custom-fetch.ts',
          name: 'customFetch',
        },
      },
    },
  },
})
