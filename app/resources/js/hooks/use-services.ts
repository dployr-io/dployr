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
    rootDir?: string | null;
    outputDir?: string | null;    
    remote?: Remote | null;
    ciRemote?: Remote | null;
    image?: string | null;
    spec?: string | null;
    buildCommand?: string | null;
    source: ServiceSource;
    port: string;
    domain: string;
    dnsProvider: string;
    envVars: string;
    secrets: string;
    // Error states
    nameError: string;
    rootDirError: string;
    runtimeError: string;
    buildCommandError: string;
    portError: string;
    domainError: string;
    dnsProviderError: string;
}

type ServiceFormAction =
    | { type: 'SET_CURRENT_PAGE'; payload: number }
    | { type: 'SET_NAME'; payload: string }
    | { type: 'SET_ROOT_DIR'; payload: string }
    | { type: 'SET_OUTPUT_DIR'; payload: string }
    | { type: 'SET_RUNTIME'; payload: Runtime }
    | { type: 'SET_REMOTE'; payload: Remote }
    | { type: 'SET_CI_REMOTE'; payload: Remote }
    | { type: 'SET_IMAGE'; payload: string }
    | { type: 'SET_SPEC'; payload: string }
    | { type: 'SET_BUILD_COMMAND'; payload: string }
    | { type: 'SET_SOURCE'; payload: ServiceSource }
    | { type: 'SET_PORT'; payload: string }
    | { type: 'SET_DOMAIN'; payload: string }
    | { type: 'SET_DNS_PROVIDER'; payload: string }
    | { type: 'SET_ENV_VARS'; payload: string }
    | { type: 'SET_SECRETS'; payload: string }
    | { type: 'SET_NAME_ERROR'; payload: string }
    | { type: 'SET_ROOT_DIR_ERROR'; payload: string }
    | { type: 'SET_RUNTIME_ERROR'; payload: string }
    | { type: 'SET_BUILD_COMMAND_ERROR'; payload: string }
    | { type: 'SET_PORT_ERROR'; payload: string }
    | { type: 'SET_DOMAIN_ERROR'; payload: string }
    | { type: 'SET_DNS_PROVIDER_ERROR'; payload: string }
    | { type: 'CLEAR_ALL_ERRORS' }
    | { type: 'NEXT_PAGE' }
    | { type: 'PREV_PAGE' }
    | { type: 'SKIP_TO_CONFIRMATION' }
    | { type: 'SOURCE_CHANGED'; payload: ServiceSource };

const initialState: ServiceFormState = {
    currentPage: 1,
    name: '',
    rootDir: '',
    outputDir: '',
    runtime: 'node-js',

    remote: null,
    ciRemote: null,
    image: '',
    spec: '',

    buildCommand: '',
    source: 'remote',
    port: '',
    domain: '',
    dnsProvider: '',

    envVars: '',
    secrets: '',

    nameError: '',
    rootDirError: '',
    runtimeError: '',
    buildCommandError: '',
    portError: '',
    domainError: '',
    dnsProviderError: '',
};

