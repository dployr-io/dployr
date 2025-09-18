import { usePage } from '@inertiajs/react';
import { useEffect } from 'react';
import { toast } from '@/lib/toast';
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Merge Tailwind and custom class names.
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 * Hook to display flash messages as toast notifications.
 */
export function getFlashToast() {
  const { flash } = usePage().props as { flash?: Record<string, string> };

  useEffect(() => {
    if (!flash) return;

    Object.entries(flash).forEach(([type, message]) => {
      if (!message) return;

      switch (type) {
        case 'success':
          toast.success(message);
          break;
        case 'error':
          toast.error(message);
          break;
        case 'warning':
          toast.warning(message);
          break;
        case 'info':
        case 'status':
        default:
          toast.info(message);
      }
    });
  }, [flash]);
}

/**
 * Convert a string to uppercase words, replacing underscores with spaces.
 */
export function toWordUpperCase(value: string) {
  return value
    .replace(/_/g, ' ')
    .split(' ')
    .map(word => word.toUpperCase())
    .join(' ');
}
