import { createInertiaApp } from '@inertiajs/react';
import createServer from '@inertiajs/react/server';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { resolvePageComponent } from 'laravel-vite-plugin/inertia-helpers';
import ReactDOMServer from 'react-dom/server';
import Wrapper from '@/wrapper';

const appName = import.meta.env.VITE_APP_NAME || 'dployr';
const queryClient = new QueryClient();

createServer((page) =>
    createInertiaApp({
        page,
        render: ReactDOMServer.renderToString,
        title: (title) => (title ? `${title} - ${appName}` : appName),
        resolve: (name) => resolvePageComponent(`./pages/${name}.tsx`, import.meta.glob('./pages/**/*.tsx')),
        setup: ({ App, props }) => {
            return (
                <QueryClientProvider client={queryClient}>
                    <App {...props}>
                        {({ Component, key, props: pageProps }) => (
                            <Wrapper>
                                <Component {...pageProps} key={key} />
                            </Wrapper>
                        )}
                    </App>
                </QueryClientProvider>
            );
        },
    }),
);
