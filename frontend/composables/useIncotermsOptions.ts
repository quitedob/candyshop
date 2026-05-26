/** 购物车/下单可用的 Incoterms 选项（与后端运费估算一致） */
export function useIncotermsOptions() {
  const { t } = useI18n()

  const incotermsOptions = computed(() => [
    { value: 'EXW', label: 'EXW', hint: t('customer.cart.incoterms_exw_hint') },
    { value: 'FOB', label: 'FOB', hint: t('customer.cart.incoterms_fob_hint') },
    { value: 'FCA', label: 'FCA', hint: '' },
    { value: 'CFR', label: 'CFR', hint: '' },
    { value: 'CIF', label: 'CIF', hint: '' },
    { value: 'DDP', label: 'DDP', hint: '' }
  ])

  const defaultIncoterms = 'FOB'

  const incotermsHint = (code: string) => {
    const opt = incotermsOptions.value.find(o => o.value === code)
    return opt?.hint || ''
  }

  return { incotermsOptions, defaultIncoterms, incotermsHint }
}
