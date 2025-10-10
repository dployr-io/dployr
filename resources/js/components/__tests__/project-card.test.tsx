import ProjectCard from '@/components/project-card';
import { Project } from '@/types';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, test, vi } from 'vitest';

vi.mock('@inertiajs/react', () => ({
    Link: ({ children, onClick, ...props }: any) => (
        <a {...props} onClick={onClick}>
            {children}
        </a>
    ),
}));

describe('ProjectCard', () => {
    beforeEach(() => {
        localStorage.clear();
    });

    const mockProject: Project = {
        id: '42',
        name: 'Test Project',
        description: 'This is a test project',
    };

    test('renders project name and description', () => {
        render(<ProjectCard project={mockProject} />);

        expect(screen.getByText('Test Project')).toBeInTheDocument();
        expect(screen.getByText('This is a test project')).toBeInTheDocument();
    });

    test('links to the correct project URL', () => {
        render(<ProjectCard project={mockProject} />);

        const link = screen.getByRole('link');
        expect(link).toHaveAttribute('href', '/projects/42');
    });

    test('stores project ID in localStorage when clicked', async () => {
        render(<ProjectCard project={mockProject} />);

        const link = screen.getByRole('link');
        await userEvent.click(link);

        expect(localStorage.getItem('current_project')).toBe('42');
    });
});
