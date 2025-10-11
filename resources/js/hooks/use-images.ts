import type { Remote } from '@/types';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { useState } from 'react';
import { z } from 'zod';

export function useImages(setOpen?: (open: boolean) => void) {
    const [searchComplete, setSearchComplete] = useState(false);
    const [error, setError] = useState<string>('');
    const [imageRegistry, setImageRegistry] = useState('');
    const queryClient = useQueryClient();

    const formSchema = z.object({
        image_registry: z
            .string()
            .min(1, 'Domain is required')
            .regex(/^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/, 'Please enter a valid docker image registry'),
        branch: z.string().optional(),
    });

    const validateForm = () => {
        const result = formSchema.safeParse({
            image_registry: imageRegistry,
        });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            setError(fieldErrors.image_registry?.[0] || 'Validation failed');
            return false;
        }

        setError('');
        return true;
    };

    const getFormAction = () => {
        return searchComplete ? '/resources/images' : '/resources/images/search';
    };

    const getFormData = () => {
        if (!validateForm()) return {};

        return { image_registry: imageRegistry };
    };

    const handleFormSuccess = (page: unknown) => {
        const pageData = page as { props?: { flash?: { data?: unknown[] } } };
        const data = pageData?.props?.flash?.data ?? [];
        if (!searchComplete && Array.isArray(data) && data.length > 0) {
            setSearchComplete(true);
        } else if (searchComplete) {
            queryClient.invalidateQueries({ queryKey: ['images'] });
            setOpen!(false);
        }
    };

    const { data, isLoading } = useQuery<Remote[]>({
        queryKey: ['remotes'],
        queryFn: async () => {
            const response = await axios.get('/resources/remotes/fetch');
            return response.data;
        },
        staleTime: 60 * 1000, // Every minute
    });

    return {
        // State
        images: data,
        isLoading,
        searchComplete,
        validationError: error,
        imageRegistry,

        // Actions
        setImageRegistry,
        getFormAction,
        getFormData,
        handleFormSuccess,
    };
}
