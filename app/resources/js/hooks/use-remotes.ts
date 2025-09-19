import type { Project, Remote } from '@/types';
import { router } from '@inertiajs/react';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { z } from 'zod';

export function useRemotes(onSearchCallback?: () => void | null) {
    const [branches, setBranches] = useState<string[]>([]);
    const [searchComplete, setSearchComplete] = useState(false);
    const [error, setError] = useState<string>('');
    const [remoteRepo, setRemoteRepo] = useState('');
    const [selectedBranch, setSelectedBranch] = useState('');

    const formSchema = z
        .object({
            remote_repo: z.string().min(5, 'Repository URL is required'),
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
        return searchComplete ? '/projects/remotes' : '/projects/remotes/search';
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
            if (onSearchCallback) {
                onSearchCallback();
            }
        }
    };

    const reset = () => {
        setRemoteRepo('');
        setSelectedBranch('');
        setBranches([]);
        setSearchComplete(false);
        setError('');
    };

    const projects = useQuery<Remote[]>({
        queryKey: ['remotes'],
        queryFn: () =>
            new Promise((resolve) => {
                router.get(
                    '/projects/remotes',
                    {},
                    {
                        onSuccess: (page) => resolve(page.props.remotes as Remote[]),
                    },
                );
            }),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });

    return {
        // State
        projects,
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
