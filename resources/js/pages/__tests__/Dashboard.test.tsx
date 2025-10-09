import { useProjects } from '@/hooks/use-projects';
import Dashboard from '@/pages/projects/index';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';

vi.mock('@/hooks/use-projects', () => ({
    useProjects: vi.fn(),
}));

vi.mock('@/components/project-card', () => ({
    default: ({ name }: { name: string }) => <div data-testid="project-card">{name}</div>,
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

describe('Dashboard', () => {
    test('renders loading skeletons when isLoading is true', () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [],
            isLoading: true,
        });

        render(<Dashboard />);

        // skeletons should appear
        expect(screen.getAllByText(/^create a new project/i)).toBeTruthy();
    });

    test('renders project cards when projects are loaded', async () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [
                { id: 1, name: 'Project A', description: 'Desc A' },
                { id: 2, name: 'Project B', description: 'Desc B' },
            ],
            isLoading: false,
        });

        render(<Dashboard />);

        const cards = await screen.findAllByTestId('project-card');
        expect(cards).toHaveLength(2);
        expect(screen.getByText('Project A')).toBeInTheDocument();
    });

    test('opens dialog when "Create a New Project" is clicked', async () => {
        (useProjects as vi.Mock).mockReturnValue({
            projects: [],
            isLoading: false,
        });

        render(<Dashboard />);

        const button = screen.getByText(/^Create a New Project/i);
        fireEvent.click(button);

        await waitFor(() => {
            expect(screen.getByTestId('dialog')).toBeInTheDocument();
        });
    });
});
