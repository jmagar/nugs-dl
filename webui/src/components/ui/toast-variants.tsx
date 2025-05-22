import { toast as sonnerToast, ToastT } from "sonner"
import { CheckCircle, XCircle, AlertCircle, Info } from "lucide-react"
import { ReactNode } from "react"

type ToastVariant = "default" | "success" | "error" | "warning" | "info"

type ToastOptions = {
  title?: string
  description?: string
  icon?: ReactNode
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
  cancel?: {
    label: string
    onClick: () => void
  }
  position?: "top-right" | "bottom-right" | "bottom-center"
  onDismiss?: (toast: ToastT) => void
  onAutoClose?: (toast: ToastT) => void
  id?: string
}

const variantClasses: Record<ToastVariant, string> = {
  default: "border-l-0",
  success: "border-l-4 border-l-green-500",
  error: "border-l-4 border-l-red-500",
  warning: "border-l-4 border-l-amber-500",
  info: "border-l-4 border-l-blue-500",
}

const variantIcons: Record<ToastVariant, ReactNode> = {
  default: null,
  success: <CheckCircle className="h-5 w-5 text-green-500" />,
  error: <XCircle className="h-5 w-5 text-red-500" />,
  warning: <AlertCircle className="h-5 w-5 text-amber-500" />,
  info: <Info className="h-5 w-5 text-blue-500" />,
}

/**
 * A custom toast function that wraps sonner toast with styled variants
 */
export function toast(message: string, options?: ToastOptions) {
  return sonnerToast(message, {
    ...options,
    className: options?.icon ? "pl-2" : "pl-4",
  })
}

toast.success = (message: string, options?: ToastOptions) => {
  return sonnerToast(message, {
    ...options,
    icon: options?.icon || variantIcons.success,
    className: `${variantClasses.success} pl-2`,
  })
}

toast.error = (message: string, options?: ToastOptions) => {
  return sonnerToast(message, {
    ...options,
    icon: options?.icon || variantIcons.error,
    className: `${variantClasses.error} pl-2`,
  })
}

toast.warning = (message: string, options?: ToastOptions) => {
  return sonnerToast(message, {
    ...options,
    icon: options?.icon || variantIcons.warning,
    className: `${variantClasses.warning} pl-2`,
  })
}

toast.info = (message: string, options?: ToastOptions) => {
  return sonnerToast(message, {
    ...options,
    icon: options?.icon || variantIcons.info,
    className: `${variantClasses.info} pl-2`,
  })
}

// Re-export other sonner methods
toast.dismiss = sonnerToast.dismiss
toast.message = sonnerToast.message
toast.promise = sonnerToast.promise
toast.loading = sonnerToast.loading
toast.custom = sonnerToast.custom 