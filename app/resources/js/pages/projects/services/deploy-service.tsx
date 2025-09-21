import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import { projectsList, projectsShow } from '@/routes';
import type { BreadcrumbItem, Project, Runtime, ServiceSource } from '@/types';
import { runtimes } from '@/types/runtimes';
import { Form, Head, Link, usePage } from '@inertiajs/react';
import { useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';

const ViewProjectBreadcrumbs = (project :Project) => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsList().url,
        },
        {
            title: project?.name || 'Project',
            href: projectsShow(project.id).url,
        },
        {
            title: 'New Service',
            href: '',
        },
    ];

    return breadcrumbs;
};

// Page 1: Basic Service Configuration
const Page1BasicConfig = ({
    name,
    nameError,
    setName,
    rootDir,
    rootDirError,
    setRootDir,
    runtime,
    runtimeError,
    setRuntime,
    buildCommand,
    buildCommandError,
    setBuildCommad,
    source,
    sourceError,
    setSource,
    onSourceValueChanged,
    error,
    processing,
    errors,
}: any) => (
    <div className="grid items-start gap-6">
        <div className="grid gap-3">
            <Label htmlFor="source">Source</Label>
            <Select value={source} onValueChange={onSourceValueChanged}>
                <SelectTrigger id="source" disabled={processing}>
                    <SelectValue>
                        <div className="flex items-center gap-2">
                            <span>{source === 'image' ? 'Docker Image' : source === 'remote' ? 'Remote Repository' : source}</span>
                        </div>
                    </SelectValue>
                </SelectTrigger>
                <SelectContent>
                    {(['image', 'remote'] as ServiceSource[]).map((option) => {
                        let label = '';
                        if (option === 'image') label = 'Docker Image';
                        else if (option === 'remote') label = 'Remote Repository';
                        else label = option;
                        return (
                            <SelectItem key={option} value={option}>
                                <div className="flex items-center gap-2">
                                    <span>{label}</span>
                                </div>
                            </SelectItem>
                        );
                    })}
                </SelectContent>
            </Select>
            {(sourceError || errors.source) && <div className="text-sm text-destructive">{sourceError || errors.source}</div>}
        </div>

        <div className="grid gap-3">
            <Label htmlFor="name">Name</Label>
            <Input
                id="name"
                name="name"
                placeholder="My awesome dployr project"
                value={name}
                onChange={(e) => setName(e.target.value)}
                tabIndex={1}
                disabled={processing}
            />
            {(nameError || errors.name) && <div className="text-sm text-destructive">{nameError || errors.name}</div>}
        </div>

        <div className="grid gap-3">
            <Label htmlFor="runtime">Runtime</Label>
            <Select value={runtime} onValueChange={(value: Runtime) => setRuntime(value)}>
                <SelectTrigger id="runtime" disabled={processing}>
                    <SelectValue>
                        <div className="flex items-center gap-2">
                            {getRuntimeIcon(runtime)}
                            <span>{runtime}</span>
                        </div>
                    </SelectValue>
                </SelectTrigger>
                <SelectContent>
                    {runtimes
                        .filter((option) => {
                            const isImage = source === 'image';
                            const isRemote = source === 'remote';
                            return isImage ? option === 'k3s' || option === 'docker' : isRemote ? option !== 'k3s' && option !== 'docker' : true;
                        })
                        .map((option) => (
                            <SelectItem key={option} value={option}>
                                <div className="flex items-center gap-2">
                                    {getRuntimeIcon(option)}
                                    <span>{option}</span>
                                </div>
                            </SelectItem>
                        ))}
                </SelectContent>
            </Select>
            {(runtimeError || errors.runtime) && <div className="text-sm text-destructive">{runtimeError || errors.runtime}</div>}
        </div>

        {source === 'remote' && (
            <div className="grid gap-3">
                <Label htmlFor="root_dir">Root Directory</Label>
                <Input
                    id="root_dir"
                    name="root_dir"
                    placeholder="src"
                    value={rootDir}
                    onChange={(e) => setRootDir(e.target.value)}
                    tabIndex={2}
                    disabled={processing}
                />
                {(rootDirError || errors.root_dir) && <div className="text-sm text-destructive">{rootDirError || errors.root_dir}</div>}
            </div>
        )}

        {source === 'remote' && (
            <div className="grid gap-3">
                <Label htmlFor="build_command">Build Command</Label>
                <Input
                    id="build_command"
                    name="build_command"
                    placeholder="npm run build"
                    value={buildCommand}
                    onChange={(e) => setBuildCommad(e.target.value)}
                    tabIndex={1}
                    disabled={processing}
                />
                {(buildCommandError || errors.build_command) && (
                    <div className="text-sm text-destructive">{buildCommandError || errors.build_command}</div>
                )}
            </div>
        )}
    </div>
);

