import { Spinner, Table } from "@heroui/react";
import { ArrowsRotateRight, CircleCheck, Copy, Server } from "@gravity-ui/icons";
import { useDashboardSummary } from "../../api/dashboard";
import { formatBytes, formatDate, formatRelativeTime } from "../../lib/format";
import { PageLayout } from "../PageLayout";
import { countryOptions } from "../share/countries";

const countryNames = new Map(countryOptions.map((country) => [country.id, country.nameZh])); // 国家代码到中文名称的只读索引。

export default function Dashboard() {
    const { data, isLoading, isError } = useDashboardSummary();

    if (isLoading) {
        return <PageLayout title="概览"><div className="flex min-h-[28rem] items-center justify-center"><Spinner size="sm" /></div></PageLayout>;
    }
    if (isError || !data) {
        return <PageLayout title="概览"><div className="flex min-h-[28rem] items-center justify-center text-sm text-muted">概览加载失败</div></PageLayout>;
    }

    const countries = Object.entries(data.country_counts).sort(([, a], [, b]) => b - a).slice(0, 6);
    const highestCountryCount = countries[0]?.[1] || 1;

    return (
        <PageLayout title="概览" className="grid content-start gap-3">
            <section className="grid grid-cols-2 gap-3 lg:grid-cols-4">
                <div className="flex min-h-28 items-center rounded-2xl bg-surface p-4">
                    <div className="flex flex-1 self-start items-center gap-2 text-sm text-muted"><Server className="size-4 text-accent" /><span>总节点</span></div>
                    <div className="ml-4 self-end text-3xl font-medium tabular-nums text-foreground">{data.total_nodes}</div>
                </div>
                <div className="flex min-h-28 items-center rounded-2xl bg-surface p-4">
                    <div className="flex flex-1 self-start items-center gap-2 text-sm text-muted"><ArrowsRotateRight className="size-4 text-accent" /><span>订阅</span></div>
                    <div className="ml-4 self-end text-3xl font-medium tabular-nums text-foreground">{data.subscriptions_total}</div>
                </div>
                <div className="flex min-h-28 items-center rounded-2xl bg-surface p-4">
                    <div className="flex flex-1 self-start items-center gap-2 text-sm text-muted"><CircleCheck className="size-4 text-accent" /><span>任务</span></div>
                    <div className="ml-4 self-end text-3xl font-medium tabular-nums text-foreground">{data.tasks_total}</div>
                </div>
                <div className="flex min-h-28 items-center rounded-2xl bg-surface p-4">
                    <div className="flex flex-1 self-start items-center gap-2 text-sm text-muted"><Copy className="size-4 text-accent" /><span>分享</span></div>
                    <div className="ml-4 self-end text-3xl font-medium tabular-nums text-foreground">{data.shares_total}</div>
                </div>
            </section>

            <section>
                <Table className="bg-surface">
                    <Table.ScrollContainer>
                        <Table.Content aria-label="订阅">
                            <Table.Header className="border-0 bg-surface [&_th]:after:hidden">
                                <Table.Column isRowHeader className="px-4 py-2">名称</Table.Column>
                                <Table.Column className="px-3 py-2">节点</Table.Column>
                                <Table.Column className="px-3 py-2">流量</Table.Column>
                                <Table.Column className="px-3 py-2">到期</Table.Column>
                                <Table.Column className="px-3 py-2">刷新</Table.Column>
                            </Table.Header>
                            <Table.Body className="[&_td]:border-0" renderEmptyState={() => <div className="py-12 text-center text-sm text-muted">暂无订阅</div>}>
                                {[...data.subscriptions].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((subscription) => {
                                    const trafficPercent = subscription.traffic_total > 0 ? Math.min(subscription.traffic_used / subscription.traffic_total * 100, 100) : 0;
                                    return (
                                        <Table.Row key={subscription.id} id={subscription.id}>
                                            <Table.Cell className="max-w-56 px-4 py-3 font-medium text-foreground"><span className="block truncate" title={subscription.name}>{subscription.name || "未命名订阅"}</span></Table.Cell>
                                            <Table.Cell className="px-3 py-3 tabular-nums text-foreground">{subscription.node_num}</Table.Cell>
                                            <Table.Cell className="px-3 py-3">
                                                {subscription.traffic_total > 0 ? (
                                                    <div className="w-32">
                                                        <div className="mb-1 text-xs tabular-nums text-muted">{formatBytes(subscription.traffic_used)} / {formatBytes(subscription.traffic_total)}</div>
                                                        <div role="progressbar" aria-label={`${subscription.name}流量使用率`} aria-valuenow={trafficPercent} aria-valuemin={0} aria-valuemax={100} className="h-1.5 overflow-hidden rounded-full bg-surface-tertiary">
                                                            <div className={`h-full rounded-full ${trafficPercent >= 90 ? "bg-danger" : trafficPercent >= 75 ? "bg-warning" : "bg-accent"}`} style={{ width: `${trafficPercent}%` }} />
                                                        </div>
                                                    </div>
                                                ) : <span className="text-muted">-</span>}
                                            </Table.Cell>
                                            <Table.Cell className="px-3 py-3 tabular-nums text-muted">{formatDate(subscription.expires_at)}</Table.Cell>
                                            <Table.Cell className="px-3 py-3 text-muted">{formatRelativeTime(subscription.refreshed_at)}</Table.Cell>
                                        </Table.Row>
                                    );
                                })}
                            </Table.Body>
                        </Table.Content>
                    </Table.ScrollContainer>
                </Table>
            </section>

            <section className="rounded-2xl bg-surface p-4">
                <h2 className="font-medium text-foreground">国家分布</h2>
                {countries.length === 0 ? (
                    <div className="py-12 text-center text-sm text-muted">暂无国家检测数据</div>
                ) : (
                    <div className="mt-4 flex flex-col gap-4">
                        {countries.map(([code, count]) => (
                            <div key={code} className="flex items-center gap-3 text-sm">
                                <div className="flex w-24 shrink-0 items-center gap-2"><span className={`fi fi-${code.toLowerCase()} shrink-0 rounded-[2px]`} /><span className="truncate text-foreground">{countryNames.get(code) || code}</span></div>
                                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-surface-tertiary"><div className="h-full rounded-full bg-accent" style={{ width: `${count / highestCountryCount * 100}%` }} /></div>
                                <span className="w-8 shrink-0 text-right tabular-nums text-muted">{count}</span>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </PageLayout>
    );
}
