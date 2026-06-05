interface ErrorLike {
  message?: string
  data?: {
    message?: string
  }
}

export const notifyError = (error: unknown, fallback: string) => {
  const apiError = error as ErrorLike | undefined
  useToast().error(apiError?.data?.message || apiError?.message || fallback)
}
