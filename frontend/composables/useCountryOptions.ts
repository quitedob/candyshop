/** 结账/表单用 ISO 国家选项 */
export function useCountryOptions() {
  const { t } = useI18n()
  const countryOptions = computed(() => [
    { value: 'US', label: t('form.options.countries.US') },
    { value: 'UK', label: t('form.options.countries.UK') },
    { value: 'DE', label: t('form.options.countries.DE') },
    { value: 'FR', label: t('form.options.countries.FR') },
    { value: 'CA', label: t('form.options.countries.CA') },
    { value: 'AU', label: t('form.options.countries.AU') },
    { value: 'JP', label: t('form.options.countries.JP') },
    { value: 'KR', label: t('form.options.countries.KR') },
    { value: 'CN', label: t('form.options.countries.CN') },
    { value: 'SG', label: t('form.options.countries.SG') },
    { value: 'MY', label: t('form.options.countries.MY') },
    { value: 'TH', label: t('form.options.countries.TH') },
    { value: 'VN', label: t('form.options.countries.VN') },
    { value: 'ID', label: t('form.options.countries.ID') },
    { value: 'PH', label: t('form.options.countries.PH') },
    { value: 'OTHER', label: t('form.options.countries.OTHER') }
  ])
  return { countryOptions }
}
