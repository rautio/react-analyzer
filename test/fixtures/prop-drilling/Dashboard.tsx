import { useState } from 'react';

// Test Case: Complex dashboard with multiple configuration props drilled
// Expected: 3 violations (locale, currency, dateFormat all drilled through multiple levels)
// Realistic e-commerce or analytics dashboard scenario

type Locale = 'en-US' | 'es-ES' | 'fr-FR';
type Currency = 'USD' | 'EUR' | 'GBP';
type DateFormat = 'MM/DD/YYYY' | 'DD/MM/YYYY' | 'YYYY-MM-DD';

interface AppSettings {
    locale: Locale;
    currency: Currency;
    dateFormat: DateFormat;
}

function App() {
    const [settings, setSettings] = useState<AppSettings>({
        locale: 'en-US',
        currency: 'USD',
        dateFormat: 'MM/DD/YYYY',
    });

    return (
        <div className="app">
            <DashboardLayout
                locale={settings.locale}
                currency={settings.currency}
                dateFormat={settings.dateFormat}
            />
        </div>
    );
}

// DashboardLayout just organizes structure, doesn't use settings
interface DashboardLayoutProps {
    locale: Locale;
    currency: Currency;
    dateFormat: DateFormat;
}

function DashboardLayout({ locale, currency, dateFormat }: DashboardLayoutProps) {
    return (
        <div className="dashboard-layout">
            <NavigationBar />
            <MainArea locale={locale} currency={currency} dateFormat={dateFormat} />
            <Footer />
        </div>
    );
}

function NavigationBar() {
    return (
        <nav>
            <h1>Analytics Dashboard</h1>
            <button>Settings</button>
        </nav>
    );
}

function Footer() {
    return <footer>© 2024 Analytics Co.</footer>;
}

// MainArea doesn't use settings, just layouts panels
function MainArea({ locale, currency, dateFormat }: DashboardLayoutProps) {
    return (
        <main>
            <MetricsPanel locale={locale} currency={currency} dateFormat={dateFormat} />
            <ChartsPanel />
        </main>
    );
}

function ChartsPanel() {
    return (
        <div className="charts-panel">
            <h2>Charts</h2>
            <div>Chart placeholder</div>
        </div>
    );
}

// MetricsPanel doesn't use settings either
function MetricsPanel({ locale, currency, dateFormat }: DashboardLayoutProps) {
    return (
        <div className="metrics-panel">
            <h2>Metrics</h2>
            <MetricCards locale={locale} currency={currency} dateFormat={dateFormat} />
        </div>
    );
}

// MetricCards finally uses the settings
function MetricCards({ locale, currency, dateFormat }: DashboardLayoutProps) {
    const revenue = 45678.90;
    const lastUpdated = new Date('2024-01-15');

    // Format currency based on locale and currency
    const formattedRevenue = new Intl.NumberFormat(locale, {
        style: 'currency',
        currency: currency,
    }).format(revenue);

    // Format date based on dateFormat
    const formattedDate = formatDate(lastUpdated, dateFormat);

    return (
        <div className="metric-cards">
            <div className="card">
                <h3>Revenue</h3>
                <p className="value">{formattedRevenue}</p>
                <p className="updated">Updated: {formattedDate}</p>
            </div>
        </div>
    );
}

function formatDate(date: Date, format: DateFormat): string {
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const year = date.getFullYear();

    switch (format) {
        case 'MM/DD/YYYY':
            return `${month}/${day}/${year}`;
        case 'DD/MM/YYYY':
            return `${day}/${month}/${year}`;
        case 'YYYY-MM-DD':
            return `${year}-${month}-${day}`;
    }
}

export default App;
