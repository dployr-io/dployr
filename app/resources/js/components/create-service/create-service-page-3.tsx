import { Remote } from '@/types';

interface FormData {
    name: string;
    source: string;
    runtime: string;
    workingDir?: string | null;
    buildCmd?: string | null;
    port: string;
    domain: string;
    dnsProvider: string;
    remote?: Remote | null; // This is the remote ID, not the full object
}

interface Props {
    formData: FormData;
}

export function CreateServicePage3({ formData }: Props) {
    return (
        <div className="grid items-start gap-6">
            <div className="rounded-lg bg-muted p-4">
                <h3 className="mb-4 text-lg font-semibold">Summary</h3>
                <div className="space-y-3 text-sm">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <strong>Name:</strong> {formData.name || 'Not specified'}
                        </div>
                        <div>
                            <strong>Source:</strong>{' '}
                            {formData.source === 'image' ? 'Docker Image' : formData.source === 'remote' ? 'Remote Repository' : formData.source}
                        </div>
                    </div>

                    {formData.remote && (
                        <div>
                            <strong>Repository:</strong> {formData.remote?.name}/{formData.remote?.repository}
                        </div>
                    )}

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <strong>Runtime:</strong> {formData.runtime || 'Not specified'}
                        </div>
                        <div>
                            <strong>Port:</strong> {formData.port || 'Not specified'}
                        </div>
                    </div>

                    {formData.workingDir && (
                        <div>
                            <strong>Root Directory:</strong> {formData.workingDir}
                        </div>
                    )}

                    {formData.buildCmd && (
                        <div>
                            <strong>Build Command:</strong> {formData.buildCmd}
                        </div>
                    )}

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <strong>Domain:</strong> {formData.domain || 'Not specified'}
                        </div>
                        <div>
                            <strong>DNS Provider:</strong> {formData.dnsProvider || 'Not specified'}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
