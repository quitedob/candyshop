<template>
  <div class="shipments-page">
    <header class="shipments-page__header">
      <h1 class="shipments-page__title">{{ t('customer.nav.shipments') }}</h1>
      <p class="shipments-page__subtitle">{{ t('customer.shipments.subtitle') }}</p>
    </header>

    <div v-if="pending" class="shipments-page__center">
      <div class="shipments-page__spinner" aria-hidden="true" />
    </div>

    <div v-else-if="error" class="shipments-page__alert shipments-page__alert--error">
      <Icon name="heroicons:exclamation-triangle" class="shipments-page__alert-icon" />
      <p>{{ error }}</p>
      <button type="button" class="shipments-btn shipments-btn--ghost" @click="load">
        {{ t('customer.shipments.retry') }}
      </button>
    </div>

    <div v-else-if="!trades.length" class="shipments-page__empty">
      <Icon name="heroicons:truck" class="shipments-page__empty-icon" />
      <p>{{ t('customer.shipments.no_trades') }}</p>
      <NuxtLink :to="localePath('/customer/trades')" class="shipments-page__link">
        {{ t('customer.shipments.view_trades') }}
      </NuxtLink>
    </div>

    <div v-else class="shipments-page__body">
      <article v-for="trade in trades" :key="trade.id" class="shipments-trade">
        <div class="shipments-trade__head">
          <div class="shipments-trade__head-main">
            <h2 class="shipments-trade__title">
              {{ t('customer.shipments.trade_label') }} #{{ trade.id }}
            </h2>
            <p class="shipments-trade__meta">
              {{ enumLabel('trade_status', trade.status) }}
              <span aria-hidden="true">·</span>
              {{ trade.incoterms || trade.terms || '—' }}
            </p>
          </div>
          <NuxtLink :to="localePath(`/customer/trades/${trade.id}`)" class="shipments-page__link shipments-trade__link">
            {{ t('customer.shipments.view_trade') }}
          </NuxtLink>
        </div>

        <div class="shipments-trade__content">
          <p v-if="!shipmentMap[trade.id]?.length" class="shipments-trade__empty">
            {{ t('customer.shipments.no_shipments') }}
          </p>

          <div v-else class="shipments-items">
            <section
              v-for="s in shipmentMap[trade.id]"
              :key="s.id"
              class="shipments-item"
            >
              <div class="shipments-item__head">
                <div class="shipments-item__summary">
                  <dl class="shipments-item__facts">
                    <div class="shipments-item__fact">
                      <dt>{{ t('customer.shipments.carrier') }}</dt>
                      <dd>{{ cell(s.carrierName) }}</dd>
                    </div>
                    <div class="shipments-item__fact">
                      <dt>{{ t('customer.shipments.tracking_no') }}</dt>
                      <dd class="shipments-item__mono">{{ cell(s.billOfLadingNo) }}</dd>
                    </div>
                  </dl>
                  <div class="shipments-item__status-row">
                    <span :class="statusBadgeClass(s.status)" class="shipments-item__badge">
                      {{ enumLabel('shipment_status', s.status) }}
                    </span>
                    <time class="shipments-item__updated">{{ formatDate(s.updatedAt) }}</time>
                  </div>
                  <p v-if="s.eta" class="shipments-item__eta">
                    {{ t('customer.shipments.eta') }}: {{ formatDate(s.eta) }}
                  </p>
                </div>

                <button
                  v-if="canNudge(s.status)"
                  type="button"
                  class="shipments-btn shipments-btn--secondary shipments-item__nudge"
                  :disabled="nudgingKey === shipmentKey(trade.id, s.id)"
                  @click="nudgeShipment(trade.id, s.id)"
                >
                  {{ nudgingKey === shipmentKey(trade.id, s.id) ? t('customer.shipments.nudging') : t('customer.shipments.nudge_button') }}
                </button>
              </div>

              <p
                v-if="nudgeMessages[shipmentKey(trade.id, s.id)]"
                class="shipments-item__feedback"
                :class="nudgeErrors[shipmentKey(trade.id, s.id)] ? 'shipments-item__feedback--error' : 'shipments-item__feedback--ok'"
              >
                {{ nudgeMessages[shipmentKey(trade.id, s.id)] }}
              </p>

              <div class="shipments-item__section">
                <button
                  type="button"
                  class="shipments-item__timeline-toggle"
                  :aria-expanded="expandedTimeline[shipmentKey(trade.id, s.id)] ? 'true' : 'false'"
                  @click="toggleTimeline(trade.id, s.id)"
                >
                  <Icon
                    :name="expandedTimeline[shipmentKey(trade.id, s.id)] ? 'heroicons:chevron-up' : 'heroicons:chevron-down'"
                    class="shipments-item__timeline-icon"
                  />
                  {{ expandedTimeline[shipmentKey(trade.id, s.id)] ? t('customer.shipments.hide_timeline') : t('customer.shipments.toggle_timeline') }}
                </button>

                <div v-if="expandedTimeline[shipmentKey(trade.id, s.id)]" class="shipments-item__timeline">
                  <p v-if="timelineLoading[shipmentKey(trade.id, s.id)]" class="shipments-item__timeline-loading">
                    {{ t('customer.shipments.timeline_loading') }}
                  </p>
                  <template v-else>
                    <div
                      v-for="ev in timelineMap[shipmentKey(trade.id, s.id)] || []"
                      :key="ev.id"
                      class="shipments-item__event"
                    >
                      <p class="shipments-item__event-title">
                        <span>{{ enumLabel('shipment_event_type', ev.eventType) }}</span>
                        <span v-if="ev.location" class="shipments-item__event-loc">· {{ ev.location }}</span>
                        <span class="shipments-item__event-time">— {{ formatDate(ev.eventTime) }}</span>
                      </p>
                      <p v-if="ev.description" class="shipments-item__event-desc">{{ ev.description }}</p>
                    </div>
                    <p v-if="!(timelineMap[shipmentKey(trade.id, s.id)] || []).length" class="shipments-item__timeline-empty">
                      {{ t('customer.shipments.timeline_empty') }}
                    </p>
                  </template>
                </div>
              </div>

              <div v-if="s.customerAttachments?.length" class="shipments-item__section">
                <p class="shipments-item__section-label">{{ t('customer.shipments.attachments') }}</p>
                <ul class="shipments-item__attachments">
                  <li v-for="(url, idx) in s.customerAttachments" :key="idx">
                    <a :href="url" target="_blank" rel="noopener noreferrer" class="shipments-page__link shipments-item__attachment">
                      {{ attachmentLabel(url) }}
                    </a>
                  </li>
                </ul>
              </div>

              <div v-if="s.status !== 'DELIVERED'" class="shipments-item__section shipments-item__upload">
                <p class="shipments-item__section-label">{{ t('customer.shipments.upload_label') }}</p>
                <p class="shipments-item__upload-hint">
                  {{ t('customer.shipments.upload_hint', { max: SHIPMENT_ATTACHMENT_MAX_FILES, mb: CUSTOMER_ATTACHMENT_MAX_MB }) }}
                </p>
                <div class="shipments-item__upload-actions">
                  <input
                    :ref="(el) => setFileInputRef(shipmentKey(trade.id, s.id), el)"
                    type="file"
                    multiple
                    :accept="CUSTOMER_ATTACHMENT_ACCEPT"
                    class="sr-only"
                    @change="(e) => onFilesSelected(trade.id, s, e)"
                  />
                  <button
                    type="button"
                    class="shipments-btn shipments-btn--secondary"
                    :disabled="uploadingKey === shipmentKey(trade.id, s.id)"
                    @click="triggerFileInput(shipmentKey(trade.id, s.id))"
                  >
                    {{ uploadingKey === shipmentKey(trade.id, s.id) ? t('customer.shipments.uploading') : t('customer.shipments.upload_button') }}
                  </button>
                </div>
                <p
                  v-if="uploadMessages[shipmentKey(trade.id, s.id)]"
                  class="shipments-item__feedback"
                  :class="uploadErrors[shipmentKey(trade.id, s.id)] ? 'shipments-item__feedback--error' : 'shipments-item__feedback--ok'"
                >
                  {{ uploadMessages[shipmentKey(trade.id, s.id)] }}
                </p>
              </div>
              <p v-else class="shipments-item__delivered-note">
                {{ t('customer.shipments.status_delivered_no_upload') }}
              </p>
            </section>
          </div>
        </div>
      </article>

      <nav
        v-if="pagination && pagination.totalPages > 1"
        class="shipments-pagination"
        :aria-label="t('pagination.nav_label')"
      >
        <p class="shipments-pagination__info">
          {{ t('customer.shipments.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </p>
        <div class="shipments-pagination__controls">
          <button type="button" class="shipments-pagination__btn" :disabled="page <= 1" @click="prevPage">
            <Icon name="heroicons:chevron-left" class="h-5 w-5" />
            <span class="sm:hidden">{{ t('common.previous') }}</span>
          </button>
          <button type="button" class="shipments-pagination__btn" :disabled="page >= pagination.totalPages" @click="nextPage">
            <span class="sm:hidden">{{ t('common.next') }}</span>
            <Icon name="heroicons:chevron-right" class="h-5 w-5" />
          </button>
        </div>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  CUSTOMER_ATTACHMENT_ACCEPT,
  CUSTOMER_ATTACHMENT_MAX_MB,
  isCustomerAttachmentAllowed,
  SHIPMENT_ATTACHMENT_MAX_FILES,
} from '~/utils/customerAttachments'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { enumLabel, formatDate, cell } = useDisplay()
const localePath = useLocalePath()

const trades = ref<any[]>([])
const shipmentMap = ref<Record<string, any[]>>({})
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const expandedTimeline = ref<Record<string, boolean>>({})
const timelineMap = ref<Record<string, any[]>>({})
const timelineLoading = ref<Record<string, boolean>>({})
const nudgingKey = ref('')
const nudgeMessages = ref<Record<string, string>>({})
const nudgeErrors = ref<Record<string, boolean>>({})
const uploadingKey = ref('')
const uploadMessages = ref<Record<string, string>>({})
const uploadErrors = ref<Record<string, boolean>>({})
const fileInputRefs = ref<Record<string, HTMLInputElement | null>>({})

const shipmentKey = (tradeId: number | string, shipmentId: number | string) => `${tradeId}-${shipmentId}`

const statusBadgeClass = (status: string) => {
  switch (status) {
    case 'DELIVERED': return 'shipments-item__badge--delivered'
    case 'IN_TRANSIT': return 'shipments-item__badge--transit'
    case 'DISPATCHED': return 'shipments-item__badge--dispatched'
    case 'EXCEPTION': return 'shipments-item__badge--exception'
    default: return 'shipments-item__badge--pending'
  }
}

const canNudge = (status: string) => ['PENDING', 'DISPATCHED', 'IN_TRANSIT', 'EXCEPTION'].includes(status)

const attachmentLabel = (url: string) => {
  const name = url.split('/').pop() || url
  try { return decodeURIComponent(name) } catch { return name }
}

const setFileInputRef = (key: string, el: unknown) => {
  fileInputRefs.value[key] = (el as HTMLInputElement | null) || null
}

const triggerFileInput = (key: string) => {
  fileInputRefs.value[key]?.click()
}

const loadShipmentsForTrades = async (tradeList: any[]) => {
  await Promise.all(
    tradeList.map(async (trade: any) => {
      try {
        const s = await api.get<any>(`/user/trades/${trade.id}/shipments`)
        shipmentMap.value[trade.id] = s.data || []
      } catch {
        shipmentMap.value[trade.id] = []
      }
    })
  )
}

const load = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>(`/user/trades?page=${page.value}&limit=${pageSize}`)
    trades.value = res.data || []
    pagination.value = res.pagination || null
    shipmentMap.value = {}
    timelineMap.value = {}
    expandedTimeline.value = {}
    await loadShipmentsForTrades(trades.value)
  } catch (err: any) {
    error.value = err?.message || t('customer.shipments.load_error')
  } finally {
    pending.value = false
  }
}

