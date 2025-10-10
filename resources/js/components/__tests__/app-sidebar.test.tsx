import { AppSidebar } from '@/components/app-sidebar';
import { render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';
import { SidebarProvider } from '../ui/sidebar';

vi.mock('@inertiajs/react', async () => {
    const actual = await vi.importActual('@inertiajs/react');
    return {
        ...actual,
        usePage: vi.fn(() => ({
            url: '/projects',
            props: {
                sidebarOpen: true,
                auth: {
                    user: {
                        id: 1,
                        name: 'Test User',
                        email: 'test@example.com',
                    },
                },
            },
        })),
        Link: ({ href, children }: any) => <a href={href}>{children}</a>,
    };
});

describe('AppSidebar', () => {
    test('renders main navigation links', () => {
        render(
            <SidebarProvider>
                <AppSidebar />
            </SidebarProvider>,
        );

        expect(screen.getByRole('link', { name: /projects/i })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /deployments/i })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /logs/i })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /console/i })).toBeInTheDocument();
    });

    test('renders main secondary links', () => {
        render(
            <SidebarProvider>
                <AppSidebar />
            </SidebarProvider>,
        );

        expect(screen.getByRole('link', { name: /remotes/i })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /images/i })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /specs/i })).toBeInTheDocument();
    });
});
