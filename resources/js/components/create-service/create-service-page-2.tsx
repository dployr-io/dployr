import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Runtime } from '@/types';
import { dnsProviders } from '@/types/runtimes';

interface Props {
    port: number;
    runtime: Runtime;
    portError: string;
    domain: string;
    domainError: string;
    dnsProvider: string;
    dnsProviderError: string;
    processing: boolean;
    errors: Record<string, string>;
    setField: (field: string, value: unknown) => void;
}

export function CreateServicePage2({
    port,
    runtime,
    portError,
    domain,
    domainError,
    dnsProvider,
    dnsProviderError,
    processing,
    errors,
    setField,
}: Props) {
    return (
        <div className="grid items-start gap-6">
            <div className="grid gap-3">
                <Label htmlFor="port">Port{runtime !== 'static' && <span className="text-destructive">*</span>}</Label>
                <Input
                    id="port"
                    name="port"
                    placeholder="3000"
                    value={port || ''}
                    onChange={(e) => setField('port', Number(e.target.value))}
                    tabIndex={1}
                    disabled={processing}
                />
                {(portError || errors.port) && <div className="text-sm text-destructive">{portError || errors.port}</div>}
            </div>

            <div className="grid gap-3">
                <Label htmlFor="domain">Domain</Label>
                <Input
                    id="domain"
                    name="domain"
                    placeholder="myapp.example.com"
                    value={domain}
                    onChange={(e) => setField('domain', e.target.value)}
                    tabIndex={2}
                    disabled={processing}
                />
                {(domainError || errors.domain) && <div className="text-sm text-destructive">{domainError || errors.domain}</div>}
            </div>

            <div className="grid gap-3">
                <Label htmlFor="dns_provider">DNS Provider</Label>
                <Select value={dnsProvider} onValueChange={(value: string) => setField('dnsProvider', value)}>
                    <SelectTrigger id="dns_provider" disabled={processing}>
                        <SelectValue placeholder="Select provider" />
                    </SelectTrigger>
                    <SelectContent>
                        {dnsProviders.map((option) => (
                            <SelectItem key={option} value={option}>
                                <div className="flex items-center gap-2">
                                    <span>{option.charAt(0).toUpperCase() + option.slice(1)}</span>
                                </div>
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                {(dnsProviderError || errors.dns_provider) && (
                    <div className="text-sm text-destructive">{dnsProviderError || errors.dns_provider}</div>
                )}
            </div>
        </div>
    );
}
