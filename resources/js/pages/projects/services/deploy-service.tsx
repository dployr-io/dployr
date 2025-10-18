import { CreateServicePage1 } from '@/components/create-service/create-service-page-1';
import { CreateServicePage2 } from '@/components/create-service/create-service-page-2';
import { CreateServicePage3 } from '@/components/create-service/create-service-page-3';
import { CreateServicePage4 } from '@/components/create-service/create-service-page-4';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';

import { useRemotes } from '@/hooks/use-remotes';
import { useRuntimes } from '@/hooks/use-runtimes';
import { useServiceForm } from '@/hooks/use-service-form';
import AppLayout from '@/layouts/app-layout';
import { deploymentsIndex, projectsIndex, projectsShow } from '@/routes';
import type { BreadcrumbItem, DnsProvider, Project } from '@/types';
import { Form, Head, Link, router, usePage } from '@inertiajs/react';
import { useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';

const ViewProjectBreadcrumbs = (project?: Project) => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsIndex().url,
        },
        {
            title: project && project.id ? project.name || 'Project' : 'Project',
            href: project && project.id ? projectsShow({ project: project.id }).url : '',
        },
        {
            title: 'New Service',
            href: '',
        },
    ];

    return breadcrumbs;
};

export default function DeployService() {
    const { props } = usePage();
    const project = (props.project as Project) || null;
    const breadcrumbs = ViewProjectBreadcrumbs(project);
    const [showSkipDialog, setShowSkipDialog] = useState(false);
    const queryClient = useQueryClient();
    const params = new URLSearchParams(window.location.search);
    const spec = params.get('spec');

    const onCreatedSuccess = () => {
        queryClient.invalidateQueries({ queryKey: ['deployments', spec] });
        router.visit(deploymentsIndex().url);
    };

    const {
        currentPage,
        name,
        nameError,
        remoteError,
        workingDir,
        workingDirError,
        staticDir,
        staticDirError,
        runtime,
        runtimeError,
        version,
        versionError,
        remote,
        runCmd,
        runCmdError,
        buildCmd,
        buildCmdError,
        source,
        port,
        portError,
        domain,
        domainError,
        dnsProvider,
        dnsProviderError,
        blueprintFormat,
        runCmdPlaceholder,
        buildCmdPlaceholder,
        envVars,
        secrets,

        // Unified handlers
        setField,
        getFormSubmissionData,
        onSourceValueChanged,
        onRemoteValueChanged,
        onRuntimeValueChanged,
        onVersionValueChanged,
        nextPage,
        prevPage,
        validateSkip,
        handleBlueprintCopy,
        setBlueprintFormat,
        yamlConfig,
        jsonConfig,
        updateEnvVar,
        updateSecret,
        removeEnvVar,
        removeSecret,
    } = useServiceForm();

    const { runtimes: versions, isLoading: isRuntimesLoading } = useRuntimes(runtime);

    const { remotes, isLoading: isRemotesLoading } = useRemotes();

    const getPageTitle = () => {
        switch (currentPage) {
            case 1:
                return 'Basic Configuration';
            case 2:
                return 'Network Configuration';
            case 3:
                return 'Environment Variables';
            case 4:
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
                return 'Configure your environment variables and secrets';
            case 4:
                return 'Review your service configuration before creating';
            default:
                return 'Set up your new service';
        }
    };

    const renderCurrentPage = (processing: boolean, errors: Record<string, string>) => {
        switch (currentPage) {
            case 1:
                return (
                    <CreateServicePage1
                        name={name}
                        nameError={nameError}
                        remoteError={remoteError}
                        workingDir={workingDir!}
                        workingDirError={workingDirError}
                        staticDir={staticDir}
                        staticDirError={staticDirError}
                        runtime={runtime}
                        runtimeError={runtimeError}
                        remote={remote}
                        isRemotesLoading={isRemotesLoading}
                        version={version}
                        versions={versions || []}
                        versionError={versionError}
                        isRuntimesLoading={isRuntimesLoading}
                        remotes={remotes || []}
                        runCmd={runCmd!}
                        runCmdError={runCmdError}
                        buildCmd={buildCmd!}
                        buildCmdError={buildCmdError}
                        source={source}
                        processing={processing}
                        errors={errors}
                        runCmdPlaceholder={runCmdPlaceholder}
                        buildCmdPlaceholder={buildCmdPlaceholder}
                        setField={setField}
                        onSourceValueChanged={onSourceValueChanged}
                        onRemoteValueChanged={onRemoteValueChanged}
                        onVersionValueChanged={onVersionValueChanged}
                        onRuntimeValueChanged={onRuntimeValueChanged}
                    />
                );
            case 2:
                return (
                    <CreateServicePage2
                        port={port!}
                        runtime={runtime}
                        portError={portError}
                        domain={domain!}
                        domainError={domainError}
                        dnsProvider={dnsProvider!}
                        dnsProviderError={dnsProviderError}
                        processing={processing}
                        errors={errors}
                        setField={setField}
                    />
                );
            case 3:
                return (
                    <CreateServicePage3
                        envVars={envVars}
                        secrets={secrets}
                        onUpdateEnvVar={updateEnvVar}
                        onUpdateSecret={updateSecret}
                        onRemoveEnvVar={removeEnvVar}
                        onRemoveSecret={removeSecret}
                    />
                );
            case 4:
                return (
                    <CreateServicePage4
                        name={name}
                        yamlConfig={yamlConfig}
                        jsonConfig={jsonConfig}
                        blueprintFormat={blueprintFormat}
                        setBlueprintFormat={setBlueprintFormat}
                        handleBlueprintCopy={handleBlueprintCopy}
                    />
                );
            default:
                return null;
        }
    };

    const renderNavigationButtons = (processing: boolean, port: string, provider: DnsProvider) => {
        return (
            <div className="mt-8 flex gap-2">
                <div className="ml-auto flex justify-end gap-2">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            if (runtime === 'static') {
                                setField('currentPage', Math.min(1, currentPage - 2));
                            } else {
                                if (currentPage > 1) {
                                    prevPage();
                                }
                            }
                        }}
                    >
                        <Link
                            href={currentPage === 1 ? (project && project.id ? projectsShow({ project: project.id }).url : projectsIndex().url) : ''}
                        >
                            {currentPage === 1 ? 'Cancel' : 'Back'}
                        </Link>
                    </Button>
                    {currentPage < 4 ? (
                        currentPage === 2 && ((runtime !== 'static' && port.length < 4) || !provider || domain.length < 4) ? (
                            <>
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        if (validateSkip()) setShowSkipDialog(true);
                                    }}
                                    disabled={processing}
                                >
                                    Skip
                                </Button>
                                <AlertDialog open={showSkipDialog} onOpenChange={setShowSkipDialog}>
                                    <AlertDialogContent>
                                        <AlertDialogHeader>
                                            <AlertDialogTitle>Proceed without HTTPS?</AlertDialogTitle>
                                            <AlertDialogDescription>
                                                Your app will only be accessible over unsecured HTTP. You can enable HTTPS later, but it’s strongly
                                                recommended to set it up now.
                                            </AlertDialogDescription>
                                        </AlertDialogHeader>
                                        <AlertDialogFooter>
                                            <AlertDialogCancel onClick={() => setShowSkipDialog(false)}>Go back</AlertDialogCancel>
                                            <AlertDialogAction
                                                onClick={() => {
                                                    setShowSkipDialog(false);
                                                    const payload = { port };
                                                    router.post('/projects/services/check-port', payload, {
                                                        onSuccess: () => (runtime === 'static' ? setField('currentPage', 4) : nextPage()),
                                                        onError: (errors) => {
                                                            if (errors.port) {
                                                                setField('portError', errors.port);
                                                            }
                                                        },
                                                    });
                                                }}
                                            >
                                                Continue
                                            </AlertDialogAction>
                                        </AlertDialogFooter>
                                    </AlertDialogContent>
                                </AlertDialog>
                            </>
                        ) : (
                            <Button type="button" onClick={nextPage} disabled={processing}>
                                Next
                            </Button>
                        )
                    ) : (
                        <Button type="submit" disabled={processing}>
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
            <div className="flex h-full flex-1 flex-col items-center gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full max-w-4xl flex-col gap-6 px-9 py-6">
                    <div className="flex items-start justify-between">
                        <div className="flex flex-col gap-1">
                            <p className="text-2xl font-black">{getPageTitle()}</p>
                            <p className="text-sm font-normal text-muted-foreground">{getPageDescription()}</p>
                        </div>
                    </div>

                    {currentPage !== 4 && (
                        <div className="mb-4 flex justify-center">
                            <div className="flex items-center space-x-2">
                                <div className="h-2 w-64 overflow-hidden rounded-full bg-muted">
                                    <div
                                        className="h-full bg-primary transition-all duration-300"
                                        style={{
                                            width: `${(currentPage / 4) * 100}%`,
                                        }}
                                    />
                                </div>
                            </div>
                        </div>
                    )}
                    <Form
                        action={currentPage === 2 ? '/projects/services/check-port' : `/projects/${project && project.id ? project.id : ''}/services`}
                        transform={() => getFormSubmissionData()}
                        method="post"
                        onSuccess={onCreatedSuccess}
                    >
                        {({ processing, errors }) => (
                            <>
                                {renderCurrentPage(processing, errors)}
                                {renderNavigationButtons(processing, String(port), dnsProvider!)}
                            </>
                        )}
                    </Form>
                </div>
            </div>
        </AppLayout>
    );
}
