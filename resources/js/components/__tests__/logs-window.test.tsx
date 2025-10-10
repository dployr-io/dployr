import { LogsWindow } from '@/components/logs-window';
import { Log } from '@/types';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';

vi.mock('@/components/ui/button', () => ({
    Button: (props: any) => <button {...props}>{props.children}</button>,
}));

vi.mock('@/components/ui/input', () => ({
    Input: (props: any) => <input {...props} />,
}));

vi.mock('@/components/ui/separator', () => ({
    Separator: () => <hr data-testid="separator" />,
}));

vi.mock('@/components/ui/dropdown-menu', () => ({
    DropdownMenu: (props: any) => <div>{props.children}</div>,
    DropdownMenuTrigger: (props: any) => <div>{props.children}</div>,
    DropdownMenuContent: (props: any) => <div>{props.children}</div>,
    DropdownMenuItem: (props: any) => (
        <div role="menuitem" onClick={props.onClick} data-testid={`menuitem-${props.children}`}>
            {props.children}
        </div>
    ),
    DropdownMenuSeparator: () => <div data-testid="menu-separator" />,
}));

vi.mock('lucide-react', () => ({
    ChevronDown: () => <span data-testid="chevron-icon" />,
}));

// Sample data
const mockLogs: Log[] = [
    { id: '1', message: 'First log', level_name: 'INFO', datetime: new Date() },
    { id: '2', message: 'Error log', level_name: 'ERROR', datetime: new Date() },
];

describe('LogsWindow', () => {
    const setup = (filtered = mockLogs) => {
        const setSelectedLevel = vi.fn();
        const setSearchQuery = vi.fn();
        const logsEndRef = { current: null };

        render(
            <LogsWindow
                logs={mockLogs}
                filteredLogs={filtered}
                selectedLevel="ALL"
                setSelectedLevel={setSelectedLevel}
                searchQuery=""
                setSearchQuery={setSearchQuery}
                logsEndRef={logsEndRef}
            />,
        );

        return { setSelectedLevel, setSearchQuery };
    };

    test('renders logs and footer count', () => {
        setup();
        expect(screen.getByText('First log')).toBeInTheDocument();
        expect(screen.getByText('Error log')).toBeInTheDocument();
        expect(screen.getByText(/^Showing 2 of 2 log entries/i)).toBeInTheDocument();
    });

    test('shows "No logs entries" when filteredLogs is empty', () => {
        setup([]);
        expect(screen.getByText(/^No logs entries/i)).toBeInTheDocument();
    });

    test('calls setSelectedLevel when a filter is clicked', () => {
        const { setSelectedLevel } = setup();
        fireEvent.click(screen.getByTestId('menuitem-Error'));
        expect(setSelectedLevel).toHaveBeenCalledWith('ERROR');
    });

    test('calls setSearchQuery on input change', () => {
        const { setSearchQuery } = setup();
        const input = screen.getByRole('searchbox');
        fireEvent.change(input, { target: { value: 'test' } });
        expect(setSearchQuery).toHaveBeenCalledWith('test');
    });

    test('renders button with correct label for "ALL"', () => {
        setup();
        expect(screen.getByRole('button', { name: /^All logs/i })).toBeInTheDocument();
    });
});
