import AppLayout from '@/layouts/app-layout';
import { console } from '@/routes';
import { type BreadcrumbItem } from '@/types';
import { Head } from '@inertiajs/react';
import { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Console',
        href: console().url,
    },
]

export default function Console() {
    const terminalRef = useRef<HTMLDivElement>(null);
    const terminal = useRef<Terminal | null>(null);
    const fitAddon = useRef<FitAddon | null>(null);

    useEffect(() => {
        if (!terminalRef.current) return;

        // Initialize terminal
        terminal.current = new Terminal({
            cursorBlink: true,
            theme: {
                background: '#171717',
                foreground: '#ffffff',
                cursor: '#ffffff',
            },
        });

        // Initialize fit addon
        fitAddon.current = new FitAddon();
        terminal.current.loadAddon(fitAddon.current);

        // Open terminal in the container
        terminal.current.open(terminalRef.current);

        // Focus and setup
        terminal.current.focus();

        // Enable basic shell-like behavior
        let currentLine = '';

        terminal.current.onData((data: string) => {
            // Handle backspace
            if (data === '\u007F') {
                if (currentLine.length > 0) {
                    currentLine = currentLine.slice(0, -1);
                    terminal.current?.write('\b \b');
                }
                return;
            }

            // Handle enter
            if (data === '\r') {
                terminal.current?.writeln('');
                if (currentLine.trim()) {
                    terminal.current?.writeln(`Command: ${currentLine}`);
                }
                currentLine = '';
                terminal.current?.write('$ ');
                return;
            }

            // Handle regular characters
            if (data >= ' ' || data === '\t') {
                currentLine += data;
                terminal.current?.write(data);
            }
        });

        // Initial prompt
        terminal.current.writeln('Welcome to Terminal Console');
        terminal.current.write('$ ');

        // Fit terminal to container
        setTimeout(() => {
            fitAddon.current?.fit();
        }, 0);

        // Setup resize observer
        const resizeObserver = new ResizeObserver(() => {
            fitAddon.current?.fit();
        });

        resizeObserver.observe(terminalRef.current);

        // Cleanup
        return () => {
            resizeObserver.disconnect();
            terminal.current?.dispose();
        };
    }, []);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Logs" />
            <style dangerouslySetInnerHTML={{
                __html: `
                    .terminal-scrollbar::-webkit-scrollbar {
                        width: 8px;
                    }
                    .terminal-scrollbar::-webkit-scrollbar-track {
                        background: #262626;
                        border-radius: 4px;
                    }
                    .terminal-scrollbar::-webkit-scrollbar-thumb {
                        background: #525252;
                        border-radius: 4px;
                    }
                    .terminal-scrollbar::-webkit-scrollbar-thumb:hover {
                        background: #737373;
                    }
                    .terminal-scrollbar {
                        scrollbar-width: thin;
                        scrollbar-color: #525252 #262626;
                    }
                `
            }} />
            <div className="flex min-h-0 h-full flex-col gap-4 overflow-y-hidden rounded-xl p-4">
                <div className="flex min-h-0 flex-1 auto-rows-min gap-4 p-8">
                    <div className="flex min-h-0 w-full flex-1 flex-col gap-6">
                        <div className="flex min-h-0 flex-1 flex-col overflow-hidden border border-sidebar-border rounded-xl">
                            <div className="flex h-full min-h-0 flex-col">
                                <div className="flex h-11 items-center justify-between border-b border-sidebar-border bg-neutral-50 px-4 py-2 dark:bg-neutral-800">
                                    <span className="text-sm font-medium text-neutral-700 dark:text-neutral-300">Terminal Console</span>
                                </div>
                                <div
                                    ref={terminalRef}
                                    className="flex-1 px-2 min-h-0 bg-neutral-900 terminal-scrollbar"
                                    onClick={() => terminal.current?.focus()}
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
