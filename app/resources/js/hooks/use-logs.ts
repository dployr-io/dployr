import { Log } from "@/types";
import { useState, useRef, useEffect } from "react";

export function useLogs() {
    const [logs, setLogs] = useState<Log[]>([]);
    const logsEndRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        const eventSource = new EventSource('/logs/stream');

        eventSource.onmessage = (event) => {
            const newLog = JSON.parse(event.data);
            setLogs((prevLogs) => [...prevLogs, newLog]);
        };

        return () => {
            eventSource.close();
        };
    }, []);

    return { logs, logsEndRef };
}