import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import AppLayout from '@/layouts/app-layout';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import { index, projectsList } from '@/routes';
import type { BreadcrumbItem, Project, Service } from  '@/types';
import { Head, Link, usePage } from '@inertiajs/react';
import { ChevronLeft, ChevronRight, CirclePlus, Settings } from 'lucide-react';
import { useState } from 'react';


const ViewProjectBreadcrumbs = (project: Project) => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsList().url,
        },
        {
            title: project?.name || 'Project',
            href: '',
        },
    ];

    return breadcrumbs;
};

const services: Service[] = [
    {
        id: '1',
        source: 'remote',
        name: 'releeva-web',
        status: 'running',
        runtime: 'custom',
        region: 'Africa/Lagos',
        last_deployed: new Date(),
    },
    {
        id: '2',
        source: 'remote',
        name: 'api-gateway',
        status: 'running',
        runtime: 'node-js',
        region: 'America/NewYork',
        last_deployed: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    },
    {
        id: '3',
        source: 'remote',
        name: 'user-service',
        status: 'stopped',
        runtime: 'python',
        region: 'Europe/London',
        last_deployed: new Date(Date.now() - 24 * 60 * 60 * 1000), // 1 day ago
    },
    {
        id: '4',
        source: 'remote',
        name: 'notification-worker',
        status: 'running',
        runtime: 'ruby',
        region: 'Asia/Tokyo',
        last_deployed: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
    },
    {
        id: '5',
        source: 'remote',
        name: 'analytics-dashboard',
        status: 'deploying',
        runtime: 'php',
        region: 'Australia/Sydney',
        last_deployed: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    },
    {
        id: '6',
        source: 'remote',
        name: 'analytics-dashboard',
        status: 'deploying',
        runtime: 'java',
        region: 'Australia/Sydney',
        last_deployed: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    },
    {
        id: '7',
        source: 'remote',
        name: 'releeva-web',
        status: 'running',
        runtime: 'go',
        region: 'Africa/Lagos',
        last_deployed: new Date(),
    },
    {
        id: '8',
        source: 'remote',
        name: 'api-gateway',
        status: 'running',
        runtime: 'node-js',
        region: 'America/NewYork',
        last_deployed: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    },
    {
        id: '9',
        source: 'image',
        name: 'user-service',
        status: 'stopped',
        runtime: 'docker',
        region: 'Europe/London',
        last_deployed: new Date(Date.now() - 24 * 60 * 60 * 1000), // 1 day ago
    },
    {
        id: '10',
        source: 'remote',
        name: 'notification-worker',
        status: 'running',
        runtime: 'dotnet',
        region: 'Asia/Tokyo',
        last_deployed: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
    },
    {
        id: '15',
        source: 'remote',
        name: 'analytics-dashboard',
        status: 'deploying',
        runtime: 'php',
        region: 'Australia/Sydney',
        last_deployed: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    },
    {
        id: '16',
        source: 'image',
        name: 'analytics-dashboard',
        status: 'deploying',
        runtime: 'k3s',
        region: 'Australia/Sydney',
        last_deployed: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    },
];

export default function ViewProject() {
    const { props } = usePage();
    const project = props.project as Project;
    const breadcrumbs = ViewProjectBreadcrumbs(project);

    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;

    const totalPages = Math.ceil(services.length / itemsPerPage);
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedServices = services.slice(startIndex, endIndex);

    const goToPage = (page: number) => {
        setCurrentPage(Math.max(1, Math.min(page, totalPages)));
    };

    const goToPreviousPage = () => {
        setCurrentPage((prev) => Math.max(1, prev - 1));
    };

    const goToNextPage = () => {
        setCurrentPage((prev) => Math.min(totalPages, prev + 1));
    };

    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="Project" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex items-start justify-between">
                        <div className="flex flex-col gap-1">
                            <p className="text-2xl font-black">{project?.name}</p>
                            <p className="text-sm font-normal text-muted-foreground">{project?.description} {project?.description}</p>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button variant="outline" className="flex items-center gap-2">
                                <Settings className="h-4 w-4" />
                                Configure
                            </Button>
                            <Button
                                className="flex items-center gap-2"
                                asChild
                            >
                                <Link href={index({ project: project.id }).url}>
                                    <CirclePlus className="h-4 w-4" />
                                    Deploy Service
                                </Link>
                            </Button>
                        </div>
                    </div>

                    <Table className="overflow-hidden rounded-t-lg">
                        <TableHeader className="gap-2 rounded-t-xl bg-neutral-50 p-2 dark:bg-neutral-900">
                            <TableRow className="h-14">
                                <TableHead className="h-14 w-[240px] align-middle">Name</TableHead>
                                <TableHead className="h-14 align-middle">Status</TableHead>
                                <TableHead className="h-14 align-middle">Runtime</TableHead>
                                <TableHead className="h-14 align-middle">Location</TableHead>
                                <TableHead className="h-14 w-[200px] text-right align-middle">Last Deployed</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {paginatedServices.map((service) => (
                                <TableRow key={service.id} className="h-16">
                                    <TableCell className="h-16 align-middle font-medium">{service.name}</TableCell>
                                    <TableCell className="h-16 align-middle">
                                        <span
                                            className={`inline-block rounded-full px-2 py-1 text-xs font-semibold ${
                                                service.status === 'running'
                                                    ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                    : service.status === 'deploying'
                                                      ? 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200'
                                                      : service.status === 'stopped'
                                                        ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                                                        : 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                                            }`}
                                        >
                                            {service.status.charAt(0).toUpperCase() + service.status.slice(1)}
                                        </span>
                                    </TableCell>
                                    <TableCell className="h-16 align-middle">
                                        <div className="flex items-center gap-2">
                                            {getRuntimeIcon(service.runtime)}
                                            <span>{service.runtime}</span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="h-16 align-middle">{service.region}</TableCell>
                                    <TableCell className="h-16 w-[200px] text-right align-middle">
                                        {service.last_deployed instanceof Date ? service.last_deployed.toLocaleString() : service.last_deployed}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>

                    <div className="flex items-center justify-between px-2 py-4">
                        <div className="text-sm text-muted-foreground">
                            Showing {startIndex + 1} to {Math.min(endIndex, services.length)} of {services.length} services
                        </div>
                        <div className="flex items-center space-x-2">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={goToPreviousPage}
                                disabled={currentPage === 1}
                                className="flex items-center gap-1"
                            >
                                <ChevronLeft className="h-4 w-4" />
                                Previous
                            </Button>

                            <div className="flex items-center space-x-1">
                                {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                                    <Button
                                        key={page}
                                        variant={currentPage === page ? 'default' : 'outline'}
                                        size="sm"
                                        onClick={() => goToPage(page)}
                                        className="h-8 w-8 p-0"
                                    >
                                        {page}
                                    </Button>
                                ))}
                            </div>

                            <Button
                                variant="outline"
                                size="sm"
                                onClick={goToNextPage}
                                disabled={currentPage === totalPages}
                                className="flex items-center gap-1"
                            >
                                Next
                                <ChevronRight className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        </AppLayout>
    );
}
