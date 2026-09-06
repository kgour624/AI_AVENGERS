import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getAdminClients, updateAdminClient } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatRelativeTime } from '@/utils/format'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 11 ("Client Management -
 * View all clients, Enable/disable clients, View client projects and
 * usage").
 *
 * Feature #7 fix (docs bug list): ListClients previously returned
 * only id/email/fullName/isActive/lastLogin/createdAt - the "usage"
 * part of the wireframe had nothing to render. Backend now returns
 * real projectCount/messageCount aggregates (admin_handler.go), shown
 * below instead of being fabricated client-side.
 */
function AdminClients() {
  const queryClient = useQueryClient()
  const { data: clients, isLoading } = useQuery({
    queryKey: ['admin', 'clients'],
    queryFn: getAdminClients,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ clientId, isActive }: { clientId: string; isActive: boolean }) =>
      updateAdminClient(clientId, isActive),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'clients'] })
    },
  })

  if (isLoading) {
    return (
      <div className="space-y-3 p-6">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="mb-4 text-xl font-semibold">Clients</h1>
      <div className="space-y-2">
        {clients?.map((client) => (
          <Card key={client.id} className="flex items-center justify-between">
            <div>
              <p className="font-medium text-text-primary">{client.fullName}</p>
              <p className="text-xs text-text-secondary">{client.email}</p>
              <p className="text-xs text-text-disabled">
                {client.projectCount} project{client.projectCount === 1 ? '' : 's'}
                {' \u00b7 '}
                {client.messageCount} message{client.messageCount === 1 ? '' : 's'} sent
                {client.lastLogin && (
                  <>
                    {' \u00b7 Last login: '}
                    {formatRelativeTime(client.lastLogin)}
                  </>
                )}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={client.isActive ? 'brand' : 'neutral'}>
                {client.isActive ? 'Active' : 'Disabled'}
              </Badge>
              <Button
                variant="secondary"
                size="sm"
                isLoading={toggleMutation.isPending && toggleMutation.variables?.clientId === client.id}
                onClick={() => toggleMutation.mutate({ clientId: client.id, isActive: !client.isActive })}
              >
                {client.isActive ? 'Disable' : 'Enable'}
              </Button>
            </div>
          </Card>
        ))}
      </div>
    </div>
  )
}

export const Component = AdminClients