const prevPage = () => {
  if (page.value <= 1) return
  page.value -= 1
  load()
}

const nextPage = () => {
  if (!pagination.value || page.value >= pagination.value.totalPages) return
  page.value += 1
  load()
}

const fetchTimeline = async (tradeId: number | string, shipmentId: number | string) => {
  const key = shipmentKey(tradeId, shipmentId)
  timelineLoading.value[key] = true
  try {
    const res = await api.customerGetShipmentTimeline(tradeId, shipmentId)
    const events = (res.data || []).slice().sort(
      (a: any, b: any) => new Date(b.eventTime).getTime() - new Date(a.eventTime).getTime()
    )
    timelineMap.value[key] = events
  } catch {
    timelineMap.value[key] = []
  } finally {
    timelineLoading.value[key] = false
  }
}

const toggleTimeline = async (tradeId: number | string, shipmentId: number | string) => {
  const key = shipmentKey(tradeId, shipmentId)
  const next = !expandedTimeline.value[key]
  expandedTimeline.value[key] = next
  if (next && !timelineMap.value[key]) {
    await fetchTimeline(tradeId, shipmentId)
  }
}

const nudgeShipment = async (tradeId: number | string, shipmentId: number | string) => {
  const key = shipmentKey(tradeId, shipmentId)
  nudgingKey.value = key
  nudgeMessages.value[key] = ''
  nudgeErrors.value[key] = false
  try {
    await api.customerNudgeShipment(tradeId, shipmentId)
    nudgeMessages.value[key] = t('customer.shipments.nudge_success')
  } catch (err: any) {
    nudgeErrors.value[key] = true
    nudgeMessages.value[key] = err?.data?.message || err?.message || t('customer.shipments.nudge_error')
  } finally {
    nudgingKey.value = ''
  }
}

