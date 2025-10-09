import type { Remote } from '@/types';
import { Link } from '@inertiajs/react';
import { FaGitlab } from 'react-icons/fa6';
import { RxGithubLogo } from 'react-icons/rx';
import { Badge } from './ui/badge';

export default function RemoteCard(remote?: Remote) {
    return (
        <Link
            href={`/projects/${remote?.id}`}
            className="flex h-28 flex-col rounded-xl border border-sidebar-border/70 p-4 no-underline hover:cursor-pointer hover:border-muted-foreground dark:border-sidebar-border dark:hover:border-muted-foreground"
        >
            <div className="mb-2 flex gap-2">
                <img className="bg-gra h-6 w-6 flex-shrink-0 rounded-full" src={remote?.avatar_url || '../img/default-project.png'} />
                <div className="min-w-0 flex-1">
                    <p className="truncate">
                        {remote?.name}/{remote?.repository}
                    </p>
                </div>
            </div>

            <Badge className="mb-2 w-fit">
                {remote?.provider?.includes('github') ? <RxGithubLogo /> : <FaGitlab />}
                <span className="max-w-36 truncate">{remote?.branch}</span>
            </Badge>

            <p className="line-clamp-1 flex-1 truncate overflow-hidden text-xs text-muted-foreground">{remote?.commit_message}</p>
        </Link>
    );
}
