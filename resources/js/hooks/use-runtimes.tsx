import type { Runtime } from '@/types/runtimes';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

export function useRuntimes(params?: { runtime?: Runtime }) {
    const query = useQuery<string[]>({
        queryKey: ['runtimes', params],
        queryFn: async () => {
            const response = await axios.get(`/runtimes`, { params: { runtime: params } });
            return response.data;
        },
        staleTime: 5 * 60 * 1000,
    });

    return {
        runtimes: query.data,
        isLoading: query.isLoading,
    };
}
