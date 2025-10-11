import type { Service } from '@/types';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

export function useServices(projectId?: string) {
    const query = useQuery<Service[]>({
        queryKey: ['services', projectId],
        queryFn: async () => {
            const response = await axios.get(`/projects/${projectId}/services/fetch`);
            return response.data;
        },
        enabled: !!projectId,
        staleTime: 5 * 60 * 1000,
    });

    return {
        services: query.data,
        isLoading: query.isLoading,
    };
}
