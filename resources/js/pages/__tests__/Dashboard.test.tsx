import Dashboard from '@/pages/projects/index';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';

vi.mock('@/hooks/use-projects', () => ({
    useProjects: vi.fn(),
}));

vi.mock('@/components/project-card', () => ({
    default: ({ project }: { project: { name: string; description: string } }) => (
        <div data-testid="project-card">
            {project.name} - {project.description}
        </div>
    ),
}));

vi.mock('@/components/project-create-dialog', () => ({
    default: ({ open }: { open: boolean }) => (open ? <div data-testid="dialog">Dialog Open</div> : null),
}));

vi.mock('@/layouts/app-layout', () => ({
    default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('@inertiajs/react', () => ({
    Head: ({ title }: { title: string }) => <title>{title}</title>,
}));

import { useProjects } from '@/hooks/use-projects';

describe('Dashboard', () => {
    test('renders loading skeletons when isLoading is true', () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [],
            isLoading: true,
        });

        render(<Dashboard />);

        expect(screen.getAllByText(/^create a new project/i)).toBeTruthy();
    });

    test('renders project cards when projects are loaded', async () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [
                { id: '1', name: 'Project A', description: 'Desc A' },
                { id: '2', name: 'Project B', description: 'Desc B' },
            ],
            isLoading: false,
        });

        render(<Dashboard />);

        const cards = await screen.findAllByTestId('project-card');
        expect(cards).toHaveLength(2);
        await screen.findByText((content) => content.includes('Project A'));
        await screen.findByText((content) => content.includes('Desc A'));
    });

    test('opens dialog when "Create a New Project" is clicked', async () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [],
            isLoading: false,
        });

        render(<Dashboard />);

        const button = screen.getByText(/^Create a New Project/i);
        fireEvent.click(button);

        expect(await screen.findByTestId('dialog')).toBeInTheDocument();
    });
});
