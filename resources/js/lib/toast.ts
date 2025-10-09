import { toast as sonnerToast } from 'sonner';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastOptions {
    duration?: number;
    position?: 'top-left' | 'top-right' | 'top-center' | 'bottom-left' | 'bottom-right' | 'bottom-center';
    description?: string;
    action?: {
        label: string;
        onClick: () => void;
    };
    persist?: boolean;
}

const defaultOptions: ToastOptions = {
    duration: 4000,
    position: 'bottom-center',
};

const toastStyle = {
    fontFamily: 'Instrument Sans, ui-sans-serif',
    fontSize: '12px',
    fontWeight: 700,
    borderRadius: '8px',
    padding: '12px 16px',
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
};

export const toast = {
    success: (message: string, options?: ToastOptions) => {
        if (options?.persist) {
            const params = new URLSearchParams(window.location.search);
            params.set('success', message);
            if (options.description) params.set('success_description', options.description);
            window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`);
        }
        sonnerToast.success(message, {
            ...defaultOptions,
            ...options,
            className: 'toast-success',
            style: {
                ...toastStyle,
                background: 'var(--success-bg, #f0fdf4)',
                color: 'var(--success-fg, #166534)',
                border: '1px solid var(--success-border, #bbf7d0)',
            },
        });
    },
    error: (message: string, options?: ToastOptions) => {
        if (options?.persist) {
            const params = new URLSearchParams(window.location.search);
            params.set('error', message);
            if (options.description) params.set('error_description', options.description);
            window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`);
        }
        sonnerToast.error(message, {
            ...defaultOptions,
            ...options,
            className: 'toast-error',
            style: {
                ...toastStyle,
                background: 'var(--error-bg, #fef2f2)',
                color: 'var(--error-fg, #991b1b)',
                border: '1px solid var(--error-border, #fecaca)',
            },
        });
    },
    warning: (message: string, options?: ToastOptions) => {
        sonnerToast.warning(message, {
            ...defaultOptions,
            ...options,
            className: 'toast-warning',
            style: {
                ...toastStyle,
                background: 'var(--warning-bg, #fffbeb)',
                color: 'var(--warning-fg, #92400e)',
                border: '1px solid var(--warning-border, #fef3c7)',
            },
        });
    },
    info: (message: string, options?: ToastOptions) => {
        sonnerToast.info(message, {
            ...defaultOptions,
            ...options,
            className: 'toast-info',
            style: {
                ...toastStyle,
                background: 'var(--info-bg, #eff6ff)',
                color: 'var(--info-fg, #1e40af)',
                border: '1px solid var(--info-border, #bfdbfe)',
            },
        });
    },
    custom: (message: string, type: ToastType, options?: ToastOptions) => {
        sonnerToast[type](message, {
            ...defaultOptions,
            ...options,
            className: `toast-${type}`,
            style: {
                ...toastStyle,
                background: `var(--${type}-bg)`,
                color: `var(--${type}-fg)`,
                border: `1px solid var(--${type}-border)`,
            },
        });
    },
};
