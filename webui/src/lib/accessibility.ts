/**
 * Accessibility utilities and helpers
 * 
 * This file contains utilities to improve application accessibility
 * including keyboard navigation, screen reader support, and focus management.
 */

// Key codes for keyboard navigation
export const KEYS = {
  TAB: 'Tab',
  ENTER: 'Enter',
  SPACE: ' ',
  ESCAPE: 'Escape',
  ARROW_UP: 'ArrowUp',
  ARROW_DOWN: 'ArrowDown',
  ARROW_LEFT: 'ArrowLeft',
  ARROW_RIGHT: 'ArrowRight',
  HOME: 'Home',
  END: 'End',
  PAGE_UP: 'PageUp',
  PAGE_DOWN: 'PageDown',
};

// Common ARIA roles
export const ROLES = {
  BUTTON: 'button',
  CHECKBOX: 'checkbox',
  DIALOG: 'dialog',
  HEADING: 'heading',
  LINK: 'link',
  LISTBOX: 'listbox',
  OPTION: 'option',
  MENU: 'menu',
  MENUITEM: 'menuitem',
  PROGRESSBAR: 'progressbar',
  RADIO: 'radio',
  SWITCH: 'switch',
  TAB: 'tab',
  TABLIST: 'tablist',
  TABPANEL: 'tabpanel',
  TEXTBOX: 'textbox',
};

// ARIA live region properties
export const ARIA_LIVE = {
  OFF: 'off',
  POLITE: 'polite',
  ASSERTIVE: 'assertive',
};

// Focus related utility classes
export const FOCUS_STYLES = {
  // High contrast focus ring for keyboard navigation
  VISIBLE: 'focus-visible:ring-2 focus-visible:ring-primary-accent focus-visible:ring-offset-2 focus-visible:outline-none',
  // Custom focus style for specific elements
  CARD: 'focus-visible:ring-2 focus-visible:ring-primary-accent focus-visible:outline-none',
  // Focus style for inputs
  INPUT: 'focus-visible:ring-2 focus-visible:ring-primary-accent focus-visible:border-primary-accent focus-visible:outline-none',
  // Focus style for buttons
  BUTTON: 'focus-visible:ring-2 focus-visible:ring-primary-accent/70 focus-visible:ring-offset-2 focus-visible:outline-none',
};

// Skip link class for keyboard users to bypass navigation
export const SKIP_TO_CONTENT_CLASS = "sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:p-4 focus:bg-card-background focus:text-text-primary focus:rounded-md focus:shadow-lg focus:border focus:border-border-primary";

// Helper to handle keyboard navigation for custom components
export const handleKeyboardNavigation = (
  event: React.KeyboardEvent,
  options: {
    onEnter?: () => void;
    onSpace?: () => void;
    onEscape?: () => void;
    onArrowUp?: () => void;
    onArrowDown?: () => void;
    onArrowLeft?: () => void;
    onArrowRight?: () => void;
    onHome?: () => void;
    onEnd?: () => void;
    onTab?: () => void;
  }
) => {
  const { key } = event;

  switch (key) {
    case KEYS.ENTER:
      options.onEnter?.();
      break;
    case KEYS.SPACE:
      options.onSpace?.();
      break;
    case KEYS.ESCAPE:
      options.onEscape?.();
      break;
    case KEYS.ARROW_UP:
      options.onArrowUp?.();
      break;
    case KEYS.ARROW_DOWN:
      options.onArrowDown?.();
      break;
    case KEYS.ARROW_LEFT:
      options.onArrowLeft?.();
      break;
    case KEYS.ARROW_RIGHT:
      options.onArrowRight?.();
      break;
    case KEYS.HOME:
      options.onHome?.();
      break;
    case KEYS.END:
      options.onEnd?.();
      break;
    case KEYS.TAB:
      options.onTab?.();
      break;
    default:
      return;
  }
};

// Helper to create accessible status announcements
export const createStatusAnnouncement = (message: string, level: 'polite' | 'assertive' = 'polite') => {
  // Create or get the status announcement element
  let statusElement = document.getElementById('status-announcer');
  
  if (!statusElement) {
    statusElement = document.createElement('div');
    statusElement.id = 'status-announcer';
    statusElement.className = 'sr-only';
    statusElement.setAttribute('aria-live', level);
    statusElement.setAttribute('aria-atomic', 'true');
    document.body.appendChild(statusElement);
  }
  
  // Set the message
  statusElement.textContent = message;
  
  // Clear the message after a delay to prevent screen readers from repeating it
  setTimeout(() => {
    statusElement.textContent = '';
  }, 5000);
};

