import { toJson, toYaml } from '@/lib/utils';
import type { Blueprint, BlueprintFormat, DnsProvider, Remote, Runtime, Service, ServiceSource } from '@/types';
import { dnsProviders, runtimes } from '@/types/runtimes';
import { useMemo, useReducer, useState } from 'react';
import { z } from 'zod';

interface ServiceFormState {
    currentPage: number;
    name: string;
    runtime: Runtime;
    workingDir?: string | null;
    outputDir?: string | null;
    remote?: Remote | null;
    ciRemote?: Remote | null;
    image?: string | null;
    spec?: string | null;
    runCmd?: string | null;
    source: ServiceSource;
    port?: number | null;
    domain: string;
    dnsProvider?: DnsProvider | null;
    envVars: string;
    secrets: string;
    // Error states
    nameError: string;
    remoteError: string;
    workingDirError: string;
    outputDirError: string;
    runtimeError: string;
    runCmdError: string;
    portError: string;
    domainError: string;
    dnsProviderError: string;
    runCmdPlaceholder?: string;
}

type ServiceFormAction =
    | { type: 'SET_CURRENT_PAGE'; payload: number }
    | { type: 'SET_FIELD'; payload: { field: string; value: unknown } }
    | { type: 'SET_ERROR'; payload: { field: string; value: string } }
    | { type: 'CLEAR_ALL_ERRORS' }
    | { type: 'NEXT_PAGE' }
    | { type: 'PREV_PAGE' }
    | { type: 'SKIP_TO_CONFIRMATION' }
    | { type: 'SOURCE_CHANGED'; payload: ServiceSource };

const initialState: ServiceFormState = {
    currentPage: 1,
    name: '',
    workingDir: '',
    outputDir: '',
    runtime: 'node-js',
    remote: null,
    ciRemote: null,
    image: '',
    spec: '',
    runCmd: '',
    source: 'remote',
    port: null,
    domain: '',
    dnsProvider: null,
    envVars: '',
    secrets: '',
    nameError: '',
    remoteError: '',
    workingDirError: '',
    outputDirError: '',
    runtimeError: '',
    runCmdError: '',
    portError: '',
    domainError: '',
    dnsProviderError: '',
    runCmdPlaceholder: 'npm run start',
};

function serviceFormReducer(state: ServiceFormState, action: ServiceFormAction): ServiceFormState {
    switch (action.type) {
        case 'SET_CURRENT_PAGE':
            return { ...state, currentPage: action.payload };
        case 'SET_FIELD': {
            const { field, value } = action.payload;
            const errorField = `${field}Error` as keyof ServiceFormState;
            return {
                ...state,
                [field]: value,
                ...(errorField in state && { [errorField]: '' }),
            };
        }
        case 'SET_ERROR':
            return { ...state, [action.payload.field]: action.payload.value };
        case 'CLEAR_ALL_ERRORS':
            return {
                ...state,
                nameError: '',
                remoteError: '',
                workingDirError: '',
                outputDirError: '',
                runtimeError: '',
                runCmdError: '',
                portError: '',
                domainError: '',
                dnsProviderError: '',
            };
        case 'NEXT_PAGE':
            return { ...state, currentPage: Math.min(state.currentPage + 1, 3) };
        case 'PREV_PAGE':
            return {
                ...state,
                currentPage: Math.max(state.currentPage - 1, 1),
                nameError: '',
                workingDirError: '',
                outputDirError: '',
                runtimeError: '',
                runCmdError: '',
                portError: '',
                domainError: '',
                dnsProviderError: '',
            };
        case 'SKIP_TO_CONFIRMATION':
            return {
                ...state,
                currentPage: 3,
                nameError: '',
                workingDirError: '',
                outputDirError: '',
                runtimeError: '',
                runCmdError: '',
                portError: '',
                domainError: '',
                dnsProviderError: '',
            };
        case 'SOURCE_CHANGED':
            return {
                ...state,
                source: action.payload,
                runtime: action.payload === 'image' ? 'docker' : 'node-js',
            };
        default:
            return state;
    }
}

