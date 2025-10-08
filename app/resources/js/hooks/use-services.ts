import type { Service } from '@/types';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

export function useServices() {
    const getServices = (projectId: string) =>
        useQuery<Service[]>({
            queryKey: ['services'],
            queryFn: async () => {
                const response = await axios.get(`/projects/${projectId}/fetch`);
                return response.data;
            },

            staleTime: 5 * 60 * 1000,
        });

    return {
        getServices,
    };
}
