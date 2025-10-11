import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { toWordUpperCase } from '@/lib/utils';
import { config } from '@/routes';
import { type BreadcrumbItem, type SharedData } from '@/types';
import { Head, router, usePage } from '@inertiajs/react';
import { useState } from 'react';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Configuration',
        href: config().url,
    },
];

export default function Config() {
    const { config }: Record<string, any> = usePage().props;
    const [editingKey, setEditingKey] = useState<string | null>(null);
    const [editValue, setEditValue] = useState<string>('');

    const handleEdit = (key: string) => {
        setEditingKey(key);
        setEditValue('');
    };

    const handleSave = (key: string) => {
        router.post(
            '/system/config',
            {
                [key]: editValue,
            },
            {
                onSuccess: () => {
                    setEditingKey(null);
                    setEditValue('');
                },
            },
        );
    };

    const handleCancel = () => {
        setEditingKey(null);
        setEditValue('');
    };

    const handleKeyPress = (e: React.KeyboardEvent, key: string) => {
        if (e.key === 'Enter') {
            handleSave(key);
        } else if (e.key === 'Escape') {
            handleCancel();
        }
    };

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Configuration" />

            <SettingsLayout>
                <div className="space-y-4">
                    <HeadingSmall title="Configuration settings" description="Update your dployr configuration settings" />

                    <Table className="overflow-hidden rounded-t-lg">
                        <TableHeader className="gap-2 rounded-t-xl bg-neutral-50 p-2 dark:bg-neutral-900">
                            <TableRow>
                                <TableHead className="ml w-[200px]">Key</TableHead>
                                <TableHead>Last Updated</TableHead>
                                <TableHead className="w-[100px] text-right"></TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {Object.entries(config || {}).map(([key, value]) => {
                                const isSet =
                                    value !== null &&
                                    value !== undefined &&
                                    value !== '' &&
                                    value !== false &&
                                    value !== 0;
                                
                                let lastUpdated: string | null = null;
                                if (isSet) {
                                    if (
                                        typeof value === 'object' &&
                                        value !== null &&
                                        Object.prototype.hasOwnProperty.call(value, 'updated_at') &&
                                        typeof (value as { updated_at?: unknown }).updated_at === 'string'
                                    ) {
                                        lastUpdated = (value as { updated_at: string }).updated_at;
                                    } else {
                                        lastUpdated = 'Recently';
                                    }
                                } else {
                                    lastUpdated = null;
                                }

                                return (
                                    <TableRow key={key}>
                                        <TableCell className="font-medium">{toWordUpperCase(key)}</TableCell>
                                        <TableCell>
                                            {editingKey === key ? (
                                                <Input
                                                    value={editValue}
                                                    onChange={(e) => setEditValue(e.target.value)}
                                                    onKeyDown={(e) => handleKeyPress(e, key)}
                                                    placeholder="Enter new value"
                                                    className="h-8 w-full"
                                                    autoFocus
                                                />
                                            ) : (
                                                <span className="text-sm text-neutral-600">{isSet ? lastUpdated : 'Unset'}</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {editingKey === key ? (
                                                <div className="flex justify-end gap-2">
                                                    <Button size="sm" onClick={() => handleSave(key)} className="h-8 px-3">
                                                        Save
                                                    </Button>
                                                    <Button size="sm" variant="outline" onClick={handleCancel} className="h-8 px-3">
                                                        Cancel
                                                    </Button>
                                                </div>
                                            ) : (
                                                <Button size="sm" variant="outline" onClick={() => handleEdit(key)} className="h-8 px-3">
                                                    {isSet ? 'Edit' : 'Set'}
                                                </Button>
                                            )}
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </div>
            </SettingsLayout>
        </AppLayout>
    );
}
