"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

// Define our own interface instead of importing from next-themes/dist/types
interface ThemeProviderProps {
  children: React.ReactNode;
  defaultTheme?: string;
  storageKey?: string;
  // Use Record<string, unknown> instead of any
  [key: string]: React.ReactNode | string | undefined | Record<string, unknown>;
}

export function ThemeProvider({ children, ...props }: ThemeProviderProps) {
  React.useEffect(() => {
    // Force dark theme by adding the 'dark' class to the html element
    document.documentElement.classList.add('dark');
  }, []);
  
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>
} 