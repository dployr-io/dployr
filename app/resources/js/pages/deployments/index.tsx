import { StatusChip } from '@/components/status-chip';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useProjects } from '@/hooks/use-projects';
import { useRemotes } from '@/hooks/use-remotes';

import { useServices } from '@/hooks/use-services';
import AppLayout from '@/layouts/app-layout';
import { getRuntimeIcon } from '@/lib/runtime-icon';
import { deploymentsList, deploymentsShow, servicesList } from '@/routes';
import type { BreadcrumbItem, Service } from '@/types';
import { Head, Link, router } from '@inertiajs/react';
import { ArrowUpRightIcon, ChevronLeft, ChevronRight, Factory } from 'lucide-react';
import { useMemo, useState } from 'react';
import { FaGitlab } from 'react-icons/fa6';
import { RxGithubLogo } from 'react-icons/rx';
import TimeAgo from 'react-timeago';

const formatWithoutSuffix = (value: number, unit: string): string => {
    const pluralizedUnit = value === 1 ? unit : `${unit}s`;
    return `${value} ${pluralizedUnit}`;
};

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Deployments',
        href: deploymentsList().url,
    },
];
export default function Deployments() {
    const { deployments } = useServices();
    const { defaultProject } = useProjects();
    const deploymentsData = deployments.data;

    const { remotes } = useRemotes();
    const remotesData = remotes.data || [];

    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;
    const totalPages = Math.ceil((deploymentsData?.length ?? 0) / itemsPerPage);
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedDeployments = deploymentsData?.slice(startIndex, endIndex);

    const normalizedDeployments = useMemo(
        () =>
            (paginatedDeployments || []).map((deployment) => {
                let config = deployment?.config ?? {};
                config = JSON.parse(config as string);
                const remote = (config as Partial<Service>)?.remote;
                let resolvedRemote = remote;
                if (remote) {
                    resolvedRemote = remotesData.find((r) => r) || remote;
                }
                return { ...deployment, config: { ...config, remote: resolvedRemote } };
            }),
        [paginatedDeployments, remotesData],
    );

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
            <Head title="Projects" />
            <div className="flex h-full flex-1 flex-col gap-4 overflow-x-auto rounded-xl p-4">
                <div className="flex w-full flex-col gap-6 px-9 py-6">
                    <div className="flex items-start justify-between">
                        <div className="flex flex-col gap-1">
                            <p className="text-2xl font-black">Deployments</p>
                            <p className="text-sm font-normal text-muted-foreground">Manage your deployments here</p>
                        </div>
                    </div>

                    {deploymentsData?.length === 0 ? (
                        <Empty>
                            <EmptyHeader>
                                <EmptyMedia variant="icon">
                                    <Factory />
                                </EmptyMedia>
                                <EmptyTitle>No Deployments Yet</EmptyTitle>
                                <EmptyDescription>
                                    You don&apos;t have any deployments yet. Get started by deploying your first service.
                                </EmptyDescription>
                            </EmptyHeader>
                            <EmptyContent>
                                <div className="flex gap-2">
                                    <Button>
                                        <Link href={defaultProject && defaultProject.id ? servicesList({ project: defaultProject.id }).url : '#'}>
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
                    ) : (
                        <>
                            <Table className="overflow-hidden rounded-t-lg">
                                <TableHeader className="gap-2 rounded-t-xl bg-neutral-50 p-2 dark:bg-neutral-900">
                                    <TableRow className="h-14">
                                        <TableHead className="h-14 w-[240px] align-middle">Name</TableHead>
                                        <TableHead className="h-14 align-middle">Duration</TableHead>
                                        <TableHead className="h-14 align-middle">Status</TableHead>
                                        <TableHead className="h-14 align-middle">Runtime</TableHead>
                                        <TableHead className="h-14 align-middle">Remote</TableHead>
                                        <TableHead className="h-14 w-[200px] text-right align-middle">Run Command</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {normalizedDeployments && normalizedDeployments.length > 0
                                        ? normalizedDeployments.map((deployment) => (
                                              <TableRow
                                                  key={deployment.id}
                                                  className="h-16 cursor-pointer"
                                                  onClick={() => router.visit(deploymentsShow(deployment.id).url)}
                                              >
                                                  <TableCell className="h-16 w-[240px] overflow-hidden align-middle font-medium">
                                                      <span className="block truncate">{deployment.config?.name || '-'}</span>
                                                  </TableCell>
                                                  <TableCell className="h-16 w-[120px] align-middle">
                                                      {deployment.status === 'completed' || deployment.status === 'failed' ? (
                                                          deployment.updated_at && deployment.created_at ? (
                                                              <span className="inline-block">
                                                                  {(() => {
                                                                      const ms =
                                                                          new Date(deployment.updated_at).getTime() -
                                                                          new Date(deployment.created_at).getTime();
                                                                      const seconds = Math.floor(ms / 1000);
                                                                      const minutes = Math.floor(seconds / 60);
                                                                      const hours = Math.floor(minutes / 60);
                                                                      const days = Math.floor(hours / 24);

                                                                      if (days > 0) return `${days} day${days !== 1 ? 's' : ''}`;
                                                                      if (hours > 0) return `${hours} hour${hours !== 1 ? 's' : ''}`;
                                                                      if (minutes > 0) return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
                                                                      return `${seconds} second${seconds !== 1 ? 's' : ''}`;
                                                                  })()}
                                                              </span>
                                                          ) : (
                                                              <>-</>
                                                          )
                                                      ) : (
                                                          <>
                                                              <TimeAgo date={deployment.created_at} formatter={formatWithoutSuffix} />
                                                          </>
                                                      )}
                                                  </TableCell>
                                                  <TableCell className="h-16 w-[120px] gap-2 align-middle">
                                                      <StatusChip status={deployment.status} />
                                                  </TableCell>
                                                  <TableCell className="h-16 w-[120px] align-middle">
                                                      <div className="flex items-center gap-2">
                                                          {getRuntimeIcon(deployment.config?.runtime)}
                                                          <span>{deployment.config?.runtime || '-'}</span>
                                                      </div>
                                                  </TableCell>
                                                  <TableCell className="h-16 max-w-[320px] overflow-hidden align-middle">
                                                      <div className="flex min-w-0 items-center gap-2">
                                                          {deployment.config?.remote?.provider?.includes('github') ? <RxGithubLogo /> : <FaGitlab />}
                                                          <span className="truncate">
                                                              {deployment.config?.remote
                                                                  ? `${deployment.config.remote.name}/${deployment.config.remote.repository}`
                                                                  : '-'}
                                                          </span>
                                                      </div>
                                                  </TableCell>
                                                  <TableCell className="h-16 w-[200px] overflow-hidden text-right align-middle">
                                                      <span className="block truncate text-right">{deployment.config?.run_cmd || '-'}</span>
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
                                    {(deploymentsData || []).length === 0
                                        ? 'No deployments found'
                                        : deploymentsData!.length === 1
                                          ? 'Showing 1 of 1 deployment'
                                          : `Showing ${startIndex + 1} to ${Math.min(endIndex, (deploymentsData || []).length || 0)} of ${(deploymentsData || []).length} deployments`}{' '}
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
