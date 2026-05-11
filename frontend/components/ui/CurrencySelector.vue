<template>
  <div class="currency-selector">
    <select
      :value="currency.selectedCurrency.value"
      class="currency-selector__dropdown"
      @change="currency.setCurrency(($event.target as HTMLSelectElement).value)"
    >
      <option
        v-for="code in currency.supportedCurrencies.value"
        :key="code"
        :value="code"
      >
        {{ code }} ({{ currencySymbols[code] || code }})
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { useCurrency } from '~/composables/useCurrency'

const currency = useCurrency()

const currencySymbols: Record<string, string> = {
  USD: '$', EUR: '€', CNY: '¥', GBP: '£', JPY: '¥',
  SAR: '﷼', AED: 'د.إ', KRW: '₩', SGD: 'S$',
  MYR: 'RM', THB: '฿', VND: '₫', IDR: 'Rp', PHP: '₱'
}
</script>

<style scoped>
.currency-selector__dropdown {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: var(--radius-md, 0.5rem);
  background-color: var(--color-bg, #fff);
  color: var(--color-text, #1a202c);
  font-size: 0.875rem;
  cursor: pointer;
  outline: none;
  transition: border-color 0.2s;
}
.currency-selector__dropdown:focus {
  border-color: var(--color-highlight, #f97316);
  box-shadow: 0 0 0 2px rgba(249, 115, 22, 0.2);
}
</style>