// Color contrast checker
export const hasGoodContrast = (foreground: string, background: string): boolean => {
  // This is a simplified implementation
  // For a production app, use a proper color contrast library
  
  // Convert hex colors to RGB
  const hexToRgb = (hex: string): { r: number; g: number; b: number } => {
    // Remove # if present
    hex = hex.replace('#', '');
    
    // Parse hex
    const r = parseInt(hex.substring(0, 2), 16);
    const g = parseInt(hex.substring(2, 4), 16);
    const b = parseInt(hex.substring(4, 6), 16);
    
    return { r, g, b };
  };
  
  // Calculate relative luminance
  const calculateLuminance = (color: { r: number; g: number; b: number }): number => {
    const { r, g, b } = color;
    
    // Normalize RGB values
    const normalizedR = r / 255;
    const normalizedG = g / 255;
    const normalizedB = b / 255;
    
    // Calculate RGB values
    const rValue = normalizedR <= 0.03928 
      ? normalizedR / 12.92 
      : Math.pow((normalizedR + 0.055) / 1.055, 2.4);
    
    const gValue = normalizedG <= 0.03928 
      ? normalizedG / 12.92 
      : Math.pow((normalizedG + 0.055) / 1.055, 2.4);
    
    const bValue = normalizedB <= 0.03928 
      ? normalizedB / 12.92 
      : Math.pow((normalizedB + 0.055) / 1.055, 2.4);
    
    // Calculate luminance
    return 0.2126 * rValue + 0.7152 * gValue + 0.0722 * bValue;
  };
  
  // Calculate contrast ratio
  const calculateContrastRatio = (
    luminance1: number,
    luminance2: number
  ): number => {
    const lighter = Math.max(luminance1, luminance2);
    const darker = Math.min(luminance1, luminance2);
    
    return (lighter + 0.05) / (darker + 0.05);
  };
  
  // Process colors
  const foregroundRgb = hexToRgb(foreground);
  const backgroundRgb = hexToRgb(background);
  
  const foregroundLuminance = calculateLuminance(foregroundRgb);
  const backgroundLuminance = calculateLuminance(backgroundRgb);
  
  const contrastRatio = calculateContrastRatio(
    foregroundLuminance,
    backgroundLuminance
  );
  
  // WCAG 2.0 level AA requires a contrast ratio of at least 4.5:1 for normal text
  // and 3:1 for large text
  return contrastRatio >= 4.5;
};

// Generate aria-label helper
export const createAriaLabel = (description: string, state?: string): string => {
  if (state) {
    return `${description}, ${state}`;
  }
  return description;
};

// Touch target size checker
export const hasSufficientTouchTarget = (
  width: number,
  height: number
): boolean => {
  // WCAG recommends touch targets to be at least 44x44 pixels
  return width >= 44 && height >= 44;
};

// Keyboard only user detection
export const setupKeyboardUserDetection = (): void => {
  window.addEventListener('keydown', handleFirstTab);
  
  function handleFirstTab(e: KeyboardEvent) {
    if (e.key === 'Tab') {
      document.body.classList.add('user-is-tabbing');
      window.removeEventListener('keydown', handleFirstTab);
    }
  }
};

// Adds necessary classes for keyboard users
export const applyKeyboardUserStyles = (): void => {
  const styleElement = document.createElement('style');
  styleElement.textContent = `
    .user-is-tabbing :focus:not(.focus-visible) {
      outline: none;
      box-shadow: none;
    }
    
    .user-is-tabbing .focus-visible, 
    .user-is-tabbing :focus-visible {
      outline: 2px solid #a855f7;
      outline-offset: 2px;
    }
  `;
  document.head.appendChild(styleElement);
};

// Export all for easy importing
export default {
  KEYS,
  ROLES,
  ARIA_LIVE,
  FOCUS_STYLES,
  SKIP_TO_CONTENT_CLASS,
  handleKeyboardNavigation,
  createStatusAnnouncement,
  hasGoodContrast,
  createAriaLabel,
  hasSufficientTouchTarget,
  setupKeyboardUserDetection,
  applyKeyboardUserStyles,
}; 