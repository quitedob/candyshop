<template>
  <div class="order-detail">
    <!-- 返回 -->
    <NuxtLink :to="localePath('/customer/orders')" class="order-detail__back">
      <Icon name="heroicons:arrow-left" class="h-4 w-4" aria-hidden="true" />
      {{ t('customer.orders.back') }}
    </NuxtLink>

    <!-- 加载 -->
    <div v-if="pending" class="order-detail__state">
      <Icon name="heroicons:arrow-path" class="h-8 w-8 animate-spin text-orange-500" aria-hidden="true" />
      <p>{{ t('customer.orders.loading_details') }}</p>
    </div>

    <!-- 错误 -->
    <div v-else-if="error && !order" class="order-detail__error">
      <Icon name="heroicons:exclamation-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
      <p>{{ error }}</p>
    </div>

    <template v-else-if="order">
      <!-- 页头 -->
      <header class="order-detail__hero">
        <div class="order-detail__hero-main">
          <p class="order-detail__eyebrow">{{ t('customer.nav.orders') }}</p>
          <h1 class="order-detail__title">
            #{{ order.orderNumber || order.id.substring(0, 8) }}
          </h1>
          <p class="order-detail__meta">
            <Icon name="heroicons:calendar-days" class="h-4 w-4" aria-hidden="true" />
            {{ t('customer.orders.placed_on') }} {{ formatDate(order.createdAt) }}
          </p>
        </div>
        <StatusBadge
          :status="order.status"
          type="order"
          :label="enumLabel('order_status', order.status)"
          class="order-detail__status"
        />
      </header>

      <!-- 状态提示条 -->
      <section v-if="order.status === 'pending_approval' && canApprove" class="callout callout--warning">
        <Icon name="heroicons:shield-exclamation" class="callout__icon" aria-hidden="true" />
        <div class="callout__body">
          <p class="callout__title">{{ t('customer.orders.approval_pending_title') }}</p>
          <p class="callout__text">{{ t('customer.orders.approval_pending_body') }}</p>
          <input
            v-model="approvalComment"
            type="text"
            :placeholder="t('customer.orders.approval_comment_placeholder')"
            class="order-input mt-3 max-w-md"
          />
          <p v-if="approvalMessage" class="mt-2 text-sm" :class="approvalError ? 'text-red-600' : 'text-green-600'">{{ approvalMessage }}</p>
        </div>
        <div class="callout__actions">
          <button type="button" :disabled="approvalProcessing" class="btn-primary" @click="approveOrderAction">
            {{ approvalProcessing ? t('customer.orders.approval_processing') : t('customer.orders.approval_approve') }}
          </button>
          <button type="button" :disabled="approvalProcessing" class="btn-ghost-danger" @click="rejectOrderAction">
            {{ t('customer.orders.approval_reject') }}
          </button>
        </div>
      </section>

      <section v-else-if="order.status === 'pending_approval' && isOwner && !canApprove" class="callout callout--warning">
        <Icon name="heroicons:clock" class="callout__icon" aria-hidden="true" />
        <div class="callout__body">
          <p class="callout__title">{{ t('customer.orders.approval_pending_title') }}</p>
          <p class="callout__text">{{ t('customer.orders.approval_pending_body') }}</p>
        </div>
      </section>

      <!-- H23: gate the confirm control on the backend-authoritative status, not a
           source allowlist. pending_confirmation orders arrive as ai_assist / bulk
           (requisition & reorder) / inquiry, and cart-sourced orders reach this
           status after the buyer-org approval flow (pending_approval →
           pending_confirmation). The backend CustomerConfirmOrder endpoint only
           checks status + ownership, so a source allowlist would leave a
           confirmable order without any actionable control. -->
      <section v-if="order.status === 'pending_confirmation'" class="callout callout--amber">
        <Icon name="heroicons:sparkles" class="callout__icon" aria-hidden="true" />
        <div class="callout__body">
          <p class="callout__title">{{ t('customer.orders.ai_draft_notice') }}</p>
          <p class="callout__text">{{ t('customer.orders.ai_draft_review') }}</p>
          <div v-if="!order.complianceOfficialEvidence" class="callout callout--danger mt-3 !p-3">
            <p class="callout__title text-sm">{{ t('customer.orders.compliance_required') }}</p>
            <p class="callout__text text-sm">{{ t('customer.orders.compliance_no_evidence') }}</p>
            <label class="mt-2 inline-flex items-start gap-2 text-sm">
              <input v-model="complianceAck" type="checkbox" class="mt-0.5 rounded border-red-300 text-red-600 focus:ring-red-500" />
              <span>{{ t('customer.orders.compliance_ack') }}</span>
            </label>
          </div>
        </div>
        <button
          :disabled="confirming || (!order.complianceOfficialEvidence && !complianceAck)"
          class="btn-primary shrink-0"
          @click="confirmOrder"
        >
          <Icon name="heroicons:check-circle" class="h-4 w-4" aria-hidden="true" />
          {{ confirming ? t('customer.orders.confirming') : t('customer.orders.confirm_order') }}
        </button>
      </section>

      <section v-if="order.status === 'pending'" class="callout callout--warning">
        <Icon name="heroicons:clock" class="callout__icon" aria-hidden="true" />
        <div class="callout__body">
          <p class="callout__title">{{ t('customer.orders.pending_request_title') }}</p>
          <p class="callout__text">{{ t('customer.orders.pending_request_body') }}</p>
        </div>
      </section>

      <div v-if="error" class="order-detail__error mb-4">
        <Icon name="heroicons:exclamation-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
        <p>{{ error }}</p>
      </div>

      <!-- 主布局 -->
      <div class="order-detail__grid">
        <div class="order-detail__main space-y-6">
          <!-- 进度 -->
          <section v-if="progress?.steps?.length" class="panel">
            <h2 class="panel__title">{{ t('customer.orders.progress_title') }}</h2>
            <div class="progress-track hidden sm:flex">
              <template v-for="(step, idx) in progress.steps" :key="step">
                <div class="progress-track__step" :class="{ 'progress-track__step--done': idx <= progress.currentStep, 'progress-track__step--current': idx === progress.currentStep }">
                  <div class="progress-track__dot">
                    <Icon v-if="idx < progress.currentStep" name="heroicons:check" class="h-4 w-4" aria-hidden="true" />
                    <span v-else>{{ Number(idx) + 1 }}</span>
                  </div>
                  <p class="progress-track__label">{{ enumLabel('order_status', step) }}</p>
                  <p v-if="stepTimestamp(step)" class="progress-track__date">{{ formatDate(stepTimestamp(step)) }}</p>
                </div>
                <div v-if="Number(idx) < progress.steps.length - 1" class="progress-track__line" :class="{ 'progress-track__line--done': Number(idx) < progress.currentStep }" />
              </template>
            </div>
            <!-- 移动端纵向时间线 -->
            <div class="sm:hidden progress-vertical">
              <div v-for="(step, idx) in progress.steps" :key="`m-${step}`" class="progress-vertical__item">
                <div class="progress-vertical__rail">
                  <div class="progress-vertical__dot" :class="{ 'progress-vertical__dot--done': idx <= progress.currentStep }" />
                  <div v-if="Number(idx) < progress.steps.length - 1" class="progress-vertical__line" :class="{ 'progress-vertical__line--done': Number(idx) < progress.currentStep }" />
                </div>
                <div class="pb-4">
                  <p class="text-sm font-medium text-gray-900">{{ enumLabel('order_status', step) }}</p>
                  <p v-if="stepTimestamp(step)" class="text-xs text-gray-500 mt-0.5">{{ formatDate(stepTimestamp(step)) }}</p>
                </div>
              </div>
            </div>
            <p v-if="progress.trackingCode" class="mt-4 text-sm text-gray-600 flex items-center gap-2">
              <Icon name="heroicons:truck" class="h-4 w-4 text-orange-500" aria-hidden="true" />
              {{ t('customer.orders.tracking_number') }}:
              <span class="font-mono font-medium text-gray-900">{{ progress.trackingCode }}</span>
            </p>
          </section>

          <!-- 商品明细 -->
          <section class="panel">
            <h2 class="panel__title">{{ t('customer.orders.order_summary') }}</h2>
            <div v-if="order.items?.length" class="table-wrap">
              <table class="order-table">
                <thead>
                  <tr>
                    <th>{{ t('customer.orders.col_product_id') }}</th>
                    <th>{{ t('customer.orders.col_quantity') }}</th>
                    <th>{{ t('customer.orders.col_unit_price') }}</th>
                    <th>{{ t('customer.orders.col_subtotal') }}</th>
                    <th class="hidden md:table-cell">{{ t('customer.orders.col_specifications') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in order.items" :key="`${item.productId}-${idx}`">
                    <td><span class="font-mono text-orange-700">{{ item.productId }}</span></td>
                    <td>{{ item.quantity }}</td>
                    <td>{{ cur(order.currency) }} {{ formatNumber(item.unitPrice || 0) }}</td>
                    <td class="font-semibold">{{ cur(order.currency) }} {{ formatNumber((item.quantity || 0) * (item.unitPrice || 0)) }}</td>
                    <td class="hidden md:table-cell text-gray-500">{{ item.specifications || t('customer.orders.spec_na') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <EmptyState
              v-else
              icon="heroicons:cube"
              :title="t('customer.orders.no_items')"
              class="!py-10"
            />
          </section>

          <!-- 消息 -->
          <section class="panel">
            <h2 class="panel__title">{{ t('customer.orders.messages_title') }}</h2>
            <div ref="messageListRef" class="chat-box">
              <p v-if="messages.length === 0" class="chat-box__empty">{{ t('customer.orders.no_messages') }}</p>
              <div
                v-for="msg in messages"
                :key="msg.id"
                class="chat-bubble-wrap"
                :class="msg.senderType === 'customer' ? 'chat-bubble-wrap--end' : 'chat-bubble-wrap--start'"
              >
                <div class="chat-bubble" :class="msg.senderType === 'customer' ? 'chat-bubble--mine' : 'chat-bubble--theirs'">
                  <p class="chat-bubble__sender">
                    {{ msg.senderType === 'customer' ? t('customer.orders.sender_customer') : t('customer.orders.sender_admin') }}
                  </p>
                  <p v-if="msg.message?.trim()" class="chat-bubble__text">{{ msg.message }}</p>
                  <ul v-if="msg.attachments?.length" class="chat-bubble__files">
                    <li v-for="(url, fileIdx) in msg.attachments" :key="fileIdx">
                      <a :href="url" target="_blank" rel="noopener noreferrer" class="chat-bubble__file-link">
                        <Icon name="heroicons:paper-clip" class="h-3.5 w-3.5" aria-hidden="true" />
                        {{ attachmentLabel(url) }}
                      </a>
                    </li>
                  </ul>
                  <p class="chat-bubble__time">{{ formatDate(msg.createdAt) }}</p>
                  <span
                    v-if="msg.senderType === 'customer'"
                    class="chat-bubble__receipt"
                    :class="{ 'chat-bubble__receipt--read': msg.readAt }"
                    :title="msg.readAt ? formatDate(msg.readAt) : ''"
                    aria-hidden="true"
                  >{{ msg.readAt ? '✓✓' : '✓' }}</span>
                </div>
              </div>
            </div>
            <div class="chat-compose">
              <input
                v-model="newMessage"
                :placeholder="t('customer.orders.message_placeholder')"
                class="order-input flex-1"
                :disabled="sending"
                @keyup.enter="sendMessage"
              />
              <label class="chat-compose__attach" :title="t('customer.orders.add_attachment')">
                <Icon name="heroicons:paper-clip" class="h-5 w-5" aria-hidden="true" />
                <input
                  type="file"
                  class="sr-only"
                  multiple
                  :accept="CUSTOMER_ATTACHMENT_ACCEPT"
                  :disabled="sending"
                  @change="handleMessageFiles"
                />
              </label>
              <button
                type="button"
                class="btn-primary !px-4"
                :disabled="sending || (!newMessage.trim() && !messageFiles.length)"
                @click="sendMessage"
              >
                <Icon name="heroicons:paper-airplane" class="h-4 w-4" aria-hidden="true" />
                <span class="hidden sm:inline">{{ sending ? t('customer.orders.sending') : t('customer.orders.send_message') }}</span>
              </button>
            </div>
            <p class="text-xs text-gray-500 mt-2">{{ t('customer.orders.message_attachments_hint') }}</p>
            <p v-if="messageFileError" class="text-xs text-red-600 mt-1">{{ messageFileError }}</p>
            <ul v-if="messageFiles.length" class="mt-2 space-y-1">
              <li
                v-for="(file, idx) in messageFiles"
                :key="`${file.name}-${idx}`"
                class="flex items-center justify-between text-xs text-gray-600 bg-gray-50 rounded px-2 py-1"
              >
                <span class="truncate">{{ file.name }}</span>
                <button type="button" class="text-gray-400 hover:text-red-600 ml-2" @click="removeMessageFile(idx)">
                  <Icon name="heroicons:x-mark" class="h-3.5 w-3.5" />
                </button>
              </li>
            </ul>
          </section>

          <!-- 在线支付（Stripe / PayPal） -->
          <section v-if="order.paymentStatus !== 'paid' && order.paymentStatus !== 'refunded'" class="panel">
            <h2 class="panel__title">{{ t('customer.orders.gateway_payment') }}</h2>
            <p class="text-sm text-gray-500 mb-3">{{ t('customer.orders.gateway_payment_hint') }}</p>
            <div class="flex flex-wrap gap-2">
              <button type="button" class="btn-secondary" :disabled="Boolean(gatewayLoading)" @click="startGatewayPayment('stripe')">
                {{ gatewayLoading === 'stripe' ? '...' : t('customer.orders.pay_with_stripe') }}
              </button>
              <button type="button" class="btn-secondary" :disabled="Boolean(gatewayLoading)" @click="startGatewayPayment('paypal')">
                {{ gatewayLoading === 'paypal' ? '...' : t('customer.orders.pay_with_paypal') }}
              </button>
            </div>
            <p v-if="gatewayMessage" class="mt-2 text-sm" :class="gatewayError ? 'text-red-600' : 'text-green-600'">{{ gatewayMessage }}</p>
          </section>

          <!-- 付款凭证 -->
          <section v-if="order.paymentStatus !== 'paid' && order.paymentStatus !== 'refunded'" class="panel">
            <h2 class="panel__title">{{ t('customer.orders.upload_payment_proof') }}</h2>
            <form class="space-y-4 max-w-lg" @submit.prevent="uploadPaymentProof">
              <div>
                <label class="field-label">{{ t('customer.orders.payment_method') }}</label>
                <select v-model="paymentForm.method" class="order-input">
                  <option value="bank_transfer">{{ t('customer.orders.payment_method_bank') }}</option>
                  <option value="swift">{{ t('customer.orders.payment_method_swift') }}</option>
                  <option value="other">{{ t('customer.orders.payment_method_other') }}</option>
                </select>
              </div>
              <div>
                <label class="field-label">{{ t('customer.orders.payment_proof') }} *</label>
                <p class="text-xs text-gray-500 mb-2">{{ t('customer.orders.payment_proof_hint') }}</p>
                <InputFile
                  id="payment-proof-file"
                  :accept="PAYMENT_PROOF_ACCEPT"
                  :max-files="1"
                  :max-size="PAYMENT_PROOF_MAX_MB"
                  :error="paymentFileError"
                  @files-selected="handlePaymentFileSelected"
                  @error="paymentFileError = $event"
                />
              </div>
              <p v-if="uploadMessage" class="text-sm" :class="uploadError ? 'text-red-600' : 'text-green-600'">{{ uploadMessage }}</p>
              <button type="submit" :disabled="uploadingPayment" class="btn-primary">
                {{ uploadingPayment ? t('customer.orders.uploading') : t('customer.orders.upload') }}
              </button>
            </form>
          </section>
        </div>

        <!-- 侧栏 -->
        <aside class="order-detail__aside space-y-4">
          <!-- 金额摘要 -->
          <div class="summary-card">
            <p class="summary-card__label">
              {{ proformaInvoice ? t('customer.orders.pricing_pi_confirmed') : t('customer.orders.total_amount') }}
            </p>
            <p class="summary-card__amount">
              {{ cur(displayPricing.currency) }} {{ formatNumber(displayPricing.total) }}
            </p>
            <div class="summary-card__breakdown">
              <div class="summary-card__row">
                <span>{{ t('customer.orders.pricing_subtotal') }}</span>
                <span>{{ cur(displayPricing.currency) }} {{ formatNumber(displayPricing.subtotal) }}</span>
              </div>
              <div class="summary-card__row">
                <span>{{ t('customer.orders.pricing_tax') }}</span>
                <span>{{ cur(displayPricing.currency) }} {{ formatNumber(displayPricing.tax) }}</span>
              </div>
              <div v-if="displayPricing.shipping > 0 || !proformaInvoice" class="summary-card__row">
                <span>{{ t('customer.orders.pricing_shipping') }}</span>
                <span>{{ cur(displayPricing.currency) }} {{ formatNumber(displayPricing.shipping) }}</span>
              </div>
            </div>
            <p v-if="proformaInvoice" class="summary-card__notice summary-card__notice--pi">
              {{ t('customer.orders.pricing_pi_total') }}: {{ proformaInvoice.invoiceNo || proformaInvoice.id }}
              <NuxtLink
                :to="localePath(`/customer/invoices/${proformaInvoice.id}`)"
                class="summary-card__pi-link"
              >
                {{ t('customer.orders.pricing_view_pi') }}
              </NuxtLink>
            </p>
            <p v-else class="summary-card__notice">{{ t('customer.orders.pricing_pi_notice') }}</p>
            <details v-if="proformaInvoice && orderSnapshotDiffers" class="summary-card__snapshot">
              <summary>{{ t('customer.orders.pricing_snapshot_reference') }}</summary>
              <div class="summary-card__row mt-2">
                <span>{{ t('customer.orders.pricing_subtotal') }}</span>
                <span>{{ cur(order.currency) }} {{ formatNumber(order.subtotal ?? 0) }}</span>
              </div>
              <div class="summary-card__row">
                <span>{{ t('customer.orders.pricing_tax') }}</span>
                <span>{{ cur(order.currency) }} {{ formatNumber(order.taxAmount ?? 0) }}</span>
              </div>
              <div class="summary-card__row">
                <span>{{ t('customer.orders.pricing_shipping') }}</span>
                <span>{{ cur(order.currency) }} {{ formatNumber(order.shippingAmount ?? 0) }}</span>
              </div>
            </details>
            <div class="summary-card__row">
              <span>{{ t('customer.orders.payment_status') }}</span>
              <StatusBadge
                :status="order.paymentStatus || 'unpaid'"
                type="payment"
                :label="enumLabel('payment_status', order.paymentStatus, 'unpaid')"
              />
            </div>
          </div>

          <!-- 订单信息 -->
          <div class="panel panel--compact">
            <h3 class="panel__subtitle">{{ t('customer.orders.reference_number') }}</h3>
            <p class="font-mono text-xs text-gray-600 break-all">{{ order.id }}</p>

            <h3 class="panel__subtitle mt-4">{{ t('customer.orders.shipping_address') }}</h3>
            <p class="text-sm text-gray-700 leading-relaxed whitespace-pre-line">{{ formatAddress(order.shippingAddress) }}</p>

            <h3 class="panel__subtitle mt-4">{{ t('customer.orders.tracking_number') }}</h3>
            <p class="text-sm text-gray-700">{{ order.trackingNumber || t('customer.orders.not_available') }}</p>
          </div>

          <!-- 时间节点 -->
          <div class="panel panel--compact">
            <h3 class="panel__subtitle">{{ t('customer.orders.progress_title') }}</h3>
            <dl class="date-list">
              <div class="date-list__item">
                <dt>{{ t('customer.orders.confirmed_at') }}</dt>
                <dd>{{ formatDate(order.confirmedAt) }}</dd>
              </div>
              <div class="date-list__item">
                <dt>{{ t('customer.orders.shipped_at') }}</dt>
                <dd>{{ formatDate(order.shippedAt) }}</dd>
              </div>
              <div class="date-list__item">
                <dt>{{ t('customer.orders.estimated_completion') }}</dt>
                <dd>{{ formatDate(order.estimatedCompletion) }}</dd>
              </div>
              <div class="date-list__item">
                <dt>{{ t('customer.orders.delivered_at') }}</dt>
                <dd>{{ formatDate(order.deliveredAt) }}</dd>
              </div>
            </dl>
          </div>

          <!-- 操作 -->
          <div v-if="['pending', 'confirmed', 'production'].includes(order.status)" class="panel panel--compact panel--action">
            <p class="text-sm text-gray-700">{{ t('customer.orders.nudge_description') }}</p>
            <button type="button" class="btn-secondary w-full mt-3" :disabled="nudging" @click="nudgeOrder">
              <Icon name="heroicons:bell-alert" class="h-4 w-4" aria-hidden="true" />
              {{ nudging ? t('customer.orders.nudging') : t('customer.orders.nudge_button') }}
            </button>
            <p v-if="nudgeMessage" class="mt-2 text-xs" :class="nudgeError ? 'text-red-600' : 'text-green-600'">{{ nudgeMessage }}</p>
          </div>

          <div v-if="['delivered', 'shipped'].includes(order.status)" class="panel panel--compact panel--action">
            <p class="text-sm font-medium text-gray-900">{{ t('customer.returns.request_title') }}</p>
            <p class="text-xs text-gray-500 mt-1">{{ t('customer.returns.request_desc') }}</p>
            <button type="button" class="btn-secondary w-full mt-3" @click="showReturnModal = true">
              <Icon name="heroicons:arrow-uturn-left" class="h-4 w-4" aria-hidden="true" />
              {{ t('customer.returns.request_button') }}
            </button>
          </div>

          <div v-if="['pending', 'pending_confirmation'].includes(order.status)" class="panel panel--compact panel--danger">
            <p class="text-sm font-medium text-red-800">{{ t('customer.orders.cancel_order') }}</p>
            <p class="text-xs text-red-600/80 mt-1">{{ t('customer.orders.cancel_order_desc') }}</p>
            <button type="button" class="btn-danger w-full mt-3" :disabled="cancelling" @click="cancelOrder">
              {{ cancelling ? t('customer.orders.cancelling') : t('customer.orders.cancel_order') }}
            </button>
            <p v-if="cancelError" class="mt-2 text-xs text-red-600">{{ cancelError }}</p>
          </div>
        </aside>
      </div>
    </template>

    <!-- 退货弹窗 -->
    <Teleport to="body">
      <div v-if="showReturnModal" class="modal-root">
        <button type="button" class="modal-backdrop" :aria-label="t('common.close')" @click="showReturnModal = false" />
        <div class="modal-panel" role="dialog" aria-modal="true">
          <div class="modal-panel__header">
            <h3>{{ t('customer.returns.modal_title') }}</h3>
            <button type="button" class="modal-panel__close" :aria-label="t('common.close')" @click="showReturnModal = false">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <form class="modal-panel__body space-y-4" @submit.prevent="submitReturn">
            <div>
              <label class="field-label">{{ t('customer.returns.col_reason') }}</label>
              <select v-model="returnForm.reason" required class="order-input">
                <option value="damaged">{{ t('customer.returns.reason_damaged') }}</option>
                <option value="wrong_item">{{ t('customer.returns.reason_wrong_item') }}</option>
                <option value="defective">{{ t('customer.returns.reason_defective') }}</option>
                <option value="not_as_described">{{ t('customer.returns.reason_not_as_described') }}</option>
                <option value="expired">{{ t('customer.returns.reason_expired') }}</option>
                <option value="other">{{ t('customer.returns.reason_other') }}</option>
              </select>
            </div>
            <div>
              <label class="field-label">{{ t('customer.returns.notes') }}</label>
              <textarea v-model="returnForm.notes" rows="2" class="order-input" />
            </div>
            <div v-if="order?.items?.length">
              <p class="field-label">{{ t('customer.returns.items_title') }}</p>
              <div v-for="(item, idx) in returnForm.items" :key="idx" class="return-item-row">
                <input v-model="item.selected" type="checkbox" class="rounded border-gray-300 text-orange-600" />
                <span class="font-mono text-sm flex-1 truncate">{{ item.productId }}</span>
                <input v-model.number="item.quantity" type="number" min="1" :max="item.maxQty" class="order-input !w-20 !py-1.5 text-sm" />
              </div>
            </div>
            <p v-if="returnError" class="text-sm text-red-600">{{ returnError }}</p>
            <div class="modal-panel__footer">
              <button type="button" class="btn-ghost" @click="showReturnModal = false">{{ t('customer.returns.cancel') }}</button>
              <button type="submit" :disabled="submittingReturn" class="btn-primary">
                {{ submittingReturn ? t('customer.returns.submitting') : t('customer.returns.submit') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, onUnmounted } from 'vue'
import {
  CUSTOMER_ATTACHMENT_ACCEPT,
  ORDER_MESSAGE_MAX_FILES,
  PAYMENT_PROOF_ACCEPT,
  PAYMENT_PROOF_MAX_MB,
  isCustomerAttachmentAllowed,
  isPaymentProofAllowed,
} from '~/utils/customerAttachments'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, enumLabel, formatNumber, formatDate } = useDisplay()
const api = useApi()
const id = route.params.id as string
const order = ref<any>(null)
const proformaInvoice = ref<any>(null)

/** 选取本订单最优 PI（形式发票，非 voided） */
const pickProformaInvoice = (invoices: any[], orderId: string) => {
  const list = (invoices || []).filter((inv) => {
    if (String(inv.orderId) !== String(orderId)) return false
    const type = String(inv.type || inv.invoiceType || '').toLowerCase()
    if (type !== 'proforma') return false
    const st = String(inv.status || '').toLowerCase()
    return st !== 'voided'
  })
  const rank = (inv: any) => {
    const st = String(inv.status || '').toLowerCase()
    if (st === 'paid') return 0
    if (st === 'sent') return 1
    if (st === 'overdue') return 2
    if (st === 'draft') return 3
    return 4
  }
  return list.sort((a, b) => rank(a) - rank(b))[0] || null
}

const displayPricing = computed(() => {
  const o = order.value
  const pi = proformaInvoice.value
  if (pi) {
    const shipping = Math.max(0, Number(o?.shippingAmount) || 0)
    const piTotal = Number(pi.totalAmount) || 0
    return {
      currency: pi.currency || o?.currency || 'USD',
      subtotal: Number(pi.amount) || 0,
      tax: Number(pi.taxAmount) || 0,
      shipping,
      total: piTotal + shipping
    }
  }
  return {
    currency: o?.currency || 'USD',
    subtotal: Number(o?.subtotal) || 0,
    tax: Number(o?.taxAmount) || 0,
    shipping: Number(o?.shippingAmount) || 0,
    total: Number(o?.totalAmount) || 0
  }
})

const orderSnapshotDiffers = computed(() => {
  const pi = proformaInvoice.value
  const o = order.value
  if (!pi || !o) return false
  const eps = 0.01
  return Math.abs((Number(pi.amount) || 0) - (Number(o.subtotal) || 0)) > eps
    || Math.abs((Number(pi.taxAmount) || 0) - (Number(o.taxAmount) || 0)) > eps
    || Math.abs((Number(pi.totalAmount) || 0) - (Number(o.subtotal) || 0) - (Number(o.taxAmount) || 0)) > eps
})
const canApprove = ref(false)
const isOwner = ref(true)
const approvalComment = ref('')
const approvalProcessing = ref(false)
const approvalMessage = ref('')
const approvalError = ref(false)
const pending = ref(true)
const error = ref('')
const confirming = ref(false)
const complianceAck = ref(false)
const cancelling = ref(false)
const cancelError = ref('')
const uploadingPayment = ref(false)
const uploadMessage = ref('')
const uploadError = ref(false)
const selectedFile = ref<File | null>(null)
const paymentFileError = ref('')

const paymentForm = reactive({ method: 'bank_transfer' })
const gatewayLoading = ref('')
const gatewayMessage = ref('')
const gatewayError = ref(false)

const nudging = ref(false)
const nudgeMessage = ref('')
const nudgeError = ref(false)
const messages = ref<any[]>([])
const newMessage = ref('')
const messageFiles = ref<File[]>([])
const messageFileError = ref('')
const sending = ref(false)
const messageListRef = ref<HTMLElement | null>(null)
const progress = ref<any>(null)
const showReturnModal = ref(false)
const submittingReturn = ref(false)
const returnError = ref('')
const returnForm = reactive({
  reason: 'damaged',
  notes: '',
  items: [] as { selected: boolean; productId: string; orderItemIdx: number; quantity: number; maxQty: number }[],
})

const stepTimestamp = (step: string) => {
  if (!progress.value?.timestamps) return null
  if (step === 'confirmed') return progress.value.timestamps.confirmedAt
  if (step === 'shipped') return progress.value.timestamps.shippedAt
  if (step === 'delivered') return progress.value.timestamps.deliveredAt
  return null
}

const fetchProgress = async () => {
  if (!order.value?.id) return
  try {
    progress.value = await api.get<any>(`/user/orders/${order.value.id}/progress`)
  } catch { progress.value = null }
}

const initReturnItems = () => {
  if (!order.value?.items) return
  returnForm.items = order.value.items.map((item: any, idx: number) => ({
    selected: idx === 0,
    productId: item.productId,
    orderItemIdx: idx,
    quantity: item.quantity || 1,
    maxQty: item.quantity || 1,
  }))
}

const submitReturn = async () => {
  if (!order.value?.id) return
  const selected = returnForm.items.filter(i => i.selected)
  if (!selected.length) {
    returnError.value = t('customer.returns.error_no_items')
    return
  }
  submittingReturn.value = true
  returnError.value = ''
  try {
    const payload = {
      reason: returnForm.reason,
      notes: returnForm.notes,
      items: selected.map(i => ({
        orderItemIdx: i.orderItemIdx,
        productId: i.productId,
        quantity: i.quantity,
        reasonCode: returnForm.reason,
        condition: 'opened',
      })),
    }
    const ret = await api.post<any>(`/user/orders/${order.value.id}/returns`, payload)
    showReturnModal.value = false
    await navigateTo(localePath(`/customer/returns/${ret.id}`))
  } catch (err: any) {
    returnError.value = err?.message || t('errors.api.save_failed')
  } finally {
    submittingReturn.value = false
  }
}

const nudgeOrder = async () => {
  if (!order.value?.id) return
  nudging.value = true; nudgeMessage.value = ''; nudgeError.value = false
  try {
    await api.post(`/user/orders/${order.value.id}/nudge`, {})
    nudgeMessage.value = t('customer.orders.nudge_success')
  } catch (err: any) {
    nudgeError.value = true
    nudgeMessage.value = err?.data?.message || err?.message || t('customer.orders.nudge_error')
  } finally { nudging.value = false }
}

const fetchMessages = async () => {
  if (!order.value?.id) return
  try {
    const res = await api.get<any>(`/user/orders/${order.value.id}/messages`)
    messages.value = res.data || []
  } catch (err: any) {
    messageFileError.value = err?.message || t('errors.api.load_failed')
  }
}

const markMessagesRead = async () => {
  if (!order.value?.id) return
  try {
    await api.post(`/user/orders/${order.value.id}/messages/read`, {})
  } catch {
    // A reconnect REST fetch also marks messages read.
  }
}

const appendMessage = (msg: any) => {
  if (!msg?.id || messages.value.some((existing) => existing.id === msg.id)) return
  messages.value.push(msg)
  nextTick(() => {
    if (messageListRef.value) messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  })
}

const applyReadReceipt = (payload: any) => {
  if (!payload?.readerType || !payload?.readAt) return
  for (const msg of messages.value) {
    if (msg.senderType !== payload.readerType && !msg.readAt) msg.readAt = payload.readAt
  }
}

const messageStream = useOrderMessageStream(
  `/user/orders/${id}/messages/ws`,
  (event) => {
    if (event.type === 'order_message') {
      appendMessage(event.payload)
      if (event.payload?.senderType === 'admin') void markMessagesRead()
    } else if (event.type === 'order_messages_read') {
      applyReadReceipt(event.payload)
    }
  },
  fetchMessages,
)

const attachmentLabel = (url: string) => {
  const name = url.split('/').pop() || url
  return decodeURIComponent(name.replace(/^\d+_/, ''))
}

const handleMessageFiles = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  messageFileError.value = ''
  for (const file of Array.from(input.files)) {
    if (messageFiles.value.length >= ORDER_MESSAGE_MAX_FILES) {
      messageFileError.value = t('customer.orders.message_max_files', { count: ORDER_MESSAGE_MAX_FILES })
      break
    }
    if (!isCustomerAttachmentAllowed(file)) {
      messageFileError.value = t('customer.orders.message_file_type_invalid')
      continue
    }
    messageFiles.value.push(file)
  }
  input.value = ''
}

const removeMessageFile = (index: number) => {
  messageFiles.value.splice(index, 1)
}

const sendMessage = async () => {
  if (!order.value?.id) return
  if (!newMessage.value.trim() && !messageFiles.value.length) return
  sending.value = true
  try {
    const msg = await api.customerSendOrderMessage(order.value.id, {
      message: newMessage.value.trim(),
      files: messageFiles.value.length ? messageFiles.value : undefined,
    })
    appendMessage(msg)
    newMessage.value = ''
    messageFiles.value = []
    nextTick(() => {
      if (messageListRef.value) messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    })
  } catch (err: any) {
    notifyError(err, t('customer.orders.message_send_error'))
  } finally { sending.value = false }
}

onMounted(async () => {
  await fetchOrder()
  await fetchOrderInvoices()
  await fetchProgress()
  initReturnItems()
  await fetchMessages()
  messageStream.connect()
})
onUnmounted(messageStream.stop)

const fetchOrder = async () => {
  pending.value = true; error.value = ''
  try {
    const data = await api.getOrder(id)
    order.value = data
    canApprove.value = !!data?.canApprove
    isOwner.value = data?.isOwner !== false
    complianceAck.value = false
    initReturnItems()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const fetchOrderInvoices = async () => {
  try {
    const res = await api.getCustomerInvoices()
    proformaInvoice.value = pickProformaInvoice(res?.data || [], id)
  } catch {
    proformaInvoice.value = null
  }
}

const approveOrderAction = async () => {
  if (!order.value?.id) return
  approvalProcessing.value = true
  approvalMessage.value = ''
  approvalError.value = false
  try {
    await api.approveOrder(order.value.id, approvalComment.value)
    approvalMessage.value = t('customer.orders.approval_success_approved')
    await fetchOrder()
  } catch (err: any) {
    approvalError.value = true
    approvalMessage.value = err?.message || t('customer.orders.approval_error')
  } finally { approvalProcessing.value = false }
}

const rejectOrderAction = async () => {
  if (!order.value?.id) return
  if (!confirm(t('customer.orders.approval_reject'))) return
  approvalProcessing.value = true
  approvalMessage.value = ''
  approvalError.value = false
  try {
    await api.rejectOrder(order.value.id, approvalComment.value)
    approvalMessage.value = t('customer.orders.approval_success_rejected')
    await fetchOrder()
  } catch (err: any) {
    approvalError.value = true
    approvalMessage.value = err?.message || t('customer.orders.approval_error')
  } finally { approvalProcessing.value = false }
}

const confirmOrder = async () => {
  if (!order.value?.id) return
  if (!order.value.complianceOfficialEvidence && !complianceAck.value) {
    error.value = t('customer.orders.error_compliance_ack')
    return
  }
  confirming.value = true; error.value = ''
  try {
    await api.confirmOrder(order.value.id, order.value.complianceOfficialEvidence ? true : complianceAck.value)
    await fetchOrder()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.confirm_failed')
  } finally { confirming.value = false }
}

const cancelOrder = async () => {
  if (!order.value?.id) return
  if (!confirm(t('customer.orders.cancel_confirm'))) return
  cancelling.value = true; cancelError.value = ''
  try {
    await api.post(`/user/orders/${order.value.id}/cancel`, {})
    await fetchOrder()
  } catch (err: any) {
    cancelError.value = err?.message || t('errors.api.cancel_failed')
  } finally { cancelling.value = false }
}

const formatAddress = (address: any) => {
  if (!address || typeof address !== 'object') return t('customer.orders.date_na')
  const fields = [address.street, address.city, address.state, address.zipCode, address.country]
    .filter((v) => typeof v === 'string' && v.trim() !== '')
  if (!fields.length) return t('customer.orders.date_na')
  return fields.join(', ')
}

const handlePaymentFileSelected = (files: FileList) => {
  paymentFileError.value = ''
  const file = files[0]
  if (!file) return
  if (!isPaymentProofAllowed(file)) {
    paymentFileError.value = t('customer.orders.payment_proof_type_invalid')
    selectedFile.value = null
    return
  }
  if (file.size > PAYMENT_PROOF_MAX_MB * 1024 * 1024) {
    paymentFileError.value = t('customer.orders.payment_proof_too_large', { max: PAYMENT_PROOF_MAX_MB })
    selectedFile.value = null
    return
  }
  selectedFile.value = file
}

const startGatewayPayment = async (method: 'stripe' | 'paypal') => {
  if (!order.value?.id) return
  gatewayLoading.value = method
  gatewayMessage.value = ''
  gatewayError.value = false
  try {
    const res = await api.customerCreateGatewayPayment(order.value.id, { method })
    if (method === 'paypal' && res.approvalUrl) {
      window.location.href = res.approvalUrl
      return
    }
    if (method === 'stripe' && res.clientSecret) {
      gatewayMessage.value = t('customer.orders.stripe_checkout_ready')
    } else {
      gatewayMessage.value = t('customer.orders.gateway_session_created')
    }
  } catch (err: any) {
    gatewayError.value = true
    gatewayMessage.value = err?.message || t('errors.api.payment_failed')
  } finally {
    gatewayLoading.value = ''
  }
}

const uploadPaymentProof = async () => {
  if (!selectedFile.value) {
    uploadError.value = true
    uploadMessage.value = t('customer.orders.error_select_file')
    return
  }
  uploadingPayment.value = true; uploadMessage.value = ''; uploadError.value = false
  try {
    const formData = new FormData()
    formData.append('method', paymentForm.method)
    formData.append('proof', selectedFile.value)
    await api.customerUploadPaymentProof(id, formData)
    uploadMessage.value = t('customer.orders.payment_uploaded')
    selectedFile.value = null
    await fetchOrder()
  } catch (err: any) {
    uploadError.value = true
    uploadMessage.value = err?.data?.message || err?.message || t('errors.api.upload_failed')
  } finally { uploadingPayment.value = false }
}
</script>

<style scoped>
.order-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.order-detail__back {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-bottom: 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-highlight);
  text-decoration: none;
  transition: color 0.15s;
}

.order-detail__back:hover {
  color: var(--color-highlight-hover);
}

.order-detail__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 4rem 1rem;
  color: var(--color-text-lighter);
}

.order-detail__error {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.875rem 1rem;
  border-radius: var(--radius-md);
  background: rgba(var(--color-error-rgb), 0.08);
  color: var(--color-error);
  font-size: 0.875rem;
}

.order-detail__hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 1.5rem;
  margin-bottom: 1.25rem;
  background: linear-gradient(135deg, rgba(var(--color-highlight-rgb), 0.08) 0%, var(--color-bg) 55%);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.order-detail__eyebrow {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-lighter);
  margin: 0 0 0.25rem;
}

.order-detail__title {
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 700;
  color: var(--color-primary);
  margin: 0;
  line-height: 1.2;
}

.order-detail__meta {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
}

.order-detail__status {
  font-size: 0.8125rem !important;
  padding: 0.375rem 0.875rem !important;
}

.order-detail__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

@media (min-width: 1024px) {
  .order-detail__grid {
    grid-template-columns: minmax(0, 1fr) 320px;
    align-items: start;
  }

  .order-detail__aside {
    position: sticky;
    top: 1rem;
  }
}

.panel {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  padding: 1.25rem 1.5rem;
  box-shadow: var(--shadow-sm);
}

.panel--compact {
  padding: 1rem 1.25rem;
}

.panel__title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-primary);
  margin: 0 0 1rem;
}

