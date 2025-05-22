/**
 * Animation utility classes and constants
 * These can be reused throughout the application for consistent animations
 */

// Transition durations (in ms)
export const DURATIONS = {
  fast: 150,
  normal: 250,
  slow: 350,
  extraSlow: 500,
};

// Easing functions
export const EASINGS = {
  default: 'cubic-bezier(0.4, 0, 0.2, 1)', // Equivalent to 'ease'
  in: 'cubic-bezier(0.4, 0, 1, 1)',
  out: 'cubic-bezier(0, 0, 0.2, 1)',
  inOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
  bounce: 'cubic-bezier(0.68, -0.55, 0.27, 1.55)',
};

// Transform origin options
export const ORIGINS = {
  center: 'center',
  top: 'top',
  bottom: 'bottom',
  left: 'left',
  right: 'right',
  topLeft: 'top left',
  topRight: 'top right',
  bottomLeft: 'bottom left',
  bottomRight: 'bottom right',
};

// Common animation variants
export const VARIANTS = {
  // Fade animations
  fadeIn: 'animate-fadeIn',
  fadeInFast: 'animate-fadeInFast',
  fadeOut: 'animate-fadeOut',
  
  // Scale animations
  scaleIn: 'animate-scaleIn',
  scaleOut: 'animate-scaleOut',
  
  // Slide animations
  slideInDown: 'animate-slideInDown',
  slideInUp: 'animate-slideInUp',
  slideInLeft: 'animate-slideInLeft',
  slideInRight: 'animate-slideInRight',
  
  // Combination animations
  popIn: 'animate-popIn',
  
  // Micro-interactions
  pulse: 'animate-pulse',
  bounce: 'animate-bounce',
  wiggle: 'animate-wiggle',
  spin: 'animate-spin',
};

// Hover and active state transitions
export const STATE_TRANSITIONS = {
  button: 'transition-all duration-200 ease-out',
  input: 'transition-all duration-200 ease-out',
  card: 'transition-all duration-200 ease-out',
  link: 'transition-colors duration-150 ease-in-out',
  opacity: 'transition-opacity duration-150 ease-in-out',
  transform: 'transition-transform duration-200 ease-out',
  shadow: 'transition-shadow duration-200 ease-out',
  border: 'transition-border duration-150 ease-in',
  background: 'transition-background duration-200 ease-out',
  all: 'transition-all duration-250 ease-out',
};

// Specific hover effects
export const HOVER_EFFECTS = {
  scale: 'hover:scale-105',
  scaleDown: 'hover:scale-98',
  brighten: 'hover:brightness-110',
  darken: 'hover:brightness-90',
  elevate: 'hover:shadow-md',
  glow: 'hover:shadow-glow',
};

// Focus effects
export const FOCUS_EFFECTS = {
  ring: 'focus:ring-2 focus:ring-primary-accent/50 focus:outline-none',
  ringOffset: 'focus:ring-2 focus:ring-primary-accent/50 focus:ring-offset-2 focus:outline-none',
  outline: 'focus:outline-2 focus:outline-primary-accent focus:outline-offset-2',
};

// Active effects
export const ACTIVE_EFFECTS = {
  scale: 'active:scale-95',
  shadow: 'active:shadow-inner',
  darken: 'active:brightness-90',
};

// Combined states for different element types
export const INTERACTIVE = {
  button: `${STATE_TRANSITIONS.button} hover:brightness-110 active:scale-95 focus:outline-none focus:ring-2 focus:ring-primary-accent/50`,
  link: `${STATE_TRANSITIONS.link} hover:text-primary-accent focus:outline-none focus:underline`,
  card: `${STATE_TRANSITIONS.card} hover:border-border-secondary hover:shadow-md`,
  iconButton: `${STATE_TRANSITIONS.all} hover:text-primary-accent active:scale-90 focus:outline-none`,
}; 