import AppLayout from '@/layouts/app-layout';
import { projectsList } from '@/routes';
import type { BreadcrumbItem, Project } from '@/types';
import { Head } from '@inertiajs/react';

import { usePage } from '@inertiajs/react';

const ViewProjectBreadcrumbs = () => {
    const { props } = usePage();
    const project = props.project as Project;

    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsList().url,
        },
        {
            title: project?.name || 'Project',
            href: '',
        },
    ];

    return breadcrumbs;
};

export default function ViewProject() {
    const breadcrumbs = ViewProjectBreadcrumbs();

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Project" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4"></div>
        </AppLayout>
    );
}
