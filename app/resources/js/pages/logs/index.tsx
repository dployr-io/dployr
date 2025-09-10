import { Badge } from '@/components/ui/badge';
import AppLayout from '@/layouts/app-layout';
import { logs } from '@/routes';
import { type BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';
import { FaGitlab } from 'react-icons/fa6';
import { RxGithubLogo } from 'react-icons/rx';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Logs',
        href: logs().url,
    },
];

export default function Logs() {
    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Logs" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="grid h-full auto-rows-min gap-4 p-8 md:grid-cols-3">
                    <div className="flex flex-col gap-6">
                        <p className="text-3xl font-black">Overview</p>

                        
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
