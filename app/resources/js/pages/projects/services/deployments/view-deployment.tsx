import BlueprintCard from '@/components/blueprint-card';

import AppLayout from '@/layouts/app-layout';
import { deploymentsShow } from '@/routes';
import type { Blueprint, BreadcrumbItem } from '@/types';
import { Head, usePage } from '@inertiajs/react';

const ViewProjectBreadcrumbs = (blueprint: Blueprint) => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Deployments',
            href: deploymentsShow(blueprint.id).url,
        },
        {
            title: blueprint?.config.name || 'Service',
            href: '',
        },
    ];

    return breadcrumbs;
};

export default function ViewDeployment() {
    const { props } = usePage();
    const blueprint = props.blueprint as Blueprint;
    const breadcrumbs = ViewProjectBreadcrumbs(blueprint);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-2xl font-black">Deployment 101</p>
                        <p className="text-sm font-normal text-muted-foreground"> This is a single deployment</p>
                    </div>

                    <div className="grid w-full grid-cols-3 gap-3">
                        <BlueprintCard key={blueprint.id} id={blueprint.id} config={blueprint.config} status={blueprint.status} />
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
