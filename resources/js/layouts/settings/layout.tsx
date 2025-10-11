import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import { config, system } from '@/routes';
import { edit as editPassword } from '@/routes/password';
import { edit } from '@/routes/profile';
import { type NavItem } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import { type PropsWithChildren } from 'react';

const sidebarNavItems: NavItem[] = [
    {
        title: 'Profile',
        href: edit(),
        icon: null,
    },
    {
        title: 'Password',
        href: editPassword(),
        icon: null,
    },
    {
        title: 'System',
        href: system(),
        icon: null,
    },
    {
        title: 'Configuration',
        href: config(),
        icon: null,
    },
];

export default function SettingsLayout({ children }: PropsWithChildren) {
    // When server-side rendering, we only render the layout on the client...
    if (typeof window === 'undefined') {
        return null;
    }

    const currentPath = window.location.pathname;
    const { version } = usePage().props;

    return (
        <div className="flex flex-1 flex-col items-center px-4 py-6">
            <div className="flex w-full max-w-6xl flex-col lg:flex-row lg:space-x-12">
                <aside className="w-full max-w-xl lg:w-48">
                    <nav className="flex flex-col space-y-1 space-x-0">
                        {sidebarNavItems.map((item, index) => (
                            <Button
                                key={`${typeof item.href === 'string' ? item.href : item.href.url}-${index}`}
                                size="sm"
                                variant="ghost"
                                asChild
                                className={cn('w-full justify-start', {
                                    'bg-muted': currentPath === (typeof item.href === 'string' ? item.href : item.href.url),
                                })}
                            >
                                <Link href={item.href} prefetch>
                                    {item.icon && <item.icon className="h-4 w-4" />}
                                    {item.title}
                                </Link>
                            </Button>
                        ))}
                    </nav>
                </aside>

                <Separator className="my-6 lg:hidden" />

                <div className="flex-1">
                    <section className="w-full space-y-12">{children}</section>
                </div>
            </div>
            <div className="mt-auto flex w-full justify-center">
                <p className="text-xs text-muted-foreground">Version {(version as string)}</p>
            </div>
        </div>
    );
}
