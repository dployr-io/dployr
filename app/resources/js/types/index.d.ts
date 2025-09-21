import { InertiaLinkProps } from '@inertiajs/react';
import { LucideIcon } from 'lucide-react';
import { Runtime } from './';


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
    data:  unknown[] | Record<string, unknown>;
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
    description: string;
    id: string;
    name: string;
    repository: string;
    branch: string;
    provider: string;
    commit_message : string;
    avatar_url: string;
}

export type ServiceSource = 'image' | 'remote';

export type Status = 'running' | 'stopped' | 'deploying';

export type AddOn = 'next-js' | 'composer' | 'laravel' | 'hono' | 'bun';

export interface Service {
    id: string;
    name: string;
    status: Status;
    runtime: Runtime;
    region: string;
    source: ServiceSource;
    last_deployed: Date;
}

export type Runtime = (typeof runtimes)[number];

export type Status = 'running' | 'stopped' | 'deploying';
export type AddOn = 'next-js' | 'composer' | 'laravel' | 'hono' | 'bun';
