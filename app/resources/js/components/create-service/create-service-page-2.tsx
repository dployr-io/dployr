import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface Props {
    // Form state
    port: string;
    portError: string;
    domain: string;
    domainError: string;
    dnsProvider: string;
    dnsProviderError: string;
    processing: boolean;
    errors: any;

    // Unified handler
    setField: (field: string, value: any) => void;
}

export function CreateServicePage2({ port, portError, domain, domainError, dnsProvider, dnsProviderError, processing, errors, setField }: Props) {
    return (
        <div className="grid items-start gap-6">
            <div className="grid gap-3">
                <Label htmlFor="port">Port</Label>
                <Input
                    id="port"
                    name="port"
                    placeholder="3000"
                    value={port}
                    onChange={(e) => setField('port', e.target.value)}
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
                        <SelectValue placeholder="Select DNS provider" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="cloudflare">Cloudflare</SelectItem>
                        <SelectItem value="aws-route53">AWS Route 53</SelectItem>
                        <SelectItem value="google-cloud-dns">Google Cloud DNS</SelectItem>
                        <SelectItem value="digitalocean">DigitalOcean</SelectItem>
                    </SelectContent>
                </Select>
                {(dnsProviderError || errors.dns_provider) && (
                    <div className="text-sm text-destructive">{dnsProviderError || errors.dns_provider}</div>
                )}
            </div>
        </div>
    );
}
