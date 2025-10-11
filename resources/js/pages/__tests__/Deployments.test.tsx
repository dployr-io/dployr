import Deployments from '@/pages/deployments';
import type { Project, Remote } from '@/types';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, test, vi } from 'vitest';

vi.mock('@inertiajs/react', async () => {
    const actual = await vi.importActual('@inertiajs/react');
    return {
        ...actual,
        usePage: vi.fn(() => ({
            url: '/deployments',
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
        Head: () => null,
        router: {
            visit: vi.fn(),
        },
    };
});

vi.mock('@/hooks/use-deployments', () => ({
    useDeployments: vi.fn(() => ({ deployments: [], isLoading: false })),
}));

vi.mock('@/hooks/use-projects', () => ({
    useProjects: vi.fn(() => ({ defaultProject: null })),
}));

vi.mock('@/hooks/use-remotes', () => ({
    useRemotes: vi.fn(() => ({ remotes: [], isLoading: false })),
}));

import { useDeployments } from '@/hooks/use-deployments';
import { useProjects } from '@/hooks/use-projects';
import { useRemotes } from '@/hooks/use-remotes';
import { router } from '@inertiajs/react';

const deployments = [
    {
        id: '1',
        status: 'completed',
        created_at: new Date('2024-01-01T10:00:00Z').toISOString(),
        updated_at: new Date('2024-01-01T10:05:00Z').toISOString(),
        config: {
            name: 'my-deployment',
            runtime: 'node-js',
            remote: { provider: 'github', name: 'org', repository: 'repo' },
            run_cmd: 'npm start',
        },
    },
];

const remotes = [
    {
        id: 'r1',
        name: 'org',
        provider: 'github',
        repository: 'repo',
    },
] as Remote[];

describe('Deployments', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    test('renders deployments table', async () => {
        (useDeployments as vi.Mock).mockReturnValue({
            deployments: deployments,
            normalizedDeployments: deployments,
            isLoading: false,
        });

        (useProjects as vi.Mock).mockReturnValue({
            defaultProject: { id: '42', name: 'Test Project' } as Project,
        });

        (useRemotes as vi.Mock).mockReturnValue({
            remotes: remotes,
            isLoading: false,
        });

        render(<Deployments />);
        expect(await screen.findByText('my-deployment')).toBeInTheDocument();
        expect(await screen.findByText('node-js')).toBeInTheDocument();
    });

    test('empty state shows correct buttons', () => {
        (useDeployments as vi.Mock).mockReturnValue({
            deployments: [],
            isLoading: false,
        });

        (useProjects as vi.Mock).mockReturnValue({
            defaultProject: { id: '42', name: 'Test Project' } as Project,
        });

        (useRemotes as vi.Mock).mockReturnValue({
            remotes: [],
            isLoading: false,
        });

        render(<Deployments />);
        expect(screen.getByText('No Deployments Yet')).toBeInTheDocument();
        expect(screen.getByText('Deploy Service')).toBeInTheDocument();
    });

    test('clicking a deployment calls routes to its page when clicked', async () => {
        (useDeployments as vi.Mock).mockReturnValue({
            deployments: deployments,
            normalizedDeployments: deployments,
            isLoading: false,
        });

        (useProjects as vi.Mock).mockReturnValue({
            defaultProject: { id: '42', name: 'Test Project' } as Project,
        });

        (useRemotes as vi.Mock).mockReturnValue({
            remotes: remotes,
            isLoading: false,
        });

        render(<Deployments />);

        const deploymentText = await screen.findByText('my-deployment');
        const tableRow = deploymentText.closest('tr');

        expect(tableRow).toBeInTheDocument();

        if (tableRow) {
            await userEvent.click(tableRow);
        }

        expect(router.visit).toHaveBeenCalledWith(expect.stringContaining('/deployments/1'));
    });
});
