import ImageAddDialog from '@/components/image-add-dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { useImages } from '@/hooks/use-images';
import AppLayout from '@/layouts/app-layout';
import { imagesIndex } from '@/routes';
import type { BreadcrumbItem, DockerImage } from '@/types';
import { Head } from '@inertiajs/react';
import { PlusCircle } from 'lucide-react';
import { useState } from 'react';
import ImageCard from '@/components/image-card';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Images',
        href: imagesIndex().url,
    },
];

export default function Image() {
    const { images, isLoading } = useImages();
    const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex flex-col gap-1">
                        <p className="text-2xl font-black">Docker Images</p>
                        <p className="text-sm font-normal text-muted-foreground">Manage your docker images</p>
                    </div>

                    <div className="grid w-full grid-cols-3 gap-3">
                        {isLoading ? (
                            <>
                                <div className="flex flex-col gap-2 rounded-xl border border-sidebar-border/70 p-4 dark:border-sidebar-border">
                                    <div className="mb-2 flex items-center gap-2">
                                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
                                            <Skeleton className="h-8 w-8 rounded-full bg-muted-foreground/20" />
                                        </div>
                                        <div className="flex-1">
                                            <Skeleton className="h-4 w-24 rounded bg-muted-foreground/20" />
                                        </div>
                                    </div>
                                    <Skeleton className="mb-1 h-3 w-32 rounded bg-muted-foreground/20" />
                                    <Skeleton className="h-3 w-20 rounded bg-muted-foreground/20" />
                                </div>
                            </>
                        ) : (
                            <>
                                {images?.map((image: DockerImage) => (
                                    <ImageCard
                                        key={image.id}
                                        image={image}                                  
                                    />
                                ))}
                            </>
                        )}
                        <div
                            className="flex flex-col gap-2 rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-accent-foreground md:min-h-min dark:border-sidebar-border dark:hover:border-muted-foreground"
                            onClick={() => setIsDialogOpen(true)}
                        >
                            <div className="flex items-center gap-2">
                                <PlusCircle size={20} className="text-muted-foreground" />
                                <p>Import a new docker image</p>
                            </div>
                            <p className="text-sm text-muted-foreground">Click to pull a new docker image to your library</p>
                        </div>
                    </div>

                    <ImageAddDialog open={isDialogOpen} setOpen={setIsDialogOpen} />
                </div>
            </div>
        </AppLayout>
    );
}
