import type { Remote } from '@/types';
import { router } from '@inertiajs/react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { z } from 'zod';

export function useRemotes(setOpen?: (open: boolean) => void) {
    const [branches, setBranches] = useState<string[]>([]);
    const [searchComplete, setSearchComplete] = useState(false);
    const [error, setError] = useState<string>('');
    const [remoteRepo, setRemoteRepo] = useState('');
    const [selectedBranch, setSelectedBranch] = useState('');
    const queryClient = useQueryClient();

    const formSchema = z
        .object({
            remote_repo: z
                .string()
                .min(1, 'Domain is required')
                .regex(/^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/, 'Please enter a valid remote address'),
            branch: z.string().optional(),
        })
        .refine((data) => {
            if (branches.length > 0 && !data.branch) {
                return false;
            }
            return true;
        });

    const validateForm = () => {
        const result = formSchema.safeParse({
            remote_repo: remoteRepo,
            branch: selectedBranch,
        });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            setError(fieldErrors.remote_repo?.[0] || fieldErrors.branch?.[0] || 'Validation failed');
            return false;
        }

        setError('');
        return true;
    };

    const getFormAction = () => {
        return searchComplete ? '/resources/remotes' : '/resources/remotes/search';
    };

    const getFormData = () => {
        if (!validateForm()) return {};

        return searchComplete ? { remote_repo: remoteRepo, branch: selectedBranch } : { remote_repo: remoteRepo };
    };

    const handleFormSuccess = (page: any) => {
        const data = page?.props?.flash?.data ?? [];
        if (!searchComplete && Array.isArray(data) && data.length > 0) {
            setBranches(data);
            setSearchComplete(true);
        } else if (searchComplete) {
            queryClient.invalidateQueries({ queryKey: ['remotes'] });
            setOpen!(false);
        }
    };

    const reset = () => {
        setRemoteRepo('');
        setSelectedBranch('');
        setBranches([]);
        setSearchComplete(false);
        setError('');
    };

    const remotes = useQuery<Remote[]>({
        queryKey: ['remotes'],
        queryFn: () =>
            new Promise((resolve) => {
                router.get(
                    '/resources/remotes',
                    {},
                    {
                        onSuccess: (page) => resolve(page.props.remotes as Remote[]),
                    },
                );
            }),
        staleTime: 60 * 1000, // Every minute
    });

    return {
        // State
        remotes,
        branches,
        searchComplete,
        validationError: error,
        remoteRepo,
        selectedBranch,

        // Actions
        setRemoteRepo,
        setSelectedBranch,
        getFormAction,
        getFormData,
        handleFormSuccess,
    };
}
