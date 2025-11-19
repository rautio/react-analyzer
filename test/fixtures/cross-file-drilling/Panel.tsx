// Named export - consumer component
export function Panel({ config }: { config: any }) {
    return <div style={{ color: config.color }}>Panel</div>;
}
