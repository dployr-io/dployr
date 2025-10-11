import { StatusChip } from '@/components/status-chip';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import { projectsIndex, servicesIndex } from '@/routes';
import type { BreadcrumbItem, Project } from '@/types';
import { Head, Link, usePage } from '@inertiajs/react';
import { ArrowUpRightIcon, ChevronLeft, ChevronRight, CirclePlus, Hexagon, Settings } from 'lucide-react';
import { useState } from 'react';

const ViewProjectBreadcrumbs = (project?: Project) => {
    const breadcrumbs: BreadcrumbItem[] = [
        {
            title: 'Projects',
            href: projectsIndex().url,
        },
        {
            title: project && project.name ? project.name : 'Project',
            href: '',
        },
    ];

    return breadcrumbs;
};

export default function Services() {
    const { props } = usePage();
    const project = (props.project as Project) || null;
    const { services, isLoading } = useServices(project?.id);
    const breadcrumbs = ViewProjectBreadcrumbs(project);
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;
    const totalPages = Math.max(1, Math.ceil((services?.length ?? 0) / itemsPerPage));
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedServices = services?.slice(startIndex, endIndex);

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
                            <p className="text-2xl font-black">{project?.name || 'Project'}</p>
                            <p className="text-sm font-normal text-muted-foreground">
                                {project?.description ? `${project.description} ${project.description}` : ''}
                            </p>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button variant="outline" className="flex items-center gap-2">
                                <Settings className="h-4 w-4" />
                                Configure
                            </Button>
                            <Button className="flex items-center gap-2" asChild>
                                <Link href={project && project.id ? servicesIndex({ project: project.id }).url : '#'}>
                                    <CirclePlus className="h-4 w-4" />
                                    Deploy Service
                                </Link>
                            </Button>
                        </div>
                    </div>

                    {paginatedServices && paginatedServices.length === 0 ? (
                        <div className="flex min-h-[400px] flex-1 items-center justify-center">
                            <Empty>
                                <EmptyHeader>
                                    <EmptyMedia variant="icon">
                                        <Hexagon />
                                    </EmptyMedia>
                                    <EmptyTitle>No Services Yet</EmptyTitle>
                                    <EmptyDescription>
                                        You haven&apos;t deployed any services yet. Get started by deploying your first service.
                                    </EmptyDescription>
                                </EmptyHeader>
                                <EmptyContent>
                                    <div className="flex justify-center gap-2">
                                        <Button>
                                            <Link href={project && project.id ? servicesIndex({ project: project.id }).url : '#'}>
                                                Deploy Service
                                            </Link>
                                        </Button>
                                        <Button variant="link" asChild className="text-muted-foreground" size="sm">
                                            <Link href="#">
                                                Learn More <ArrowUpRightIcon />
                                            </Link>
                                        </Button>
                                    </div>
                                </EmptyContent>
                            </Empty>
                        </div>
                    ) : (
                        <>
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
                                    {!isLoading
                                        ? paginatedServices!.map((service) => (
                                            <TableRow key={service.id} className="h-16">
                                                <TableCell className="h-16 align-middle font-medium">{service.name}</TableCell>
                                                <TableCell className="h-16 align-middle">
                                                    <StatusChip status={service.status} />
                                                </TableCell>
                                                <TableCell className="h-16 align-middle">
                                                    <div className="flex items-center gap-2">
                                                        {getRuntimeIcon(service.runtime)}
                                                        <span>{service.runtime}</span>
                                                    </div>
                                                </TableCell>
                                                <TableCell className="h-16 align-middle">{service.region}</TableCell>
                                                <TableCell className="h-16 w-[200px] text-right align-middle">
                                                    {service.last_deployed instanceof Date
                                                        ? service.last_deployed.toLocaleString()
                                                        : service.last_deployed}
                                                </TableCell>
                                            </TableRow>
                                        ))
                                        : Array.from({ length: 3 }).map((_, idx) => (
                                            <TableRow key={`skeleton-${idx}`} className="h-16">
                                                <TableCell className="h-16 max-w-[240px] overflow-hidden align-middle font-medium">
                                                    <div className="h-4 w-32 animate-pulse rounded bg-muted" />
                                                </TableCell>
                                                <TableCell className="h-16 align-middle">
                                                    <div className="h-4 w-16 animate-pulse rounded bg-muted" />
                                                </TableCell>
                                                <TableCell className="h-16 align-middle">
                                                    <div className="h-4 w-20 animate-pulse rounded bg-muted" />
                                                </TableCell>
                                                <TableCell className="h-16 max-w-[320px] overflow-hidden align-middle">
                                                    <div className="h-4 w-40 animate-pulse rounded bg-muted" />
                                                </TableCell>
                                                <TableCell className="h-16 w-[200px] overflow-hidden text-right align-middle">
                                                    <div className="ml-auto h-4 w-24 animate-pulse rounded bg-muted" />
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                </TableBody>
                            </Table>

                            <div className="flex items-center justify-between px-2 py-4">
                                <div className="text-sm text-muted-foreground">
                                    {(services ?? []).length === 0
                                        ? 'No services found'
                                        : services!.length === 1
                                            ? 'Showing 1 of 1 service'
                                            : `Showing ${startIndex + 1} to ${Math.min(endIndex, services?.length ?? 0)} of ${services?.length ?? 0} services`}{' '}
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
                        </>
                    )}
                </div>
            </div>
        </AppLayout>
    );
}
