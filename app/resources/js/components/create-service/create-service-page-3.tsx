import { Button } from '@/components/ui/button';
import type { BlueprintFormat } from '@/types';

interface Props {
    name: string;
    yamlConfig: string;
    jsonConfig: string;
    blueprintFormat: BlueprintFormat;
    setBlueprintFormat: (arg0: BlueprintFormat) => void;
    handleBlueprintCopy: () => void;
}

export function CreateServicePage3({ name, yamlConfig, jsonConfig, blueprintFormat, setBlueprintFormat, handleBlueprintCopy }: Props) {
    return (
        <div className="grid items-start gap-6">
            <div className="rounded-lg bg-muted p-4">
                <div className="mb-4 flex items-center justify-between">
                    <h3 className="text-lg font-semibold">Configuration Blueprint</h3>
                    <div className="flex items-center gap-2">
                        <div className="flex rounded-md bg-background p-1">
                            <Button
                                type="button"
                                variant={blueprintFormat === 'yaml' ? 'default' : 'ghost'}
                                size="sm"
                                onClick={() => setBlueprintFormat('yaml')}
                                className="h-7 px-3"
                            >
                                YAML
                            </Button>
                            <Button
                                type="button"
                                variant={blueprintFormat === 'json' ? 'default' : 'ghost'}
                                size="sm"
                                onClick={() => setBlueprintFormat('json')}
                                className="h-7 px-3"
                            >
                                JSON
                            </Button>
                        </div>
                        <Button type="button" variant="outline" size="sm" onClick={handleBlueprintCopy} className="h-7 px-3">
                            Copy
                        </Button>
                    </div>
                </div>

                <div className="rounded border bg-background">
                    <pre className="overflow-x-auto p-4 font-mono text-sm whitespace-pre-wrap">
                        <code>{blueprintFormat === 'yaml' ? yamlConfig : jsonConfig}</code>
                    </pre>
                </div>

                <p className="mt-2 text-xs text-muted-foreground">
                    This configuration will be saved as{' '}
                    <code className="rounded bg-background px-1 text-xs">
                        {name}.{blueprintFormat}
                    </code>
                </p>
            </div>
        </div>
    );
}
