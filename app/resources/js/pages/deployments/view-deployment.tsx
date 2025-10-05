import { Blueprint as BlueprintSection } from '@/components/blueprint';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useLogs } from '@/hooks/use-logs';
import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { toJson, toYaml } from '@/lib/utils';
import { deploymentsList, deploymentsShow } from '@/routes';
import type { Blueprint, BreadcrumbItem, Log } from '@/types';
import { Head, usePage } from '@inertiajs/react';
import { Loader2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

const ViewProjectBreadcrumbs = (blueprint?: Blueprint) => {
    const config = JSON.parse(blueprint!.config as string);

    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Deployments',
            href: deploymentsList().url,
        },
        {
            title: config.name || 'Deployment',
            href: blueprint && blueprint.id ? deploymentsShow(blueprint.id).url : '',
        },
    ];

    return breadcrumbs;
};

function LogEntry({ log }: { log: Log }) {
    return (
        <div className="border-b border-accent px-4 py-2 last:border-0">
            <p className="text-sm text-muted-foreground">{log.message}</p>
        </div>
    );
}

export default function ViewDeployment() {
    const { props } = usePage();
    const blueprint = (props.blueprint as Blueprint) || null;
    const config = JSON.parse(blueprint.config as string);
    const breadcrumbs = ViewProjectBreadcrumbs(blueprint);

    const { logs, logsEndRef } = useLogs();

    const { blueprintFormat, setBlueprintFormat } = useServices();

    const yamlConfig = toYaml(config);
    const jsonConfig = toJson(config);
    const handleBlueprintCopy = async () => {
        try {
            if (!blueprint) return;
            await navigator.clipboard.writeText(blueprintFormat === 'yaml' ? yamlConfig : jsonConfig);
        } catch (err) {}
    };

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-xl font-semibold">{config?.name || 'Deployment'}</p>
                    </div>
                    <Tabs defaultValue="logs">
                        <TabsList>
                            <TabsTrigger value="logs">Logs</TabsTrigger>
                            <TabsTrigger value="blueprint">Blueprint</TabsTrigger>
                        </TabsList>
                        <TabsContent value="logs">
                            <div className="mt-8 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-sidebar-border">
                                <div className="flex flex-shrink-0 gap-2 bg-neutral-50 p-2 dark:bg-neutral-900">
                                    <p className="text-xs font-medium text-muted-foreground">Startup Logs</p>
                                </div>
                                <Separator />
                                <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
                                    <div className="min-h-0 flex-1 overflow-y-auto">
                                        {logs?.length === 0 ? (
                                            <div className="flex h-120 items-center justify-center gap-1">
                                                <Loader2 width={12} height={12} className="animate-spin" />
                                                <p className="text-sm text-muted-foreground">Retrieving logs</p>
                                            </div>
                                        ) : (
                                            logs?.map((log: Log) => <LogEntry key={log.id} log={log} />)
                                        )}
                                        <div ref={logsEndRef} />
                                    </div>
                                    <div className="border-t border-accent bg-neutral-50 p-2 text-center text-xs text-muted-foreground dark:bg-neutral-800">
                                        {logs.length > 0 ? `Showing ${logs.length} of ${logs.length} log entries` : 'No logs yet'}
                                    </div>
                                </div>
                            </div>
                        </TabsContent>
                        <TabsContent value="blueprint">
                            <BlueprintSection
                                name={config.name}
                                blueprintFormat={blueprintFormat}
                                yamlConfig={yamlConfig}
                                jsonConfig={jsonConfig}
                                setBlueprintFormat={setBlueprintFormat}
                                handleBlueprintCopy={handleBlueprintCopy}
                            />
                        </TabsContent>
                    </Tabs>
                </div>
            </div>
        </AppLayout>
    );
}