const onFilesSelected = async (tradeId: number | string, shipment: any, e: Event) => {
  const input = e.target as HTMLInputElement
  const files = input.files ? Array.from(input.files) : []
  input.value = ''
  if (!files.length) return

  const key = shipmentKey(tradeId, shipment.id)
  const existing = (shipment.customerAttachments || []).length
  if (existing + files.length > SHIPMENT_ATTACHMENT_MAX_FILES) {
    uploadErrors.value[key] = true
    uploadMessages.value[key] = t('customer.shipments.file_count_exceeded')
    return
  }

  for (const f of files) {
    if (!isCustomerAttachmentAllowed(f)) {
      uploadErrors.value[key] = true
      uploadMessages.value[key] = t('customer.shipments.file_type_invalid')
      return
    }
    if (f.size > CUSTOMER_ATTACHMENT_MAX_MB * 1024 * 1024) {
      uploadErrors.value[key] = true
      uploadMessages.value[key] = t('customer.shipments.file_size_exceeded')
      return
    }
  }

  uploadingKey.value = key
  uploadMessages.value[key] = ''
  uploadErrors.value[key] = false
  try {
    const formData = new FormData()
    files.forEach((f) => formData.append('files', f))
    const res = await api.customerUploadShipmentAttachment(tradeId, shipment.id, formData)
    shipment.customerAttachments = [
      ...(shipment.customerAttachments || []),
      ...(res.files || []),
    ]
    uploadMessages.value[key] = t('customer.shipments.upload_success')
  } catch (err: any) {
    uploadErrors.value[key] = true
    uploadMessages.value[key] = err?.data?.message || err?.message || t('customer.shipments.upload_error')
  } finally {
    uploadingKey.value = ''
  }
}

