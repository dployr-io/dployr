import { NavFooter } from '@/components/nav-footer';
import { NavMain } from '@/components/nav-main';
import { NavUser } from '@/components/nav-user';
import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { logs, console, projectsList } from '@/routes';
import { type NavItem } from '@/types';
import { Link } from '@inertiajs/react';
import { BookOpen, Container, Folder, FolderGit2, LayoutGrid, Logs, SquareChevronRight } from 'lucide-react';
import AppLogo from './app-logo';
import { NavSecondary } from './nav-secondary';

const mainNavItems: NavItem[] = [
    {
        title: 'Projects',
        href: projectsList(),
        icon: LayoutGrid,
    },
    {
        title: 'Logs',
        href: logs(),
        icon: Logs,
    },
    {
        title: 'Console',
        href: console(),
        icon: SquareChevronRight,
    },
];

const secondaryNavItems: NavItem[] = [  
    {
        title: 'Remotes',
        href: projectsList(),
        icon: FolderGit2,
    },
    {
        title: 'Images',
        href: logs(),
        icon: Container,
    },
]

const footerNavItems: NavItem[] = [
    {
        title: 'Repository',
        href: 'https://github.com/laravel/react-starter-kit',
        icon: Folder,
    },
    {
        title: 'Documentation',
        href: 'https://laravel.com/docs/starter-kits#react',
        icon: BookOpen,
    },
];

export function AppSidebar() {
    return (
        <Sidebar collapsible="icon" variant="inset">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" asChild>
                            <Link href={projectsList()} prefetch>
                                <AppLogo />
                            </Link>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent>
                <NavMain items={mainNavItems} />
                <div className='h-4' />
                <NavSecondary items={secondaryNavItems} />
            </SidebarContent>

            <SidebarFooter>
                <NavFooter items={footerNavItems} className="mt-auto" />
                <NavUser />
            </SidebarFooter>
        </Sidebar>
    );
}
