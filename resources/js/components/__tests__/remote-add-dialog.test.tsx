import RemoteAddDialog from '@/components/remote-add-dialog';
import { useRemotes } from '@/hooks/use-remotes';
import { Form } from '@inertiajs/react';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, test, vi } from 'vitest';

vi.mock('@inertiajs/react', () => ({
    Form: vi.fn(({ children, onSuccess }) =>
        children({
            processing: false,
            errors: {},
        }),
    ),
}));

vi.mock('@/hooks/use-remotes', () => ({
    useRemotes: vi.fn(),
}));

vi.mock('@radix-ui/react-portal', () => ({
    Root: ({ children }: any) => children,
}));

describe('RemoteAddDialog', () => {
    const setOpen = vi.fn();
    const setSelectedBranch = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();

        (Form as vi.Mock).mockImplementation(({ children }) =>
            children({
                processing: false,
                errors: {},
            }),
        );

        (useRemotes as vi.Mock).mockReturnValue({
            branches: [],
            searchComplete: false,
            validationError: '',
            remoteRepo: '',
            selectedBranch: '',
            setRemoteRepo: vi.fn(),
            setSelectedBranch: vi.fn(),
            getFormAction: vi.fn(() => '/mock-action'),
            getFormData: vi.fn(() => ({})),
            handleFormSuccess: vi.fn(),
        });
    });

    test('renders dialog content', () => {
        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        expect(screen.getByText('Add Remote')).toBeInTheDocument();
        expect(screen.getByText(/^Enter a link to remote repository/i)).toBeInTheDocument();
    });

    test('renders input and updates value', () => {
        const setRemoteRepo = vi.fn();
        (useRemotes as vi.Mock).mockReturnValue({
            ...useRemotes(),
            setRemoteRepo,
        });

        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        const input = screen.getByPlaceholderText(/^github.com\/username/i);
        fireEvent.change(input, { target: { value: 'github.com/test/repo.git' } });

        expect(setRemoteRepo).toHaveBeenCalledWith('github.com/test/repo.git');
    });

    test('shows validation error message', () => {
        (useRemotes as vi.Mock).mockReturnValue({
            ...useRemotes(),
            validationError: 'Invalid repository URL',
        });

        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        expect(screen.getByText('Invalid repository URL')).toBeInTheDocument();
    });

    test('shows loader when processing', () => {
        (Form as vi.Mock).mockImplementation(({ children }) =>
            children({
                processing: true,
                errors: {},
            }),
        );

        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        expect(screen.getByText(/^Fetching remote repository.../)).toBeInTheDocument();
    });

    test('renders branch dropdown when branches exist', async () => {
        (useRemotes as vi.Mock).mockReturnValue({
            ...useRemotes(),
            branches: ['main', 'dev'],
            setSelectedBranch,
            processing: false,
        });

        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        const dropdownButton = await screen.findByRole('button', { name: /^select branch/i });
        expect(dropdownButton).toBeInTheDocument();

        await userEvent.click(dropdownButton);

        expect(await screen.findByText('main', { exact: false })).toBeInTheDocument();
        expect(await screen.findByText('dev', { exact: false })).toBeInTheDocument();
    });

    test('calls setOpen(false) when cancel is clicked', () => {
        render(<RemoteAddDialog open={true} setOpen={setOpen} />);

        fireEvent.click(screen.getByRole('button', { name: /^cancel/i }));

        expect(setOpen).toHaveBeenCalledWith(false);
    });
});