.panel__subtitle {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-lighter);
  margin: 0 0 0.375rem;
}

.summary-card {
  padding: 1.25rem;
  background: var(--color-primary-container);
  border-radius: var(--radius-xl);
  /* H4: primary-container flips to a light surface in dark mode, so fixed
     white text becomes invisible — use the paired on-container token. */
  color: var(--color-on-primary-container);
}

.summary-card__label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  opacity: 0.75;
  margin: 0;
}

.summary-card__amount {
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 700;
  margin: 0.375rem 0 0.75rem;
  line-height: 1.2;
}

.summary-card__breakdown {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.summary-card__notice {
  font-size: 0.6875rem;
  line-height: 1.45;
  opacity: 0.8;
  margin: 0 0 0.75rem;
}

.summary-card__notice--pi {
  opacity: 0.95;
}

.summary-card__pi-link {
  display: inline-block;
  margin-top: 0.25rem;
  text-decoration: underline;
  color: inherit;
  opacity: 0.95;
}

.summary-card__snapshot {
  font-size: 0.6875rem;
  opacity: 0.75;
  margin-bottom: 0.75rem;
}

.summary-card__snapshot summary {
  cursor: pointer;
  list-style: none;
}

.summary-card__snapshot summary::-webkit-details-marker {
  display: none;
}

.summary-card__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  font-size: 0.8125rem;
  opacity: 0.9;
}

