import { LogsWindow } from '@/components/logs-window';
import { useLogs } from '@/hooks/use-logs';
import AppLayout from '@/layouts/app-layout';
import { logs } from '@/routes';
import { type BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Logs',
        href: logs().url,
    },
];

export default function Logs() {
    const { 
        logs, 
        filteredLogs, 
        selectedLevel, 
        searchQuery, 
        logsEndRef, 
        setSelectedLevel, 
        setSearchQuery 
    } = useLogs();

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Logs" />
            <div className="flex h-full min-h-0 flex-col gap-4 overflow-y-hidden rounded-xl p-4">
                <div className="flex min-h-0 flex-1 auto-rows-min gap-4 p-8">
                    <div className="flex min-h-0 w-full flex-1 flex-col gap-6">
                        <LogsWindow
                            logs={logs}
                            filteredLogs={filteredLogs}
                            selectedLevel={selectedLevel}
                            searchQuery={searchQuery}
                            logsEndRef={logsEndRef}
                            setSelectedLevel={setSelectedLevel}
                            setSearchQuery={setSearchQuery}
                        />
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
