import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

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

export function toYaml(obj: Record<string, any>): string {
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
    function processObject(o: Record<string, any>, indent: number) {
        for (const key in o) {
            const value = o[key];
            const indentation = '  '.repeat(indent);
            if (typeof value === 'object' && value !== null) {
                yamlLines.push(`${indentation}${key}:`);
                processObject(value, indent + 1);
            } else {
                yamlLines.push(`${indentation}${key}: ${value}`);
            }
        }
    }
    processObject(parsed, 0);
    return yamlLines.join('\n');
}

export function toJson(obj: Record<string, any>): string {
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
