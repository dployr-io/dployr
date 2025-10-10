import type { Project } from '@/types';
import { Link } from '@inertiajs/react';

interface Props {
    project: Project;
}

export default function ProjectCard({ project }: Props) {
    return (
        <Link
            href={`/projects/${project.id}`}
            onClick={() => localStorage.setItem('current_project', project.id.toString())}
            className="flex h-28 flex-col rounded-xl border border-sidebar-border/70 p-4 no-underline hover:cursor-pointer hover:border-muted-foreground dark:border-sidebar-border dark:hover:border-muted-foreground"
        >
            <div className="mb-2 flex gap-2">
                <img className="bg-gra h-6 w-6 flex-shrink-0 rounded-full" src="img/default-project.png" />
                <div className="min-w-0 flex-1">
                    <p className="truncate">{project.name}</p>
                </div>
            </div>

            <p className="line-clamp-1 flex-1 overflow-hidden text-xs text-muted-foreground">{project.description}</p>
        </Link>
    );
}
