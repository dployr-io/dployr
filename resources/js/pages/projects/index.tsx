import ProjectCard from '@/components/project-card';
import ProjectCreateDialog from '@/components/project-create-dialog';
import { Skeleton } from '@/components/ui/skeleton';

import { useProjects } from '@/hooks/use-projects';
import AppLayout from '@/layouts/app-layout';
import { projectsIndex } from '@/routes';
import type { BreadcrumbItem, Project } from '@/types';
import { Head } from '@inertiajs/react';
import { PlusCircle } from 'lucide-react';
import { useState } from 'react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Projects',
        href: projectsIndex().url,
    },
];
export default function Dashboard() {
    const { projects, isLoading } = useProjects();
    const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <p className="text-3xl font-black">Overview</p>

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
                                {projects?.map((project: Project) => (
                                    <ProjectCard key={project.id} project={project} />
                                ))}
                            </>
                        )}
                        <div
                            className="flex flex-col gap-2 rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-accent-foreground md:min-h-min dark:border-sidebar-border dark:hover:border-muted-foreground"
                            onClick={() => setIsDialogOpen(true)}
                        >
                            <div className="flex items-center gap-2">
                                <PlusCircle size={20} className="text-muted-foreground" />
                                <p>Create a New Project</p>
                            </div>

                            <p className="text-sm text-muted-foreground">Click to create a new project and get started with deploying with dployr</p>
                        </div>
                    </div>

                    <ProjectCreateDialog open={isDialogOpen} setOpen={(open) => setIsDialogOpen(open)} />
                </div>
            </div>
        </AppLayout>
    );
}
