import type { DockerImage } from '@/types';
import { Link } from '@inertiajs/react';

interface Props {
    image: DockerImage
}

export default function ImageCard({ image }: Props) {
    return (
        <Link
            href={`/projects/${image.id}`}
            className="flex h-28 flex-col rounded-xl border border-sidebar-border/70 p-4 no-underline hover:cursor-pointer hover:border-muted-foreground dark:border-sidebar-border dark:hover:border-muted-foreground"
        >
            <div className="mb-2 flex gap-2">
                <div className="min-w-0 flex-1">
                    <p className="truncate">
                        {image.name}
                    </p>
                </div>
            </div>
        </Link>
    );
}
