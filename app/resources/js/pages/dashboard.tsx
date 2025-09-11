import { Badge } from '@/components/ui/badge';
import AppLayout from '@/layouts/app-layout';
import { dashboard } from '@/routes';
import { type BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';
import { FaGitlab } from 'react-icons/fa6';
import { RxGithubLogo } from 'react-icons/rx';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Dashboard',
        href: dashboard().url,
    },
];

interface Props {
    projects: Project[];
}

interface Project {
    id: string;
    name: string;
    remoteRepo: string;
    lastCommitMessage: string;
}

export default function Dashboard({ projects }: Props) {
    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Dashboard" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="grid h-full auto-rows-min gap-4 p-8 md:grid-cols-3">
                    <div className="flex flex-col gap-6">
                        <p className="text-3xl font-black">Overview</p>

                        {projects?.map((project) => (
                            <div
                                key={project.id}
                                className="rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-accent-foreground md:min-h-min dark:border-sidebar-border"
                            >
                                <div className="flex gap-2">
                                    <img className="h-6 w-6 rounded-full" />
                                    <div>
                                        <p>{project.name}</p>
                                    </div>
                                </div>

                                <Badge>
                                    {project.remoteRepo.includes('github') ? <RxGithubLogo /> : <FaGitlab />}
                                    {project.name}
                                </Badge>

                                <p className="text-xs">{project.lastCommitMessage}</p>

                                <div></div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
