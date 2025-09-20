import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useProjects } from '@/hooks/use-projects';
import { Form } from '@inertiajs/react';
import { useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';

interface Props {
    open: boolean;
    setOpen: (open: boolean) => void;
}

export default function ProjectCreateDialog({ open, setOpen }: Props) {
    const queryClient = useQueryClient();

    const onCreatedSuccess = () => {
        queryClient.invalidateQueries({ queryKey: ['projects'] });
        setOpen(false);
    };

    const {
        name,
        description,
        validationError,
        setName,
        setDescription,
        getFormData,
        handleFormSuccess,
    } = useProjects(onCreatedSuccess);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Create Project</DialogTitle>
                    <DialogDescription>Add a new project to your library.</DialogDescription>
                </DialogHeader>

                <Form
                    action="/projects"
                    transform={() => getFormData()}
                    method="post"
                    onSuccess={handleFormSuccess}
                    className="grid items-start gap-6"
                >
                    {({ processing, errors }) => (
                        <>
                            <div className="grid gap-3">
                                <Label htmlFor="name">Name</Label>
                                <Input
                                    id="name"
                                    name="name"
                                    placeholder="My awesome dployr project"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    tabIndex={1}
                                    disabled={processing}
                                />
                                {(validationError || errors.name) && (
                                    <div className="text-sm text-destructive">{validationError || errors.name}</div>
                                )}
                            </div>

                            <div className="grid gap-3">
                                <Label htmlFor="description">Description</Label>
                                <Input
                                    id="description"
                                    name="description"
                                    placeholder="My app, my server, my rules!"
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    tabIndex={2}
                                    disabled={processing}
                                />
                                {(validationError || errors.description) && (
                                    <div className="text-sm text-destructive">{validationError || errors.description}</div>
                                )}
                            </div>

                            <div className="flex justify-end gap-2">
                                <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                                    Cancel
                                </Button>
                                <Button type="submit" disabled={processing}>
                                    {processing && <Loader2 className="h-4 w-4 animate-spin" />}
                                    {processing ? 'Creating' : 'Create'}
                                </Button>
                            </div>
                        </>
                    )}
                </Form>
            </DialogContent>
        </Dialog>
    );
}
