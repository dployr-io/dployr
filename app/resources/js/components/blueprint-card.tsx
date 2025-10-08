import { getRuntimeIcon } from '@/lib/runtime-icon';
import type { Blueprint } from '@/types';
import { Link } from '@inertiajs/react';

export default function BlueprintCard(blueprint: Blueprint) {
    const config = typeof blueprint.config === 'string' ? JSON.parse(blueprint.config) : blueprint.config;

    const status = typeof blueprint.status === 'string' ? blueprint.status : '';

    return (
        <Link
            href={`/deployments/${blueprint.id}`}
            className="flex h-28 flex-col justify-between rounded-xl border border-sidebar-border/70 p-4 no-underline hover:cursor-pointer hover:border-muted-foreground dark:border-sidebar-border dark:hover:border-muted-foreground"
        >
            <div>
                <p className="mb-1 truncate text-base font-semibold">{config?.name || 'Deployment'}</p>
                <div className="flex items-center gap-2">
                    <span className="flex-shrink-0">{getRuntimeIcon(config?.runtime)}</span>
                    <span className="truncate text-sm text-muted-foreground">{config?.runtime || 'Deployment'}</span>
                </div>
            </div>
            <div className="flex justify-end">
                <span
                    className={`inline-block rounded-full px-2 py-1 text-xs font-semibold ${(() => {
                        switch (status) {
                            case 'completed':
                                return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
                            case 'pending':
                                return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200';
                            case 'failed':
                                return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
                            case 'in_progress':
                                return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
                            default:
                                return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
                        }
                    })()}`}
                >
                    {status.charAt(0).toUpperCase() + status.slice(1)}
                </span>
            </div>
        </Link>
    );
}
