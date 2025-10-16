import { InertiaLinkProps } from '@inertiajs/react';
import { LucideIcon } from 'lucide-react';
import { Runtime } from '.';
import { dnsProviders } from './runtimes';

export interface Auth {
    user: User;
}

export interface BreadcrumbItem {
    title: string;
    href: string;
}

export interface NavGroup {
    title: string;
    items: NavItem[];
}

export interface NavItem {
    title: string;
    href: NonNullable<InertiaLinkProps['href']>;
    icon?: LucideIcon | null;
    isActive?: boolean;
}

export interface SharedData {
    name: string;
    quote: { message: string; author: string };
    auth: Auth;
    sidebarOpen: boolean;
    [key: string]: unknown;
}

export interface ApiResponse {
    success: boolean;
    data: unknown[] | Record<string, unknown>;
    error?: string;
}

export interface User {
    id: string;
    name: string;
    email: string;
    avatar?: string;
    email_verified_at: string | null;
    created_at: string;
    updated_at: string;
    [key: string]: unknown; // This allows for additional properties...
}

export interface Project {
    id: string;
    name: string;
    description: string;
}

export interface Remote {
    id: string;
    name: string;
    repository: string;
    branch: string;
    provider: string;
    commit_message?: string | null;
    avatar_url?: string | null;
}

export interface DockerImage {
    id: string;
    name: string;
}

export type ServiceSource = 'image' | 'remote';

export type Status = 'pending' | 'in_progress' | 'failed' | 'completed';

export interface Service {
    id: string;
    name: string;
    status: Status;
    runtime: RuntimeObj;
    remote?: Remote | null;
    ci_remote?: Remote | null;
    run_cmd?: string | null;
    working_dir?: string | null;
    static_dir?: string | null;
    env_vars?: Record<string, string> | null;
    build_cmd?: string | null;
    image?: string | null;
    port?: number | null;
    dns_provider?: DnsProvider | null;
    domain?: string | null;
    region: string;
    source: ServiceSource;
    last_deployed: Date;
    created_at: Date;
    updated_at: Date;
}

export interface Blueprint {
    id: string;
    config: Partial<Service>;
    status: Status;
    created_at: Date;
    updated_at: Date;
}

export type Runtime = (typeof runtimes)[number];

export interface RuntimeObj {
    type: Runtime;
    version?: string | null;   
}

export type DnsProvider = (typeof dnsProviders)[number];

export type BlueprintFormat = 'yaml' | 'json';

export type LogChannel = 'production' | 'local';

export type LogLevel = (typeof logLevels)[number];

export interface Log {
    id: string;
    message: string;
    level?: number | null;
    level_name?: LogLevel | null;
    datetime?: Date | null;
    channe?: LogChannel;
    context?: Record<string, unknown> | null;
}
