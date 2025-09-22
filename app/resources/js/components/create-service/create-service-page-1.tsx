import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import type { Remote, Runtime, ServiceSource } from '@/types';
import { runtimes } from '@/types/runtimes';
import { FaGitlab } from 'react-icons/fa6';
import { RxGithubLogo } from 'react-icons/rx';
import { ulid } from 'ulid';

interface Props {
    // Form state
    name: string;
    nameError: string;
    remoteError: string;
    workingDir?: string | null;
    workingDirError: string;
    runtime: Runtime;
    runtimeError: string;
    remote?: Remote | null;
    remotes: Remote[];
    runCmd?: string | null;
    runCmdError: string;
    source: ServiceSource;
    processing: boolean;
    errors: any;

    // Unified handlers
    setField: (field: string, value: any) => void;
    onSourceValueChanged: (arg0: ServiceSource) => void;
    onRemoteValueChanged: (arg0: Remote) => void;
}

export function CreateServicePage1({
    name,
    nameError,
    remoteError,
    workingDir,
    workingDirError,
    runtime,
    runtimeError,
    remote,
    remotes,
    runCmd,
    runCmdError,
    source,
    processing,
    errors,
    setField,
    onSourceValueChanged,
    onRemoteValueChanged,
}: Props) {
    return (
        <div className="grid items-start gap-6">
            <div className="grid gap-3">
                <Label htmlFor="source">
                    Source <span className="text-destructive">*</span>
                </Label>
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
            </div>

            {source === 'remote' && (
                <div className="grid gap-3">
                    <Label htmlFor="remote">
                        Remote Repository <span className="text-destructive">*</span>
                    </Label>
                    <Select
                        value={remote ? `${remote.name}/${remote.repository}` : 'Select remote'}
                        onValueChange={(value: string) => {
                            const selected = remotes?.find((item: Remote) => `${item.name}/${item.repository}` === value);
                            if (selected) {
                                onRemoteValueChanged(selected);
                                const suffix = ulid().slice(-4).toLowerCase();
                                setField('name', `${selected?.repository}-${suffix}`);
                            }
                        }}
                    >
                        <SelectTrigger id="remote" disabled={processing}>
                            <SelectValue>
                                <div className="flex items-center gap-2">
                                    {remote && (remote?.provider?.includes('github') ? <RxGithubLogo /> : <FaGitlab />)}
                                    <span className={!remote ? 'text-muted-foreground' : undefined}>
                                        {remote ? `${remote.name}/${remote.repository}` : 'Select a remote repository'}
                                    </span>
                                </div>
                            </SelectValue>
                        </SelectTrigger>
                        <SelectContent>
                            {Array.isArray(remotes) &&
                                remotes.map((item: Remote) => {
                                    return (
                                        <SelectItem key={item.id} value={`${item.name}/${item.repository}`}>
                                            <div className="flex items-center gap-2">
                                                {item?.provider?.includes('github') ? <RxGithubLogo /> : <FaGitlab />}
                                                <span>{`${item.name}/${item.repository}`}</span>
                                            </div>
                                        </SelectItem>
                                    );
                                })}
                        </SelectContent>
                    </Select>
                    {(remoteError || errors.remote) && <div className="text-sm text-destructive">{remoteError || errors.remote}</div>}
                </div>
            )}

            <div className="grid gap-3">
                <Label htmlFor="name">
                    Name <span className="text-destructive">*</span>
                </Label>
                <Input
                    id="name"
                    name="name"
                    placeholder="My awesome dployr project"
                    value={name}
                    onChange={(e) => setField('name', e.target.value)}
                    tabIndex={1}
                    disabled={processing}
                />
                {(nameError || errors.name) && <div className="text-sm text-destructive">{nameError || errors.name}</div>}
            </div>

            <div className="grid gap-3">
                <Label htmlFor="runtime">
                    Runtime <span className="text-destructive">*</span>
                </Label>
                <Select value={runtime} onValueChange={(value: Runtime) => setField('runtime', value)}>
                    <SelectTrigger id="runtime" disabled={processing}>
                        <SelectValue>
                            <div className="flex items-center gap-2">
                                {getRuntimeIcon(runtime)}
                                <span>{runtime && runtime.trim() !== '' ? runtime : 'Select a runtime'}</span>
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

            {source === 'remote' && runtime !== 'static' && (
                <div className="grid gap-3">
                    <Label htmlFor="run_cmd">
                        Run Command <span className="text-destructive">*</span>
                    </Label>
                    <Input
                        id="run_cmd"
                        name="run_cmd"
                        placeholder="npm run start"
                        value={runCmd!}
                        onChange={(e) => setField('runCmd', e.target.value)}
                        tabIndex={1}
                        disabled={processing}
                    />
                    {(runCmdError || errors.run_cmd) && <div className="text-sm text-destructive">{runCmdError || errors.run_cmd}</div>}
                </div>
            )}

            {source === 'remote' && (
                <div className="grid gap-3">
                    <Label htmlFor="working_dir">
                        Working Directory <span className="text-xs text-muted-foreground">(Defaults to root directory)</span>
                    </Label>
                    <Input
                        id="working_dir"
                        name="working_dir"
                        placeholder="src"
                        value={workingDir!}
                        onChange={(e) => setField('workingDir', e.target.value)}
                        tabIndex={2}
                        disabled={processing}
                    />
                    {(workingDirError || errors.working_dir) && (
                        <div className="text-sm text-destructive">{workingDirError || errors.working_dir}</div>
                    )}
                </div>
            )}
        </div>
    );
}