onMounted(load)
</script>

<style scoped>
/* 页面容器 */
.shipments-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  min-width: 0;
}

.shipments-page__header {
  min-width: 0;
}

.shipments-page__title {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-primary);
  line-height: 1.25;
}

.shipments-page__subtitle {
  margin-top: var(--spacing-xs);
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.5;
}

.shipments-page__body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  min-width: 0;
}

.shipments-page__center {
  display: flex;
  justify-content: center;
  padding: 3rem 0;
}

.shipments-page__spinner {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 9999px;
  border: 4px solid rgba(var(--color-highlight-rgb), 0.2);
  border-top-color: var(--color-highlight);
  animation: shipments-spin 0.8s linear infinite;
}

@keyframes shipments-spin {
  to { transform: rotate(360deg); }
}

.shipments-page__empty {
  text-align: center;
  padding: 4rem var(--spacing-md);
  color: var(--color-text-light);
}

.shipments-page__empty-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto var(--spacing-sm);
  color: var(--color-border);
}

.shipments-page__link {
  color: var(--color-highlight);
  font-size: var(--text-sm);
  font-weight: 500;
  text-decoration: none;
}

.shipments-page__link:hover {
  color: var(--color-highlight-hover);
  text-decoration: underline;
}

.shipments-page__alert {
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  text-align: center;
  font-size: var(--text-sm);
}

