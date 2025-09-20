import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import { projectsList, projectsShow } from '@/routes';
import { runtimes } from '@/types/runtimes';
import type { BreadcrumbItem, Project, Runtime } from '@/types';
import { Form, Head, Link, usePage } from '@inertiajs/react';
import { useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';


const ViewProjectBreadcrumbs = () => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsList().url,
        },
        {
            title: 'New Service',
            href: '',
        },
    ];

    return breadcrumbs;
};




export default function DeployService() {
    const breadcrumbs = ViewProjectBreadcrumbs();
    const { props } = usePage();
    const project = props.project as Project;


    const queryClient = useQueryClient();

    const onCreatedSuccess = () => {
        queryClient.invalidateQueries({ queryKey: ['projects'] });
    };

    const {
        name,
        rootDir,
        runtime,
        error,
        buildCommand,

        setName,
        setRootDir,
        setRuntime,
        setBuildCommad,
        getFormData,
        handleFormSuccess
    } = useServices(onCreatedSuccess);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Project" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex items-start justify-between">
                        <div className="flex flex-col gap-1">
                            <p className="text-2xl font-black">New Service</p>
                            <p className="text-sm font-normal text-muted-foreground">Choose a new service</p>
                        </div>    
                    </div>

                    <Form
                        action={`/projects/${project.id}/services`}
                        transform={() => getFormData()}
                        method="post"
                        onSuccess={handleFormSuccess}
                        className="grid items-start gap-6"
                    >
                        {({ processing, errors }) => (
                            <>
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
                                    {(error || errors.name) && (
                                        <div className="text-sm text-destructive">{error || errors.name}</div>
                                    )}
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
                                            {runtimes.map((option) => (
                                                <SelectItem key={option} value={option}>
                                                    <div className="flex items-center gap-2">
                                                        {getRuntimeIcon(option)}
                                                        <span>{option}</span>
                                                    </div>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    {(error || errors.runtime) && (
                                        <div className="text-sm text-destructive">{error || errors.runtime}</div>
                                    )}
                                </div>
                                    {
                                        !(runtime === 'k3s' || runtime === 'docker') && 
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
                                            {(error || errors.description) && (
                                                <div className="text-sm text-destructive">{error || errors.description}</div>
                                            )}
                                        </div>
                                    }
                                <div className="grid gap-3">
                                    <Label htmlFor="build_command">Build Command</Label>
                                    <Input
                                        id="build_command"
                                        name="build_command"
                                        placeholder="npm run build"
                                        value={name}
                                        onChange={(e) => setName(e.target.value)}
                                        tabIndex={1}
                                        disabled={processing}
                                    />
                                    {(error || errors.name) && (
                                        <div className="text-sm text-destructive">{error || errors.name}</div>
                                    )}
                                </div>

                                <div className="flex justify-end gap-2">
                                    <Button type="button" variant="outline">
                                        <Link href={projectsShow({ project: project.id }).url}>
                                            Cancel
                                        </Link>
                                    </Button>
                                    <Button type="submit" disabled={processing}>
                                        {processing && <Loader2 className="h-4 w-4 animate-spin" />}
                                        {processing ? 'Creating' : 'Create'}
                                    </Button>
                                </div>
                            </>
                        )}
                    </Form>

                </div>
            </div>
        </AppLayout>
    );
}
