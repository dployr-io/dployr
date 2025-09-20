import { runtimes } from '@/types/runtimes';
import type { Runtime, Service } from '@/types';
import { router } from '@inertiajs/react';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { z } from 'zod';

export function useServices(onCreateServiceCallback?: () => void | null) {
    const [error, setError] = useState<string>('');
    const [name, setName] = useState('');
    const [rootDir, setRootDir] = useState('');
    const [runtime, setRuntime] = useState<Runtime>('node-js');
    const [buildCommand, setBuildCommad] = useState('');

    const formSchema = z.object({
        name: z.string().min(3, 'Name with a minimum of three (3) characters is required'),
        rootDir: z.string().optional(),
        runtime: z.enum(runtimes),
    });

    const validateForm = () => {
        const result = formSchema.safeParse({ name, rootDir });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            setError(fieldErrors.name?.[0] || fieldErrors.rootDir?.[0] || 'Validation failed');
            return false;
        }

        setError('');
        return true;
    };

    const getFormData = () => {
        if (!validateForm()) return {};

        return { name, rootDir };
    };

    const handleFormSuccess = () => onCreateServiceCallback!();

    const services = useQuery<Service[]>({
        queryKey: ['projects'],
        queryFn: () =>
            new Promise((resolve) => {
                router.get(
                    '/projects',
                    {},
                    {
                        onSuccess: (page) => resolve(page.props.services as Service[]),
                    },
                );
            }),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });

    return {
        name,
        rootDir,
        error,
        runtime,
        buildCommand,

        setName,
        setRootDir,
        setRuntime,
        setBuildCommad,
        getFormData,
        handleFormSuccess,
    };
}