.shipments-page__alert--error {
  background: rgba(var(--color-error-rgb), 0.08);
  border: 1px solid rgba(var(--color-error-rgb), 0.25);
  color: var(--color-error);
}

.shipments-page__alert-icon {
  width: 2rem;
  height: 2rem;
  margin: 0 auto var(--spacing-sm);
}

/* 贸易卡片 */
.shipments-trade {
  background: var(--color-bg-elevated, #fff);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  min-width: 0;
}

.shipments-trade__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-sm) var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  background: var(--color-bg-alt);
}

.shipments-trade__head-main {
  min-width: 0;
  flex: 1 1 12rem;
}

.shipments-trade__title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-primary);
}

.shipments-trade__meta {
  margin-top: 0.25rem;
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.shipments-trade__link {
  flex-shrink: 0;
  white-space: nowrap;
}

.shipments-trade__content {
  padding: var(--spacing-md) var(--spacing-lg) var(--spacing-lg);
}

.shipments-trade__empty {
  font-size: var(--text-sm);
  color: var(--color-text-lighter);
}

/* 发货条目 */
.shipments-items {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.shipments-item {
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background: var(--color-bg);
  min-width: 0;
}

.shipments-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-sm) var(--spacing-md);
}

.shipments-item__summary {
  flex: 1 1 14rem;
  min-width: 0;
}

.shipments-item__facts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
  gap: var(--spacing-sm) var(--spacing-md);
  margin: 0;
}

.shipments-item__fact dt {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-lighter);
}

.shipments-item__fact dd {
  margin: 0.125rem 0 0;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-primary);
  word-break: break-word;
}

.shipments-item__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
}

.shipments-item__status-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: var(--spacing-sm);
}

.shipments-item__badge {
  display: inline-flex;
  align-items: center;
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.6875rem;
  font-weight: 600;
}

