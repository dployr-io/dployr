import BlueprintCard from '@/components/blueprint-card';

import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { deploymentsList } from '@/routes';
import type { BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Deployments',
        href: deploymentsList().url,
    },
];
export default function Deployments() {
    const { deployments } = useServices();

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-2xl font-black">Deployments</p>
                        <p className="text-sm font-normal text-muted-foreground"> Your deployments at a glance</p>
                    </div>

                    <div className="grid w-full grid-cols-3 gap-3">
                        {deployments?.data?.map((service) => (
                            <BlueprintCard key={service.id} id={service.id} config={service.config} status={service.status} />
                        ))}
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