.callout {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.875rem 1rem;
  padding: 1rem 1.25rem;
  margin-bottom: 1rem;
  border-radius: var(--radius-lg);
  border: 1px solid transparent;
}

.callout--warning {
  background: rgba(var(--color-warning-rgb), 0.1);
  border-color: rgba(var(--color-warning-rgb), 0.25);
}

.callout--amber {
  background: rgba(var(--color-warning-rgb), 0.12);
  border-color: rgba(var(--color-warning-rgb), 0.3);
}

.callout--danger {
  background: rgba(var(--color-error-rgb), 0.08);
  border-color: rgba(var(--color-error-rgb), 0.2);
}

.callout__icon {
  width: 1.375rem;
  height: 1.375rem;
  flex-shrink: 0;
  color: var(--color-warning-hover);
  margin-top: 0.125rem;
}

.callout--danger .callout__icon { color: var(--color-error); }

.callout__body { flex: 1; min-width: 200px; }

.callout__title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-primary);
  margin: 0;
}

.callout__text {
  font-size: 0.8125rem;
  color: var(--color-text-light);
  margin: 0.25rem 0 0;
  line-height: 1.5;
}

.callout__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  width: 100%;
}

@media (min-width: 640px) {
  .callout__actions { width: auto; margin-left: auto; }
}