function serviceFormReducer(state: ServiceFormState, action: ServiceFormAction): ServiceFormState {
    switch (action.type) {
        case 'SET_CURRENT_PAGE':
            return { ...state, currentPage: action.payload };
        case 'SET_NAME':
            return { ...state, name: action.payload, nameError: '' };
        case 'SET_ROOT_DIR':
            return { ...state, rootDir: action.payload, rootDirError: '' };
        case 'SET_OUTPUT_DIR':
            return { ...state, outputDir: action.payload };
        case 'SET_RUNTIME':
            return { ...state, runtime: action.payload, runtimeError: '' };
        case 'SET_REMOTE':
            return { ...state, remote: action.payload };
        case 'SET_CI_REMOTE':
            return { ...state, ciRemote: action.payload };
        case 'SET_IMAGE':
            return { ...state, image: action.payload };
        case 'SET_SPEC':
            return { ...state, spec: action.payload };
        case 'SET_BUILD_COMMAND':
            return { ...state, buildCommand: action.payload, buildCommandError: '' };
        case 'SET_SOURCE':
            return { ...state, source: action.payload };
        case 'SET_PORT':
            return { ...state, port: action.payload, portError: '' };
        case 'SET_DOMAIN':
            return { ...state, domain: action.payload, domainError: '' };
        case 'SET_DNS_PROVIDER':
            return { ...state, dnsProvider: action.payload, dnsProviderError: '' };
        case 'SET_ENV_VARS':
            return { ...state, envVars: action.payload };
        case 'SET_SECRETS':
            return { ...state, secrets: action.payload };
        case 'SET_NAME_ERROR':
            return { ...state, nameError: action.payload };
        case 'SET_ROOT_DIR_ERROR':
            return { ...state, rootDirError: action.payload };
        case 'SET_RUNTIME_ERROR':
            return { ...state, runtimeError: action.payload };
        case 'SET_BUILD_COMMAND_ERROR':
            return { ...state, buildCommandError: action.payload };
        case 'SET_PORT_ERROR':
            return { ...state, portError: action.payload };
        case 'SET_DOMAIN_ERROR':
            return { ...state, domainError: action.payload };
        case 'SET_DNS_PROVIDER_ERROR':
            return { ...state, dnsProviderError: action.payload };
        case 'CLEAR_ALL_ERRORS':
            return {
                ...state,
                nameError: '',
                rootDirError: '',
                runtimeError: '',
                buildCommandError: '',
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
                rootDirError: '',
                runtimeError: '',
                buildCommandError: '',
                portError: '',
                domainError: '',
                dnsProviderError: '',
            };
        case 'SKIP_TO_CONFIRMATION':
            return {
                ...state,
                currentPage: 3,
                nameError: '',
                rootDirError: '',
                runtimeError: '',
                buildCommandError: '',
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
        name: z.string().min(3, 'Name with a minimum of three (3) characters is required'),
        rootDir: z.string().optional(),
        runtime: z.enum(runtimes, 'Runtime value selected must be a supported type. Read the docs for more info'),
        buildCommand: z
            .string()
            .optional()
            .refine(
                (val) =>
                    !val ||
                    /[a-zA-Z]/.test(val),
                {
                    message: "That build command doesn't look right. Be sure to double-check",
                }
            ),
    });

    const page2Schema = z.object({
        port: z.string()
            .min(1, 'Port is required')
            .regex(/^\d+$/, 'Port must be a number')
            .refine(
                (val) => {
                    const num = Number(val);
                    return num >= 1024 && num <= 9999;
                },
                { message: 'Port must be between 1024 and 9999' }
            ),
        domain: z.string()
            .min(1, 'Domain is required')
            .regex(
                /^(?!:\/\/)([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/,
                'Please enter a valid domain (e.g., domain.com)'
            ),
        dnsProvider: z.enum(dnsProviders, 'DNS provider value selected must be a supported type. Read the docs for more info'),
    });

    const validateCurrentPage = () => {
        let result;

        dispatch({ type: 'CLEAR_ALL_ERRORS' });

        if (state.currentPage === 1) {
            result = page1Schema.safeParse({ 
                name: state.name, 
                rootDir: state.rootDir, 
                runtime: state.runtime,
                buildCommand: state.buildCommand
            });

            if (!result.success) {
                const fieldErrors = result.error.flatten().fieldErrors;

                if (fieldErrors.name) dispatch({ type: 'SET_NAME_ERROR', payload: fieldErrors.name[0] });
                if (fieldErrors.rootDir) dispatch({ type: 'SET_ROOT_DIR_ERROR', payload: fieldErrors.rootDir[0] });
                if (fieldErrors.runtime) dispatch({ type: 'SET_RUNTIME_ERROR', payload: fieldErrors.runtime[0] });
                if (fieldErrors.buildCommand) dispatch({ type: 'SET_BUILD_COMMAND_ERROR', payload: fieldErrors.buildCommand[0] });

                return false;
            }
        } else if (state.currentPage === 2) {
            result = page2Schema.safeParse({ 
                port: state.port, 
                domain: state.domain, 
                dnsProvider: state.dnsProvider 
            });

            if (!result.success) {
                const fieldErrors = result.error.flatten().fieldErrors;

                if (fieldErrors.port) dispatch({ type: 'SET_PORT_ERROR', payload: fieldErrors.port[0] });
                if (fieldErrors.domain) dispatch({ type: 'SET_DOMAIN_ERROR', payload: fieldErrors.domain[0] });
                if (fieldErrors.dnsProvider) dispatch({ type: 'SET_DNS_PROVIDER_ERROR', payload: fieldErrors.dnsProvider[0] });

                return false;
            }
        }

        return true;
    };

    const getFormData = () => {
        return {
            name: state.name,
            rootDir: state.rootDir,
            outputDir: state.outputDir,
            runtime: state.runtime,
            remote: state.remote,
            ciRemote: state.ciRemote,
            image: state.image,
            spec: state.spec,
            buildCommand: state.buildCommand,
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

    const onSourceValueChanged = (value: ServiceSource) => {
        dispatch({ type: 'SOURCE_CHANGED', payload: value });
    };

    return {
        currentPage: state.currentPage,
        name: state.name,
        rootDir: state.rootDir,
        outputDir: state.outputDir,
        runtime: state.runtime,
        remote: state.remote,
        ciRemote: state.ciRemote,
        image: state.image,
        spec: state.spec,
        buildCommand: state.buildCommand,
        source: state.source,
        port: state.port,
        domain: state.domain,
        dnsProvider: state.dnsProvider,
        envVars: state.envVars,
        secrets: state.secrets,

        // Individual error states
        nameError: state.nameError,
        rootDirError: state.rootDirError,
        runtimeError: state.runtimeError,
        buildCommandError: state.buildCommandError,
        portError: state.portError,
        domainError: state.domainError,
        dnsProviderError: state.dnsProviderError,

        // Setters for form fields
        setName: (value: string) => dispatch({ type: 'SET_NAME', payload: value }),
        setRootDir: (value: string) => dispatch({ type: 'SET_ROOT_DIR', payload: value }),
        setOutputDir: (value: string) => dispatch({ type: 'SET_OUTPUT_DIR', payload: value }),
        setRuntime: (value: Runtime) => dispatch({ type: 'SET_RUNTIME', payload: value }),
        setRemote: (value: Remote) => dispatch({ type: 'SET_REMOTE', payload: value }),
        setCiRemote: (value: Remote) => dispatch({ type: 'SET_CI_REMOTE', payload: value }),
        setImage: (value: string) => dispatch({ type: 'SET_IMAGE', payload: value }),
        setSpec: (value: string) => dispatch({ type: 'SET_SPEC', payload: value }),
        setBuildCommand: (value: string) => dispatch({ type: 'SET_BUILD_COMMAND', payload: value }),
        setSource: (value: ServiceSource) => dispatch({ type: 'SET_SOURCE', payload: value }),
        setPort: (value: string) => dispatch({ type: 'SET_PORT', payload: value }),
        setDomain: (value: string) => dispatch({ type: 'SET_DOMAIN', payload: value }),
        setDnsProvider: (value: string) => dispatch({ type: 'SET_DNS_PROVIDER', payload: value }),
        setEnvVars: (value: string) => dispatch({ type: 'SET_ENV_VARS', payload: value }),
        setSecrets: (value: string) => dispatch({ type: 'SET_SECRETS', payload: value }),

        // Utility functions
        getFormData,
        handleFormSuccess,
        onSourceValueChanged,
        nextPage,
        prevPage,
        skipToConfirmation,
        handleCreate,
        clearAllErrors: () => dispatch({ type: 'CLEAR_ALL_ERRORS' }),
    };
}
