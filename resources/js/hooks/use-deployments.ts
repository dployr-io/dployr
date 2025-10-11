import { useRemotes } from '@/hooks/use-remotes';
import type { Blueprint, Service } from '@/types';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { useMemo, useState } from 'react';

export function useDeployments() {
    const params = new URLSearchParams(window.location.search);
    const spec = params.get('spec');

    const { data: deployments, isLoading } = useQuery<Blueprint[]>({
        queryKey: ['deployments', spec],
        queryFn: async () => {
            try {
                const response = await axios.get('/deployments/fetch', {
                    params: Object.fromEntries(params),
                });
                return response.data;
            } catch (error) {}
        },
        staleTime: 5 * 60 * 1000,
    });

    const { remotes } = useRemotes();
    const remotesData = remotes || [];
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;
    const totalPages = Math.ceil((deployments?.length ?? 0) / itemsPerPage);
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedDeployments = deployments?.slice(startIndex, endIndex);

    const normalizedDeployments = useMemo(
        () =>
            (paginatedDeployments || []).map((deployment) => {
                let config = deployment?.config ?? {};
                config = JSON.parse(config as string);
                const remote = (config as Partial<Service>)?.remote;
                let resolvedRemote = remote;
                if (remote) {
                    resolvedRemote = remotesData.find((r) => r) || remote;
                }
                return { ...deployment, config: { ...config, remote: resolvedRemote } };
            }),
        [paginatedDeployments, remotesData],
    );

    const goToPage = (page: number) => {
        setCurrentPage(Math.max(1, Math.min(page, totalPages)));
    };

    const goToPreviousPage = () => {
        setCurrentPage((prev) => Math.max(1, prev - 1));
    };

    const goToNextPage = () => {
        setCurrentPage((prev) => Math.min(totalPages, prev + 1));
    };

    return {
        deployments,
        normalizedDeployments,
        currentPage,
        totalPages,
        startIndex,
        endIndex,
        isLoading,

        goToPage,
        goToNextPage,
        goToPreviousPage,
    };
}