.progress-track {
  align-items: flex-start;
  gap: 0;
}

.progress-track__step {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  min-width: 0;
}

.progress-track__dot {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--color-bg-alt);
  color: var(--color-text-lighter);
  border: 2px solid var(--color-border);
  transition: all 0.2s;
}

.progress-track__step--done .progress-track__dot,
.progress-track__step--current .progress-track__dot {
  background: var(--color-highlight);
  border-color: var(--color-highlight);
  color: white;
}

.progress-track__label {
  margin-top: 0.5rem;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.3;
  padding: 0 0.25rem;
}

.progress-track__date {
  font-size: 0.625rem;
  color: var(--color-text-lighter);
  margin-top: 0.125rem;
}

.progress-track__line {
  flex: 1;
  height: 2px;
  background: var(--color-border);
  margin-top: 1rem;
  min-width: 0.5rem;
}

.progress-track__line--done {
  background: var(--color-highlight);
}

.progress-vertical__item {
  display: flex;
  gap: 0.75rem;
}

.progress-vertical__rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 1rem;
}

.progress-vertical__dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 9999px;
  background: var(--color-border);
  flex-shrink: 0;
}

.progress-vertical__dot--done {
  background: var(--color-highlight);
}

.progress-vertical__line {
  flex: 1;
  width: 2px;
  min-height: 1.5rem;
  background: var(--color-border);
  margin: 0.25rem 0;
}

