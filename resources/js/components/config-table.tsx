import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toWordUpperCase } from '@/lib/utils';
import { useState } from 'react';

interface Props {
    config: Record<string, any>;
    onUpdateConfig: (key: string, value: string) => void;
    onRemoveConfig?: (key: string) => void;
}

export function ConfigTable({ config, onUpdateConfig, onRemoveConfig }: Props) {
    const [editingKey, setEditingKey] = useState<string | null>(null);
    const [editValue, setEditValue] = useState<string>('');
    const [newKey, setNewKey] = useState<string>('');
    const [newValue, setNewValue] = useState<string>('');
    const [isAddingNew, setIsAddingNew] = useState<boolean>(false);

    const handleEdit = (key: string) => {
        setEditingKey(key);
        setEditValue(config[key] || '');
        setIsAddingNew(false);
    };

    const handleSave = (key: string, value: string) => {
        onUpdateConfig(key, value);
        setEditingKey(null);
        setEditValue('');
    };

    const handleCancel = () => {
        setEditingKey(null);
        setEditValue('');
        setIsAddingNew(false);
        setNewKey('');
        setNewValue('');
    };

    const handleAddNew = () => {
        setIsAddingNew(true);
        setNewKey('');
        setNewValue('');
    };

    const handleSaveNew = () => {
        if (newKey.trim() && newValue.trim()) {
            onUpdateConfig(newKey.trim(), newValue.trim());
            setNewKey('');
            setNewValue('');
            setIsAddingNew(false);
        }
    };

    const handleKeyPress = (e: React.KeyboardEvent, key: string) => {
        if (e.key === 'Enter') {
            handleSave(key, editValue);
        } else if (e.key === 'Escape') {
            handleCancel();
        }
    };

    const handleNewKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleSaveNew();
        } else if (e.key === 'Escape') {
            handleCancel();
        }
    };

    const handleRemove = (key: string) => {
        if (onRemoveConfig) {
            onRemoveConfig(key);
        }
    };

    return (
        <Table className="overflow-hidden rounded-t-lg">
            <TableHeader className="gap-2 rounded-t-xl bg-neutral-50 p-2 dark:bg-neutral-900">
                <TableRow>
                    <TableHead className="w-[200px]">Key</TableHead>
                    <TableHead>Value</TableHead>
                    <TableHead className="w-[100px] text-right"></TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {Object.entries(config || {}).map(([key, value]) => {
                    const isSet = value !== null && value !== undefined && value !== '';

                    return (
                        <TableRow key={key}>
                            <TableCell className="font-medium">{toWordUpperCase(key)}</TableCell>
                            <TableCell>
                                {editingKey === key ? (
                                    <Input
                                        value={editValue}
                                        onChange={(e) => setEditValue(e.target.value)}
                                        onKeyDown={(e) => handleKeyPress(e, key)}
                                        placeholder="Enter value"
                                        className="h-8 w-full"
                                        autoFocus
                                    />
                                ) : (
                                    <span className="text-sm text-neutral-600">{isSet ? (typeof value === 'string' ? value : 'Set') : 'Unset'}</span>
                                )}
                            </TableCell>
                            <TableCell className="text-right">
                                {editingKey === key ? (
                                    <div className="flex justify-end gap-2">
                                        <Button size="sm" onClick={() => handleSave(key, editValue)} className="h-8 px-3">
                                            Save
                                        </Button>
                                        <Button size="sm" variant="outline" onClick={handleCancel} className="h-8 px-3">
                                            Cancel
                                        </Button>
                                    </div>
                                ) : (
                                    <div className="flex justify-end gap-2">
                                        <Button size="sm" variant="outline" onClick={() => handleEdit(key)} className="h-8 px-3">
                                            {isSet ? 'Edit' : 'Set'}
                                        </Button>
                                        {onRemoveConfig && (
                                            <Button
                                                size="sm"
                                                variant="outline"
                                                onClick={() => handleRemove(key)}
                                                className="h-8 px-3 text-red-600 hover:bg-red-50 hover:text-red-700 dark:hover:bg-red-950"
                                            >
                                                Remove
                                            </Button>
                                        )}
                                    </div>
                                )}
                            </TableCell>
                        </TableRow>
                    );
                })}

                {/* Add new entry row */}
                <TableRow className="border-t-2 border-dashed border-neutral-200 dark:border-neutral-700">
                    <TableCell className="font-medium">
                        {isAddingNew ? (
                            <Input
                                value={newKey}
                                onChange={(e) => setNewKey(e.target.value)}
                                onKeyDown={handleNewKeyPress}
                                placeholder="Enter key"
                                className="h-8 w-full"
                                autoFocus
                            />
                        ) : (
                            <span className="text-sm text-neutral-500">Enter key</span>
                        )}
                    </TableCell>
                    <TableCell>
                        {isAddingNew ? (
                            <Input
                                value={newValue}
                                onChange={(e) => setNewValue(e.target.value)}
                                onKeyDown={handleNewKeyPress}
                                placeholder="Enter value"
                                className="h-8 w-full"
                            />
                        ) : (
                            <span className="text-sm text-neutral-500">Enter value</span>
                        )}
                    </TableCell>
                    <TableCell className="text-right">
                        {isAddingNew ? (
                            <div className="flex justify-end gap-2">
                                <Button size="sm" onClick={handleSaveNew} className="h-8 px-3" disabled={!newKey.trim() || !newValue.trim()}>
                                    Add
                                </Button>
                                <Button size="sm" variant="outline" onClick={handleCancel} className="h-8 px-3">
                                    Cancel
                                </Button>
                            </div>
                        ) : (
                            <Button size="sm" variant="outline" onClick={handleAddNew} className="h-8 px-3">
                                + Add
                            </Button>
                        )}
                    </TableCell>
                </TableRow>
            </TableBody>
        </Table>
    );
}
