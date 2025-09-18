import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import AppLayout from '@/layouts/app-layout';
import { logs } from '@/routes';
import { type BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';
import { AlertTriangle, ChevronDown, Circle, Info, XCircle } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Logs',
        href: logs().url,
    },
];

type LogLevel = 'info' | 'warning' | 'error';
type RefreshInterval = 'live' | '10s' | '30s' | '1m';

type Log = {
    id: number;
    level: LogLevel;
    message: string;
    timestamp: string;
};

// Mock log entries for demonstration
const mockLogs = [
    { id: 1, level: 'info', message: 'User authentication successful', timestamp: '2025-01-15 10:30:25' },
    { id: 2, level: 'warning', message: 'High memory usage detected', timestamp: '2025-01-15 10:31:12' },
    { id: 3, level: 'error', message: 'Database connection failed', timestamp: '2025-01-15 10:32:45' },
    { id: 4, level: 'info', message: 'Cache cleared successfully', timestamp: '2025-01-15 10:33:10' },
    { id: 5, level: 'error', message: 'Payment processing failed', timestamp: '2025-01-15 10:34:22' },
] as Log[];

const logLevels: Record<LogLevel, { label: string; icon: typeof Info; color: string }> = {
    info: { label: 'Info', icon: Info, color: '' },
    warning: { label: 'Warning', icon: AlertTriangle, color: 'text-yellow-600' },
    error: { label: 'Error', icon: XCircle, color: 'text-red-500' },
};

const refreshIntervals: Record<RefreshInterval, { label: string; value: number }> = {
    live: { label: 'Live', value: 1000 },
    '10s': { label: 'Every 10 seconds', value: 10000 },
    '30s': { label: 'Every 30 seconds', value: 30000 },
    '1m': { label: 'Every minute', value: 60000 },
};

