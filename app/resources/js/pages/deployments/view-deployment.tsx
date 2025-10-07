import { Blueprint as BlueprintSection } from '@/components/blueprint';
import { LogsWindow } from '@/components/logs-window';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useLogs } from '@/hooks/use-logs';
import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { toJson, toYaml } from '@/lib/utils';
import { deploymentsList, deploymentsShow } from '@/routes';
import type { Blueprint, BreadcrumbItem, Log } from '@/types';
import { Head, usePage } from '@inertiajs/react';

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

export default function ViewDeployment() {
    const { props } = usePage();
    const blueprint = (props.blueprint as Blueprint) || null;
    const config = JSON.parse(blueprint.config as string);
    const breadcrumbs = ViewProjectBreadcrumbs(blueprint);

    const { 
        logs, 
        filteredLogs, 
        selectedLevel, 
        searchQuery, 
        logsEndRef, 
        setSelectedLevel, 
        setSearchQuery,
    } = useLogs(blueprint);

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
                            <LogsWindow
                                logs={logs}
                                filteredLogs={filteredLogs}
                                selectedLevel={selectedLevel}
                                searchQuery={searchQuery}
                                logsEndRef={logsEndRef}
                                setSelectedLevel={setSelectedLevel}
                                setSearchQuery={setSearchQuery}
                            />
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
