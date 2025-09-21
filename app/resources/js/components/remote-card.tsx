import type { Remote } from "@/types";
import { Badge } from "./ui/badge";
import { RxGithubLogo } from "react-icons/rx";
import { FaGitlab } from "react-icons/fa6";
import { Link } from "@inertiajs/react";

export default function RemoteCard(remote?: Remote) {
    return (
        <Link
            href={`/projects/${remote?.id}`}
            className="rounded-xl border border-sidebar-border/70 p-4 hover:cursor-pointer hover:border-muted-foreground dark:hover:border-muted-foreground h-28 flex flex-col dark:border-sidebar-border no-underline"
        >
            <div className="flex gap-2 mb-2">
                <img
                    className="h-6 w-6 bg-gra rounded-full flex-shrink-0"
                    src={remote?.avatar_url}
                />
                <div className="min-w-0 flex-1">
                    <p className="truncate">{remote?.name}/{remote?.repository}</p>
                </div>
            </div>

            <Badge className="mb-2 w-fit">
                {remote?.provider?.includes('github') ? <RxGithubLogo /> : <FaGitlab />}
                <span className="truncate max-w-36">{remote?.branch}</span>
            </Badge>

            <p className="text-xs text-muted-foreground flex-1 line-clamp-1 overflow-hidden truncate">
                {remote?.commit_message}
            </p>
        </Link>
    );
}