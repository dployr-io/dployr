import type { Blueprint } from '@/types';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

export function useDeployments() {
    const { data, isLoading } = useQuery<Blueprint[]>({
        queryKey: ['deployments'],
        queryFn: async () => {
            const response = await axios.get('/deployments/fetch');
            return response.data;
        },
        staleTime: 5 * 60 * 1000,
    });

    return {
        deployments: data,
        isLoading,
    };
}
