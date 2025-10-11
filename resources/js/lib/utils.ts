import { Log, LogLevel } from '@/types';
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { v4 as uuidv4 } from 'uuid';

/**
 * Merge Tailwind and custom class names.
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 * Convert a string to uppercase words, replacing underscores with spaces.
 */
export function toWordUpperCase(value: string) {
    return value
        .replace(/_/g, ' ')
        .split(' ')
        .map((word) => word.toUpperCase())
        .join(' ');
}

export function toYaml(obj: Record<string, unknown>): string {
    // Parse nested JSON strings first
    const parsed = JSON.parse(JSON.stringify(obj), (key, value) => {
        if (typeof value === 'string') {
            try {
                return JSON.parse(value);
            } catch {
                return value;
            }
        }
        return value;
    });

    const yamlLines: string[] = [];
    function processObject(o: Record<string, unknown>, indent: number) {
        for (const key in o) {
            const value = o[key];
            const indentation = '  '.repeat(indent);
            if (typeof value === 'object' && value !== null) {
                yamlLines.push(`${indentation}${key}:`);
                processObject(value as Record<string, unknown>, indent + 1);
            } else {
                yamlLines.push(`${indentation}${key}: ${value}`);
            }
        }
    }
    processObject(parsed, 0);
    return yamlLines.join('\n');
}

export function toJson(obj: Record<string, unknown>): string {
    const parsed = JSON.parse(JSON.stringify(obj), (key, value) => {
        if (typeof value === 'string') {
            try {
                return JSON.parse(value);
            } catch {
                return value;
            }
        }
        return value;
    });

    return JSON.stringify(parsed, null, 2);
}

/** Parse a log line into a Log object.
 * e.g.: [2025-10-07 02:14:06] local.ERROR: Something went wrong {"context":"value"}
 * defaults to INFO level_name.
 */
export function parseLog(raw: string): Log {
    const id = uuidv4();
    if (!raw) return { id, message: '', level_name: 'INFO' };

    try {
        const logData = JSON.parse(JSON.parse(raw).message);

        if (logData.level_name && logData.message) {
            return {
                id,
                message: logData.message,
                level: parseInt(logData.level),
                level_name: logData.level_name.toUpperCase() as LogLevel,
                datetime: logData.datetime ? new Date(logData.datetime) : undefined,
                context: logData.context,
            };
        }
    } catch (error) {
        console.error((error as Error).message || "An unknown error occoured while parsing log")
    }

    return { id, message: raw, level_name: 'INFO' };
}

/** Tiny wrapper for fetch api to timeout requests and avoid zombie processes */
export function fetchWithTimeout(url: string, options = {}, timeout = 10000) {
    return Promise.race([fetch(url, options), new Promise((_, reject) => setTimeout(() => reject(new Error('Request timed out')), timeout))]);
}
