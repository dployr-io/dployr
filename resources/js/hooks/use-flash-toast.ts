/**
 * useFlashToast makes it easy to retrieve session messages on the frontend
 *
 * @see app/Http/Middleware/HandleInertiaRequests.php
 */
import { toast } from '@/lib/toast';
import { usePage } from '@inertiajs/react';
import { useEffect } from 'react';

export function useFlashToast() {
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
                    toast.info(message);
                    break;
                case 'status':
                    toast.info(message);
                    break;
            }
        });
    }, [flash]);

    return { type: flash?.type };
}
