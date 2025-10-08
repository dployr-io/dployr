import type { Project } from "@/types";
import { Link } from "@inertiajs/react";

export default function ProjectCard(project: Project) {
    return (
        <Link
            href={`/projects/${project.id}`}
            onClick={() => localStorage.setItem('current_project', project.id.toString())}
            className="rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-muted-foreground dark:hover:border-muted-foreground h-28 flex flex-col dark:border-sidebar-border no-underline"
        >
            <div className="flex gap-2 mb-2">
                <img
                    className="h-6 w-6 bg-gra rounded-full flex-shrink-0"
                    src="img/default-project.png"
                />
                <div className="min-w-0 flex-1">
                    <p className="truncate">{project.name}</p>
                </div>
            </div>

            <p className="text-xs text-muted-foreground flex-1 line-clamp-1 overflow-hidden">
                {project.description}
            </p>
        </Link>
    );
}