import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useRemotes } from '@/hooks/use-remotes';
import { Form } from '@inertiajs/react';
import { ChevronDown, Loader2 } from 'lucide-react';

interface Props {
    open: boolean;
    setOpen: (open: boolean) => void;
}

export default function RemoteAddDialog({ open, setOpen }: Props) {
    const {
        branches,
        searchComplete,
        validationError,
        remoteRepo,
        selectedBranch,
        setRemoteRepo,
        setSelectedBranch,
        getFormAction,
        getFormData,
        handleFormSuccess,
    } = useRemotes(setOpen);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Add Remote</DialogTitle>
                    <DialogDescription>Enter a link to remote repository to import to your project.</DialogDescription>
                </DialogHeader>

                <Form
                    action={getFormAction()}
                    transform={() => getFormData()}
                    method="post"
                    onSuccess={handleFormSuccess}
                    className="grid items-start gap-6"
                >
                    {({ processing, errors }) => (
                        <>
                            <div className="grid gap-3">
                                <Label htmlFor="remote_repo">Remote Repository</Label>
                                <Input
                                    id="remote_repo"
                                    name="remote_repo"
                                    placeholder="github.com/username/repository.git"
                                    value={remoteRepo}
                                    onChange={(e) => setRemoteRepo(e.target.value)}
                                    disabled={processing}
                                />
                                {(validationError || errors.remote_repo) && (
                                    <div className="text-sm text-destructive">{validationError || errors.remote_repo}</div>
                                )}
                            </div>

                            <div className="grid gap-3">
                                {processing && (
                                    <div className="flex justify-center">
                                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            {searchComplete ? 'Creating project' : 'Fetching remote repository'}
                                        </div>
                                    </div>
                                )}

                                {branches.length > 0 && !processing && (
                                    <>
                                        <Label htmlFor="branch">Branch</Label>
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button
                                                    type="button"
                                                    size="default"
                                                    variant="outline"
                                                    className="group min-w-40 text-sidebar-accent-foreground data-[state=open]:bg-sidebar-accent"
                                                >
                                                    {selectedBranch || 'Select branch'}
                                                    <ChevronDown className="ml-auto size-4 transition-transform group-data-[state=open]:rotate-180" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent
                                                className="w-[--radix-dropdown-menu-trigger-width] min-w-40 rounded-lg"
                                                align="start"
                                            >
                                                {branches.map((branch, index) => (
                                                    <div key={branch}>
                                                        <DropdownMenuItem onClick={() => setSelectedBranch(branch)}>{branch}</DropdownMenuItem>
                                                        {index < branches.length - 1 && <DropdownMenuSeparator />}
                                                    </div>
                                                ))}
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                        <input type="hidden" name="branch" value={selectedBranch} />
                                        {errors.branch && <div className="text-sm text-destructive">{errors.branch}</div>}
                                    </>
                                )}
                            </div>

                            <div className="flex justify-end gap-2">
                                <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                                    Cancel
                                </Button>
                                <Button type="submit" disabled={processing || (searchComplete && !selectedBranch)}>
                                    {processing && <Loader2 className="h-4 w-4 animate-spin" />}
                                    {processing ? (searchComplete ? 'Importing' : 'Searching') : searchComplete ? 'Import' : 'Search'}
                                </Button>
                            </div>
                        </>
                    )}
                </Form>
            </DialogContent>
        </Dialog>
    );
}