.progress-vertical__line--done {
  background: var(--color-highlight);
}

.table-wrap {
  overflow-x: auto;
  margin: 0 -0.25rem;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
}

.order-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.order-table th {
  padding: 0.625rem 1rem;
  text-align: left;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-lighter);
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border-light);
}

.order-table td {
  padding: 0.75rem 1rem;
  color: var(--color-text);
  border-bottom: 1px solid var(--color-border-light);
}

.order-table tbody tr:last-child td {
  border-bottom: none;
}

.order-table tbody tr:hover {
  background: rgba(var(--color-highlight-rgb), 0.03);
}

.chat-box {
  max-height: 18rem;
  overflow-y: auto;
  padding: 0.75rem;
  margin-bottom: 0.75rem;
  background: var(--color-bg-alt);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
}

.chat-box__empty {
  text-align: center;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
  padding: 2rem 0;
}

.chat-bubble-wrap {
  display: flex;
  margin-bottom: 0.625rem;
}

.chat-bubble-wrap--end { justify-content: flex-end; }
.chat-bubble-wrap--start { justify-content: flex-start; }

.chat-bubble {
  max-width: min(85%, 22rem);
  padding: 0.625rem 0.875rem;
  border-radius: var(--radius-lg);
}

.chat-bubble--mine {
  background: var(--color-highlight);
  color: white;
  border-bottom-right-radius: 0.25rem;
}

