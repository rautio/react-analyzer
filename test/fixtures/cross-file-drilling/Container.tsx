import { Panel } from './Panel';

// Named export - passthrough component
export function Container({ config }: { config: any }) {
    return <Panel config={config} />;
}
