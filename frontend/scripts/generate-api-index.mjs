/**
 * 扫描 generated/api/endpoints 下各 tag 模块，生成 barrel index.ts
 */
import { readdirSync, writeFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const endpointsDir = join(__dirname, '../generated/api/endpoints')
const indexPath = join(endpointsDir, 'index.ts')

const tags = readdirSync(endpointsDir, { withFileTypes: true })
  .filter((d) => d.isDirectory())
  .map((d) => d.name)
  .sort()

const lines = [
  '/**',
  ' * 自动生成：orval tags-split 聚合入口，勿手动编辑',
  ' * 由 scripts/generate-api-index.mjs 维护',
  ' */',
  ...tags.map((tag) => `export * from './${tag}/${tag}'`),
  '',
]

writeFileSync(indexPath, lines.join('\n'), 'utf8')
console.log(`Wrote ${indexPath} (${tags.length} tags)`)