export default function Logs() {
    const [logs, setLogs] = useState<Log[]>(mockLogs);
    const [filteredLogs, setFilteredLogs] = useState<Log[]>(mockLogs);
    const [selectedLevel, setSelectedLevel] = useState<'all' | LogLevel>('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [refreshInterval, setRefreshInterval] = useState<RefreshInterval>('live');
    const [isLive, setIsLive] = useState(false);
    const logsEndRef = useRef<HTMLDivElement | null>(null);
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    // Filter logs based on level and search query
    useEffect(() => {
        let filtered = logs;

        if (selectedLevel !== 'all') {
            filtered = filtered.filter((log) => log.level === selectedLevel);
        }

        if (searchQuery) {
            filtered = filtered.filter((log) => log.message.toLowerCase().includes(searchQuery.toLowerCase()));
        }

        setFilteredLogs(filtered);
    }, [logs, selectedLevel, searchQuery]);

    // Auto-scroll to bottom when new logs arrive
    useEffect(() => {
        logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [filteredLogs]);

    // Mock SSE connection - replace with actual SSE implementation
    useEffect(() => {
        if (isLive) {
            intervalRef.current = setInterval(() => {
                // Simulate new log entry
                const newLog: Log = {
                    id: Date.now(),
                    level: (['info', 'warning', 'error'] as LogLevel[])[Math.floor(Math.random() * 3)],
                    message: `New log entry at ${new Date().toLocaleTimeString()}`,
                    timestamp: new Date().toLocaleString('sv-SE').replace('T', ' ').substring(0, 19),
                };

                setLogs((prev) => [...prev, newLog]);
            }, refreshIntervals[refreshInterval].value);
        } else {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        }

        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        };
    }, [isLive, refreshInterval]);

    const LogEntry = ({ log }: { log: Log }) => {
        const levelConfig = logLevels[log.level];
        const IconComponent = levelConfig.icon;

        return (
            <div className="flex items-start gap-3 border-b p-3">
                <IconComponent className={`size-4 ${levelConfig.color}`} />

                <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between">
                        <span className={`text-xs font-medium ${levelConfig.color} uppercase`}>{levelConfig.label}</span>
                        <span className="text-xs text-muted-foreground">{log.timestamp}</span>
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">{log.message}</p>
                </div>
            </div>
        );
    };

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Logs" />
            <div className="flex min-h-0 h-full flex-col gap-4 overflow-y-hidden rounded-xl p-4">
                <div className="flex min-h-0 flex-1 auto-rows-min gap-4 p-8">
                    <div className="flex min-h-0 w-full flex-1 flex-col gap-6">
                        <div className="flex min-h-0 flex-1 flex-col overflow-hidden border border-sidebar-border">
                            <div className="flex flex-shrink-0 gap-2 p-2 bg-neutral-50 dark:bg-neutral-900">
                                {/* Log Level Filter */}
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button
                                            size="default"
                                            variant={'outline'}
                                            className="group min-w-40 text-sidebar-accent-foreground data-[state=open]:bg-sidebar-accent"
                                        >
                                            {selectedLevel === 'all' ? 'All logs' : logLevels[selectedLevel]?.label}
                                             <ChevronDown className="ml-auto size-4 transition-transform group-data-[state=open]:rotate-180" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width] min-w-40 rounded-lg" align="start">
                                        <DropdownMenuItem onClick={() => setSelectedLevel('info')}>Info</DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => setSelectedLevel('warning')}>Warning</DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => setSelectedLevel('error')}>Error</DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => setSelectedLevel('all')}>All logs</DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>

                                {/* Search Input */}
                                <Input
                                    id="search"
                                    type="search"
                                    name="search"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    autoFocus
                                    tabIndex={1}
                                    autoComplete="search"
                                    placeholder="Search for a log entry..."
                                    className='dark:bg-neutral-950'
                                />

                                {/* Refresh Interval */}
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button
                                            size="default"
                                            variant={'outline'}
                                            className="group min-w-40 text-sidebar-accent-foreground data-[state=open]:bg-sidebar-accent"
                                        >
                                            {refreshIntervals[refreshInterval]?.label}
                                            <ChevronDown className="ml-auto size-4 transition-transform group-data-[state=open]:rotate-180" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent className="w-[--radix-dropdown-menu-trigger-width] min-w-40 rounded-lg" align="end">
                                        <DropdownMenuItem
                                            onClick={() => {
                                                setRefreshInterval('10s');
                                                setIsLive(true);
                                            }}
                                        >
                                            Every 10 seconds
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            onClick={() => {
                                                setRefreshInterval('30s');
                                                setIsLive(true);
                                            }}
                                        >
                                            Every 30 seconds
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            onClick={() => {
                                                setRefreshInterval('1m');
                                                setIsLive(true);
                                            }}
                                        >
                                            Every minute
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem
                                            onClick={() => {
                                                setRefreshInterval('live');
                                                setIsLive(true);
                                            }}
                                        >
                                            Live
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>

                                {/* Live Toggle */}
                                <Button variant={isLive ? 'default' : 'outline'} onClick={() => setIsLive(!isLive)} className="min-w-20">
                                    <Circle className={`size-3 ${isLive ? 'animate-pulse fill-current' : ''}`} />
                                    {isLive ? 'Live' : 'Paused'}
                                </Button>
                            </div>
                            <Separator />
                            <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
                                <div className="min-h-0 flex-1 overflow-y-auto">
                                    {filteredLogs.length === 0 ? (
                                        <div className="flex h-32 items-center justify-center text-gray-500">No logs found</div>
                                    ) : (
                                        filteredLogs.map((log) => <LogEntry key={log.id} log={log} />)
                                    )}
                                    <div ref={logsEndRef} />
                                </div>
                                <div className="border-t border-accent p-2 text-center text-xs text-muted-foreground bg-neutral-50 dark:bg-neutral-800">
                                    Showing {filteredLogs.length} of {logs.length} log entries
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
