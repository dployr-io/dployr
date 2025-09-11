import { Head } from '@inertiajs/react';

import AppearanceTabs from '@/components/appearance-tabs';
import HeadingSmall from '@/components/heading-small';
import { type BreadcrumbItem } from '@/types';

import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { system } from '@/routes';

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'System settings',
        href: system().url,
    },
];

export default function System() {
    return (
        <AppLayout breadcrumbs={breadcrumbs}>
            <Head title="System settings" />

            <SettingsLayout>
                <div className="space-y-6">
                    <HeadingSmall title="System settings" description="Update your account's system settings" />
                    <AppearanceTabs />
                </div>
            </SettingsLayout>
        </AppLayout>
    );
}
