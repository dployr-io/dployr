import SpecCard from '@/components/spec-card';
import { Skeleton } from '@/components/ui/skeleton';
import { useDeployments } from '@/hooks/use-deployments';
import AppLayout from '@/layouts/app-layout';
import { specsIndex } from '@/routes';
import type { Blueprint, BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Specs',
        href: specsIndex().url,
    },
];

export default function Specs() {
    const { normalizedDeployments, isLoading } = useDeployments();

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-2xl font-black">Specs</p>
                        <p className="text-sm font-normal text-muted-foreground">Manage your saved deployments and redeploy them anytime</p>
                    </div>

                    <div className="grid w-full grid-cols-3 gap-3">
                        {isLoading ? (
                            <>
                                <div className="flex flex-col gap-2 rounded-xl border border-sidebar-border/70 p-4 dark:border-sidebar-border">
                                    <div className="mb-2 flex items-center gap-2">
                                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
                                            <Skeleton className="h-8 w-8 rounded-full bg-muted-foreground/20" />
                                        </div>
                                        <div className="flex-1">
                                            <Skeleton className="h-4 w-24 rounded bg-muted-foreground/20" />
                                        </div>
                                    </div>
                                    <Skeleton className="mb-1 h-3 w-32 rounded bg-muted-foreground/20" />
                                    <Skeleton className="h-3 w-20 rounded bg-muted-foreground/20" />
                                </div>
                            </>
                        ) : (
                            <>
                                {normalizedDeployments?.map((blueprint: Blueprint) => (
                                    <SpecCard key={blueprint.id} blueprint={blueprint} />
                                ))}
                            </>
                        )}
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
