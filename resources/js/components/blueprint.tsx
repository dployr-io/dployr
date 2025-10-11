import { Button } from '@/components/ui/button';
import type { BlueprintFormat } from '@/types';
import { Copy } from 'lucide-react';

interface Props {
    name: string;
    yamlConfig: string;
    jsonConfig: string;
    blueprintFormat: BlueprintFormat;
    setBlueprintFormat: (arg0: BlueprintFormat) => void;
    handleBlueprintCopy: () => void;
}

export function Blueprint({ yamlConfig, jsonConfig, blueprintFormat, setBlueprintFormat, handleBlueprintCopy }: Props) {
    return (
        <>
            <div className="mb-4 flex items-center justify-between">
                <div className="flex w-full items-center justify-end gap-2">
                    <div className="flex items-center gap-1 rounded-md bg-background p-1">
                        <Button
                            type="button"
                            variant={blueprintFormat === 'yaml' ? 'default' : 'ghost'}
                            size="sm"
                            onClick={() => setBlueprintFormat('yaml')}
                            className="h-5 px-2 text-xs"
                        >
                            YAML
                        </Button>
                        <Button
                            type="button"
                            variant={blueprintFormat === 'json' ? 'default' : 'ghost'}
                            size="sm"
                            onClick={() => setBlueprintFormat('json')}
                            className="h-5 px-2 text-xs"
                        >
                            JSON
                        </Button>
                    </div>
                    <Button type="button" variant="ghost" size="sm" onClick={handleBlueprintCopy}>
                        <Copy className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            <div className="rounded border bg-background">
                <pre className="overflow-x-auto p-4 font-mono text-sm whitespace-pre-wrap">
                    <code>{blueprintFormat === 'yaml' ? yamlConfig : jsonConfig}</code>
                </pre>
            </div>
        </>
    );
}
