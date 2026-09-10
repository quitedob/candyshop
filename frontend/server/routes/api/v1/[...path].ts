import { getRequestURL, proxyRequest } from 'h3'

const API_PREFIX = '/api/v1'

export default defineEventHandler(event => {
  // Resolve on each request so a built image can use its deployment's backend.
  // Keep browser requests and SSR data fetching on the same API configuration.
  const backendApiBase = useRuntimeConfig(event).internalApiBase.replace(/\/+$/, '')
  const requestUrl = getRequestURL(event)
  const apiPath = requestUrl.pathname.slice(API_PREFIX.length)
  return proxyRequest(event, `${backendApiBase}${apiPath}${requestUrl.search}`)
})