export function useServiceForm() {
    const [state, dispatch] = useReducer(serviceFormReducer, initialState);
    const [blueprintFormat, setBlueprintFormat] = useState<BlueprintFormat>('yaml');

    const page1Schema = z.object({
        name: z.string().min(3, 'Enter a name with at least three (3) characters'),
        workingDir: z.string().optional(),
        outputDir: z.string().optional(),
        runtime: z.enum(runtimes, 'Select a supported runtime. Read the docs for more details.'),
        runCmd: z
            .string()
            .optional()
            .refine((val) => !val || /[a-zA-Z]/.test(val), {
                message: 'Enter a valid build command',
            }),
    });

    const page2Schema = z.object({
        port: z.number().min(1024, 'Port must be a minimum of 1024').max(9999, 'Port must be a maximum of 9999').optional(),
        domain: z
            .string()
            .min(1, 'Domain is required')
            .regex(/^(?!:\/\/)([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/, 'Enter a valid domain (e.g., domain.com)')
            .optional(),
        dnsProvider: z.enum(dnsProviders, 'Select a supported DNS provider. Read the docs for more details.'),
    });

    const validatePage1 = () => {
        dispatch({ type: 'CLEAR_ALL_ERRORS' });
        let hasErrors = false;

        const result = page1Schema.safeParse({
            name: state.name,
            workingDir: state.workingDir,
            outputDir: state.outputDir,
            runtime: state.runtime,
            runCmd: state.runCmd,
        });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            if (fieldErrors.name) dispatch({ type: 'SET_ERROR', payload: { field: 'nameError', value: fieldErrors.name[0] } });
            if (fieldErrors.workingDir) dispatch({ type: 'SET_ERROR', payload: { field: 'workingDirError', value: fieldErrors.workingDir[0] } });
            if (fieldErrors.outputDir) dispatch({ type: 'SET_ERROR', payload: { field: 'outputDirError', value: fieldErrors.outputDir[0] } });
            if (fieldErrors.runtime) dispatch({ type: 'SET_ERROR', payload: { field: 'runtimeError', value: fieldErrors.runtime[0] } });
            if (fieldErrors.runCmd) dispatch({ type: 'SET_ERROR', payload: { field: 'runCmdError', value: fieldErrors.runCmd[0] } });
            hasErrors = true;
        }

        if (state.source === 'remote' && !state.remote) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'remoteError', value: 'Select a remote repository' } });
            hasErrors = true;
        }

        if (
            state.source === 'remote' &&
            state.runtime !== 'static' &&
            (!state.runCmd || !/[a-zA-Z].*[a-zA-Z]/.test(state.runCmd)) // Ensure there's at least 2 alphabetic characters
        ) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'runCmdError', value: 'Enter a valid build command' } });
            hasErrors = true;
        }

        if (state.source === 'image' && (!state.image || state.image.trim() === '')) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'imageError', value: 'Select a Docker image name' } });
            hasErrors = true;
        }

        return !hasErrors;
    };

    const validatePage2 = () => {
        dispatch({ type: 'CLEAR_ALL_ERRORS' });
        let hasErrors = false;

        const result = page2Schema.safeParse({
            port: state.port,
            domain: state.domain,
            dnsProvider: state.dnsProvider,
        });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            if (fieldErrors.port) dispatch({ type: 'SET_ERROR', payload: { field: 'portError', value: fieldErrors.port[0] } });
            if (fieldErrors.domain) dispatch({ type: 'SET_ERROR', payload: { field: 'domainError', value: fieldErrors.domain[0] } });
            if (fieldErrors.dnsProvider) dispatch({ type: 'SET_ERROR', payload: { field: 'dnsProviderError', value: fieldErrors.dnsProvider[0] } });
            return false;
        }

        if (state.runtime !== 'static' && (state.port! < 1024 || state.port! > 9999)) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'portError', value: 'Port must be between 1024 and 9999' } });
            hasErrors = true;
        }

        return !hasErrors;
    };

    const validateCurrentPage = () => {
        if (state.currentPage === 1) return validatePage1();
        if (state.currentPage === 2) return validatePage2();
        return true;
    };

    const getFormData = () => {
        return {
            name: state.name,
            working_dir: state.workingDir,
            output_dir: state.outputDir,
            runtime: state.runtime,
            remote: state.remote,
            ci_remote: state.ciRemote,
            image: state.image,
            spec: state.spec,
            run_cmd: state.runCmd,
            source: state.source,
            port: state.port,
            domain: state.domain,
            dns_provider: state.dnsProvider,
            env_vars: state.envVars,
            secrets: state.secrets,
        };
    };

    const getFormSubmissionData = () => {
        return {
            name: state.name,
            working_dir: state.workingDir,
            output_dir: state.outputDir,
            runtime: state.runtime,
            remote: state.remote?.id || null,
            ci_remote: state.ciRemote?.id || null,
            image: state.image,
            spec: state.spec,
            run_cmd: state.runCmd,
            source: state.source,
            port: state.port,
            domain: state.domain,
            dns_provider: state.dnsProvider,
            env_vars: state.envVars,
            secrets: state.secrets,
        };
    };

    const nextPage = () => {
        if (validateCurrentPage()) {
            dispatch({ type: 'NEXT_PAGE' });
        }
    };

    const validateSkip = () => {
        dispatch({ type: 'CLEAR_ALL_ERRORS' });

        if (state.runtime === 'static') {
            return true;
        }

        if (!state.port || state.port < 1024 || state.port > 9999) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'portError', value: 'Enter a valid port between 1024 and 9999' } });
            return false;
        }

        return true;
    };

    const prevPage = () => {
        dispatch({ type: 'PREV_PAGE' });
    };

    const skipToConfirmation = () => {
        dispatch({ type: 'SKIP_TO_CONFIRMATION' });
    };

    const setField = (field: string, value: unknown) => {
        dispatch({ type: 'SET_FIELD', payload: { field, value } });
    };

    const setError = (field: string, value: string) => {
        dispatch({ type: 'SET_ERROR', payload: { field, value } });
    };

    const onSourceValueChanged = (value: ServiceSource) => {
        dispatch({ type: 'SOURCE_CHANGED', payload: value });
    };

    const onRemoteValueChanged = (value: Remote) => {
        setField('remote', value);
    };

    const onRuntimeValueChanged = (value: Runtime) => {
        setField('runtime', value);

        switch (value) {
            case 'node-js': {
                setField('runCmdPlaceholder', 'npm run start');
                break;
            }
            case 'python': {
                setField('runCmdPlaceholder', 'python app.py');
                break;
            }
            case 'ruby': {
                setField('runCmdPlaceholder', 'ruby app.rb');
                break;
            }
            case 'php': {
                setField('runCmdPlaceholder', 'php -S localhost:8000 -t public');
                break;
            }
            case 'go': {
                setField('runCmdPlaceholder', './app');
                break;
            }
            case 'static': {
                setField('runCmdPlaceholder', 'npm run build');
                setField('port', 80);
                break;
            }
            default: {
                setField('runCmdPlaceholder', 'npm run start');
            }
        }
    };

    const config = useMemo(() => {
        const cleanConfig: Partial<Service> = {
            name: state.name || 'my-dployr-app',
            source: state.source,
            runtime: state.runtime,
        };

        if (state.workingDir) cleanConfig.working_dir = state.workingDir;
        if (state.runCmd) cleanConfig.run_cmd = state.runCmd;
        if (state.port) cleanConfig.port = Number(state.port);
        if (state.domain) cleanConfig.domain = state.domain;
        if (state.dnsProvider) cleanConfig.dns_provider = state.dnsProvider;
        if (state.remote) cleanConfig.remote = state.remote;

        return cleanConfig;
    }, [state.name, state.source, state.runtime, state.workingDir, state.runCmd, state.port, state.domain, state.dnsProvider, state.remote]);

    const yamlConfig = useMemo(() => {
        return toYaml(config);
    }, [config]);

    const jsonConfig = useMemo(() => {
        return toJson(config);
    }, [config]);

    const currentBlueprint: Blueprint | null = localStorage.getItem('current_blueprint')
        ? JSON.parse(localStorage.getItem('current_blueprint') as string)
        : null;

    const handleBlueprintCopy = async () => {
        try {
            if (!currentBlueprint) return;
            await navigator.clipboard.writeText(blueprintFormat === 'yaml' ? yamlConfig : jsonConfig);
        } catch (error) {
            console.error((error as Error).message || 'An unknown error occoured while retrieving blueprints');
        }
    };

    return {
        currentPage: state.currentPage,
        name: state.name,
        workingDir: state.workingDir,
        outputDir: state.outputDir,
        runtime: state.runtime,
        remote: state.remote,
        ciRemote: state.ciRemote,
        image: state.image,
        spec: state.spec,
        runCmd: state.runCmd,
        source: state.source,
        port: state.port,
        domain: state.domain,
        dnsProvider: state.dnsProvider,
        envVars: state.envVars,
        secrets: state.secrets,
        runCmdPlaceholder: state.runCmdPlaceholder,

        // Blueprint helpers (Service page 3)
        blueprintFormat,
        config,
        yamlConfig,
        jsonConfig,
        handleBlueprintCopy,
        setBlueprintFormat,

        // Error states
        nameError: state.nameError,
        remoteError: state.remoteError,
        workingDirError: state.workingDirError,
        outputDirError: state.outputDirError,
        runtimeError: state.runtimeError,
        runCmdError: state.runCmdError,
        portError: state.portError,
        domainError: state.domainError,
        dnsProviderError: state.dnsProviderError,

        // Handlers
        setField,
        setError,
        setName: (value: string) => setField('name', value),
        setWorkingDir: (value: string) => setField('workingDir', value),
        setOutputDir: (value: string) => setField('outputDir', value),
        setRuntime: (value: Runtime) => setField('runtime', value),
        setRemote: (value: Remote) => setField('remote', value),
        setCiRemote: (value: Remote) => setField('ciRemote', value),
        setImage: (value: string) => setField('image', value),
        setSpec: (value: string) => setField('spec', value),
        setBuildCommand: (value: string) => setField('runCmd', value),
        setSource: (value: ServiceSource) => setField('source', value),
        setPort: (value: number) => setField('port', value),
        setDomain: (value: string) => setField('domain', value),
        setDnsProvider: (value: string) => setField('dnsProvider', value),
        setEnvVars: (value: string) => setField('envVars', value),
        setSecrets: (value: string) => setField('secrets', value),

        // Error setters
        setNameError: (value: string) => setError('nameError', value),
        setRemoteError: (value: string) => setError('remoteError', value),
        setWorkingDirError: (value: string) => setError('workingDirError', value),
        setRuntimeError: (value: string) => setError('runtimeError', value),
        setBuildCommandError: (value: string) => setError('runCmdError', value),
        setPortError: (value: string) => setError('portError', value),
        setDomainError: (value: string) => setError('domainError', value),
        setDnsProviderError: (value: string) => setError('dnsProviderError', value),

        // Utility functions
        getFormData,
        getFormSubmissionData,
        onSourceValueChanged,
        onRemoteValueChanged,
        onRuntimeValueChanged,
        nextPage,
        prevPage,
        validateSkip,
        skipToConfirmation,
        clearAllErrors: () => dispatch({ type: 'CLEAR_ALL_ERRORS' }),
    };
}
