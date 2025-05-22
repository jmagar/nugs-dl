"use client"

import { useTheme } from "next-themes"
import { Toaster as Sonner } from "sonner"

type ToasterProps = React.ComponentProps<typeof Sonner>

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "dark" } = useTheme()

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      position="bottom-right"
      toastOptions={{
        classNames: {
          toast: "group bg-card-background border border-border-primary rounded-lg shadow-lg p-4 text-white w-[350px] relative",
          title: "text-sm font-semibold",
          description: "text-xs text-text-secondary",
          actionButton: "bg-primary-accent text-white text-xs px-2 py-1 rounded hover:bg-primary-accent/90",
          cancelButton: "bg-secondary-background text-white text-xs px-2 py-1 rounded hover:bg-secondary-background/90",
          closeButton: "absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-text-secondary hover:text-white",
          success: "border-l-2 border-l-success pl-2",
          error: "border-l-2 border-l-error pl-2",
          warning: "border-l-2 border-l-warning pl-2",
          info: "border-l-2 border-l-info pl-2",
        },
        // Apply custom transition and animation styles
        style: {
          '--animation-enter': 'slide-in-from-right',
          '--animation-exit': 'fade-out',
          '--animation-duration': '300ms',
        } as React.CSSProperties,
      }}
      closeButton
      {...props}
    />
  )
}

export { Toaster }
