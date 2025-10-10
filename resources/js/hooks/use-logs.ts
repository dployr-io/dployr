import { parseLog } from '@/lib/utils';
import type { Blueprint, LogLevel, Service } from '@/types';
import { Log } from '@/types';
import { useEffect, useRef, useState } from 'react';

export function useLogs(filterItem?: Blueprint | Service | null) {
    const [logs, setLogs] = useState<Log[]>([]);
    const [filteredLogs, setFilteredLogs] = useState<Log[]>([]);
    const [selectedLevel, setSelectedLevel] = useState<'ALL' | LogLevel>('ALL');
    const [searchQuery, setSearchQuery] = useState('');
    const logsEndRef = useRef<HTMLDivElement | null>(null);
    const hasSeenValidLog = useRef(false);

    useEffect(() => {
        // Reset the flag when filterItem changes
        hasSeenValidLog.current = false;

        const eventSource = new EventSource('/logs/stream');

        eventSource.onmessage = (event) => {
            const log = parseLog(event.data);

            if (filterItem) {
                const logTime = log.datetime!.getTime();
                const createdTime = new Date(filterItem.created_at).getTime();

                // If log has a timestamp, check if it's in range
                if (logTime) {
                    // Filter out logs before creation time
                    if (logTime < createdTime) {
                        return;
                    }

                    const status = filterItem.status;

                    if (status === 'completed' || status === 'failed') {
                        const logTimeInSecs = Math.floor(new Date(logTime).getTime() / 1000);
                        const updatedTimeInSecs = Math.floor(new Date(filterItem.updated_at).getTime() / 1000);

                        if (logTimeInSecs > updatedTimeInSecs) {
                            return;
                        }
                    }

                    // This log is in range, mark that we've seen a valid log
                    hasSeenValidLog.current = true;
                } else {
                    // Log has no timestamp - only include if
                    // we've already seen a valid log in range
                    // This is useful in tracking stack traces that
                    // sometimes get broken into chucks in transit
                    if (!hasSeenValidLog.current) {
                        return;
                    }
                }
            }

            setLogs((prevLogs) => [...prevLogs, log]);
        };

        eventSource.onerror = (error) => {
            console.error('SSE connection error:', error);
            eventSource.close();
        };

        return () => {
            eventSource.close();
        };
    }, [filterItem]);

    // Filter logs based on level and search query
    useEffect(() => {
        let filtered = logs;

        if (selectedLevel !== 'ALL') {
            filtered = filtered.filter((log) => log.level_name === selectedLevel);
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

    return {
        logs,
        filteredLogs,
        selectedLevel,
        searchQuery,
        logsEndRef,
        setSelectedLevel,
        setSearchQuery,
    };
}