// Page 2: Port, Domain, DNS Configuration
const Page2NetworkConfig = ({
    port,
    portError,
    setPort,
    domain,
    domainError,
    setDomain,
    dnsProvider,
    dnsProviderError,
    setDnsProvider,
    processing,
    errors,
}: any) => (
    <div className="grid items-start gap-6">
        <div className="grid gap-3">
            <Label htmlFor="port">Port</Label>
            <Input
                id="port"
                name="port"
                placeholder="3000"
                value={port}
                onChange={(e) => setPort(e.target.value)}
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
                onChange={(e) => setDomain(e.target.value)}
                tabIndex={2}
                disabled={processing}
            />
            {(domainError || errors.domain) && <div className="text-sm text-destructive">{domainError || errors.domain}</div>}
        </div>

        <div className="grid gap-3">
            <Label htmlFor="dns_provider">DNS Provider</Label>
            <Select value={dnsProvider} onValueChange={setDnsProvider}>
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
            {(dnsProviderError || errors.dns_provider) && <div className="text-sm text-destructive">{dnsProviderError || errors.dns_provider}</div>}
        </div>
    </div>
);

// Page 3: Confirmation
const Page3Confirmation = ({ formData }: any) => (
    <div className="grid items-start gap-6">
        <div className="rounded-lg bg-muted p-4">
            <h3 className="mb-4 text-lg font-semibold">Summary</h3>
            <div className="space-y-2 text-sm">
                <div>
                    <strong>Name:</strong> {formData.name}
                </div>
                <div>
                    <strong>Source:</strong> {formData.source}
                </div>
                <div>
                    <strong>Runtime:</strong> {formData.runtime}
                </div>
                {formData.rootDir && (
                    <div>
                        <strong>Root Directory:</strong> {formData.rootDir}
                    </div>
                )}
                {formData.buildCommand && (
                    <div>
                        <strong>Build Command:</strong> {formData.buildCommand}
                    </div>
                )}
                <div>
                    <strong>Port:</strong> {formData.port}
                </div>
                <div>
                    <strong>Domain:</strong> {formData.domain}
                </div>
                <div>
                    <strong>DNS Provider:</strong> {formData.dnsProvider}
                </div>
            </div>
        </div>
    </div>
);

