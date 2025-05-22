import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-full border px-2.5 py-1 text-xs font-semibold w-fit whitespace-nowrap shrink-0 [&>svg]:size-4 [&>svg]:mr-1 [&>svg]:pointer-events-none transition-colors",
  {
    variants: {
      variant: {
        // Content type variants
        album: "bg-primary-accent/20 text-purple-300 border-primary-accent/30",
        video: "bg-info/20 text-blue-300 border-info/30",
        livestream: "bg-success/20 text-green-300 border-success/30",
        playlist: "bg-warning/20 text-amber-300 border-warning/30",
        
        // Status variants
        downloading: "bg-info/20 text-blue-300 border-info/30",
        queued: "bg-gray-500/20 text-gray-300 border-gray-500/30",
        paused: "bg-warning/20 text-amber-300 border-warning/30",
        completed: "bg-success/20 text-green-300 border-success/30",
        error: "bg-error/20 text-red-300 border-error/30",
        
        // Legacy variants - keeping for backward compatibility
        default: "border-transparent bg-primary-accent text-text-primary",
        secondary: "border-transparent bg-secondary-background text-text-secondary",
        destructive: "border-transparent bg-error text-white",
        outline: "text-text-primary border-border-primary",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span"

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
