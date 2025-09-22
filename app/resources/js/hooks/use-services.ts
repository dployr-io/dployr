import type { Remote, Runtime, Service, ServiceSource } from '@/types';
import { dnsProviders, runtimes } from '@/types/runtimes';
import { router } from '@inertiajs/react';
import { useQuery } from '@tanstack/react-query';
import { useReducer } from 'react';
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
    buildCmd?: string | null;
    source: ServiceSource;
    port: string;
    domain: string;
    dnsProvider: string;
    envVars: string;
    secrets: string;
    // Error states
    nameError: string;
    remoteError: string;
    workingDirError: string;
    runtimeError: string;
    buildCmdError: string;
    portError: string;
    domainError: string;
    dnsProviderError: string;
}

type ServiceFormAction =
    | { type: 'SET_CURRENT_PAGE'; payload: number }
    | { type: 'SET_FIELD'; payload: { field: string; value: any } }
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
    buildCmd: '',
    source: 'remote',
    port: '',
    domain: '',
    dnsProvider: '',
    envVars: '',
    secrets: '',
    nameError: '',
    remoteError: '',
    workingDirError: '',
    runtimeError: '',
    buildCmdError: '',
    portError: '',
    domainError: '',
    dnsProviderError: '',
};

function serviceFormReducer(state: ServiceFormState, action: ServiceFormAction): ServiceFormState {
    switch (action.type) {
        case 'SET_CURRENT_PAGE':
            return { ...state, currentPage: action.payload };
        case 'SET_FIELD':
            const { field, value } = action.payload;
            const errorField = `${field}Error` as keyof ServiceFormState;
            return {
                ...state,
                [field]: value,
                ...(errorField in state && { [errorField]: '' }),
            };
        case 'SET_ERROR':
            return { ...state, [action.payload.field]: action.payload.value };
        case 'CLEAR_ALL_ERRORS':
            return {
                ...state,
                nameError: '',
                remoteError: '',
                workingDirError: '',
                runtimeError: '',
                buildCmdError: '',
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
                runtimeError: '',
                buildCmdError: '',
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
                runtimeError: '',
                buildCmdError: '',
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

export function useServices(onCreateServiceCallback?: () => void | null) {
    const [state, dispatch] = useReducer(serviceFormReducer, initialState);

    const page1Schema = z.object({
        name: z.string().min(3, 'Enter a name with at least three (3) characters'),
        workingDir: z.string().optional(),
        runtime: z.enum(runtimes, 'Select a supported runtime. Read the docs for more details.'),
        buildCmd: z
            .string()
            .optional()
            .refine((val) => !val || /[a-zA-Z]/.test(val), {
                message: "Enter a valid build command",
            }),
    });

    const page2Schema = z.object({
        port: z
            .string()
            .min(1, 'Port is required')
            .regex(/^\d+$/, 'Port must be a number')
            .refine(
                (val) => {
                    const num = Number(val);
                    return num >= 1024 && num <= 9999;
                },
                { message: 'Port must be between 1024 and 9999' },
            ),
        domain: z
            .string()
            .min(1, 'Domain is required')
            .regex(/^(?!:\/\/)([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/, 'Enter a valid domain (e.g., domain.com)'),
        dnsProvider: z.enum(dnsProviders, 'Select a supported DNS provider. Read the docs for more details.'),
    });

    const validatePage1 = () => {
        dispatch({ type: 'CLEAR_ALL_ERRORS' });
        let hasErrors = false;

        // Zod validation first
        const zodResult = page1Schema.safeParse({
            name: state.name,
            workingDir: state.workingDir,
            runtime: state.runtime,
            buildCmd: state.buildCmd,
        });

        if (!zodResult.success) {
            const fieldErrors = zodResult.error.flatten().fieldErrors;
            if (fieldErrors.name) dispatch({ type: 'SET_ERROR', payload: { field: 'nameError', value: fieldErrors.name[0] } });
            if (fieldErrors.workingDir) dispatch({ type: 'SET_ERROR', payload: { field: 'workingDirError', value: fieldErrors.workingDir[0] } });
            if (fieldErrors.runtime) dispatch({ type: 'SET_ERROR', payload: { field: 'runtimeError', value: fieldErrors.runtime[0] } });
            if (fieldErrors.buildCmd) dispatch({ type: 'SET_ERROR', payload: { field: 'buildCmdError', value: fieldErrors.buildCmd[0] } });
            hasErrors = true;
        }

        if (state.source === 'remote' && !state.remote) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'remoteError', value: 'Select a remote repository' } });
            hasErrors = true;
        }

        if (
            state.source === 'remote' &&
            (!state.buildCmd || !/[a-zA-Z].*[a-zA-Z]/.test(state.buildCmd)) // Ensure there's at least 2 alphabetic characters
        ) {
            dispatch({ type: 'SET_ERROR', payload: { field: 'buildCmdError', value: 'Enter a valid build command' } });
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

        return true;
    };

    const validateCurrentPage = () => {
        if (state.currentPage === 1) return validatePage1();
        if (state.currentPage === 2) return validatePage2();
        return true;
    };

    const getFormData = () => {
        return {
            name: state.name,
            workingDir: state.workingDir,
            outputDir: state.outputDir,
            runtime: state.runtime,
            remote: state.remote,
            ciRemote: state.ciRemote,
            image: state.image,
            spec: state.spec,
            buildCmd: state.buildCmd,
            source: state.source,
            port: state.port,
            domain: state.domain,
            dnsProvider: state.dnsProvider,
            envVars: state.envVars,
            secrets: state.secrets,
        };
    };

    const getFormSubmissionData = () => {
        return {
            name: state.name,
            workingDir: state.workingDir,
            outputDir: state.outputDir,
            runtime: state.runtime,
            remote: state.remote?.id || null,
            ciRemote: state.ciRemote?.id || null,
            image: state.image,
            spec: state.spec,
            buildCmd: state.buildCmd,
            source: state.source,
            port: state.port,
            domain: state.domain,
            dnsProvider: state.dnsProvider,
            envVars: state.envVars,
            secrets: state.secrets,
        };
    };

    const handleFormSuccess = () => onCreateServiceCallback?.();

    const nextPage = () => {
        if (validateCurrentPage()) {
            dispatch({ type: 'NEXT_PAGE' });
        }
    };

    const prevPage = () => {
        dispatch({ type: 'PREV_PAGE' });
    };

    const skipToConfirmation = () => {
        dispatch({ type: 'SKIP_TO_CONFIRMATION' });
    };

    const handleCreate = () => {
        const formData = getFormData();
        alert(JSON.stringify(formData, null, 2));
        onCreateServiceCallback?.();
    };

    const services = useQuery<Service[]>({
        queryKey: ['projects'],
        queryFn: () =>
            new Promise((resolve) => {
                router.get(
                    '/projects',
                    {},
                    {
                        onSuccess: (page) => resolve(page.props.services as Service[]),
                    },
                );
            }),
        staleTime: 5 * 60 * 1000,
    });

    const setField = (field: string, value: any) => {
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
        buildCmd: state.buildCmd,
        source: state.source,
        port: state.port,
        domain: state.domain,
        dnsProvider: state.dnsProvider,
        envVars: state.envVars,
        secrets: state.secrets,

        // Error states
        nameError: state.nameError,
        remoteError: state.remoteError,
        workingDirError: state.workingDirError,
        runtimeError: state.runtimeError,
        buildCmdError: state.buildCmdError,
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
        setBuildCommand: (value: string) => setField('buildCmd', value),
        setSource: (value: ServiceSource) => setField('source', value),
        setPort: (value: string) => setField('port', value),
        setDomain: (value: string) => setField('domain', value),
        setDnsProvider: (value: string) => setField('dnsProvider', value),
        setEnvVars: (value: string) => setField('envVars', value),
        setSecrets: (value: string) => setField('secrets', value),

        // Error setters
        setNameError: (value: string) => setError('nameError', value),
        setRemoteError: (value: string) => setError('remoteError', value),
        setWorkingDirError: (value: string) => setError('workingDirError', value),
        setRuntimeError: (value: string) => setError('runtimeError', value),
        setBuildCommandError: (value: string) => setError('buildCmdError', value),
        setPortError: (value: string) => setError('portError', value),
        setDomainError: (value: string) => setError('domainError', value),
        setDnsProviderError: (value: string) => setError('dnsProviderError', value),

        // Utility functions
        getFormData,
        getFormSubmissionData,
        handleFormSuccess,
        onSourceValueChanged,
        onRemoteValueChanged,
        nextPage,
        prevPage,
        skipToConfirmation,
        handleCreate,
        clearAllErrors: () => dispatch({ type: 'CLEAR_ALL_ERRORS' }),
    };
}
