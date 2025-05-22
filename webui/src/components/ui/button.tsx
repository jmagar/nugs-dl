import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-all duration-300 disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
  {
    variants: {
      variant: {
        default: // Standard primary button (solid, if needed, else primary-gradient is the main one)
          "bg-primary-accent text-text-primary shadow-lg hover:bg-primary-accent/90 active:translate-y-px",
        "primary-gradient":
          "bg-gradient-to-r from-purple-600 to-pink-600 text-text-primary font-medium shadow-lg hover:from-purple-700 hover:to-pink-700 hover:shadow-purple-glow-md active:translate-y-px transform hover:-translate-y-px",
        destructive:
          "bg-error text-text-primary shadow-sm hover:bg-error/90 active:translate-y-px",
        outline:
          "border border-border-primary bg-transparent text-text-secondary shadow-sm hover:bg-primary-accent/10 hover:text-primary-accent hover:border-primary-accent/30 active:translate-y-px transition-all duration-200",
        secondary: // Kept for general use, might not perfectly match a spec item
          "bg-secondary-background text-text-secondary shadow-sm hover:bg-secondary-background/80 active:translate-y-px",
        ghost:
          "text-text-secondary hover:bg-secondary-background hover:text-text-primary active:translate-y-px transition-all duration-200",
        link: "text-primary-accent underline-offset-4 hover:underline",
      },
      size: {
        default: "h-11 px-3 py-2", // 2.75rem height, 0.75rem horizontal padding
        sm: "h-9 px-3 py-1",    // 2.25rem height, 0.75rem horiz, 0.25rem vert (spec wanted 0.5rem horiz for outline small)
        lg: "h-12 px-8", // Example large size, spec has h-10 for icon button, can adjust if specific lg needed
        icon: "h-10 w-10 p-0", // 2.5rem x 2.5rem
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
