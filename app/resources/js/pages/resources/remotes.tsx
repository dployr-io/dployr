import RemoteAddDialog from '@/components/remote-add-dialog';
import RemoteCard from '@/components/remote-card';

import { useRemotes } from '@/hooks/use-remotes';
import AppLayout from '@/layouts/app-layout';
import { remotesList } from '@/routes';
import type { BreadcrumbItem, Remote } from '@/types';
import { Head } from '@inertiajs/react';
import { PlusCircle } from 'lucide-react';
import { useState } from 'react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Remotes',
        href: remotesList().url,
    },
];

export default function Remotes() {
    const { remotes } = useRemotes();
    const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-2xl font-black">Remote Repositories</p>
                        <p className="text-sm font-normal text-muted-foreground">Manage your remote repositories</p>
                    </div>

                    <div className="grid w-full grid-cols-3 gap-3">
                        {remotes?.data?.map((remote: Remote) => (
                            <RemoteCard
                                key={remote.id}
                                id={remote.id}
                                name={remote.name}
                                repository={remote.repository}
                                branch={remote.branch}
                                provider={remote.provider}
                                commit_message={remote.commit_message}
                                avatar_url={remote.avatar_url}
                            />
                        ))}
                        <div
                            className="flex flex-col gap-2 rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-accent-foreground md:min-h-min dark:border-sidebar-border dark:hover:border-muted-foreground"
                            onClick={() => setIsDialogOpen(true)}
                        >
                            <div className="flex items-center gap-2">
                                <PlusCircle size={20} className="text-muted-foreground" />
                                <p>Import a New Repository</p>
                            </div>

                            <p className="text-sm text-muted-foreground">Click to import a new remote repository to your library</p>
                        </div>
                    </div>

                    <RemoteAddDialog open={isDialogOpen} setOpen={setIsDialogOpen} />
                </div>
            </div>
        </AppLayout>
    );
}