.chat-bubble--theirs {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-bottom-left-radius: 0.25rem;
}

.chat-bubble__sender {
  font-size: 0.6875rem;
  font-weight: 600;
  opacity: 0.85;
  margin: 0 0 0.25rem;
}

.chat-bubble__text {
  font-size: 0.875rem;
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.45;
}

.chat-bubble__time {
  font-size: 0.625rem;
  opacity: 0.7;
  margin: 0.375rem 0 0;
}

.chat-bubble__receipt {
  display: block;
  margin-top: 0.125rem;
  text-align: right;
  font-size: 0.6875rem;
  color: rgba(255, 255, 255, 0.65);
}

.chat-bubble__receipt--read {
  color: #d1fae5;
}

.chat-compose {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.chat-compose__attach {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
  border-radius: 0.5rem;
  border: 1px solid var(--color-border);
  color: var(--color-text-lighter);
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}

.chat-compose__attach:hover {
  color: var(--color-highlight);
  border-color: var(--color-highlight);
}

.chat-bubble__files {
  list-style: none;
  margin: 0.375rem 0 0;
  padding: 0;
}

.chat-bubble__file-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  text-decoration: underline;
  word-break: break-all;
}

.chat-bubble--mine .chat-bubble__file-link {
  color: inherit;
}

