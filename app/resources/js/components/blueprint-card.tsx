import type { Blueprint, Project } from "@/types";
import { Link } from "@inertiajs/react";

export default function BlueprintCard(blueprint: Blueprint) {
    return (
        <Link
            href={`/projects/${blueprint.id}`}
            className="rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-muted-foreground dark:hover:border-muted-foreground h-28 flex flex-col dark:border-sidebar-border no-underline"
        >
            <div className="flex gap-2 mb-2">
                <img
                    className="h-6 w-6 bg-gra rounded-full flex-shrink-0"
                    src="img/default-project.png"
                />
                <div className="min-w-0 flex-1">
                    <p className="truncate">{blueprint.config.name}</p>
                </div>
            </div>
        </Link>
    );
}