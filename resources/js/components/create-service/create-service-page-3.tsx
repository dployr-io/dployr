import { ConfigTable } from '@/components/config-table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

interface Props {
    envVars: Record<string, string>;
    secrets: Record<string, string>;
    onUpdateEnvVar: (key: string, value: string) => void;
    onUpdateSecret: (key: string, value: string) => void;
    onRemoveEnvVar?: (key: string) => void;
    onRemoveSecret?: (key: string) => void;
}

export function CreateServicePage3({ envVars, secrets, onUpdateEnvVar, onUpdateSecret, onRemoveEnvVar, onRemoveSecret }: Props) {
    return (
        <div className="grid items-start gap-8">
            <Tabs defaultValue="env_vars" className="flex min-h-0 w-full flex-col">
                <TabsList className="self-start">
                    <TabsTrigger value="env_vars">Variables</TabsTrigger>
                    <TabsTrigger value="secrets">Secrets</TabsTrigger>
                </TabsList>
                <TabsContent value="env_vars" className="flex min-h-0 flex-1 flex-col">
                    <ConfigTable config={envVars} onUpdateConfig={onUpdateEnvVar} onRemoveConfig={onRemoveEnvVar} />
                </TabsContent>
                <TabsContent value="secrets">
                    <ConfigTable config={secrets} onUpdateConfig={onUpdateSecret} onRemoveConfig={onRemoveSecret} />
                </TabsContent>
            </Tabs>
        </div>
    );
}
