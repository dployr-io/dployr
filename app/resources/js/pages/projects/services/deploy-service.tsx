import AppLayout from '@/layouts/app-layout';
import { projectsList } from '@/routes';
import type { BreadcrumbItem, Project } from '@/types';
import { Head, usePage } from '@inertiajs/react';


const ViewProjectBreadcrumbs = (project: Project) => {
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




export default function DeployService() {
    const { props } = usePage();
    const project = props.project as Project;
    const breadcrumbs = ViewProjectBreadcrumbs(project);



    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Project" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex items-start justify-between">
                        <div className="flex flex-col gap-1">
                            <p className="text-2xl font-black">New Service</p>
                            <p className="text-sm font-normal text-muted-foreground">Choose a new service</p>
                        </div>    
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
