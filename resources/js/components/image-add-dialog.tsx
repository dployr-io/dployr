import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useRemotes } from '@/hooks/use-remotes';
import { Form } from '@inertiajs/react';
import { Loader2 } from 'lucide-react';

interface Props {
    open: boolean;
    setOpen: (open: boolean) => void;
}

export default function ImageAddDialog({ open, setOpen }: Props) {
    const { searchComplete, validationError, remoteRepo, selectedBranch, setRemoteRepo, getFormAction, getFormData, handleFormSuccess } =
        useRemotes(setOpen);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Add Image</DialogTitle>
                    <DialogDescription>Enter a link to docker image to import to your project.</DialogDescription>
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
                                <Label htmlFor="image_registry">Docker image</Label>
                                <Input
                                    id="image_registry"
                                    name="image_registry"
                                    placeholder="redis:latest"
                                    value={remoteRepo}
                                    onChange={(e) => setRemoteRepo(e.target.value)}
                                    disabled={processing}
                                />
                                {(validationError || errors.image_registry) && (
                                    <div className="text-sm text-destructive">{validationError || errors.image_registry}</div>
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
