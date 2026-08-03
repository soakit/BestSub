import { useState } from "react"
import { Button } from "@/src/components/ui/button"
import { Copy, Plus, Upload } from "lucide-react"
import { toast } from "sonner"
import { SubForm } from "./sub-form"
import { SubDetail } from "./sub-detail"
import { SubList } from "./sub-list"
import { BatchSubForm } from "./batch-sub-form"
import { copyToClipboard } from "@/src/components/features/share/utils"
import { useSubs } from "@/src/lib/queries/sub-queries"
import type { SubResponse } from "@/src/types/sub"

export function SubPage() {
    const { data: subscriptions = [], isLoading } = useSubs()
    const [detailSubscription, setDetailSubscription] = useState<SubResponse | null>(null)
    const [isDetailDialogOpen, setIsDetailDialogOpen] = useState(false)
    const [isFormDialogOpen, setIsFormDialogOpen] = useState(false)
    const [isBatchFormDialogOpen, setIsBatchFormDialogOpen] = useState(false)
    const [editingSubscription, setEditingSubscription] = useState<SubResponse | null>(null)


    const handleEdit = (subscription: SubResponse) => {
        setEditingSubscription(subscription)
        setIsFormDialogOpen(true)
    }

    const handleCreate = () => {
        setEditingSubscription(null)
        setIsFormDialogOpen(true)
    }

    const handleBatchCreate = () => {
        setIsBatchFormDialogOpen(true)
    }

    // 将所有订阅链接按每行一个写入剪贴板。
    const handleExport = async () => {
        const success = await copyToClipboard(subscriptions.map(subscription => subscription.config.url).join('\n'))
        if (!success) {
            toast.error('导出失败，无法写入剪贴板')
            return
        }

        toast.success(`已复制 ${subscriptions.length} 个订阅链接`)
    }

    const handleFormSuccess = () => {
        setIsFormDialogOpen(false)
        setEditingSubscription(null)
    }

    const handleBatchFormSuccess = () => {
        setIsBatchFormDialogOpen(false)
    }


    const showDetail = (subscription: SubResponse) => {
        setDetailSubscription(subscription)
        setIsDetailDialogOpen(true)
    }

    return (
        <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            <div className="flex flex-col gap-3 px-4 sm:flex-row sm:items-center sm:justify-between lg:px-6">
                <div>
                    <h1 className="text-2xl font-bold">订阅管理</h1>
                </div>

                <div className="flex flex-wrap gap-2">
                    <Button variant="outline" onClick={handleExport} disabled={isLoading || subscriptions.length === 0}>
                        <Copy className="h-4 w-4 mr-2" />
                        导出
                    </Button>
                    <Button variant="outline" onClick={handleBatchCreate}>
                        <Upload className="h-4 w-4 mr-2" />
                        批量添加
                    </Button>
                    <Button onClick={handleCreate}>
                        <Plus className="h-4 w-4 mr-2" />
                        添加订阅
                    </Button>
                </div>
            </div>

            <SubForm
                initialData={editingSubscription ? {
                    name: editingSubscription.name,
                    tags: editingSubscription.tags || [],
                    enable: editingSubscription.enable,
                    cron_expr: editingSubscription.cron_expr,
                    config: {
                        url: editingSubscription.config.url || '',
                        proxy: editingSubscription.config.proxy || false,
                        timeout: editingSubscription.config.timeout || 10,
                    }
                } : undefined}
                formTitle={editingSubscription ? "编辑订阅" : "添加订阅"}
                isOpen={isFormDialogOpen}
                onClose={handleFormSuccess}
                editingSubId={editingSubscription?.id}
            />

            <BatchSubForm
                isOpen={isBatchFormDialogOpen}
                onClose={handleBatchFormSuccess}
            />

            <div className="px-4 lg:px-6">
                <SubList
                    onEdit={(sub) => handleEdit(sub)}
                    onShowDetail={showDetail}
                />
            </div>

            <SubDetail
                subscription={detailSubscription}
                isOpen={isDetailDialogOpen}
                onOpenChange={setIsDetailDialogOpen}
            />
        </div>
    )
}
