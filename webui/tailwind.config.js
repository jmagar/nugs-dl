/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ["class"],
  content: [
    './pages/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './app/**/*.{ts,tsx}',
    './src/**/*.{ts,tsx}',
    './index.html', // Make sure index.html is scanned too
	],
  prefix: "",
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        'primary-background-start': '#111827', // gray-900
        'primary-background-end': '#000000',   // black
        'secondary-background': 'rgba(31, 41, 55, 0.5)', // gray-800 50% opacity
        'tertiary-background': 'rgba(17, 24, 39, 0.3)', // gray-900 30% opacity
        'footer-background': 'rgba(17, 24, 39, 0.5)', // Added for footer spec
        'input-field-background': 'rgba(17, 24, 39, 0.5)', // Added for input/textarea
        'card-background': 'rgba(31, 41, 55, 0.5)',      // gray-800 50% opacity
        'primary-accent': '#a855f7',      // purple-500
        'secondary-accent': '#ec4899',    // pink-500
        'text-primary': '#ffffff',        // white
        'text-secondary': '#9ca3af',      // gray-400
        'text-muted': '#6b7280',         // gray-500
        
        // Updated border colors to match UI specification
        'border-primary': 'rgba(55, 65, 81, 0.5)', // gray-700 with 50% opacity per ui.md
        'border-secondary': 'rgba(31, 41, 55, 0.5)', // gray-800 with 50% opacity per ui.md
        
        success: '#10b981',             // green-500
        warning: '#f59e0b',             // amber-500
        error: '#ef4444',               // red-500
        info: '#3b82f6',                // blue-500
        
        // Added badge colors
        'purple-300': '#d8b4fe',        // For album badge text
        'blue-300': '#93c5fd',          // For video/downloading badge text
        'green-300': '#86efac',         // For livestream/completed badge text
        'amber-300': '#fcd34d',         // For playlist/paused badge text
        'red-300': '#fca5a5',           // For error badge text
        'gray-300': '#d1d5db',          // For queued badge text
        'gray-500': '#6b7280',          // For queued badge bg/border
        'gray-700': '#374151',          // Used in border colors
        'gray-800': '#1f2937',          // Used in backgrounds
        'gray-900': '#111827',          // Used in backgrounds
        
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        'accent-blue': 'hsl(210, 90%, 50%)',
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        mono: ['monospace'], // As per spec for URLs/code
      },
      fontSize: {
        'h1-mobile': '3rem',
        'h1-desktop': '5rem',
        'h2': '1.5rem',
        'h3': '1.125rem',
        'body': '1rem',
        'small': '0.875rem',
        'xs': '0.75rem', // Extra Small Text
      },
      fontWeight: {
        'h1': '700',
        'h2': '700',
        'h3': '500',
        'body': '400',
        'small': '400',
        'xs': '400',
      },
      letterSpacing: {
        'h1': '-0.025em', // tracking-tight
      },
      borderRadius: {
        lg: "0.75rem", // rounded-xl from spec (was var(--radius))
        md: "0.5rem",  // rounded-lg (calc(var(--radius) - 2px))
        sm: "0.375rem", // rounded-md (calc(var(--radius) - 4px))
        xs: "0.25rem", // rounded-sm
        full: "9999px",
      },
      boxShadow: {
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)', // Default lg
        xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)', // Default xl
        'purple-glow-sm': '0 0 15px 5px rgba(168, 85, 247, 0.05)', // purple-900 at 5% for card hover
        'purple-glow-md': '0 0 20px 10px rgba(168, 85, 247, 0.2)', // purple-500 at 20% for button hover
        'glow': '0 0 10px 2px rgba(168, 85, 247, 0.15)', // Purple glow effect for hover states
        'focus-ring': '0 0 0 2px rgba(168, 85, 247, 0.3)', // Custom focus ring
        'card-hover': '0 8px 30px rgba(0, 0, 0, 0.12)', // Elevated card shadow
      },
      height: {
        '11': '2.75rem', // Primary button height
        '9': '2.25rem',  // Outline button height
        'h-10': '2.5rem', // Icon button height
        '64px': '64px',   // Header height
      },
      minHeight: {
        '150px': '150px',
      },
      maxWidth: {
        '36rem': '36rem',
        '5xl': '64rem', // Content max-width
        '1400px': '1400px',
        '500px': '500px',
        '350px': '350px',
      },
      animation: {
        fadeIn: 'fadeIn 500ms ease-in-out',
        fadeInFast: 'fadeIn 300ms ease-in-out',
        fadeOut: 'fadeOut 300ms ease-in-out',
        scaleIn: 'scaleIn 300ms ease-out',
        scaleOut: 'scaleOut 200ms ease-in',
        slideInDown: 'slideInDown 300ms ease-out',
        slideInUp: 'slideInUp 300ms ease-out',
        slideInLeft: 'slideInLeft 300ms ease-out',
        slideInRight: 'slideInRight 300ms ease-out',
        slideOutLeft: 'slideOutLeft 200ms ease-in',
        slideOutRight: 'slideOutRight 200ms ease-in',
        popIn: 'popIn 400ms cubic-bezier(0.68, -0.55, 0.27, 1.55)',
        wiggle: 'wiggle 300ms ease-in-out',
        progressGrow: 'progressGrow 800ms ease-out',
        pulse: 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        bounce: 'bounce 1s infinite',
        // For elements that need to expand/collapse
        collapsibleOpen: 'collapsibleOpen 300ms ease-out',
        collapsibleClose: 'collapsibleClose 200ms ease-in',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        scaleOut: {
          '0%': { transform: 'scale(1)', opacity: '1' },
          '100%': { transform: 'scale(0.95)', opacity: '0' },
        },
        slideInDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInLeft: {
          '0%': { transform: 'translateX(-10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideOutLeft: {
          '0%': { transform: 'translateX(0)', opacity: '1' },
          '100%': { transform: 'translateX(-10px)', opacity: '0' },
        },
        slideOutRight: {
          '0%': { transform: 'translateX(0)', opacity: '1' },
          '100%': { transform: 'translateX(10px)', opacity: '0' },
        },
        popIn: {
          '0%': { transform: 'scale(0.8)', opacity: '0' },
          '50%': { transform: 'scale(1.05)', opacity: '0.8' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        wiggle: {
          '0%, 100%': { transform: 'rotate(-2deg)' },
          '50%': { transform: 'rotate(2deg)' },
        },
        progressGrow: {
          '0%': { width: '0%' },
          '100%': { width: 'var(--progress-value, 100%)' },
        },
        pulse: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
        bounce: {
          '0%, 100%': { transform: 'translateY(-25%)', animationTimingFunction: 'cubic-bezier(0.8, 0, 1, 1)' },
          '50%': { transform: 'translateY(0)', animationTimingFunction: 'cubic-bezier(0, 0, 0.2, 1)' },
        },
        collapsibleOpen: {
          '0%': { maxHeight: '0px', opacity: '0' },
          '100%': { maxHeight: 'var(--radix-collapsible-content-height)', opacity: '1' },
        },
        collapsibleClose: {
          '0%': { maxHeight: 'var(--radix-collapsible-content-height)', opacity: '1' },
          '100%': { maxHeight: '0px', opacity: '0' },
        },
      },
      transitionProperty: {
        'height': 'height',
        'spacing': 'margin, padding',
        'width': 'width',
        'border': 'border-color, border-width',
        'background': 'background-color, background-image, background-position',
        'colors': 'color, background-color, border-color, text-decoration-color, fill, stroke',
      },
      transitionTimingFunction: {
        'bounce': 'cubic-bezier(0.68, -0.55, 0.27, 1.55)',
      },
      transitionDuration: {
        '250': '250ms',
        '350': '350ms',
        '400': '400ms',
      },
      scale: {
        '98': '0.98',
        '102': '1.02',
      },
    },
  },
  plugins: [
    require('tailwindcss-animate') // For Radix UI animations if not already included by shadcn/ui base
  ],
} 