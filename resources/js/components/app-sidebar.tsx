import { NavFooter } from '@/components/nav-footer';
import { NavMain } from '@/components/nav-main';
import { NavUser } from '@/components/nav-user';
import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import {
    console,
    deploymentsIndex,
    imagesIndex,
    logs,
    notificationsIndex,
    projectsIndex,
    remotesIndex,
    resourceManagerIndex,
    specsIndex,
} from '@/routes';
import { type NavItem } from '@/types';
import { Link } from '@inertiajs/react';
import {
    Bell,
    BookOpen,
    CircleGauge,
    Container,
    Factory,
    FileSliders,
    FolderGit2,
    LayoutGrid,
    Logs,
    MessageCircleQuestion,
    SquareChevronRight,
} from 'lucide-react';
import AppLogo from './app-logo';

const mainNavItems: NavItem[] = [
    {
        title: 'Projects',
        href: projectsIndex(),
        icon: LayoutGrid,
    },
    {
        title: 'Deployments',
        href: deploymentsIndex(),
        icon: Factory,
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
        title: 'Resources',
        href: resourceManagerIndex(),
        icon: CircleGauge,
    },
    {
        title: 'Remotes',
        href: remotesIndex(),
        icon: FolderGit2,
    },
    {
        title: 'Images',
        href: imagesIndex(),
        icon: Container,
    },
    {
        title: 'Specs',
        href: specsIndex({ query: { spec: true } }).url,
        icon: FileSliders,
    },
];

const footerNavItems: NavItem[] = [
    {
        title: 'Notifications',
        href: notificationsIndex(),
        icon: Bell,
    },
    {
        title: 'Support',
        href: 'https://dployr.io/support',
        icon: MessageCircleQuestion,
    },
    {
        title: 'Docs',
        href: 'https://dployr.io/docs',
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
                            <Link href={projectsIndex()} prefetch>
                                <AppLogo />
                            </Link>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent>
                <NavMain items={mainNavItems} title="Platform" />
                <div className="h-4" />
                <NavMain items={secondaryNavItems} title="Resources" />
            </SidebarContent>

            <SidebarFooter>
                <NavFooter items={footerNavItems} className="mt-auto" />
                <NavUser />
            </SidebarFooter>
        </Sidebar>
    );
}
