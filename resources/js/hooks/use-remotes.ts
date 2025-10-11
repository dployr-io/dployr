import type { Remote } from '@/types';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
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
                .regex(/^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/, 'Please enter a valid remote address'),
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

    const handleFormSuccess = (page: unknown) => {
        const pageData = page as { props?: { flash?: { data?: string[] } } };
        const data = pageData?.props?.flash?.data ?? [];
        if (!searchComplete && Array.isArray(data) && data.length > 0) {
            setBranches(data);
            setSearchComplete(true);
        } else if (searchComplete) {
            queryClient.invalidateQueries({ queryKey: ['remotes'] });
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
        remotes: data || [],
        isLoading,
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
