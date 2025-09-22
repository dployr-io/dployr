import type { Project } from '@/types';
import { router } from '@inertiajs/react';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { z } from 'zod';

export function useProjects() {
    const [error, setError] = useState<string>('');
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');

    const formSchema = z.object({
        name: z.string().min(3, 'Name with a minimum of three (3) characters is required'),
        description: z.string().optional(),
    });

    const validateForm = () => {
        const result = formSchema.safeParse({ name, description });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            setError(fieldErrors.name?.[0] || fieldErrors.description?.[0] || 'Validation failed');
            return false;
        }

        setError('');
        return true;
    };

    const getFormData = () => {
        if (!validateForm()) return {};

        return { name, description };
    };

    const projects = useQuery<Project[]>({
        queryKey: ['projects'],
        queryFn: () =>
            new Promise((resolve) => {
                router.get(
                    '/projects',
                    {},
                    {
                        onSuccess: (page) => resolve(page.props.projects as Project[]),
                    },
                );
            }),
        staleTime: 60 * 1000, // Every minute
    });

    return {
        projects,
        name,
        description,
        validationError: error,

        setName,
        setDescription,
        getFormData,
    };
}