export default function DeployService() {
    const { props } = usePage();
    const project = props.project as Project;
    const breadcrumbs = ViewProjectBreadcrumbs(project);

    const queryClient = useQueryClient();

    const onCreatedSuccess = () => {
        queryClient.invalidateQueries({ queryKey: ['projects'] });
    };

    const {
        currentPage,
        name,
        nameError,
        rootDir,
        rootDirError,
        runtime,
        runtimeError,
        buildCommand,
        buildCommandError,
        source,
        port,
        portError,
        domain,
        domainError,
        dnsProvider,
        dnsProviderError,

        setName,
        setRootDir,
        setRuntime,
        setBuildCommand,
        setSource,
        setPort,
        setDomain,
        setDnsProvider,
        getFormData,
        handleFormSuccess,
        onSourceValueChanged,
        nextPage,
        prevPage,
        skipToConfirmation,
        handleCreate,
    } = useServices(onCreatedSuccess);

    const getPageTitle = () => {
        switch (currentPage) {
            case 1:
                return 'Basic Configuration';
            case 2:
                return 'Network Configuration';
            case 3:
                return 'Confirmation';
            default:
                return 'New Service';
        }
    };

    const getPageDescription = () => {
        switch (currentPage) {
            case 1:
                return 'Configure your new service below';
            case 2:
                return 'Set up networking and domain configuration';
            case 3:
                return 'Review your service configuration before creating';
            default:
                return 'Set up your new service';
        }
    };

    const renderCurrentPage = (processing: boolean, errors: any) => {
        switch (currentPage) {
            case 1:
                return (
                    <Page1BasicConfig
                        name={name}
                        nameError={nameError}
                        setName={setName}
                        rootDir={rootDir}
                        rootDirError={rootDirError}
                        setRootDir={setRootDir}
                        runtime={runtime}
                        runtimeError={runtimeError}
                        setRuntime={setRuntime}
                        buildCommand={buildCommand}
                        buildCommandError={buildCommandError}
                        setBuildCommad={setBuildCommand}
                        source={source}
                        setSource={setSource}
                        onSourceValueChanged={onSourceValueChanged}
                        processing={processing}
                        errors={errors}
                    />
                );
            case 2:
                return (
                    <Page2NetworkConfig
                        port={port}
                        portError={portError}
                        setPort={setPort}
                        domain={domain}
                        domainError={domainError}
                        setDomain={setDomain}
                        dnsProvider={dnsProvider}
                        dnsProviderError={dnsProviderError}
                        setDnsProvider={setDnsProvider}
                        processing={processing}
                        errors={errors}
                    />
                );
            case 3:
                return <Page3Confirmation formData={getFormData()} />;
            default:
                return null;
        }
    };

    const renderNavigationButtons = (processing: boolean) => {
        return (
            <div className="mt-8 flex gap-2">
                <div className="ml-auto flex justify-end gap-2">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            if (currentPage > 1) {
                                prevPage();
                            }
                        }}
                    >
                        <Link href={currentPage === 1 ? projectsShow({ project: project.id }).url : ''}>{currentPage === 1 ? 'Cancel' : 'Back'}</Link>
                    </Button>
                    {currentPage === 2 && (
                        <Button type="button" variant="outline" onClick={skipToConfirmation} disabled={processing}>
                            Skip
                        </Button>
                    )}
                    {currentPage < 3 ? (
                        <Button type="button" onClick={nextPage} disabled={processing}>
                            Next
                        </Button>
                    ) : (
                        <Button type="button" onClick={handleCreate} disabled={processing}>
                            {processing && <Loader2 className="h-4 w-4 animate-spin" />}
                            {processing ? 'Deploying' : 'Deploy'}
                        </Button>
                    )}
                </div>
            </div>
        );
    };

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Project" />
            <div className="flex justify-around">
                <div className="flex h-full max-w-4xl flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                    <div className="flex w-full flex-col gap-6 px-9 py-6">
                        <div className="flex items-start justify-between">
                            <div className="flex flex-col gap-1">
                                <p className="text-2xl font-black">{getPageTitle()}</p>
                                <p className="text-sm font-normal text-muted-foreground">{getPageDescription()}</p>
                            </div>
                        </div>

                        {
                            currentPage !== 3 && <div className="mb-4 flex justify-center">
                                <div className="flex items-center space-x-2">
                                    <div className="h-2 w-64 overflow-hidden rounded-full bg-muted">
                                        <div
                                            className="h-full bg-primary transition-all duration-300"
                                            style={{
                                                width: `${(currentPage / 3) * 100}%`,
                                            }}
                                        />
                                    </div>
                                </div>
                            </div>
                        }
                        <Form action={`/projects/${project.id}/services`} transform={() => getFormData()} method="post" onSuccess={handleFormSuccess}>
                            {({ processing, errors }) => (
                                <>
                                    {renderCurrentPage(processing, errors)}
                                    {renderNavigationButtons(processing)}
                                </>
                            )}
                        </Form>
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
