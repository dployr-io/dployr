import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { Log, LogLevel } from '@/types';
import { ChevronDown } from 'lucide-react';

interface Props {
    logs: Log[];
    filteredLogs: Log[];
    selectedLevel: 'all' | LogLevel;
    setSelectedLevel: (level: 'all' | LogLevel) => void;
    searchQuery: string;
    setSearchQuery: (query: string) => void;
    logsEndRef: React.RefObject<HTMLDivElement | null>;
}

const logLevels: Record<LogLevel, string> = {
    info: 'Info',
    warning: 'Warning',
    error: 'Error',
};

const LogEntry = ({ log }: { log: Log }) => {
    return (
        <div className="flex items-start gap-3 border-b p-3">
            <div className="flex items-center gap-2">
                {log.datetime && (
                    <span
                        className={`text-xs ${(() => {
                            switch (log.level_name) {
                                case 'INFO':
                                    return 'text-muted-foreground';
                                case 'WARNING':
                                    return 'text-orange-400';
                                case 'ERROR':
                                case 'CRITICAL':
                                case 'ALERT':
                                case 'EMERGENCY':
                                    return 'text-red-500';
                            }
                        })()}`}
                    >
                        {log.datetime.toLocaleTimeString()}
                    </span>
                )}
                <span
                    className={`text-sm ${(() => {
                        switch (log.level_name) {
                            case 'INFO':
                                return 'text-muted-foreground';
                            case 'WARNING':
                                return 'text-orange-400';
                            case 'ERROR':
                            case 'CRITICAL':
                            case 'ALERT':
                            case 'EMERGENCY':
                                return 'text-red-500';
                        }
                    })()}`}
                >
                    {log.message}
                </span>
            </div>
        </div>
    );
};

export function LogsWindow({ logs, filteredLogs, selectedLevel, setSelectedLevel, searchQuery, setSearchQuery, logsEndRef }: Props) {
    return (
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-sidebar-border">
            <div className="flex flex-shrink-0 gap-2 bg-neutral-50 p-2 dark:bg-neutral-900">
                {/* Log Level Filter */}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            size="default"
                            variant={'outline'}
                            className="group min-w-40 text-sidebar-accent-foreground data-[state=open]:bg-sidebar-accent"
                        >
                            {selectedLevel === 'all' ? 'All logs' : logLevels[selectedLevel]}
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
                    className="dark:bg-neutral-950"
                />
            </div>
            <Separator />
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
                <div className="min-h-0 flex-1 overflow-y-auto">
                    {filteredLogs.length === 0 ? (
                        <div className="flex h-32 items-center justify-center">
                            <p className="text-sm text-muted-foreground">No logs entries</p>
                        </div>
                    ) : (
                        filteredLogs.map((log) => <LogEntry key={log.id} log={log} />)
                    )}
                    <div ref={logsEndRef} />
                </div>
                <div className="border-t border-accent bg-neutral-50 p-2 dark:bg-neutral-800">
                    <p className="text-center text-xs text-muted-foreground">
                        Showing {filteredLogs.length} of {logs.length} log entries
                    </p>
                </div>
            </div>
        </div>
    );
}