.chat-bubble--theirs .chat-bubble__file-link {
  color: var(--color-highlight);
}

.date-list__item {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.375rem 0;
  font-size: 0.8125rem;
  border-bottom: 1px solid var(--color-border-light);
}

.date-list__item:last-child { border-bottom: none; }

.date-list__item dt {
  color: var(--color-text-lighter);
  flex-shrink: 0;
}

.date-list__item dd {
  text-align: right;
  color: var(--color-text);
  margin: 0;
}

.panel--action {
  background: var(--color-bg-alt);
}

.panel--danger {
  background: rgba(var(--color-error-rgb), 0.06);
  border-color: rgba(var(--color-error-rgb), 0.2);
}

.field-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: 0.375rem;
}

.order-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: var(--color-text);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.order-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.12);
}

.file-input::file-selector-button {
  margin-right: 0.75rem;
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-highlight);
  background: var(--color-bg-alt);
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.btn-primary,
.btn-secondary,
.btn-danger,
.btn-ghost,
.btn-ghost-danger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: var(--radius-md);
  border: none;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
}

.btn-primary {
  background: var(--color-highlight);
  color: white;
}

.btn-primary:hover:not(:disabled) { background: var(--color-highlight-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  /* H4: hardcoded white bg + near-white primary text in dark mode was
     unreadable — use the theme surface token. */
  background: var(--color-bg);
  color: var(--color-primary);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover:not(:disabled) { background: var(--color-bg-alt); }

.btn-danger {
  background: var(--color-error);
  color: white;
}

.btn-danger:hover:not(:disabled) { filter: brightness(0.92); }

.btn-ghost {
  background: transparent;
  color: var(--color-text-light);
  border: 1px solid var(--color-border);
}

.btn-ghost-danger {
  background: transparent;
  color: var(--color-error);
  border: 1px solid rgba(var(--color-error-rgb), 0.35);
}

.return-item-row {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.625rem;
  margin-bottom: 0.375rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  background: var(--color-bg-alt);
}

.modal-root {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  border: none;
  cursor: pointer;
}

.modal-panel {
  position: relative;
  width: 100%;
  max-width: 28rem;
  background: var(--color-bg);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  overflow: hidden;
}

.modal-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-light);
}

.modal-panel__header h3 {
  font-size: 1.0625rem;
  font-weight: 600;
  margin: 0;
  color: var(--color-primary);
}

.modal-panel__close {
  display: flex;
  padding: 0.25rem;
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
  border-radius: var(--radius-sm);
}

.modal-panel__body {
  padding: 1.25rem;
}

.modal-panel__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding-top: 0.5rem;
}
</style>
