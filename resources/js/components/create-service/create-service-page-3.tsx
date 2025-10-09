import { Blueprint } from '@/components/blueprint';
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
                <Blueprint
                    name={name}
                    yamlConfig={yamlConfig}
                    jsonConfig={jsonConfig}
                    blueprintFormat={blueprintFormat}
                    setBlueprintFormat={setBlueprintFormat}
                    handleBlueprintCopy={handleBlueprintCopy}
                />

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
