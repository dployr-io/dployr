import type { Service } from '@/types';
import { router } from '@inertiajs/react';
import { useQuery } from '@tanstack/react-query';

export function useServices() {
    const getServices = (id: string) =>
        useQuery<Service[]>({
            queryKey: ['services'],
            queryFn: () =>
                new Promise((resolve, reject) => {
                    router.get(
                        `/projects/${id}`,
                        {},
                        {
                            onSuccess: (page) => resolve(page.props.services as Service[]),
                            onError: (errors) => reject(errors),
                        },
                    );
                }),
            staleTime: 5 * 60 * 1000,
        });

    return {
        getServices,
    };
}