.shipments-item__badge--pending { background: #fef3c7; color: #92400e; }
.shipments-item__badge--dispatched { background: #e0e7ff; color: #3730a3; }
.shipments-item__badge--transit { background: #dbeafe; color: #1e40af; }
.shipments-item__badge--delivered { background: #dcfce7; color: #166534; }
.shipments-item__badge--exception { background: #fee2e2; color: #991b1b; }

.shipments-item__updated,
.shipments-item__eta {
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

.shipments-item__eta {
  margin-top: 0.375rem;
}

.shipments-item__nudge {
  flex-shrink: 0;
  align-self: flex-start;
}

.shipments-item__section {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
}

.shipments-item__section-label {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text);
}

.shipments-item__timeline-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: var(--text-xs);
  font-weight: 500;
  color: var(--color-highlight);
  cursor: pointer;
}

.shipments-item__timeline-toggle:hover {
  color: var(--color-highlight-hover);
}

.shipments-item__timeline-icon {
  width: 1rem;
  height: 1rem;
}

.shipments-item__timeline {
  margin-top: var(--spacing-sm);
  padding-left: var(--spacing-sm);
  border-left: 2px solid rgba(var(--color-highlight-rgb), 0.2);
}

.shipments-item__timeline-loading,
.shipments-item__timeline-empty {
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

.shipments-item__event + .shipments-item__event {
  margin-top: var(--spacing-sm);
}

.shipments-item__event-title {
  font-size: var(--text-xs);
  color: var(--color-text);
}

.shipments-item__event-title span:first-child {
  font-weight: 600;
}

.shipments-item__event-loc {
  color: var(--color-text-light);
}

.shipments-item__event-time {
  color: var(--color-text-lighter);
}

.shipments-item__event-desc {
  margin-top: 0.125rem;
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.shipments-item__attachments {
  margin: var(--spacing-xs) 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.shipments-item__attachment {
  word-break: break-all;
  font-size: var(--text-xs);
}

.shipments-item__upload-hint {
  margin-top: 0.25rem;
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
  line-height: 1.45;
}

.shipments-item__upload-actions {
  margin-top: var(--spacing-sm);
}

.shipments-item__delivered-note {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

.shipments-item__feedback {
  margin-top: var(--spacing-sm);
  font-size: var(--text-xs);
}

.shipments-item__feedback--ok { color: var(--color-success); }
.shipments-item__feedback--error { color: var(--color-error); }

/* 按钮 */
.shipments-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  padding: 0.4375rem 0.875rem;
  border-radius: var(--radius-md);
  font-size: var(--text-xs);
  font-weight: 600;
  border: 1px solid transparent;
  cursor: pointer;
  transition: background-color 0.15s, border-color 0.15s, opacity 0.15s;
}

.shipments-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.shipments-btn--secondary {
  background: var(--color-bg-alt);
  color: var(--color-text);
  border-color: var(--color-border);
}

.shipments-btn--secondary:hover:not(:disabled) {
  background: var(--color-bg);
  border-color: var(--color-border-dark, var(--color-border));
}

.shipments-btn--ghost {
  margin-top: var(--spacing-sm);
  background: transparent;
  color: var(--color-error);
  border-color: rgba(var(--color-error-rgb), 0.35);
}

/* 分页 */
.shipments-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-sm) var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--color-bg-elevated, #fff);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.shipments-pagination__info {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.shipments-pagination__controls {
  display: flex;
  gap: var(--spacing-xs);
}

.shipments-pagination__btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  padding: 0 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg);
  color: var(--color-text-light);
  font-size: var(--text-sm);
  cursor: pointer;
}

.shipments-pagination__btn:hover:not(:disabled) {
  background: var(--color-bg-alt);
}

.shipments-pagination__btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

@media (max-width: 639px) {
  .shipments-trade__head,
  .shipments-item__head {
    flex-direction: column;
    align-items: stretch;
  }

  .shipments-item__nudge,
  .shipments-trade__link {
    width: 100%;
    justify-content: center;
    text-align: center;
  }

  .shipments-item__facts {
    grid-template-columns: 1fr;
  }

  .shipments-pagination {
    flex-direction: column;
    align-items: stretch;
  }

  .shipments-pagination__controls {
    justify-content: space-between;
  }

  .shipments-pagination__btn {
    flex: 1;
    justify-content: center;
  }
}
</style>
